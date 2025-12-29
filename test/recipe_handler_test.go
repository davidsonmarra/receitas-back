package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

func TestCreateRecipe(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário de teste
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")

	recipe := models.Recipe{
		Title:       "Bolo de Chocolate",
		Description: "Delicioso bolo de chocolate",
		PrepTime:    45,
		Servings:    8,
		Difficulty:  "média",
	}

	body, _ := json.Marshal(recipe)
	req, err := http.NewRequest("POST", "/recipes", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}

	// Adicionar contexto com userID (simular middleware RequireAuth)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	ctx = context.WithValue(ctx, middleware.UserEmailKey, user.Email)
	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.CreateRecipe)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Esperado status 201, obteve %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verificar se a receita foi criada
	var response models.Recipe
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	if response.Title != recipe.Title {
		t.Errorf("Título esperado %s, obteve %s", recipe.Title, response.Title)
	}

	if response.UserID == nil || *response.UserID != user.ID {
		t.Errorf("UserID esperado %d, obteve %v", user.ID, response.UserID)
	}
}

func TestListRecipes(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário e algumas receitas de teste
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")

	testdb.SeedRecipe(t, "Receita 1", "Descrição 1", user.ID, false)
	testdb.SeedRecipe(t, "Receita 2", "Descrição 2", user.ID, false)
	testdb.SeedRecipe(t, "Receita Geral", "Receita pública", 0, true)

	req, err := http.NewRequest("GET", "/recipes", nil)
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.ListRecipes)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obteve %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verifica Content-Type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Esperado Content-Type application/json, obteve %v", contentType)
	}

	// Verificar resposta paginada
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	if response["data"] == nil {
		t.Error("Resposta deve conter campo 'data'")
	}

	if response["pagination"] == nil {
		t.Error("Resposta deve conter campo 'pagination'")
	}

	// Verificar que temos 3 receitas
	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatal("Campo 'data' deve ser um array")
	}

	if len(data) != 3 {
		t.Errorf("Esperado 3 receitas, obteve %d", len(data))
	}
}
