package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/storage"
	"github.com/davidsonmarra/receitas-app/test/testdb"
	"github.com/go-chi/chi/v5"
)

// setupTestRecipe cria uma receita de teste no banco
func setupTestRecipe(userID uint) (*models.Recipe, error) {
	var recipeUserID *uint
	if userID > 0 {
		recipeUserID = &userID
	}
	
	recipe := &models.Recipe{
		Title:       "Receita Teste",
		Description: "Descrição teste",
		PrepTime:    30,
		Servings:    4,
		Difficulty:  "fácil",
		UserID:      recipeUserID,
	}

	if err := database.DB.Create(recipe).Error; err != nil {
		return nil, err
	}

	return recipe, nil
}

// cleanupTestRecipe remove a receita de teste
func cleanupTestRecipe(recipeID uint) {
	database.DB.Delete(&models.Recipe{}, recipeID)
}

// TestUploadRecipeImage_MissingAuth foi removido pois o endpoint legado foi descontinuado
// Use TestGenerateUploadURL_Unauthorized para testar autenticação

// TestUploadRecipeImage_MissingFile foi removido pois o endpoint legado foi descontinuado
// Use TestConfirmImageUpload_InvalidData para testar validação de dados

// TestUploadRecipeImage_RecipeNotFound foi removido pois o endpoint legado foi descontinuado
// Use TestGenerateUploadURL_RecipeNotFound para testar receita não encontrada

func TestDeleteRecipeImage_MissingAuth(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/recipes/1/image", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.DeleteRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rr.Code)
	}
}

func TestDeleteRecipeImage_RecipeNotFound(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário de teste
	user := testdb.SeedUser(t, "Test User", "test3@example.com", "hashed_password", "user")

	req, err := http.NewRequest("DELETE", "/recipes/99999/image", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.DeleteRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", rr.Code)
	}
}

func TestGetRecipeImageVariants_RecipeNotFound(t *testing.T) {
	testdb.SetupWithCleanup(t)

	req, err := http.NewRequest("GET", "/recipes/99999/image/variants", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GetRecipeImageVariants)
	handler.ServeHTTP(rr, req)

	// Deve retornar 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", rr.Code)
	}
}

func TestGetRecipeImageVariants_NoImage(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário e receita sem imagem
	user := testdb.SeedUser(t, "Test User", "test4@example.com", "hashed_password", "user")

	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("/recipes/%d/image/variants", recipe.ID), nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GetRecipeImageVariants)
	handler.ServeHTTP(rr, req)

	// Deve retornar 404 (receita não tem imagem)
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", rr.Code)
	}
}

func TestGetOptimizedRecipeImage_WithQueryParams(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Substituir o ServiceFactory por um mock
	originalFactory := storage.ServiceFactory
	defer func() { storage.ServiceFactory = originalFactory }()
	
	mockService := testdb.NewMockCloudinaryService()
	storage.ServiceFactory = func() (storage.ImageService, error) {
		return mockService, nil
	}

	// Criar usuário e receita com imagem
	user := testdb.SeedUser(t, "Test User", "test5@example.com", "hashed_password", "user")

	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	// Adicionar imagem fake à receita
	recipe.ImagePublicID = "test/recipe_123"
	recipe.ImageURL = "https://res.cloudinary.com/test/image/upload/test/recipe_123"
	database.DB.Save(recipe)

	req, err := http.NewRequest("GET", fmt.Sprintf("/recipes/%d/image/optimized?width=500&height=500&quality=80", recipe.ID), nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GetOptimizedRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rr.Code)
	}

	// Verificar resposta JSON
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	// Verificar campos esperados
	if _, ok := response["url"]; !ok {
		t.Error("resposta não contém campo 'url'")
	}
	if width, ok := response["width"].(float64); !ok || width != 500 {
		t.Errorf("esperado width 500, obteve %v", response["width"])
	}
	if height, ok := response["height"].(float64); !ok || height != 500 {
		t.Errorf("esperado height 500, obteve %v", response["height"])
	}
}
