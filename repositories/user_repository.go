package repositories

import (
	"go-jwt-backend/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserRepository implementa UserRepository usando GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository crea una nueva instancia de GormUserRepository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("nombre_de_usuario = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *GormUserRepository) Update(user *models.User, updates map[string]interface{}) error {
	return r.db.Model(user).Updates(updates).Error
}

func (r *GormUserRepository) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}

func (r *GormUserRepository) List(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Offset(offset).Limit(limit).Order("fecha_creacion DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
