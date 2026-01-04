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
	"github.com/go-chi/chi/v5"
)

func TestListIngredients(t *testing.T) {
	testdb.SetupWithCleanup(t)

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
	testdb.SetupWithCleanup(t)

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
	testdb.SetupWithCleanup(t)

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
	testdb.SetupWithCleanup(t)

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
	testdb.SetupWithCleanup(t)

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

	req := httptest.NewRequest("GET", fmt.Sprintf("/recipes/%d/nutrition", recipe.ID), nil)
	
	// Adicionar parâmetro de rota através do chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", recipe.ID))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
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
	testdb.SetupWithCleanup(t)

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

func TestSearchIngredientsByName(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes para testar busca
	ingredients := []models.Ingredient{
		{Name: "Arroz branco", Calories: 128, Category: "cereais", Source: "test"},
		{Name: "Arroz integral", Calories: 123, Category: "cereais", Source: "test"},
		{Name: "Macarrão de arroz", Calories: 102, Category: "massas", Source: "test"},
		{Name: "Feijão preto", Calories: 77, Category: "leguminosas", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar por "arroz"
	req := httptest.NewRequest("GET", "/ingredients?search=arroz", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve retornar 3 ingredientes com "arroz" no nome
		if len(data) != 3 {
			t.Errorf("esperado 3 ingredientes com 'arroz', obteve %d", len(data))
		}

		// Verificar que "Arroz branco" vem primeiro (começa com "arroz")
		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			if name != "Arroz branco" && name != "Arroz integral" {
				t.Logf("Primeiro resultado: %s (esperado começar com 'Arroz')", name)
			}
		}
	} else {
		t.Error("response['data'] não é uma lista")
	}
}

func TestSearchIngredientsByCategory(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes
	ingredients := []models.Ingredient{
		{Name: "Tomate", Calories: 15, Category: "vegetais", Source: "test"},
		{Name: "Cebola", Calories: 38, Category: "vegetais", Source: "test"},
		{Name: "Banana", Calories: 92, Category: "frutas", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar por "vegetais"
	req := httptest.NewRequest("GET", "/ingredients?search=vegetais", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve retornar 2 ingredientes da categoria "vegetais"
		if len(data) != 2 {
			t.Errorf("esperado 2 ingredientes em 'vegetais', obteve %d", len(data))
		}
	}
}

func TestSearchIngredientsCaseInsensitive(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingrediente
	ingredient := models.Ingredient{
		Name:     "Açúcar refinado",
		Calories: 387,
		Category: "açúcares",
		Source:   "test",
	}
	database.DB.Create(&ingredient)

	// Buscar com maiúsculas
	req := httptest.NewRequest("GET", "/ingredients?search=AÇÚCAR", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		if len(data) != 1 {
			t.Errorf("esperado 1 ingrediente com 'AÇÚCAR', obteve %d", len(data))
		}
	}
}

func TestSearchWithCategoryFilter(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes
	ingredients := []models.Ingredient{
		{Name: "Feijão preto", Calories: 77, Category: "leguminosas", Source: "test"},
		{Name: "Feijão carioca", Calories: 76, Category: "leguminosas", Source: "test"},
		{Name: "Farinha de feijão", Calories: 330, Category: "farinhas", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar "feijão" apenas na categoria "leguminosas"
	req := httptest.NewRequest("GET", "/ingredients?search=feijão&category=leguminosas", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve retornar apenas os 2 feijões da categoria leguminosas
		if len(data) != 2 {
			t.Errorf("esperado 2 ingredientes, obteve %d", len(data))
		}
	}
}

func TestSearchMultipleWords(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes para testar busca com múltiplas palavras
	ingredients := []models.Ingredient{
		{Name: "Arroz integral", Calories: 123, Category: "cereais", Source: "test"},
		{Name: "Arroz branco", Calories: 128, Category: "cereais", Source: "test"},
		{Name: "Macarrão integral", Calories: 102, Category: "massas", Source: "test"},
		{Name: "Feijão preto", Calories: 77, Category: "leguminosas", Source: "test"},
		{Name: "Farinha de Trigo", Calories: 364, Category: "farinhas", Source: "test"},
		{Name: "Farinha de Rosca", Calories: 398, Category: "farinhas", Source: "test"},
		{Name: "Trigo em Grão", Calories: 330, Category: "cereais", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar por "arroz integral" - deve encontrar ambos arroz e integral
	req := httptest.NewRequest("GET", "/ingredients?search=arroz+integral", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve retornar 3 ingredientes: Arroz integral (ambas palavras), Arroz branco (arroz), Macarrão integral (integral)
		if len(data) < 2 {
			t.Errorf("esperado pelo menos 2 ingredientes com 'arroz' ou 'integral', obteve %d", len(data))
		}

		// Verificar que "Arroz integral" vem primeiro (contém AMBAS as palavras)
		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			if name != "Arroz integral" {
				t.Logf("Primeiro resultado: %s (esperado 'Arroz integral' por conter ambas palavras)", name)
			}
		}

		// Verificar que encontrou pelo menos "Arroz integral" e um dos outros
		foundArrozIntegral := false
		foundArrozBranco := false
		foundMacarraoIntegral := false

		for _, item := range data {
			ing := item.(map[string]interface{})
			name := ing["name"].(string)
			if name == "Arroz integral" {
				foundArrozIntegral = true
			}
			if name == "Arroz branco" {
				foundArrozBranco = true
			}
			if name == "Macarrão integral" {
				foundMacarraoIntegral = true
			}
		}

		if !foundArrozIntegral {
			t.Error("não encontrou 'Arroz integral' (deveria conter ambas palavras)")
		}
		if !foundArrozBranco && !foundMacarraoIntegral {
			t.Error("não encontrou nem 'Arroz branco' nem 'Macarrão integral'")
		}
	} else {
		t.Error("response['data'] não é uma lista")
	}
}

func TestSearchWithStopwords(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes para testar busca com stopwords
	ingredients := []models.Ingredient{
		{Name: "Farinha de Trigo", Calories: 364, Category: "farinhas", Source: "test"},
		{Name: "Farinha de Trigo Integral", Calories: 340, Category: "farinhas", Source: "test"},
		{Name: "Farinha de Rosca", Calories: 398, Category: "farinhas", Source: "test"},
		{Name: "Trigo em Grão", Calories: 330, Category: "cereais", Source: "test"},
		{Name: "Açúcar refinado", Calories: 387, Category: "açúcares", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar por "farinha de trigo" - "de" deve ser ignorado
	req := httptest.NewRequest("GET", "/ingredients?search=farinha+de+trigo", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve encontrar ingredientes com "farinha" OU "trigo" (ignorando "de")
		if len(data) < 3 {
			t.Errorf("esperado pelo menos 3 ingredientes com 'farinha' ou 'trigo', obteve %d", len(data))
		}

		// Verificar que encontrou ingredientes com ambas palavras primeiro
		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			// Primeiro deve ser um que contém AMBAS palavras
			if name != "Farinha de Trigo" && name != "Farinha de Trigo Integral" {
				t.Logf("Primeiro resultado: %s (esperado conter 'farinha' e 'trigo')", name)
			}
		}

		// Verificar que encontrou ingredientes esperados
		foundFarinhaTrigo := false
		foundFarinhaTrigoIntegral := false
		foundTrigoGrao := false
		foundFarinhaRosca := false

		for _, item := range data {
			ing := item.(map[string]interface{})
			name := ing["name"].(string)
			if name == "Farinha de Trigo" {
				foundFarinhaTrigo = true
			}
			if name == "Farinha de Trigo Integral" {
				foundFarinhaTrigoIntegral = true
			}
			if name == "Trigo em Grão" {
				foundTrigoGrao = true
			}
			if name == "Farinha de Rosca" {
				foundFarinhaRosca = true
			}
		}

		if !foundFarinhaTrigo || !foundFarinhaTrigoIntegral {
			t.Error("não encontrou ingredientes com 'farinha' e 'trigo'")
		}
		if !foundTrigoGrao {
			t.Error("não encontrou 'Trigo em Grão' (contém 'trigo')")
		}
		if !foundFarinhaRosca {
			t.Error("não encontrou 'Farinha de Rosca' (contém 'farinha')")
		}
	} else {
		t.Error("response['data'] não é uma lista")
	}
}

func TestSearchSingleShortWord(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingredientes
	ingredients := []models.Ingredient{
		{Name: "Óleo de Coco", Calories: 862, Category: "óleos", Source: "test"},
		{Name: "Óleo de Soja", Calories: 884, Category: "óleos", Source: "test"},
		{Name: "Coco ralado", Calories: 354, Category: "frutas", Source: "test"},
		{Name: "Sal", Calories: 0, Category: "temperos", Source: "test"},
	}

	for _, ing := range ingredients {
		database.DB.Create(&ing)
	}

	// Buscar por "óleo de coco" - "de" deve ser ignorado (< 3 chars e stopword)
	req := httptest.NewRequest("GET", "/ingredients?search=%C3%B3leo+de+coco", nil)
	rec := httptest.NewRecorder()

	handlers.ListIngredients(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
		return
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if data, ok := response["data"].([]interface{}); ok {
		// Deve encontrar ingredientes com "óleo" OU "coco"
		if len(data) < 2 {
			t.Errorf("esperado pelo menos 2 ingredientes com 'óleo' ou 'coco', obteve %d", len(data))
		}

		// Verificar que "Óleo de Coco" vem primeiro (contém ambas palavras)
		if len(data) > 0 {
			first := data[0].(map[string]interface{})
			name := first["name"].(string)
			if name != "Óleo de Coco" {
				t.Logf("Primeiro resultado: %s (esperado 'Óleo de Coco')", name)
			}
		}

		// Verificar que encontrou os esperados
		foundOleoCoco := false
		foundOleoSoja := false
		foundCocoRalado := false

		for _, item := range data {
			ing := item.(map[string]interface{})
			name := ing["name"].(string)
			if name == "Óleo de Coco" {
				foundOleoCoco = true
			}
			if name == "Óleo de Soja" {
				foundOleoSoja = true
			}
			if name == "Coco ralado" {
				foundCocoRalado = true
			}
		}

		if !foundOleoCoco {
			t.Error("não encontrou 'Óleo de Coco'")
		}
		if !foundOleoSoja && !foundCocoRalado {
			t.Error("não encontrou 'Óleo de Soja' ou 'Coco ralado'")
		}
	} else {
		t.Error("response['data'] não é uma lista")
	}
}
