package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/routes"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

func TestCreateRecipe_ValidationErrors(t *testing.T) {
	testdb.SetupWithCleanup(t)
	
	// Criar usuário de teste e gerar token
	user := testdb.SeedUser(t, "Test User", "validation@test.com", "hashed_password", "user")
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}
	
	router := routes.Setup()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "título vazio",
			payload:        map[string]interface{}{"title": "", "prep_time": 30, "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "título é obrigatório",
		},
		{
			name:           "título muito curto",
			payload:        map[string]interface{}{"title": "AB", "prep_time": 30, "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "título deve ter no mínimo 3 caracteres",
		},
		{
			name:           "título muito longo",
			payload:        map[string]interface{}{"title": strings.Repeat("A", 201), "prep_time": 30, "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "título deve ter no máximo 200 caracteres",
		},
		{
			name:           "prep_time ausente",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "tempo de preparo é obrigatório",
		},
		{
			name:           "prep_time zero",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "prep_time": 0, "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "tempo de preparo",
		},
		{
			name:           "prep_time negativo",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "prep_time": -10, "servings": 4},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "tempo de preparo deve ser no mínimo 1",
		},
		{
			name:           "servings ausente",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "prep_time": 30},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "número de porções é obrigatório",
		},
		{
			name:           "servings zero",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "prep_time": 30, "servings": 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "número de porções",
		},
		{
			name:           "difficulty inválida",
			payload:        map[string]interface{}{"title": "Bolo de Chocolate", "prep_time": 30, "servings": 4, "difficulty": "impossível"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "dificuldade deve ser uma das opções",
		},
		{
			name:           "múltiplos erros - retorna apenas o primeiro",
			payload:        map[string]interface{}{"title": "AB", "prep_time": -1, "servings": 0},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			errorContains:  "título", // Apenas o primeiro erro
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError {
				var response map[string]interface{}
				json.NewDecoder(w.Body).Decode(&response)

				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					if title, ok := errorObj["title"].(string); !ok || title != "Ops, algo deu errado!" {
						t.Errorf("Expected error title 'Ops, algo deu errado!', got %v", title)
					}

					if message, ok := errorObj["message"].(string); ok {
						if !strings.Contains(strings.ToLower(message), strings.ToLower(tt.errorContains)) {
							t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorContains, message)
						}
					} else {
						t.Errorf("Expected error message, got none")
					}
				} else {
					t.Errorf("Expected error object in response, got %v", response)
				}
			}
		})
	}
}
