package config

import (
	"fmt"
	"log"
	"time"

	"go-jwt-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase establece la conexión con la base de datos usando la config centralizada
func ConnectDatabase(cfg *DatabaseConfig) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	gormLogger := logger.Default.LogMode(logger.Info)
	if getEnv("GIN_MODE", "debug") == "release" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Fatal("Error al conectar con la base de datos:", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Error al obtener la instancia de la base de datos:", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Conexión a la base de datos establecida exitosamente")

	RunMigrations()
}

// RunMigrations ejecuta las migraciones automáticas
func RunMigrations() {
	log.Println("Ejecutando migraciones...")

	err := DB.AutoMigrate(
		&models.User{},
	)

	if err != nil {
		log.Fatal("Error al ejecutar migraciones:", err)
	}

	log.Println("Migraciones ejecutadas exitosamente")

	CreateDefaultAdmin()
}

// CreateDefaultAdmin crea un usuario administrador por defecto
func CreateDefaultAdmin() {
	var adminUser models.User
	result := DB.Where("email = ?", "admin@example.com").First(&adminUser)

	if result.Error == gorm.ErrRecordNotFound {
		hashedPassword, err := hashPassword("admin123")
		if err != nil {
			log.Printf("Error al crear contraseña del admin: %v", err)
			return
		}

		admin := models.User{
			Email:           "admin@example.com",
			NombreDeUsuario: "admin",
			ContraseñaHash:  hashedPassword,
			Rol:             "admin",
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Error al crear usuario administrador: %v", err)
		} else {
			log.Println("Usuario administrador creado: admin@example.com / admin123")
		}
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
