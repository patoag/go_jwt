package services

import (
	"testing"

	"go-jwt-backend/models"
	"go-jwt-backend/utils"

	"github.com/google/uuid"
)

func createTestUser(repo *mockUserRepo, email, username string) *models.User {
	hash, _ := utils.HashPassword("password123")
	user := &models.User{
		ID:              uuid.New(),
		Email:           email,
		NombreDeUsuario: username,
		ContraseñaHash:  hash,
		Rol:             "user",
	}
	repo.Create(user)
	return user
}

func TestGetProfile_OK(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)
	user := createTestUser(repo, "test@example.com", "testuser")

	result, err := svc.GetProfile(user.ID)
	if err != nil {
		t.Fatalf("GetProfile falló: %v", err)
	}
	if result.Email != "test@example.com" {
		t.Errorf("Email no coincide")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetProfile(uuid.New())
	if err != ErrUserNotFound {
		t.Fatalf("Esperado ErrUserNotFound, obtenido: %v", err)
	}
}

func TestUpdateProfile_OK(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)
	user := createTestUser(repo, "test@example.com", "testuser")

	newEmail := "new@example.com"
	newUsername := "newuser"

	result, err := svc.UpdateProfile(user.ID, &newEmail, &newUsername)
	if err != nil {
		t.Fatalf("UpdateProfile falló: %v", err)
	}
	if result.Email != newEmail {
		t.Errorf("Email esperado %s, obtenido %s", newEmail, result.Email)
	}
	if result.NombreDeUsuario != newUsername {
		t.Errorf("Username esperado %s, obtenido %s", newUsername, result.NombreDeUsuario)
	}
}

func TestUpdateProfile_EmailDuplicado(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	createTestUser(repo, "existing@example.com", "existing")
	user2 := createTestUser(repo, "test@example.com", "testuser")

	existingEmail := "existing@example.com"
	_, err := svc.UpdateProfile(user2.ID, &existingEmail, nil)
	if err != ErrEmailAlreadyExists {
		t.Fatalf("Esperado ErrEmailAlreadyExists, obtenido: %v", err)
	}
}

func TestDeleteAccount_OK(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)
	user := createTestUser(repo, "test@example.com", "testuser")

	err := svc.DeleteAccount(user.ID)
	if err != nil {
		t.Fatalf("DeleteAccount falló: %v", err)
	}

	_, err = svc.GetProfile(user.ID)
	if err != ErrUserNotFound {
		t.Fatal("Usuario debería haber sido eliminado")
	}
}

func TestDeleteAccount_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	err := svc.DeleteAccount(uuid.New())
	if err != ErrUserNotFound {
		t.Fatalf("Esperado ErrUserNotFound, obtenido: %v", err)
	}
}

func TestListUsers(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	createTestUser(repo, "user1@example.com", "user1")
	createTestUser(repo, "user2@example.com", "user2")
	createTestUser(repo, "user3@example.com", "user3")

	result, err := svc.ListUsers(1, 2)
	if err != nil {
		t.Fatalf("ListUsers falló: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total esperado 3, obtenido %d", result.Total)
	}
	if len(result.Users) != 2 {
		t.Errorf("Usuarios en página esperados 2, obtenidos %d", len(result.Users))
	}
	if result.TotalPages != 2 {
		t.Errorf("TotalPages esperado 2, obtenido %d", result.TotalPages)
	}
}

func TestListUsers_PaginaInvalida(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	result, err := svc.ListUsers(-1, 0)
	if err != nil {
		t.Fatalf("ListUsers falló: %v", err)
	}
	// Debería normalizar page=1 y pageSize=10
	if result.Page != 1 {
		t.Errorf("Page esperado 1, obtenido %d", result.Page)
	}
	if result.PageSize != 10 {
		t.Errorf("PageSize esperado 10, obtenido %d", result.PageSize)
	}
}

func TestGetUserByID_OK(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)
	user := createTestUser(repo, "test@example.com", "testuser")

	result, err := svc.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID falló: %v", err)
	}
	if result.ID != user.ID {
		t.Errorf("ID no coincide")
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetUserByID(uuid.New())
	if err != ErrUserNotFound {
		t.Fatalf("Esperado ErrUserNotFound, obtenido: %v", err)
	}
}
