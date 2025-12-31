package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

// TestAddRecipeIngredient_Success testa adição de ingrediente com sucesso
func TestAddRecipeIngredient_Success(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário, receita e ingrediente
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)
	ingredient := testdb.SeedIngredient(t, "Farinha de Trigo", "Grãos", 364.0)

	// Request body usando o DTO
	reqBody := map[string]interface{}{
		"ingredient_id": ingredient.ID,
		"quantity":      200.0,
		"unit":          "g",
		"notes":         "Peneirada",
		"order":         1,
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/ingredients", recipe.ID), bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Erro ao criar requisição: %v", err)
	}

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)
	
	// Adicionar parâmetro de URL do Chi
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.AddRecipeIngredient)
	handler.ServeHTTP(rr, req)

	// Verificar resposta
	if rr.Code != http.StatusCreated {
		t.Errorf("Status code esperado %d, recebido %d. Body: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var result models.RecipeIngredient
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	// Validar campos
	if result.IngredientID != ingredient.ID {
		t.Errorf("IngredientID esperado %d, recebido %d", ingredient.ID, result.IngredientID)
	}
	if result.Quantity != 200.0 {
		t.Errorf("Quantity esperado 200.0, recebido %f", result.Quantity)
	}
	if result.Unit != "g" {
		t.Errorf("Unit esperado 'g', recebido '%s'", result.Unit)
	}
	if result.Notes != "Peneirada" {
		t.Errorf("Notes esperado 'Peneirada', recebido '%s'", result.Notes)
	}
}

// TestAddRecipeIngredient_ValidationErrors testa erros de validação
func TestAddRecipeIngredient_ValidationErrors(t *testing.T) {
	testdb.SetupWithCleanup(t)

	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)

	testCases := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "Quantidade negativa",
			body: map[string]interface{}{
				"ingredient_id": 1,
				"quantity":      -100.0,
				"unit":          "g",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "quantity",
		},
		{
			name: "Quantidade zero",
			body: map[string]interface{}{
				"ingredient_id": 1,
				"quantity":      0,
				"unit":          "g",
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "quantity",
		},
		{
			name: "Sem ingredient_id",
			body: map[string]interface{}{
				"quantity": 100.0,
				"unit":     "g",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Sem unit",
			body: map[string]interface{}{
				"ingredient_id": 1,
				"quantity":      100.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/ingredients", recipe.ID), bytes.NewBuffer(body))

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
			req = req.WithContext(ctx)
			ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.AddRecipeIngredient)
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("Status code esperado %d, recebido %d. Body: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// TestAddRecipeIngredient_IngredientNotFound testa ingrediente inexistente
func TestAddRecipeIngredient_IngredientNotFound(t *testing.T) {
	testdb.SetupWithCleanup(t)

	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)

	reqBody := map[string]interface{}{
		"ingredient_id": 999, // ID inexistente
		"quantity":      100.0,
		"unit":          "g",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/ingredients", recipe.ID), bytes.NewBuffer(body))

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.AddRecipeIngredient)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status code esperado %d, recebido %d", http.StatusBadRequest, rr.Code)
	}
}

// TestAddRecipeIngredient_Unauthorized testa sem permissão
func TestAddRecipeIngredient_Unauthorized(t *testing.T) {
	testdb.SetupWithCleanup(t)

	hashedPassword, _ := auth.HashPassword("senha123")
	owner := testdb.SeedUser(t, "Owner", "owner@example.com", hashedPassword, "user")
	otherUser := testdb.SeedUser(t, "Other User", "other@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", owner.ID, false)
	ingredient := testdb.SeedIngredient(t, "Farinha de Trigo", "Grãos", 364.0)

	reqBody := map[string]interface{}{
		"ingredient_id": ingredient.ID,
		"quantity":      100.0,
		"unit":          "g",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/ingredients", recipe.ID), bytes.NewBuffer(body))

	// Usar outro usuário (não o dono)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, otherUser.ID)
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.AddRecipeIngredient)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Status code esperado %d, recebido %d", http.StatusForbidden, rr.Code)
	}
}

// TestUpdateRecipeIngredient_Success testa atualização com sucesso
func TestUpdateRecipeIngredient_Success(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Setup
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)
	ingredient := testdb.SeedIngredient(t, "Farinha de Trigo", "Grãos", 364.0)

	// Criar recipe ingredient
	recipeIng := &models.RecipeIngredient{
		RecipeID:     recipe.ID,
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "g",
		Order:        1,
	}
	database.DB.Create(recipeIng)

	// Atualizar usando DTO
	updateBody := map[string]interface{}{
		"quantity": 200.0,
		"unit":     "kg",
		"notes":    "Bem peneirada",
	}

	body, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/recipes/%d/ingredients/%d", recipe.ID, recipeIng.ID), bytes.NewBuffer(body))

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "ingredient_id", fmt.Sprint(recipeIng.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateRecipeIngredient)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status code esperado %d, recebido %d. Body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var result models.RecipeIngredient
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}

	// Validar atualização
	if result.Quantity != 200.0 {
		t.Errorf("Quantity esperado 200.0, recebido %f", result.Quantity)
	}
	if result.Unit != "kg" {
		t.Errorf("Unit esperado 'kg', recebido '%s'", result.Unit)
	}
	if result.Notes != "Bem peneirada" {
		t.Errorf("Notes esperado 'Bem peneirada', recebido '%s'", result.Notes)
	}
}

