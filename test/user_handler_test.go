package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

// setupTestDB inicializa um database de teste
func setupUserTestDB(t *testing.T) {
	if database.DB == nil {
		// Para testes, usar um banco em memória ou configurar DATABASE_URL
		// Por ora, vamos pular testes que requerem DB se não estiver configurado
		t.Skip("DATABASE_URL não configurado para testes")
	}

	// Limpar tabela users para testes isolados
	database.DB.Exec("DELETE FROM users")

	// Executar migrations
	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("erro ao executar migrations: %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	setupUserTestDB(t)

	payload := map[string]string{
		"name":     "João Silva",
		"email":    "joao@test.com",
		"password": "senha123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("esperado status 201, obteve %d", rec.Code)
		t.Logf("Response body: %s", rec.Body.String())
	}

	// Verificar resposta
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	if response["token"] == nil {
		t.Error("resposta deve conter token")
	}

	if response["user"] == nil {
		t.Error("resposta deve conter dados do usuário")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupUserTestDB(t)

	// Criar primeiro usuário
	payload := map[string]string{
		"name":     "Primeiro Usuário",
		"email":    "duplicate@test.com",
		"password": "senha123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("primeiro registro falhou: %d", rec.Code)
	}

	// Tentar criar segundo usuário com mesmo email
	payload2 := map[string]string{
		"name":     "Segundo Usuário",
		"email":    "duplicate@test.com",
		"password": "outrasenha",
	}

	body2, _ := json.Marshal(payload2)
	req2 := httptest.NewRequest("POST", "/users/register", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	handlers.Register(rec2, req2)

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400 para email duplicado, obteve %d", rec2.Code)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	setupUserTestDB(t)

	tests := []struct {
		name     string
		payload  map[string]string
		expected int
	}{
		{
			name:     "nome muito curto",
			payload:  map[string]string{"name": "Jo", "email": "jo@test.com", "password": "senha123"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "email inválido",
			payload:  map[string]string{"name": "João", "email": "email-invalido", "password": "senha123"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "senha muito curta",
			payload:  map[string]string{"name": "João", "email": "joao@test.com", "password": "123"},
			expected: http.StatusBadRequest,
		},
		{
			name:     "campos faltando",
			payload:  map[string]string{"name": "João"},
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/users/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handlers.Register(rec, req)

			if rec.Code != tt.expected {
				t.Errorf("esperado status %d, obteve %d", tt.expected, rec.Code)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	setupUserTestDB(t)

	// Criar usuário primeiro
	hashedPassword, _ := auth.HashPassword("senha123")
	user := models.User{
		Name:     "Login Test",
		Email:    "login@test.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	// Tentar login
	payload := map[string]string{
		"email":    "login@test.com",
		"password": "senha123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}

	// Verificar resposta
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	if response["token"] == nil {
		t.Error("resposta deve conter token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupUserTestDB(t)

	// Criar usuário
	hashedPassword, _ := auth.HashPassword("senha-correta")
	user := models.User{
		Name:     "Wrong Password Test",
		Email:    "wrongpass@test.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	// Tentar login com senha errada
	payload := map[string]string{
		"email":    "wrongpass@test.com",
		"password": "senha-errada",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	setupUserTestDB(t)

	payload := map[string]string{
		"email":    "naoexiste@test.com",
		"password": "senha123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handlers.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestLogout_Success(t *testing.T) {
	// Gerar token válido
	token, err := auth.GenerateToken(123, "logout@test.com")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	req := httptest.NewRequest("POST", "/users/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handlers.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
	}

	// Verificar se token foi adicionado à blacklist
	if !auth.IsBlacklisted(token) {
		t.Error("token deveria estar na blacklist após logout")
	}
}

func TestLogout_NoToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/users/logout", nil)
	rec := httptest.NewRecorder()

	handlers.Logout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	req := httptest.NewRequest("POST", "/users/logout", nil)
	req.Header.Set("Authorization", "Bearer token.invalido.xyz")
	rec := httptest.NewRecorder()

	handlers.Logout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}
