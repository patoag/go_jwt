package handlers

import (
	"net/http"

	"go-jwt-backend/config"
	"go-jwt-backend/models"
	"go-jwt-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler maneja las operaciones de autenticación
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler crea una nueva instancia de AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		db: config.DB,
	}
}

// Register maneja el registro de nuevos usuarios
// @Summary Registrar nuevo usuario
// @Description Crea una nueva cuenta de usuario
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

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	// Validar datos de entrada
	if validationErrors := utils.ValidateUserCreateRequest(req.Email, req.NombreDeUsuario, req.Contraseña); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	// Verificar si el email ya existe
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Email ya registrado",
			"message": "Ya existe una cuenta con este email",
		})
		return
	}

	// Verificar si el nombre de usuario ya existe
	if err := h.db.Where("nombre_de_usuario = ?", req.NombreDeUsuario).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Nombre de usuario ya registrado",
			"message": "Ya existe una cuenta con este nombre de usuario",
		})
		return
	}

	// Hashear la contraseña
	hashedPassword, err := utils.HashPassword(req.Contraseña)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al procesar la contraseña",
		})
		return
	}

	// Crear el usuario
	user := models.User{
		Email:           req.Email,
		NombreDeUsuario: req.NombreDeUsuario,
		ContraseñaHash:  hashedPassword,
		Rol:             "user", // Rol por defecto
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al crear el usuario",
		})
		return
	}

	// Generar JWT
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Rol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al generar el token",
		})
		return
	}

	// Respuesta exitosa
	response := models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusCreated, response)
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

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"message": err.Error(),
		})
		return
	}

	// Validar datos de entrada
	if validationErrors := utils.ValidateLoginRequest(req.Email, req.Contraseña); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Errores de validación",
			"details": validationErrors,
		})
		return
	}

	// Buscar el usuario por email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Credenciales inválidas",
			"message": "Email o contraseña incorrectos",
		})
		return
	}

	// Verificar la contraseña
	if !utils.CheckPasswordHash(req.Contraseña, user.ContraseñaHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Credenciales inválidas",
			"message": "Email o contraseña incorrectos",
		})
		return
	}

	// Generar JWT
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Rol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"message": "Error al generar el token",
		})
		return
	}

	// Respuesta exitosa
	response := models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken maneja la renovación de tokens JWT
// @Summary Renovar token JWT
// @Description Genera un nuevo token JWT basado en uno válido existente
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Obtener el token del header
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

	// Generar nuevo token
	newToken, err := utils.RefreshJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token inválido",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}
