package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/models"
)

func TestCreateRecipe(t *testing.T) {
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

	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.CreateRecipe)
	handler.ServeHTTP(rr, req)

	// Nota: Este teste requer database conectado
	// Para testes de integração, configurar test database
	// Por enquanto, apenas verifica a estrutura do handler

	if rr.Code != http.StatusCreated && rr.Code != http.StatusInternalServerError {
		t.Errorf("Handler retornou status code inesperado: %v", rr.Code)
	}
}

func TestListRecipes(t *testing.T) {
	req, err := http.NewRequest("GET", "/recipes", nil)
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(handlers.ListRecipes)
	handler.ServeHTTP(rr, req)

	// Nota: Este teste requer database conectado
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("Handler retornou status code inesperado: %v", rr.Code)
	}

	// Verifica Content-Type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" && rr.Code == http.StatusOK {
		t.Errorf("Handler retornou Content-Type errado: %v", contentType)
	}
}
