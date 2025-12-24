package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
)

func TestHealthHandler(t *testing.T) {
	// Cria uma requisição HTTP de teste
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}

	// Cria um ResponseRecorder para capturar a resposta
	rr := httptest.NewRecorder()

	// Cria um handler HTTP e executa a requisição
	handler := http.HandlerFunc(handlers.HealthHandler)
	handler.ServeHTTP(rr, req)

	// Verifica o status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler retornou status code errado: esperado %v, recebido %v",
			http.StatusOK, status)
	}

	// Verifica o Content-Type
	contentType := rr.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Errorf("Handler retornou Content-Type errado: esperado %v, recebido %v",
			expectedContentType, contentType)
	}

	// Verifica o corpo da resposta
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Erro ao decodificar resposta JSON: %v", err)
	}

	// Verifica se status está presente e é "healthy"
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Handler retornou status errado: esperado 'healthy', recebido %v",
			response["status"])
	}

	// Verifica se timestamp está presente
	if _, ok := response["timestamp"]; !ok {
		t.Error("Handler não retornou timestamp")
	}
}
