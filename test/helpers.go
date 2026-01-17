package test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/davidsonmarra/receitas-app/internal/http/routes"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

// setupRouter cria um router configurado para testes
func setupRouter() *chi.Mux {
	return routes.Setup()
}

// createTestUser cria um usuário de teste usando testdb.SeedUser
func createTestUser(t *testing.T, email, password, name string) *models.User {
	t.Helper()
	
	// Hash da senha
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	
	return testdb.SeedUser(t, name, email, hashedPassword, "user")
}

// createTestRecipe cria uma receita de teste usando testdb.SeedRecipe
func createTestRecipe(t *testing.T, userID uint) *models.Recipe {
	t.Helper()
	return testdb.SeedRecipe(t, "Test Recipe", "Test Description", userID, false)
}

// loginTestUser faz login e retorna o token JWT
func loginTestUser(t *testing.T, router *chi.Mux, email, password string) string {
	t.Helper()
	
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(loginData)
	
	req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	
	if rec.Code != 200 {
		t.Fatalf("login failed with status %d: %s", rec.Code, rec.Body.String())
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}
	
	// Tentar primeiro access_token (novo formato)
	if token, ok := response["access_token"].(string); ok {
		return token
	}
	
	// Fallback para token (formato antigo)
	if token, ok := response["token"].(string); ok {
		return token
	}
	
	t.Fatal("no token found in login response")
	return ""
}

