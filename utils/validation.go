package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors representa múltiples errores de validación
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// ValidateEmail valida el formato de un email
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("el email es requerido")
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}
	
	return nil
}

// ValidatePassword valida la fortaleza de una contraseña
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("la contraseña es requerida")
	}
	
	if len(password) < 6 {
		return fmt.Errorf("la contraseña debe tener al menos 6 caracteres")
	}
	
	if len(password) > 100 {
		return fmt.Errorf("la contraseña no puede tener más de 100 caracteres")
	}
	
	return nil
}

// ValidateUsername valida el nombre de usuario
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("el nombre de usuario es requerido")
	}
	
	if len(username) < 3 {
		return fmt.Errorf("el nombre de usuario debe tener al menos 3 caracteres")
	}
	
	if len(username) > 50 {
		return fmt.Errorf("el nombre de usuario no puede tener más de 50 caracteres")
	}
	
	// Validar que solo contenga caracteres alfanuméricos y algunos especiales
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("el nombre de usuario solo puede contener letras, números, guiones, puntos y guiones bajos")
	}
	
	return nil
}

// ValidateRole valida que el rol sea válido
func ValidateRole(role string) error {
	validRoles := map[string]bool{
		"user":  true,
		"admin": true,
	}
	
	if !validRoles[role] {
		return fmt.Errorf("rol inválido. Los roles válidos son: user, admin")
	}
	
	return nil
}

// ValidateUserCreateRequest valida una solicitud de creación de usuario
func ValidateUserCreateRequest(email, username, password string) ValidationErrors {
	var errors ValidationErrors
	
	if err := ValidateEmail(email); err != nil {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: err.Error(),
		})
	}
	
	if err := ValidateUsername(username); err != nil {
		errors = append(errors, ValidationError{
			Field:   "nombre_de_usuario",
			Message: err.Error(),
		})
	}
	
	if err := ValidatePassword(password); err != nil {
		errors = append(errors, ValidationError{
			Field:   "contraseña",
			Message: err.Error(),
		})
	}
	
	return errors
}

// ValidateUserUpdateRequest valida una solicitud de actualización de usuario
func ValidateUserUpdateRequest(email, username *string) ValidationErrors {
	var errors ValidationErrors
	
	if email != nil && *email != "" {
		if err := ValidateEmail(*email); err != nil {
			errors = append(errors, ValidationError{
				Field:   "email",
				Message: err.Error(),
			})
		}
	}
	
	if username != nil && *username != "" {
		if err := ValidateUsername(*username); err != nil {
			errors = append(errors, ValidationError{
				Field:   "nombre_de_usuario",
				Message: err.Error(),
			})
		}
	}
	
	return errors
}

// ValidateLoginRequest valida una solicitud de login
func ValidateLoginRequest(email, password string) ValidationErrors {
	var errors ValidationErrors
	
	if err := ValidateEmail(email); err != nil {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: err.Error(),
		})
	}
	
	if password == "" {
		errors = append(errors, ValidationError{
			Field:   "contraseña",
			Message: "la contraseña es requerida",
		})
	}
	
	return errors
}
