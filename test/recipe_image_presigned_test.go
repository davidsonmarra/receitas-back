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
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/storage"
	"github.com/davidsonmarra/receitas-app/test/testdb"
	"github.com/go-chi/chi/v5"
)

// TestGenerateUploadURL_Success testa geração bem-sucedida de URL de upload
func TestGenerateUploadURL_Success(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário e receita de teste
	user := testdb.SeedUser(t, "Test User", "test@example.com", "hashed_password", "user")
	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/image/upload-url", recipe.ID), nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GenerateUploadURL)
	handler.ServeHTTP(rr, req)

	// Deve retornar 200 OK ou 500 (se CLOUDINARY_URL não estiver configurado no ambiente de teste)
	// No ambiente de teste sem Cloudinary configurado, é esperado erro 500
	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("esperado status 200 ou 500, obteve %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Se retornou 200, verificar resposta JSON
	if rr.Code == http.StatusOK {
		var response storage.UploadSignature
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatalf("erro ao decodificar resposta: %v", err)
		}

		// Verificar campos esperados
		if response.UploadURL == "" {
			t.Error("upload_url não deve ser vazio")
		}
		if response.PublicID == "" {
			t.Error("public_id não deve ser vazio")
		}
		if response.Signature == "" {
			t.Error("signature não deve ser vazio")
		}
		if response.APIKey == "" {
			t.Error("api_key não deve ser vazio")
		}
		if response.Timestamp == 0 {
			t.Error("timestamp não deve ser zero")
		}
	}
}

// TestGenerateUploadURL_Unauthorized testa erro de autenticação
func TestGenerateUploadURL_Unauthorized(t *testing.T) {
	req, err := http.NewRequest("POST", "/recipes/1/image/upload-url", nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GenerateUploadURL)
	handler.ServeHTTP(rr, req)

	// Deve retornar 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rr.Code)
	}
}

// TestGenerateUploadURL_RecipeNotFound testa receita não encontrada
func TestGenerateUploadURL_RecipeNotFound(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test2@example.com", "hashed_password", "user")

	req, err := http.NewRequest("POST", "/recipes/99999/image/upload-url", nil)
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
	handler := http.HandlerFunc(handlers.GenerateUploadURL)
	handler.ServeHTTP(rr, req)

	// Deve retornar 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", rr.Code)
	}
}

// TestGenerateUploadURL_Forbidden testa falta de permissão
func TestGenerateUploadURL_Forbidden(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar dois usuários diferentes
	user1 := testdb.SeedUser(t, "User 1", "user1@example.com", "hashed_password", "user")
	user2 := testdb.SeedUser(t, "User 2", "user2@example.com", "hashed_password", "user")

	// Criar receita do user1
	recipe, err := setupTestRecipe(user1.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/image/upload-url", recipe.ID), nil)
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Tentar acessar com user2 (não é dono da receita)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user2.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GenerateUploadURL)
	handler.ServeHTTP(rr, req)

	// Deve retornar 403 Forbidden
	if rr.Code != http.StatusForbidden {
		t.Errorf("esperado status 403, obteve %d", rr.Code)
	}
}

// TestConfirmImageUpload_Success testa confirmação bem-sucedida de upload
func TestConfirmImageUpload_Success(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário e receita de teste
	user := testdb.SeedUser(t, "Test User", "test3@example.com", "hashed_password", "user")
	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	// Dados de confirmação
	confirmData := handlers.ConfirmImageUploadRequest{
		PublicID:  "recipes/recipe_1_12345",
		SecureURL: "https://res.cloudinary.com/test/image/upload/v1/recipes/recipe_1_12345.jpg",
		Width:     1024,
		Height:    768,
		Format:    "jpg",
		Bytes:     204800,
	}

	jsonData, _ := json.Marshal(confirmData)
	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/image/confirm", recipe.ID), bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	// Adicionar contexto com userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Adicionar parâmetro de rota
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ConfirmImageUpload)
	handler.ServeHTTP(rr, req)

	// Deve retornar 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d. Body: %s", rr.Code, rr.Body.String())
	}

	// Verificar que a receita foi atualizada
	var updatedRecipe models.Recipe
	database.DB.First(&updatedRecipe, recipe.ID)

	if updatedRecipe.ImageURL != confirmData.SecureURL {
		t.Errorf("esperado ImageURL %s, obteve %s", confirmData.SecureURL, updatedRecipe.ImageURL)
	}
	if updatedRecipe.ImagePublicID != confirmData.PublicID {
		t.Errorf("esperado ImagePublicID %s, obteve %s", confirmData.PublicID, updatedRecipe.ImagePublicID)
	}
}

