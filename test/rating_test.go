package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRatingCreation testa a criação de uma avaliação
func TestRatingCreation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e fazer login
	user := createTestUser(t, "rating_user@test.com", "password123", "Rating User")
	token := loginTestUser(t, router, "rating_user@test.com", "password123")

	// Criar uma receita
	recipe := createTestRecipe(t, user.ID)

	// Criar avaliação
	ratingData := map[string]interface{}{
		"score":   5,
		"comment": "Receita maravilhosa! Ficou perfeita.",
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response models.Rating
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 5, response.Score)
	assert.Equal(t, "Receita maravilhosa! Ficou perfeita.", response.Comment)
	assert.Equal(t, recipe.ID, response.RecipeID)
	assert.Equal(t, user.ID, response.UserID)
}

// TestRatingUpdate testa a atualização de uma avaliação (upsert)
func TestRatingUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e fazer login
	user := createTestUser(t, "rating_update@test.com", "password123", "Rating Update User")
	token := loginTestUser(t, router, "rating_update@test.com", "password123")

	// Criar uma receita
	recipe := createTestRecipe(t, user.ID)

	// Criar avaliação inicial
	ratingData := map[string]interface{}{
		"score":   4,
		"comment": "Boa receita",
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Atualizar avaliação (mesmo endpoint)
	updatedRatingData := map[string]interface{}{
		"score":   5,
		"comment": "Receita excelente! Mudei de ideia.",
	}
	body, _ = json.Marshal(updatedRatingData)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code) // 200 para update

	var response models.Rating
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 5, response.Score)
	assert.Equal(t, "Receita excelente! Mudei de ideia.", response.Comment)

	// Verificar que só existe uma avaliação no banco
	var count int64
	database.DB.Model(&models.Rating{}).Where("recipe_id = ? AND user_id = ?", recipe.ID, user.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// TestRatingValidation testa validações de avaliação
func TestRatingValidation(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e fazer login
	user := createTestUser(t, "rating_validation@test.com", "password123", "Rating Validation User")
	token := loginTestUser(t, router, "rating_validation@test.com", "password123")

	// Criar uma receita
	recipe := createTestRecipe(t, user.ID)

	// Teste 1: Score inválido (maior que 5)
	ratingData := map[string]interface{}{
		"score":   6,
		"comment": "Teste",
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Teste 2: Score inválido (menor que 1)
	ratingData = map[string]interface{}{
		"score":   0,
		"comment": "Teste",
	}
	body, _ = json.Marshal(ratingData)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Teste 3: Comentário muito longo (mais de 1000 caracteres)
	longComment := ""
	for i := 0; i < 1001; i++ {
		longComment += "a"
	}
	ratingData = map[string]interface{}{
		"score":   5,
		"comment": longComment,
	}
	body, _ = json.Marshal(ratingData)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Teste 4: Rating válido sem comentário (comentário é opcional)
	ratingData = map[string]interface{}{
		"score": 5,
	}
	body, _ = json.Marshal(ratingData)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

// TestGetMyRating testa obter a avaliação do usuário logado
func TestGetMyRating(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e fazer login
	user := createTestUser(t, "get_rating@test.com", "password123", "Get Rating User")
	token := loginTestUser(t, router, "get_rating@test.com", "password123")

	// Criar uma receita
	recipe := createTestRecipe(t, user.ID)

	// Tentar obter avaliação antes de criar (deve retornar 404)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings/me", recipe.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Criar avaliação
	ratingData := map[string]interface{}{
		"score":   4,
		"comment": "Boa receita!",
	}
	body, _ := json.Marshal(ratingData)

	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Obter avaliação (deve funcionar agora)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings/me", recipe.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.Rating
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 4, response.Score)
	assert.Equal(t, "Boa receita!", response.Comment)
}

// TestDeleteRating testa a exclusão de uma avaliação
func TestDeleteRating(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e fazer login
	user := createTestUser(t, "delete_rating@test.com", "password123", "Delete Rating User")
	token := loginTestUser(t, router, "delete_rating@test.com", "password123")

	// Criar uma receita
	recipe := createTestRecipe(t, user.ID)

	// Criar avaliação
	ratingData := map[string]interface{}{
		"score":   3,
		"comment": "Receita ok",
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Deletar avaliação
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/recipes/%d/ratings/me", recipe.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Tentar obter avaliação deletada (deve retornar 404)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings/me", recipe.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestListRatings testa a listagem de avaliações com paginação
func TestListRatings(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário para a receita
	recipeOwner := createTestUser(t, "recipe_owner@test.com", "password123", "Recipe Owner")
	recipe := createTestRecipe(t, recipeOwner.ID)

	// Criar múltiplos usuários e avaliações
	for i := 1; i <= 5; i++ {
		user := createTestUser(t, fmt.Sprintf("rater%d@test.com", i), "password123", fmt.Sprintf("Rater %d", i))
		token := loginTestUser(t, router, fmt.Sprintf("rater%d@test.com", i), "password123")

		ratingData := map[string]interface{}{
			"score":   i, // Score de 1 a 5
			"comment": fmt.Sprintf("Comentário do usuário %d", i),
		}
		body, _ := json.Marshal(ratingData)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	// Listar avaliações (sem autenticação - endpoint público)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	assert.Equal(t, 5, len(data))

	// Testar paginação
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings?page=1&limit=2", recipe.ID), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data = response["data"].([]interface{})
	assert.Equal(t, 2, len(data))
	assert.Equal(t, float64(5), response["total"].(float64))
}

// TestRatingStats testa as estatísticas de avaliação
func TestRatingStats(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário para a receita
	recipeOwner := createTestUser(t, "stats_owner@test.com", "password123", "Stats Owner")
	recipe := createTestRecipe(t, recipeOwner.ID)

	// Criar avaliações com distribuição conhecida
	scores := []int{5, 5, 5, 4, 4, 3, 2, 1}
	for i, score := range scores {
		user := createTestUser(t, fmt.Sprintf("stats_user%d@test.com", i), "password123", fmt.Sprintf("Stats User %d", i))
		token := loginTestUser(t, router, fmt.Sprintf("stats_user%d@test.com", i), "password123")

		ratingData := map[string]interface{}{
			"score":   score,
			"comment": fmt.Sprintf("Comentário %d", i),
		}
		body, _ := json.Marshal(ratingData)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	// Obter estatísticas
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings/stats", recipe.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(8), response["total_ratings"].(float64))
	
	// Média esperada: (5+5+5+4+4+3+2+1) / 8 = 29/8 = 3.625
	averageRating := response["average_rating"].(float64)
	assert.InDelta(t, 3.625, averageRating, 0.01)

	distribution := response["distribution"].(map[string]interface{})
	assert.Equal(t, float64(1), distribution["1"].(float64))
	assert.Equal(t, float64(1), distribution["2"].(float64))
	assert.Equal(t, float64(1), distribution["3"].(float64))
	assert.Equal(t, float64(2), distribution["4"].(float64))
	assert.Equal(t, float64(3), distribution["5"].(float64))
}

// TestRecipeWithRatings testa se as receitas incluem informações de rating
func TestRecipeWithRatings(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e receita
	user := createTestUser(t, "recipe_rating@test.com", "password123", "Recipe Rating User")
	recipe := createTestRecipe(t, user.ID)

	// Criar algumas avaliações
	scores := []int{5, 4, 5}
	for i, score := range scores {
		rater := createTestUser(t, fmt.Sprintf("rater_recipe%d@test.com", i), "password123", fmt.Sprintf("Rater %d", i))
		token := loginTestUser(t, router, fmt.Sprintf("rater_recipe%d@test.com", i), "password123")

		ratingData := map[string]interface{}{
			"score": score,
		}
		body, _ := json.Marshal(ratingData)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	// Obter receita e verificar ratings
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d", recipe.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.Recipe
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Média esperada: (5+4+5) / 3 = 14/3 = 4.666...
	assert.InDelta(t, 4.666, response.AverageRating, 0.01)
	assert.Equal(t, int64(3), response.RatingCount)
}

// TestAdminDeleteRating testa a moderação de avaliações por admin
func TestAdminDeleteRating(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar admin
	admin := createTestUser(t, "admin@test.com", "password123", "Admin User")
	database.DB.Model(&admin).Update("role", "admin")
	adminToken := loginTestUser(t, router, "admin@test.com", "password123")

	// Criar usuário regular e avaliação
	user := createTestUser(t, "regular@test.com", "password123", "Regular User")
	userToken := loginTestUser(t, router, "regular@test.com", "password123")
	recipe := createTestRecipe(t, user.ID)

	ratingData := map[string]interface{}{
		"score":   3,
		"comment": "Avaliação inadequada",
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var rating models.Rating
	err := json.Unmarshal(rec.Body.Bytes(), &rating)
	require.NoError(t, err)

	// Admin deleta a avaliação
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/ratings/%d", rating.ID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verificar que foi deletada
	var count int64
	database.DB.Model(&models.Rating{}).Where("id = ?", rating.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestRatingsSorting testa ordenação de avaliações
func TestRatingsSorting(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário e receita
	recipeOwner := createTestUser(t, "sort_owner@test.com", "password123", "Sort Owner")
	recipe := createTestRecipe(t, recipeOwner.ID)

	// Criar avaliações com scores diferentes
	scores := []int{1, 5, 3}
	for i, score := range scores {
		user := createTestUser(t, fmt.Sprintf("sort_user%d@test.com", i), "password123", fmt.Sprintf("Sort User %d", i))
		token := loginTestUser(t, router, fmt.Sprintf("sort_user%d@test.com", i), "password123")

		ratingData := map[string]interface{}{
			"score": score,
		}
		body, _ := json.Marshal(ratingData)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipe.ID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	// Testar ordenação por highest
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings?sort=highest", recipe.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	firstRating := data[0].(map[string]interface{})
	assert.Equal(t, float64(5), firstRating["score"].(float64))

	// Testar ordenação por lowest
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/recipes/%d/ratings?sort=lowest", recipe.ID), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data = response["data"].([]interface{})
	firstRating = data[0].(map[string]interface{})
	assert.Equal(t, float64(1), firstRating["score"].(float64))
}

// TestRecipeListSortByRating testa ordenação de receitas por rating
func TestRecipeListSortByRating(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Criar usuário
	user := createTestUser(t, "list_sort@test.com", "password123", "List Sort User")

	// Criar 3 receitas
	recipe1 := createTestRecipe(t, user.ID)
	recipe2 := createTestRecipe(t, user.ID)
	recipe3 := createTestRecipe(t, user.ID)

	// Avaliar com ratings diferentes
	// Recipe1: 5 estrelas
	// Recipe2: 3 estrelas
	// Recipe3: 4 estrelas

	// Recipe 1
	rater1 := createTestUser(t, "rater1@test.com", "password123", "Rater 1")
	token1 := loginTestUser(t, router, "rater1@test.com", "password123")
	createRating(t, router, token1, recipe1.ID, 5)

	// Recipe 2
	rater2 := createTestUser(t, "rater2@test.com", "password123", "Rater 2")
	token2 := loginTestUser(t, router, "rater2@test.com", "password123")
	createRating(t, router, token2, recipe2.ID, 3)

	// Recipe 3
	rater3 := createTestUser(t, "rater3@test.com", "password123", "Rater 3")
	token3 := loginTestUser(t, router, "rater3@test.com", "password123")
	createRating(t, router, token3, recipe3.ID, 4)

	// Listar receitas ordenadas por rating
	req := httptest.NewRequest(http.MethodGet, "/recipes?sort_by=rating", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})
	
	// Primeira receita deve ser a de rating 5
	firstRecipe := data[0].(map[string]interface{})
	assert.Equal(t, float64(recipe1.ID), firstRecipe["id"].(float64))
	assert.InDelta(t, 5.0, firstRecipe["average_rating"].(float64), 0.01)
}

// Helper function to create a rating
func createRating(t *testing.T, router http.Handler, token string, recipeID uint, score int) {
	ratingData := map[string]interface{}{
		"score": score,
	}
	body, _ := json.Marshal(ratingData)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/recipes/%d/ratings", recipeID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

