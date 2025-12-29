package test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/go-chi/chi/v5"
)

// setupTestRecipe cria uma receita de teste no banco
func setupTestRecipe(userID uint) (*models.Recipe, error) {
	recipe := &models.Recipe{
		Title:       "Receita Teste",
		Description: "Descrição teste",
		PrepTime:    30,
		Servings:    4,
		Difficulty:  "fácil",
		UserID:      userID,
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

func TestUploadRecipeImage_MissingAuth(t *testing.T) {
	// Criar request sem autenticação
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest("POST", "/recipes/1/image", body)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UploadRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rr.Code)
	}
}

func TestUploadRecipeImage_MissingFile(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}

	// Criar usuário e receita de teste
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashed_password",
	}
	database.DB.Create(user)
	defer database.DB.Delete(user)

	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}
	defer cleanupTestRecipe(recipe.ID)

	// Criar request sem arquivo
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, err := http.NewRequest("POST", "/recipes/"+string(rune(recipe.ID))+"/image", body)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", string(rune(recipe.ID)))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UploadRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar erro de validação
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", rr.Code)
	}
}

func TestUploadRecipeImage_RecipeNotFound(t *testing.T) {
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}

	// Criar usuário de teste
	user := &models.User{
		Name:     "Test User",
		Email:    "test2@example.com",
		Password: "hashed_password",
	}
	database.DB.Create(user)
	defer database.DB.Delete(user)

	// Criar request com ID de receita inexistente
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Adicionar arquivo fake
	part, _ := writer.CreateFormFile("image", "test.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	req, err := http.NewRequest("POST", "/recipes/99999/image", body)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.UploadRecipeImage)
	handler.ServeHTTP(rr, req)

	// Deve retornar 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", rr.Code)
	}
}

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
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}

	// Criar usuário de teste
	user := &models.User{
		Name:     "Test User",
		Email:    "test3@example.com",
		Password: "hashed_password",
	}
	database.DB.Create(user)
	defer database.DB.Delete(user)

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
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}

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
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}

	// Criar usuário e receita sem imagem
	user := &models.User{
		Name:     "Test User",
		Email:    "test4@example.com",
		Password: "hashed_password",
	}
	database.DB.Create(user)
	defer database.DB.Delete(user)

	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}
	defer cleanupTestRecipe(recipe.ID)

	req, err := http.NewRequest("GET", "/recipes/"+string(rune(recipe.ID))+"/image/variants", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", string(rune(recipe.ID)))
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
	if database.DB == nil {
		t.Skip("Database não configurado - pulando teste de integração")
	}
	if os.Getenv("CLOUDINARY_URL") == "" {
		t.Skip("CLOUDINARY_URL não configurada - pulando teste de integração")
	}

	// Criar usuário e receita com imagem
	user := &models.User{
		Name:     "Test User",
		Email:    "test5@example.com",
		Password: "hashed_password",
	}
	database.DB.Create(user)
	defer database.DB.Delete(user)

	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}
	defer cleanupTestRecipe(recipe.ID)

	// Adicionar imagem fake à receita
	recipe.ImagePublicID = "test/recipe_123"
	recipe.ImageURL = "https://res.cloudinary.com/test/image/upload/test/recipe_123"
	database.DB.Save(recipe)

	req, err := http.NewRequest("GET", "/recipes/"+string(rune(recipe.ID))+"/image/optimized?width=500&height=500&quality=80", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", string(rune(recipe.ID)))
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
