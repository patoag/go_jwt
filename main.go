package main

import (
	"log"
	"os"

	"go-jwt-backend/config"
	"go-jwt-backend/handlers"
	"go-jwt-backend/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Configurar modo de Gin basado en variable de entorno
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Conectar a la base de datos
	config.ConnectDatabase()

	// Crear instancia de Gin
	r := gin.New()

	// Middleware globales
	r.Use(middleware.LoggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	// Crear handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()

	// Rutas públicas
	public := r.Group("/api/v1")
	{
		// Rutas de autenticación
		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Ruta de health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "Servidor funcionando correctamente",
				"version": "1.0.0",
			})
		})
	}

	// Rutas protegidas (requieren autenticación)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// Rutas de usuario
		users := protected.Group("/users")
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.DELETE("/profile", userHandler.DeleteAccount)
		}
	}

	// Rutas de administrador (requieren autenticación y rol admin)
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		// Rutas de administración de usuarios
		admin.GET("/users", userHandler.ListUsers)
		admin.GET("/users/:id", userHandler.GetUserByID)
	}

	// Ruta para documentación de la API
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API REST de Gestión de Usuarios con JWT",
			"version": "1.0.0",
			"endpoints": gin.H{
				"auth": gin.H{
					"POST /api/v1/auth/register": "Registrar nuevo usuario",
					"POST /api/v1/auth/login":    "Iniciar sesión",
					"POST /api/v1/auth/refresh":  "Renovar token JWT",
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
			"documentation":  "Consulta el README.md para ejemplos de uso",
		})
	})

	// Obtener puerto desde variable de entorno
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Iniciar servidor
	log.Printf("Servidor iniciando en puerto %s", port)
	log.Printf("Documentación disponible en: http://localhost:%s", port)
	log.Printf("Health check disponible en: http://localhost:%s/api/v1/health", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Error al iniciar el servidor:", err)
	}
}
