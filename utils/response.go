package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse representa una respuesta de error estándar
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse representa una respuesta exitosa estándar
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithError envía una respuesta de error estándar
func RespondWithError(c *gin.Context, statusCode int, error string, message string, details interface{}) {
	response := ErrorResponse{
		Error:   error,
		Message: message,
		Details: details,
	}
	c.JSON(statusCode, response)
}

// RespondWithSuccess envía una respuesta exitosa estándar
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

// RespondBadRequest envía una respuesta de solicitud incorrecta (400)
func RespondBadRequest(c *gin.Context, message string, details interface{}) {
	RespondWithError(c, http.StatusBadRequest, "Solicitud incorrecta", message, details)
}

// RespondUnauthorized envía una respuesta de no autorizado (401)
func RespondUnauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, "No autorizado", message, nil)
}

// RespondForbidden envía una respuesta de prohibido (403)
func RespondForbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, "Acceso denegado", message, nil)
}

// RespondNotFound envía una respuesta de no encontrado (404)
func RespondNotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, "No encontrado", message, nil)
}

// RespondConflict envía una respuesta de conflicto (409)
func RespondConflict(c *gin.Context, message string) {
	RespondWithError(c, http.StatusConflict, "Conflicto", message, nil)
}

// RespondInternalServerError envía una respuesta de error interno del servidor (500)
func RespondInternalServerError(c *gin.Context, message string) {
	RespondWithError(c, http.StatusInternalServerError, "Error interno del servidor", message, nil)
}

// RespondCreated envía una respuesta de creado exitosamente (201)
func RespondCreated(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusCreated, message, data)
}

// RespondOK envía una respuesta exitosa (200)
func RespondOK(c *gin.Context, message string, data interface{}) {
	RespondWithSuccess(c, http.StatusOK, message, data)
}
