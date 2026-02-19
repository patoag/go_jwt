package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"go-jwt-backend/utils"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const testSecret = "test-secret-for-middleware"

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	token, _ := utils.GenerateJWTWithConfig(userID, "test@example.com", "user", testSecret, 24)

	r := gin.New()
	r.Use(AuthMiddleware(testSecret, nil))
	r.GET("/test", func(c *gin.Context) {
		id, _ := GetUserIDFromContext(c)
		c.JSON(200, gin.H{"user_id": id.String()})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Esperado 200, obtenido %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["user_id"] != userID.String() {
		t.Errorf("UserID esperado %s, obtenido %s", userID.String(), resp["user_id"])
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware(testSecret, nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Esperado 401, obtenido %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware(testSecret, nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Esperado 401, obtenido %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	r := gin.New()
	r.Use(AuthMiddleware(testSecret, nil))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Esperado 401, obtenido %d", w.Code)
	}
}

func TestAdminMiddleware_Admin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	r.Use(AdminMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Esperado 200, obtenido %d", w.Code)
	}
}

func TestAdminMiddleware_NotAdmin(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_role", "user")
		c.Next()
	})
	r.Use(AdminMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatalf("Esperado 403, obtenido %d", w.Code)
	}
}

func TestAdminMiddleware_NoRole(t *testing.T) {
	r := gin.New()
	r.Use(AdminMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Fatalf("Esperado 401, obtenido %d", w.Code)
	}
}

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("No se pudo iniciar miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	r := gin.New()
	r.Use(RateLimitMiddleware(rdb, 5, time.Minute))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Hacer 5 requests (deberían pasar)
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("Request %d: Esperado 200, obtenido %d", i+1, w.Code)
		}
	}

	// El 6to request debería ser rechazado
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 429 {
		t.Fatalf("Request 6: Esperado 429, obtenido %d", w.Code)
	}
}

func TestRateLimitMiddleware_NilRedis(t *testing.T) {
	r := gin.New()
	r.Use(RateLimitMiddleware(nil, 5, time.Minute))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Sin Redis, debería pasar (fail-open)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Esperado 200 (fail-open), obtenido %d", w.Code)
	}
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("CORS debería permitir todos los orígenes")
	}
}

func TestCORSMiddleware_SpecificOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware([]string{"https://example.com", "https://app.example.com"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	// Origin permitido
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatal("CORS debería permitir el origen especificado")
	}

	// Origin no permitido
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "https://malicious.com")
	r.ServeHTTP(w2, req2)

	if w2.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("CORS no debería permitir orígenes no especificados")
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Fatalf("OPTIONS debería retornar 204, obtenido %d", w.Code)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Sin user_id
	_, err := GetUserIDFromContext(c)
	if err == nil {
		t.Fatal("Debería fallar sin user_id en contexto")
	}

	// Con user_id válido
	expectedID := uuid.New()
	c.Set("user_id", expectedID)
	id, err := GetUserIDFromContext(c)
	if err != nil {
		t.Fatalf("GetUserIDFromContext falló: %v", err)
	}
	if id != expectedID {
		t.Errorf("ID esperado %s, obtenido %s", expectedID, id)
	}
}

func TestGetUserRoleFromContext(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Sin user_role
	_, err := GetUserRoleFromContext(c)
	if err == nil {
		t.Fatal("Debería fallar sin user_role en contexto")
	}

	// Con user_role
	c.Set("user_role", "admin")
	role, err := GetUserRoleFromContext(c)
	if err != nil {
		t.Fatalf("GetUserRoleFromContext falló: %v", err)
	}
	if role != "admin" {
		t.Errorf("Rol esperado admin, obtenido %s", role)
	}
}
