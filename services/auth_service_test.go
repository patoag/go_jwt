package services

import (
	"testing"

	"go-jwt-backend/models"
	"go-jwt-backend/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Mocks ---

type mockUserRepo struct {
	users map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*models.User)}
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) FindByUsername(username string) (*models.User, error) {
	for _, u := range m.users {
		if u.NombreDeUsuario == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) FindByID(id uuid.UUID) (*models.User, error) {
	user, ok := m.users[id.String()]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return user, nil
}

func (m *mockUserRepo) Create(user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID.String()] = user
	return nil
}

func (m *mockUserRepo) Update(user *models.User, updates map[string]interface{}) error {
	for k, v := range updates {
		switch k {
		case "email":
			user.Email = v.(string)
		case "nombre_de_usuario":
			user.NombreDeUsuario = v.(string)
		case "contraseña_hash":
			user.ContraseñaHash = v.(string)
		}
	}
	m.users[user.ID.String()] = user
	return nil
}

func (m *mockUserRepo) Delete(user *models.User) error {
	delete(m.users, user.ID.String())
	return nil
}

func (m *mockUserRepo) List(offset, limit int) ([]models.User, int64, error) {
	var result []models.User
	for _, u := range m.users {
		result = append(result, *u)
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []models.User{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

type mockBlacklist struct {
	tokens map[string]bool
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{tokens: make(map[string]bool)}
}

func (m *mockBlacklist) Add(token string, remainingTTL int64) error {
	m.tokens[token] = true
	return nil
}

func (m *mockBlacklist) IsBlacklisted(token string) (bool, error) {
	return m.tokens[token], nil
}

// --- Tests ---

const testSecret = "test-secret-key-for-testing"

func TestRegister_OK(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	token, user, err := svc.Register("test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("Register falló: %v", err)
	}
	if token == "" {
		t.Fatal("Token vacío")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email esperado test@example.com, obtenido %s", user.Email)
	}
	if user.Rol != "user" {
		t.Errorf("Rol esperado user, obtenido %s", user.Rol)
	}
}

func TestRegister_EmailDuplicado(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	svc.Register("test@example.com", "user1", "password123")
	_, _, err := svc.Register("test@example.com", "user2", "password123")

	if err != ErrEmailAlreadyExists {
		t.Fatalf("Esperado ErrEmailAlreadyExists, obtenido: %v", err)
	}
}

func TestRegister_UsernameDuplicado(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	svc.Register("test1@example.com", "testuser", "password123")
	_, _, err := svc.Register("test2@example.com", "testuser", "password123")

	if err != ErrUsernameAlreadyExists {
		t.Fatalf("Esperado ErrUsernameAlreadyExists, obtenido: %v", err)
	}
}

func TestLogin_OK(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	svc.Register("test@example.com", "testuser", "password123")

	token, user, err := svc.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("Login falló: %v", err)
	}
	if token == "" {
		t.Fatal("Token vacío")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Email no coincide")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	svc.Register("test@example.com", "testuser", "password123")

	_, _, err := svc.Login("test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Fatalf("Esperado ErrInvalidCredentials, obtenido: %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	_, _, err := svc.Login("noexiste@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Fatalf("Esperado ErrInvalidCredentials, obtenido: %v", err)
	}
}

func TestLogout(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	token, _, _ := svc.Register("test@example.com", "testuser", "password123")

	err := svc.Logout(token)
	if err != nil {
		t.Fatalf("Logout falló: %v", err)
	}

	blacklisted, err := bl.IsBlacklisted(token)
	if err != nil {
		t.Fatalf("IsBlacklisted falló: %v", err)
	}
	if !blacklisted {
		t.Fatal("Token debería estar blacklisted después del logout")
	}
}

func TestChangePassword_OK(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	_, user, _ := svc.Register("test@example.com", "testuser", "password123")

	err := svc.ChangePassword(user.ID, "password123", "newpassword456")
	if err != nil {
		t.Fatalf("ChangePassword falló: %v", err)
	}

	// Verificar que se puede login con la nueva contraseña
	_, _, err = svc.Login("test@example.com", "newpassword456")
	if err != nil {
		t.Fatalf("Login con nueva contraseña falló: %v", err)
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	repo := newMockUserRepo()
	bl := newMockBlacklist()
	svc := NewAuthService(repo, bl, testSecret, 24)

	_, user, _ := svc.Register("test@example.com", "testuser", "password123")

	err := svc.ChangePassword(user.ID, "wrongpassword", "newpassword456")
	if err != ErrWrongPassword {
		t.Fatalf("Esperado ErrWrongPassword, obtenido: %v", err)
	}
}

func TestIsTokenBlacklisted(t *testing.T) {
	bl := newMockBlacklist()
	svc := NewAuthService(newMockUserRepo(), bl, testSecret, 24)

	userID := uuid.New()
	token, _ := utils.GenerateJWTWithConfig(userID, "test@example.com", "user", testSecret, 24)

	blacklisted, _ := svc.IsTokenBlacklisted(token)
	if blacklisted {
		t.Fatal("Token no debería estar blacklisted")
	}

	bl.Add(token, 3600)

	blacklisted, _ = svc.IsTokenBlacklisted(token)
	if !blacklisted {
		t.Fatal("Token debería estar blacklisted")
	}
}
