package handlers

import (
	"math"
	"net/http"
	"strconv"

	"go-jwt-backend/config"
	"go-jwt-backend/middleware"
	"go-jwt-backend/models"
	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserHandler maneja las operaciones CRUD de usuarios
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler crea una nueva instancia de UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		db: config.DB,
	}
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
// @Failure 500 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Obtener ID del usuario desde el contexto (middleware de auth)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	// Buscar el usuario en la base de datos
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuario no encontrado",
				"message": "El usuario no existe",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener el usuario",
		})
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
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Obtener ID del usuario desde el contexto
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

	// Validar datos de entrada
	if validationErrors := utils.ValidateUserUpdateRequest(req.Email, req.NombreDeUsuario); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	// Buscar el usuario actual
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuario no encontrado",
				"message": "El usuario no existe",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener el usuario",
		})
		return
	}

	// Verificar si el nuevo email ya existe (si se está actualizando)
	if req.Email != nil && *req.Email != user.Email {
		var existingUser models.User
		if err := h.db.Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email ya registrado",
				"message": "Ya existe otra cuenta con este email",
			})
			return
		}
	}

	// Verificar si el nuevo nombre de usuario ya existe (si se está actualizando)
	if req.NombreDeUsuario != nil && *req.NombreDeUsuario != user.NombreDeUsuario {
		var existingUser models.User
		if err := h.db.Where("nombre_de_usuario = ? AND id != ?", *req.NombreDeUsuario, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Nombre de usuario ya registrado",
				"message": "Ya existe otra cuenta con este nombre de usuario",
			})
			return
		}
	}

	// Actualizar campos si se proporcionaron
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.NombreDeUsuario != nil {
		updates["nombre_de_usuario"] = *req.NombreDeUsuario
	}

	// Realizar la actualización
	if len(updates) > 0 {
		if err := h.db.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error interno del servidor",
				"message": "Error al actualizar el usuario",
			})
			return
		}
	}

	// Obtener el usuario actualizado
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener el usuario actualizado",
		})
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
// @Failure 500 {object} map[string]interface{}
// @Router /users/profile [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	// Obtener ID del usuario desde el contexto
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"message": "No se pudo obtener la información del usuario",
		})
		return
	}

	// Verificar que el usuario existe
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuario no encontrado",
				"message": "El usuario no existe",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener el usuario",
		})
		return
	}

	// Eliminar el usuario
	if err := h.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al eliminar el usuario",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cuenta eliminada exitosamente",
	})
}

// ListUsers lista todos los usuarios (solo admin) con paginación
// @Summary Listar usuarios (Admin)
// @Description Obtiene una lista paginada de todos los usuarios (solo administradores)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Número de página" default(1)
// @Param page_size query int false "Tamaño de página" default(10)
// @Success 200 {object} models.PaginatedUsersResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Obtener parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// Validar parámetros de paginación
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calcular offset
	offset := (page - 1) * pageSize

	// Contar total de usuarios
	var total int64
	if err := h.db.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al contar usuarios",
		})
		return
	}

	// Obtener usuarios con paginación
	var users []models.User
	if err := h.db.Offset(offset).Limit(pageSize).Order("fecha_creacion DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener usuarios",
		})
		return
	}

	// Convertir a respuesta
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	response := models.PaginatedUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserByID obtiene un usuario por ID (solo admin)
// @Summary Obtener usuario por ID (Admin)
// @Description Obtiene los datos de un usuario específico por su ID (solo administradores)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "ID del usuario"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Obtener ID del parámetro de la URL
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID inválido",
			"message": "El ID del usuario debe ser un UUID válido",
		})
		return
	}

	// Buscar el usuario
	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Usuario no encontrado",
				"message": "No existe un usuario con el ID especificado",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al obtener el usuario",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}
