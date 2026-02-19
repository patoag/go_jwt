package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-jwt-backend/services"
	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AuthMiddleware verifica que el usuario esté autenticado y que el token no esté blacklisted
func AuthMiddleware(jwtSecret string, authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token de autorización requerido",
				"message": "Debe incluir el header Authorization con el token Bearer",
			})
			c.Abort()
			return
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Formato de token inválido",
				"message": "El header Authorization debe tener el formato: Bearer <token>",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWTWithSecret(token, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token inválido",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Verificar blacklist
		if authService != nil {
			blacklisted, err := authService.IsTokenBlacklisted(token)
			if err != nil {
				log.Printf("Error verificando blacklist: %v", err)
			}
			if blacklisted {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Token inválido",
					"message": "El token ha sido revocado",
				})
				c.Abort()
				return
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Rol)

		c.Next()
	}
}

// AdminMiddleware verifica que el usuario tenga rol de administrador
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Usuario no autenticado",
				"message": "Debe estar autenticado para acceder a este recurso",
			})
			c.Abort()
			return
		}

		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Acceso denegado",
				"message": "Se requieren permisos de administrador para acceder a este recurso",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware maneja las políticas CORS con orígenes configurables
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	originsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originsMap[strings.TrimSpace(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if originsMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implementa rate limiting con Redis sliding window por IP
func RateLimitMiddleware(redisClient *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fail-open si Redis no está disponible
		if redisClient == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", ip)
		now := time.Now().UnixMilli()
		windowMs := window.Milliseconds()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		member := fmt.Sprintf("%d:%d", now, time.Now().UnixNano())

		pipe := redisClient.Pipeline()

		// Eliminar entradas fuera de la ventana
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now-windowMs, 10))
		// Contar requests en la ventana
		countCmd := pipe.ZCard(ctx, key)
		// Agregar request actual con miembro único
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
		// Renovar TTL
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			// Fail-open: si Redis falla, dejar pasar
			log.Printf("Error en rate limiting: %v", err)
			c.Next()
			return
		}

		count := countCmd.Val()

		// Agregar headers informativos
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		remaining := maxRequests - int(count) - 1
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if int(count) >= maxRequests {
			retryAfter := int(window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Demasiadas solicitudes",
				"message": fmt.Sprintf("Límite de %d solicitudes por %v excedido. Intente nuevamente más tarde.", maxRequests, window),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromContext obtiene el ID del usuario desde el contexto
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id no encontrado en el contexto")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user_id no es un UUID válido")
	}

	return id, nil
}

// GetUserRoleFromContext obtiene el rol del usuario desde el contexto
func GetUserRoleFromContext(c *gin.Context) (string, error) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", fmt.Errorf("user_role no encontrado en el contexto")
	}

	role, ok := userRole.(string)
	if !ok {
		return "", fmt.Errorf("user_role no es un string válido")
	}

	return role, nil
}

// LoggerMiddleware personalizado para logging de requests
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}
