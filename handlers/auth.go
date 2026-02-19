package handlers

import (
	"errors"
	"net/http"
	"strings"

	"go-jwt-backend/middleware"
	"go-jwt-backend/models"
	"go-jwt-backend/services"
	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler maneja las operaciones de autenticación
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler crea una nueva instancia de AuthHandler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register maneja el registro de nuevos usuarios
// @Summary Registrar nuevo usuario
// @Description Crea una nueva cuenta de usuario y retorna un JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserCreateRequest true "Datos del usuario"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.NombreDeUsuario = strings.TrimSpace(req.NombreDeUsuario)

	if validationErrors := utils.ValidateUserCreateRequest(req.Email, req.NombreDeUsuario, req.Contraseña); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	token, user, err := h.authService.Register(req.Email, req.NombreDeUsuario, req.Contraseña)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email ya registrado",
				"message": err.Error(),
			})
		case errors.Is(err, services.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Nombre de usuario ya registrado",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error interno del servidor",
				"message": "Error al crear el usuario",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// Login maneja el inicio de sesión de usuarios
// @Summary Iniciar sesión
// @Description Autentica un usuario y devuelve un JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.UserLoginRequest true "Credenciales de login"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if validationErrors := utils.ValidateLoginRequest(req.Email, req.Contraseña); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	token, user, err := h.authService.Login(req.Email, req.Contraseña)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Credenciales inválidas",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al procesar el login",
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// RefreshToken maneja la renovación de tokens JWT
// @Summary Renovar token JWT
// @Description Genera un nuevo token JWT basado en uno válido existente
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token requerido",
			"message": "Debe incluir el token en el header Authorization",
		})
		return
	}

	token := utils.ExtractTokenFromHeader(authHeader)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Formato de token inválido",
			"message": "El header Authorization debe tener el formato: Bearer <token>",
		})
		return
	}

	newToken, err := h.authService.RefreshToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token inválido",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// Logout invalida el token actual del usuario
// @Summary Cerrar sesión
// @Description Invalida el token JWT actual agregándolo a la blacklist
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := utils.ExtractTokenFromHeader(authHeader)

	if err := h.authService.Logout(token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token inválido",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada exitosamente"})
}

// ChangePassword cambia la contraseña del usuario autenticado
// @Summary Cambiar contraseña
// @Description Cambia la contraseña del usuario autenticado
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param passwords body models.ChangePasswordRequest true "Contraseñas"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidatePassword(req.NuevaContraseña); validationErrors != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"message": validationErrors.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(userID, req.ContraseñaActual, req.NuevaContraseña); err != nil {
		switch {
		case errors.Is(err, services.ErrWrongPassword):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Contraseña incorrecta",
				"message": err.Error(),
			})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuario no encontrado",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error interno del servidor",
				"message": "Error al cambiar la contraseña",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada exitosamente"})
}
