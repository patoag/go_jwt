package services

import (
	"go-jwt-backend/models"

	"github.com/google/uuid"
)

// UserRepository define las operaciones de persistencia de usuarios
type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User, updates map[string]interface{}) error
	Delete(user *models.User) error
	List(offset, limit int) ([]models.User, int64, error)
}

// TokenBlacklist define las operaciones de blacklist de tokens
type TokenBlacklist interface {
	Add(token string, remainingTTL int64) error
	IsBlacklisted(token string) (bool, error)
}
