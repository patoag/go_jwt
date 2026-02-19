package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims representa las claims del JWT
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Rol    string    `json:"rol"`
	jwt.RegisteredClaims
}

// HashPassword hashea una contraseña usando bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifica si una contraseña coincide con su hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT genera un token JWT para un usuario (usa env vars)
func GenerateJWT(userID uuid.UUID, email, rol string) (string, error) {
	return GenerateJWTWithConfig(userID, email, rol, getJWTSecret(), 24)
}

// GenerateJWTWithConfig genera un token JWT con configuración explícita
func GenerateJWTWithConfig(userID uuid.UUID, email, rol, secret string, expirationHours int) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Rol:    rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-jwt-backend",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT valida un token JWT y retorna las claims (usa env vars)
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	return ValidateJWTWithSecret(tokenString, getJWTSecret())
}

// ValidateJWTWithSecret valida un token JWT con un secret explícito
func ValidateJWTWithSecret(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inválido")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

// RefreshJWT genera un nuevo token JWT basado en uno existente válido
func RefreshJWT(tokenString string) (string, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return "", err
	}
	return GenerateJWT(claims.UserID, claims.Email, claims.Rol)
}

// getJWTSecret obtiene la clave secreta para JWT desde variables de entorno
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "mi-super-secreto-jwt-key-2024-desarrollo"
	}
	return secret
}

// ExtractTokenFromHeader extrae el token del header Authorization
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
