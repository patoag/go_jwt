package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User representa el modelo de usuario en la base de datos
type User struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email             string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	NombreDeUsuario   string    `json:"nombre_de_usuario" gorm:"uniqueIndex;not null" validate:"required,min=3,max=50"`
	ContraseñaHash    string    `json:"-" gorm:"not null"` // El "-" evita que se serialice en JSON
	Rol               string    `json:"rol" gorm:"default:user;not null" validate:"oneof=user admin"`
	FechaCreacion     time.Time `json:"fecha_creacion" gorm:"autoCreateTime"`
	FechaActualizacion time.Time `json:"fecha_actualizacion" gorm:"autoUpdateTime"`
}

// UserResponse representa la respuesta del usuario sin datos sensibles
type UserResponse struct {
	ID                 uuid.UUID `json:"id"`
	Email              string    `json:"email"`
	NombreDeUsuario    string    `json:"nombre_de_usuario"`
	Rol                string    `json:"rol"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

// UserCreateRequest representa la solicitud para crear un usuario
type UserCreateRequest struct {
	Email           string `json:"email" validate:"required,email"`
	NombreDeUsuario string `json:"nombre_de_usuario" validate:"required,min=3,max=50"`
	Contraseña      string `json:"contraseña" validate:"required,min=6"`
}

// UserLoginRequest representa la solicitud de login
type UserLoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Contraseña string `json:"contraseña" validate:"required"`
}

// UserUpdateRequest representa la solicitud para actualizar un usuario
type UserUpdateRequest struct {
	Email           *string `json:"email,omitempty" validate:"omitempty,email"`
	NombreDeUsuario *string `json:"nombre_de_usuario,omitempty" validate:"omitempty,min=3,max=50"`
}

// LoginResponse representa la respuesta del login
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// BeforeCreate hook de GORM para generar UUID antes de crear
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ToResponse convierte un User a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                 u.ID,
		Email:              u.Email,
		NombreDeUsuario:    u.NombreDeUsuario,
		Rol:                u.Rol,
		FechaCreacion:      u.FechaCreacion,
		FechaActualizacion: u.FechaActualizacion,
	}
}

// PaginatedUsersResponse representa una respuesta paginada de usuarios
type PaginatedUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
