package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"go-jwt-backend/middleware"
	"go-jwt-backend/models"
	"go-jwt-backend/services"
	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler maneja las operaciones CRUD de usuarios
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler crea una nueva instancia de UserHandler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile obtiene el perfil del usuario autenticado
// @Summary Obtener perfil del usuario
// @Description Obtiene los datos del perfil del usuario autenticado
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		mapServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateProfile actualiza el perfil del usuario autenticado
// @Summary Actualizar perfil del usuario
// @Description Actualiza los datos del perfil del usuario autenticado
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body models.UserUpdateRequest true "Datos a actualizar"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	if validationErrors := utils.ValidateUserUpdateRequest(req.Email, req.NombreDeUsuario); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Email, req.NombreDeUsuario)
	if err != nil {
		mapServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// DeleteAccount elimina la cuenta del usuario autenticado
// @Summary Eliminar cuenta del usuario
// @Description Elimina permanentemente la cuenta del usuario autenticado
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /users/profile [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	if err := h.userService.DeleteAccount(userID); err != nil {
		mapServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cuenta eliminada exitosamente"})
}

// ListUsers lista todos los usuarios (solo admin) con paginación
// @Summary Listar usuarios (Admin)
// @Description Obtiene una lista paginada de todos los usuarios
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param page_size query int false "Tamaño de página" default(10)
// @Success 200 {object} models.PaginatedUsersResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.userService.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener usuarios",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserByID obtiene un usuario por ID (solo admin)
// @Summary Obtener usuario por ID (Admin)
// @Description Obtiene los datos de un usuario específico por su ID
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID del usuario (UUID)"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID inválido",
			"message": "El ID del usuario debe ser un UUID válido",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		mapServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// mapServiceError mapea errores de servicio a respuestas HTTP
func mapServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Usuario no encontrado",
			"message": err.Error(),
		})
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
			"message": "Error al procesar la solicitud",
		})
	}
}
