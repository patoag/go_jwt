package utils

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Email válido", "user@example.com", false},
		{"Email con subdomain", "user@sub.example.com", false},
		{"Email con +", "user+tag@example.com", false},
		{"Email vacío", "", true},
		{"Sin @", "userexample.com", true},
		{"Sin dominio", "user@", true},
		{"Sin usuario", "@example.com", true},
		{"Sin TLD", "user@example", true},
		{"Espacios", "user @example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Password válido", "password123", false},
		{"Password mínimo", "123456", false},
		{"Password vacío", "", true},
		{"Password muy corto", "12345", true},
		{"Password 100 chars", string(make([]byte, 100)), false},
		{"Password 101 chars", string(make([]byte, 101)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Username válido", "testuser", false},
		{"Con números", "user123", false},
		{"Con guión bajo", "test_user", false},
		{"Con punto", "test.user", false},
		{"Con guión", "test-user", false},
		{"Mínimo 3 chars", "abc", false},
		{"Vacío", "", true},
		{"Muy corto", "ab", true},
		{"51 chars", string(make([]byte, 51)), true},
		{"Con espacios", "test user", true},
		{"Con @", "test@user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{"Rol user", "user", false},
		{"Rol admin", "admin", false},
		{"Rol inválido", "superadmin", true},
		{"Rol vacío", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRole(%q) error = %v, wantErr %v", tt.role, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserCreateRequest(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		username  string
		password  string
		wantCount int
	}{
		{"Todo válido", "user@example.com", "testuser", "password123", 0},
		{"Todo inválido", "", "", "", 3},
		{"Solo email inválido", "invalid", "testuser", "password123", 1},
		{"Solo password corto", "user@example.com", "testuser", "12345", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateUserCreateRequest(tt.email, tt.username, tt.password)
			if len(errors) != tt.wantCount {
				t.Errorf("ValidateUserCreateRequest errores = %d, esperados %d", len(errors), tt.wantCount)
			}
		})
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		wantCount int
	}{
		{"Todo válido", "user@example.com", "password123", 0},
		{"Email inválido", "invalid", "password123", 1},
		{"Password vacío", "user@example.com", "", 1},
		{"Todo inválido", "", "", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateLoginRequest(tt.email, tt.password)
			if len(errors) != tt.wantCount {
				t.Errorf("ValidateLoginRequest errores = %d, esperados %d", len(errors), tt.wantCount)
			}
		})
	}
}

func TestValidateUserUpdateRequest(t *testing.T) {
	email := "new@example.com"
	username := "newuser"
	badEmail := "invalid"

	tests := []struct {
		name      string
		email     *string
		username  *string
		wantCount int
	}{
		{"Nada que actualizar", nil, nil, 0},
		{"Email válido", &email, nil, 0},
		{"Username válido", nil, &username, 0},
		{"Email inválido", &badEmail, nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateUserUpdateRequest(tt.email, tt.username)
			if len(errors) != tt.wantCount {
				t.Errorf("ValidateUserUpdateRequest errores = %d, esperados %d", len(errors), tt.wantCount)
			}
		})
	}
}
