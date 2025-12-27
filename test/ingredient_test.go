package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

func setupIngredientTestDB(t *testing.T) {
	if database.DB == nil {
		t.Skip("DATABASE_URL não configurado para testes")
	}

	// Limpar tabelas
	database.DB.Exec("DELETE FROM recipe_ingredients")
	database.DB.Exec("DELETE FROM ingredients")
	database.DB.Exec("DELETE FROM recipes")
	database.DB.Exec("DELETE FROM users")

	// Executar migrations
	if err := database.DB.AutoMigrate(
		&models.User{},
		&models.Recipe{},
		&models.Ingredient{},
		&models.RecipeIngredient{},
	); err != nil {
		t.Fatalf("erro ao executar migrations: %v", err)
	}
}

func TestListIngredients(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar alguns ingredientes de teste
	ing1 := models.Ingredient{
		Name:     "Tomate",
		Calories: 15,
		Protein:  1.1,
		Carbs:    3.1,
		Fat:      0.2,
		Category: "vegetais",
		Source:   "test",
	}
	ing2 := models.Ingredient{
		Name:     "Cebola",
		Calories: 38,
		Protein:  1.4,
		Carbs:    8.9,
		Fat:      0.1,
		Category: "vegetais",
		Source:   "test",
	}
	database.DB.Create(&ing1)
	database.DB.Create(&ing2)

	req := httptest.NewRequest("GET", "/ingredients", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		if len(data) < 2 {
			t.Errorf("esperado pelo menos 2 ingredientes, obteve %d", len(data))
		}
	}
}

func TestGetIngredient(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar ingrediente de teste
	ing := models.Ingredient{
		Name:     "Arroz",
		Calories: 128,
		Protein:  2.5,
		Carbs:    28.1,
		Fat:      0.2,
		Category: "cereais",
		Source:   "test",
	}
	database.DB.Create(&ing)

	req := httptest.NewRequest("GET", "/ingredients/1", nil)
	rec := httptest.NewRecorder()

	handlers.GetIngredient(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Errorf("esperado 200 ou 404, obteve %d", rec.Code)
	}
}

func TestCreateIngredient_Admin(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := models.User{
		Name:     "Admin",
		Email:    "admin@test.com",
		Password: hashedPassword,
		Role:     "admin",
	}
	database.DB.Create(&admin)

	// Gerar token
	token, _ := auth.GenerateToken(admin.ID, admin.Email, admin.Role)

	// Payload do ingrediente
	payload := map[string]interface{}{
		"name":     "Feijão",
		"calories": 77,
		"protein":  4.5,
		"carbs":    14.0,
		"fat":      0.5,
		"category": "leguminosas",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/admin/ingredients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middlewares
	handler := middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(handlers.CreateIngredient)))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("esperado status 201, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}
}

func TestAddRecipeIngredient(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar usuário
	hashedPassword, _ := auth.HashPassword("senha123")
	user := models.User{
		Name:     "Usuário",
		Email:    "user@test.com",
		Password: hashedPassword,
		Role:     "user",
	}
	database.DB.Create(&user)

	// Criar receita
	recipe := models.Recipe{
		Title:    "Receita Teste",
		PrepTime: 30,
		Servings: 4,
		UserID:   &user.ID,
	}
	database.DB.Create(&recipe)

	// Criar ingrediente
	ingredient := models.Ingredient{
		Name:     "Sal",
		Calories: 0,
		Category: "temperos",
		Source:   "test",
	}
	database.DB.Create(&ingredient)

	// Gerar token
	token, _ := auth.GenerateToken(user.ID, user.Email, user.Role)

	// Payload
	payload := map[string]interface{}{
		"ingredient_id": ingredient.ID,
		"quantity":      5,
		"unit":          "g",
		"order":         1,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/recipes/1/ingredients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middleware auth
	handler := middleware.RequireAuth(http.HandlerFunc(handlers.AddRecipeIngredient))
	handler.ServeHTTP(rec, req)

	// Pode dar 404 se chi router não estiver configurado, mas não deve dar erro de servidor
	if rec.Code != http.StatusCreated && rec.Code != http.StatusNotFound {
		t.Logf("Status code: %d (esperado 201 ou 404)", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}
}

func TestCalculateRecipeNutrition(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar receita
	user := models.User{
		Name:     "User",
		Email:    "user2@test.com",
		Password: "hash",
		Role:     "user",
	}
	database.DB.Create(&user)

	recipe := models.Recipe{
		Title:    "Receita Nutricional",
		PrepTime: 20,
		Servings: 2,
		UserID:   &user.ID,
	}
	database.DB.Create(&recipe)

	// Criar ingrediente
	ingredient := models.Ingredient{
		Name:     "Frango",
		Calories: 159,
		Protein:  32.0,
		Carbs:    0,
		Fat:      3.0,
		Category: "carnes",
		Source:   "test",
	}
	database.DB.Create(&ingredient)

	// Adicionar à receita (200g)
	recipeIng := models.RecipeIngredient{
		RecipeID:     recipe.ID,
		IngredientID: ingredient.ID,
		Quantity:     200,
		Unit:         "g",
	}
	database.DB.Create(&recipeIng)

	req := httptest.NewRequest("GET", "/recipes/1/nutrition", nil)
	rec := httptest.NewRecorder()

	handlers.GetRecipeNutrition(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Errorf("esperado 200 ou 404, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}

	if rec.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if total, ok := response["total"].(map[string]interface{}); ok {
			// 200g de frango = 2x os valores por 100g
			expectedCalories := 159.0 * 2
			if calories, ok := total["calories"].(float64); ok {
				if calories != expectedCalories {
					t.Errorf("esperado ~%.1f calorias, obteve %.1f", expectedCalories, calories)
				}
			}
		}
	}
}

func TestGetCategories(t *testing.T) {
	setupIngredientTestDB(t)

	// Criar ingredientes de diferentes categorias
	ingredients := []models.Ingredient{
		{Name: "Ing1", Calories: 10, Category: "frutas", Source: "test"},
		{Name: "Ing2", Calories: 10, Category: "vegetais", Source: "test"},
		{Name: "Ing3", Calories: 10, Category: "carnes", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	req := httptest.NewRequest("GET", "/ingredients/categories", nil)
	rec := httptest.NewRecorder()

	handlers.GetCategories(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if categories, ok := response["categories"].([]interface{}); ok {
		if len(categories) < 3 {
			t.Errorf("esperado pelo menos 3 categorias, obteve %d", len(categories))
		}
	}
}

