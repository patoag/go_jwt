package services

import (
	"strings"
	"time"

	"go-jwt-backend/models"
	"go-jwt-backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService maneja la lógica de negocio de autenticación
type AuthService struct {
	userRepo        UserRepository
	blacklist       TokenBlacklist
	jwtSecret       string
	expirationHours int
}

// NewAuthService crea una nueva instancia de AuthService
func NewAuthService(userRepo UserRepository, blacklist TokenBlacklist, jwtSecret string, expirationHours int) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		blacklist:       blacklist,
		jwtSecret:       jwtSecret,
		expirationHours: expirationHours,
	}
}

// Register registra un nuevo usuario y retorna el token JWT
func (s *AuthService) Register(email, username, password string) (string, *models.User, error) {
	email = strings.TrimSpace(email)
	username = strings.TrimSpace(username)

	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return "", nil, ErrEmailAlreadyExists
	}

	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return "", nil, ErrUsernameAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return "", nil, err
	}

	user := &models.User{
		Email:           email,
		NombreDeUsuario: username,
		ContraseñaHash:  hashedPassword,
		Rol:             "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateJWTWithConfig(user.ID, user.Email, user.Rol, s.jwtSecret, s.expirationHours)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// Login autentica un usuario y retorna el token JWT
func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	email = strings.TrimSpace(email)

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if !utils.CheckPasswordHash(password, user.ContraseñaHash) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateJWTWithConfig(user.ID, user.Email, user.Rol, s.jwtSecret, s.expirationHours)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// RefreshToken genera un nuevo token JWT basado en uno existente
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := utils.ValidateJWTWithSecret(tokenString, s.jwtSecret)
	if err != nil {
		return "", ErrInvalidToken
	}

	blacklisted, err := s.blacklist.IsBlacklisted(tokenString)
	if err != nil {
		return "", err
	}
	if blacklisted {
		return "", ErrInvalidToken
	}

	return utils.GenerateJWTWithConfig(claims.UserID, claims.Email, claims.Rol, s.jwtSecret, s.expirationHours)
}

// Logout invalida un token agregándolo a la blacklist
func (s *AuthService) Logout(tokenString string) error {
	claims, err := utils.ValidateJWTWithSecret(tokenString, s.jwtSecret)
	if err != nil {
		return ErrInvalidToken
	}

	remainingTTL := time.Until(claims.ExpiresAt.Time).Seconds()
	if remainingTTL <= 0 {
		return nil
	}

	return s.blacklist.Add(tokenString, int64(remainingTTL))
}

// ChangePassword cambia la contraseña del usuario
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return err
	}

	if !utils.CheckPasswordHash(currentPassword, user.ContraseñaHash) {
		return ErrWrongPassword
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.Update(user, map[string]interface{}{
		"contraseña_hash": hashedPassword,
	})
}

// IsTokenBlacklisted verifica si un token está en la blacklist
func (s *AuthService) IsTokenBlacklisted(tokenString string) (bool, error) {
	return s.blacklist.IsBlacklisted(tokenString)
}
