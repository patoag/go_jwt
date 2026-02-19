package services

import (
	"math"
	"strings"

	"go-jwt-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService maneja la lógica de negocio de usuarios
type UserService struct {
	userRepo UserRepository
}

// NewUserService crea una nueva instancia de UserService
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile obtiene el perfil de un usuario por ID
func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateProfile actualiza el perfil de un usuario
func (s *UserService) UpdateProfile(userID uuid.UUID, email, username *string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	updates := make(map[string]interface{})

	if email != nil {
		trimmed := strings.TrimSpace(*email)
		if trimmed != user.Email {
			if existing, err := s.userRepo.FindByEmail(trimmed); err == nil && existing.ID != userID {
				return nil, ErrEmailAlreadyExists
			}
			updates["email"] = trimmed
		}
	}

	if username != nil {
		trimmed := strings.TrimSpace(*username)
		if trimmed != user.NombreDeUsuario {
			if existing, err := s.userRepo.FindByUsername(trimmed); err == nil && existing.ID != userID {
				return nil, ErrUsernameAlreadyExists
			}
			updates["nombre_de_usuario"] = trimmed
		}
	}

	if len(updates) > 0 {
		if err := s.userRepo.Update(user, updates); err != nil {
			return nil, err
		}
		// Recargar usuario actualizado
		user, err = s.userRepo.FindByID(userID)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// DeleteAccount elimina la cuenta de un usuario
func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}
	return s.userRepo.Delete(user)
}

// ListUsers lista usuarios con paginación
func (s *UserService) ListUsers(page, pageSize int) (*models.PaginatedUsersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	users, total, err := s.userRepo.List(offset, pageSize)
	if err != nil {
		return nil, err
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &models.PaginatedUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID obtiene un usuario por ID (para admin)
func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