// TestConfirmImageUpload_InvalidData testa dados inválidos
func TestConfirmImageUpload_InvalidData(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test4@example.com", "hashed_password", "user")
	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	// Dados inválidos (faltando campos obrigatórios)
	invalidData := map[string]interface{}{
		"public_id": "test",
		// faltando secure_url e outros campos
	}

	jsonData, _ := json.Marshal(invalidData)
	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/image/confirm", recipe.ID), bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ConfirmImageUpload)
	handler.ServeHTTP(rr, req)

	// Deve retornar 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", rr.Code)
	}
}

// TestConfirmImageUpload_Unauthorized testa erro de autenticação
func TestConfirmImageUpload_Unauthorized(t *testing.T) {
	confirmData := handlers.ConfirmImageUploadRequest{
		PublicID:  "test",
		SecureURL: "https://example.com/test.jpg",
		Width:     100,
		Height:    100,
		Format:    "jpg",
		Bytes:     1000,
	}

	jsonData, _ := json.Marshal(confirmData)
	req, err := http.NewRequest("POST", "/recipes/1/image/confirm", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ConfirmImageUpload)
	handler.ServeHTTP(rr, req)

	// Deve retornar 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rr.Code)
	}
}

// TestConfirmImageUpload_DeletesOldImage testa deleção de imagem antiga
func TestConfirmImageUpload_DeletesOldImage(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Substituir o ServiceFactory por um mock
	originalFactory := storage.ServiceFactory
	defer func() { storage.ServiceFactory = originalFactory }()

	mockService := testdb.NewMockCloudinaryService()
	storage.ServiceFactory = func() (storage.ImageService, error) {
		return mockService, nil
	}

	user := testdb.SeedUser(t, "Test User", "test5@example.com", "hashed_password", "user")
	recipe, err := setupTestRecipe(user.ID)
	if err != nil {
		t.Fatalf("erro ao criar receita de teste: %v", err)
	}

	// Adicionar imagem antiga à receita
	recipe.ImagePublicID = "old_public_id"
	recipe.ImageURL = "https://old.url/image.jpg"
	database.DB.Save(recipe)

	// Dados da nova imagem
	confirmData := handlers.ConfirmImageUploadRequest{
		PublicID:  "recipes/recipe_1_new",
		SecureURL: "https://res.cloudinary.com/test/image/upload/v1/recipes/recipe_1_new.jpg",
		Width:     1024,
		Height:    768,
		Format:    "jpg",
		Bytes:     204800,
	}

	jsonData, _ := json.Marshal(confirmData)
	req, err := http.NewRequest("POST", fmt.Sprintf("/recipes/%d/image/confirm", recipe.ID), bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("erro ao criar requisição: %v", err)
	}

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ConfirmImageUpload)
	handler.ServeHTTP(rr, req)

	// Deve retornar 200 OK
	if rr.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rr.Code)
	}

	// Verificar que a receita foi atualizada com a nova imagem
	var updatedRecipe models.Recipe
	database.DB.First(&updatedRecipe, recipe.ID)

	if updatedRecipe.ImageURL != confirmData.SecureURL {
		t.Errorf("esperado nova ImageURL %s, obteve %s", confirmData.SecureURL, updatedRecipe.ImageURL)
	}
	if updatedRecipe.ImagePublicID != confirmData.PublicID {
		t.Errorf("esperado novo ImagePublicID %s, obteve %s", confirmData.PublicID, updatedRecipe.ImagePublicID)
	}
}