// TestUpdateRecipeIngredient_PartialUpdate testa atualização parcial
func TestUpdateRecipeIngredient_PartialUpdate(t *testing.T) {
	testdb.SetupWithCleanup(t)

	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)
	ingredient := testdb.SeedIngredient(t, "Farinha de Trigo", "Grãos", 364.0)

	recipeIng := &models.RecipeIngredient{
		RecipeID:     recipe.ID,
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "g",
		Notes:        "Original",
		Order:        1,
	}
	database.DB.Create(recipeIng)

	// Atualizar apenas quantity
	updateBody := map[string]interface{}{
		"quantity": 150.0,
	}

	body, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/recipes/%d/ingredients/%d", recipe.ID, recipeIng.ID), bytes.NewBuffer(body))

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "ingredient_id", fmt.Sprint(recipeIng.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateRecipeIngredient)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Status code esperado %d, recebido %d", http.StatusOK, rr.Code)
	}

	var result models.RecipeIngredient
	json.Unmarshal(rr.Body.Bytes(), &result)

	// Verificar que apenas quantity mudou
	if result.Quantity != 150.0 {
		t.Errorf("Quantity esperado 150.0, recebido %f", result.Quantity)
	}
	if result.Unit != "g" { // Deve manter original
		t.Errorf("Unit deveria manter 'g', recebido '%s'", result.Unit)
	}
	if result.Notes != "Original" { // Deve manter original
		t.Errorf("Notes deveria manter 'Original', recebido '%s'", result.Notes)
	}
}

// TestUpdateRecipeIngredient_InvalidQuantity testa validação de quantidade inválida
func TestUpdateRecipeIngredient_InvalidQuantity(t *testing.T) {
	testdb.SetupWithCleanup(t)

	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Test User", "test@example.com", hashedPassword, "user")
	recipe := testdb.SeedRecipe(t, "Bolo de Chocolate", "Delicioso bolo", user.ID, false)
	ingredient := testdb.SeedIngredient(t, "Farinha de Trigo", "Grãos", 364.0)

	recipeIng := &models.RecipeIngredient{
		RecipeID:     recipe.ID,
		IngredientID: ingredient.ID,
		Quantity:     100.0,
		Unit:         "g",
	}
	database.DB.Create(recipeIng)

	// Tentar atualizar com quantidade inválida
	updateBody := map[string]interface{}{
		"quantity": -50.0, // Negativo
	}

	body, _ := json.Marshal(updateBody)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/recipes/%d/ingredients/%d", recipe.ID, recipeIng.ID), bytes.NewBuffer(body))

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "id", fmt.Sprint(recipe.ID))
	req = req.WithContext(ctx)
	ctx = testdb.AddChiURLParam(req, "ingredient_id", fmt.Sprint(recipeIng.ID))
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UpdateRecipeIngredient)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status code esperado %d, recebido %d", http.StatusBadRequest, rr.Code)
	}
}

