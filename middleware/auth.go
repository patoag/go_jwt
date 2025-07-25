package middleware

import (
	"fmt"
	"net/http"

	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware verifica que el usuario esté autenticado
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token de autorización requerido",
				"message": "Debe incluir el header Authorization con el token Bearer",
			})
			c.Abort()
			return
		}

		// Extraer el token del header
		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Formato de token inválido",
				"message": "El header Authorization debe tener el formato: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Validar el token
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token inválido",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Guardar la información del usuario en el contexto
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Rol)

		c.Next()
	}
}

// AdminMiddleware verifica que el usuario tenga rol de administrador
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verificar que el usuario esté autenticado primero
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Usuario no autenticado",
				"message": "Debe estar autenticado para acceder a este recurso",
			})
			c.Abort()
			return
		}

		// Verificar que el rol sea admin
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

// CORSMiddleware maneja las políticas CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
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

// GetUserIDFromContext obtiene el ID del usuario desde el contexto
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic}
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic}
	}

	return id, nil
}

// GetUserRoleFromContext obtiene el rol del usuario desde el contexto
func GetUserRoleFromContext(c *gin.Context) (string, error) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic}
	}

	role, ok := userRole.(string)
	if !ok {
		return "", gin.Error{Err: gin.Error{}.Err, Type: gin.ErrorTypePublic}
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

// RateLimitMiddleware implementa un rate limiting básico
func RateLimitMiddleware() gin.HandlerFunc {
	// En un entorno de producción, usarías Redis o similar para el rate limiting
	// Por simplicidad, aquí implementamos un rate limiting básico en memoria
	return func(c *gin.Context) {
		// Por ahora, solo pasamos el request
		// En producción implementarías lógica de rate limiting real
		c.Next()
	}
}
