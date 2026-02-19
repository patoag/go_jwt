package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-jwt-backend/config"
	_ "go-jwt-backend/docs"
	"go-jwt-backend/handlers"
	"go-jwt-backend/middleware"
	"go-jwt-backend/repositories"
	"go-jwt-backend/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Go JWT Backend API
// @version 2.0.0
// @description API REST de autenticación y gestión de usuarios con JWT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Ingrese el token con el prefijo Bearer. Ejemplo: "Bearer {token}"
func main() {
	// Cargar configuración centralizada
	cfg := config.LoadConfig()

	// Configurar modo de Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Conectar a la base de datos
	config.ConnectDatabase(&cfg.Database)

	// Conectar a Redis
	config.ConnectRedis(&cfg.Redis)

	// Crear repositorios
	userRepo := repositories.NewGormUserRepository(config.DB)
	tokenBlacklist := repositories.NewRedisTokenBlacklist(config.RedisClient)

	// Crear servicios
	authService := services.NewAuthService(userRepo, tokenBlacklist, cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	userService := services.NewUserService(userRepo)

	// Crear handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// Configurar router
	router := setupRouter(cfg, authService, authHandler, userHandler)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Servidor iniciando en puerto %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor: %v", err)
		}
	}()

	// Esperar señal de terminación
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Apagando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forzando apagado del servidor: %v", err)
	}

	// Cerrar conexión Redis
	if config.RedisClient != nil {
		if err := config.RedisClient.Close(); err != nil {
			log.Printf("Error al cerrar conexión Redis: %v", err)
		}
	}

	// Cerrar conexión DB
	if sqlDB, err := config.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error al cerrar conexión DB: %v", err)
		}
	}

	log.Println("Servidor apagado correctamente")
}

func setupRouter(cfg *config.Config, authService *services.AuthService, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) *gin.Engine {
	r := gin.New()

	// Middleware globales
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))
	r.Use(middleware.RateLimitMiddleware(config.RedisClient, cfg.RateLimit.Requests, cfg.RateLimit.Window))

	// Rutas públicas
	public := r.Group("/api/v1")
	{
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Health check mejorado
		public.GET("/health", healthCheckHandler())
	}

	// Rutas de auth que requieren autenticación
	authProtected := r.Group("/api/v1/auth")
	authProtected.Use(middleware.AuthMiddleware(cfg.JWT.Secret, authService))
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.POST("/change-password", authHandler.ChangePassword)
	}

	// Rutas protegidas de usuario
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret, authService))
	{
		users := protected.Group("/users")
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.DELETE("/profile", userHandler.DeleteAccount)
		}
	}

	// Rutas de administrador
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(cfg.JWT.Secret, authService))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", userHandler.ListUsers)
		admin.GET("/users/:id", userHandler.GetUserByID)
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Ruta raíz con documentación de la API
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API REST de Gestión de Usuarios con JWT",
			"version": "2.0.0",
			"endpoints": gin.H{
				"auth": gin.H{
					"POST /api/v1/auth/register":        "Registrar nuevo usuario",
					"POST /api/v1/auth/login":            "Iniciar sesión",
					"POST /api/v1/auth/refresh":          "Renovar token JWT",
					"POST /api/v1/auth/logout":           "Cerrar sesión (requiere auth)",
					"POST /api/v1/auth/change-password":  "Cambiar contraseña (requiere auth)",
				},
				"users": gin.H{
					"GET /api/v1/users/profile":    "Obtener perfil del usuario",
					"PUT /api/v1/users/profile":    "Actualizar perfil del usuario",
					"DELETE /api/v1/users/profile": "Eliminar cuenta del usuario",
				},
				"admin": gin.H{
					"GET /api/v1/admin/users":     "Listar todos los usuarios (Admin)",
					"GET /api/v1/admin/users/:id": "Obtener usuario por ID (Admin)",
				},
				"health": gin.H{
					"GET /api/v1/health": "Health check del servidor",
				},
			},
			"authentication": "Bearer Token (JWT) en header Authorization",
		})
	})

	return r
}

func healthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "ok"
		httpStatus := http.StatusOK

		services := gin.H{}

		// Verificar base de datos
		if sqlDB, err := config.DB.DB(); err != nil {
			services["database"] = gin.H{"status": "error", "message": err.Error()}
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else if err := sqlDB.Ping(); err != nil {
			services["database"] = gin.H{"status": "error", "message": err.Error()}
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			services["database"] = gin.H{"status": "ok"}
		}

		// Verificar Redis
		if config.RedisClient == nil {
			services["redis"] = gin.H{"status": "unavailable", "message": "Redis no configurado"}
		} else if err := config.RedisClient.Ping(c.Request.Context()).Err(); err != nil {
			services["redis"] = gin.H{"status": "error", "message": err.Error()}
			status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			services["redis"] = gin.H{"status": "ok"}
		}

		c.JSON(httpStatus, gin.H{
			"status":   status,
			"version":  "2.0.0",
			"services": services,
		})
	}
}
