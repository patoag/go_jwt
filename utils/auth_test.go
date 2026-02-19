package utils

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword falló: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash vacío")
	}

	if hash == password {
		t.Fatal("Hash no debería ser igual al password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testPassword123"
	hash, _ := HashPassword(password)

	if !CheckPasswordHash(password, hash) {
		t.Fatal("CheckPasswordHash debería retornar true para password correcto")
	}

	if CheckPasswordHash("wrongpassword", hash) {
		t.Fatal("CheckPasswordHash debería retornar false para password incorrecto")
	}
}

func TestGenerateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	email := "test@example.com"
	rol := "user"

	token, err := GenerateJWT(userID, email, rol)
	if err != nil {
		t.Fatalf("GenerateJWT falló: %v", err)
	}

	if token == "" {
		t.Fatal("Token vacío")
	}
}

func TestGenerateJWTWithConfig(t *testing.T) {
	userID := uuid.New()
	token, err := GenerateJWTWithConfig(userID, "test@example.com", "user", "my-secret", 24)
	if err != nil {
		t.Fatalf("GenerateJWTWithConfig falló: %v", err)
	}

	if token == "" {
		t.Fatal("Token vacío")
	}
}

func TestValidateJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	email := "test@example.com"
	rol := "admin"

	token, _ := GenerateJWT(userID, email, rol)
	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT falló: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID esperado %s, obtenido %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Email esperado %s, obtenido %s", email, claims.Email)
	}
	if claims.Rol != rol {
		t.Errorf("Rol esperado %s, obtenido %s", rol, claims.Rol)
	}
}

func TestValidateJWTWithSecret(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()

	token, _ := GenerateJWTWithConfig(userID, "test@example.com", "user", secret, 1)

	claims, err := ValidateJWTWithSecret(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWTWithSecret falló: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID no coincide")
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()

	// Crear token expirado manualmente
	claims := JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
		Rol:    "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "go-jwt-backend",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := ValidateJWTWithSecret(tokenString, secret)
	if err == nil {
		t.Fatal("ValidateJWT debería fallar con token expirado")
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	token, _ := GenerateJWTWithConfig(userID, "test@example.com", "user", "secret-1", 24)

	_, err := ValidateJWTWithSecret(token, "secret-2")
	if err == nil {
		t.Fatal("ValidateJWT debería fallar con firma inválida")
	}
}

func TestRefreshJWT(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	token, _ := GenerateJWT(userID, "test@example.com", "user")

	newToken, err := RefreshJWT(token)
	if err != nil {
		t.Fatalf("RefreshJWT falló: %v", err)
	}

	if newToken == "" {
		t.Fatal("Nuevo token vacío")
	}

	if newToken == token {
		t.Log("Nota: los tokens podrían ser iguales si se generan en el mismo segundo")
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"Bearer válido", "Bearer abc123", "abc123"},
		{"Sin Bearer", "abc123", ""},
		{"Vacío", "", ""},
		{"Solo Bearer", "Bearer ", ""},
		{"bearer minúscula", "bearer abc123", ""},
		{"Con espacios extra", "Bearer  abc123", " abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTokenFromHeader(tt.header)
			if result != tt.expected {
				t.Errorf("ExtractTokenFromHeader(%q) = %q, esperado %q", tt.header, result, tt.expected)
			}
		})
	}
}
