package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"go-jwt-backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// DatabaseConfig contiene la configuración de la base de datos
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig obtiene la configuración de la base de datos desde variables de entorno
func GetDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres123"),
		DBName:   getEnv("DB_NAME", "go_jwt_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// ConnectDatabase establece la conexión con la base de datos
func ConnectDatabase() {
	config := GetDatabaseConfig()
	
	// Construir DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Configurar logger de GORM
	gormLogger := logger.Default.LogMode(logger.Info)
	if getEnv("GIN_MODE", "debug") == "release" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Conectar a la base de datos
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

	// Configurar pool de conexiones
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Error al obtener la instancia de la base de datos:", err)
	}

	// Configuraciones del pool de conexiones
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Conexión a la base de datos establecida exitosamente")

	// Ejecutar migraciones
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
	
	// Crear usuario administrador por defecto si no existe
	CreateDefaultAdmin()
}

// CreateDefaultAdmin crea un usuario administrador por defecto
func CreateDefaultAdmin() {
	var adminUser models.User
	result := DB.Where("email = ?", "admin@example.com").First(&adminUser)

	if result.Error == gorm.ErrRecordNotFound {
		// Usar bcrypt para hashear la contraseña
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

// getEnv obtiene una variable de entorno con un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// hashPassword es una función auxiliar para hashear contraseñas usando bcrypt
func hashPassword(password string) (string, error) {
	// Usar bcrypt real para hashear la contraseña
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
