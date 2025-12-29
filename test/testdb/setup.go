package testdb

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

// Setup inicializa um banco de dados SQLite in-memory para testes
// Retorna uma função de cleanup que deve ser chamada com defer
func Setup(t *testing.T) func() {
	t.Helper()

	// Criar banco SQLite in-memory
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("falha ao criar banco de testes: %v", err)
	}

	// Atribuir ao database global
	database.DB = db

	// Executar migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Recipe{},
		&models.Ingredient{},
		&models.RecipeIngredient{},
	); err != nil {
		t.Fatalf("falha ao executar migrations: %v", err)
	}

	// Retornar função de cleanup
	return func() {
		// Limpar todas as tabelas
		db.Exec("DELETE FROM recipe_ingredients")
		db.Exec("DELETE FROM recipes")
		db.Exec("DELETE FROM ingredients")
		db.Exec("DELETE FROM users")

		// Fechar conexão
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}

		// Resetar database.DB
		database.DB = nil
	}
}

// SetupWithCleanup é similar ao Setup mas faz cleanup automático entre testes
func SetupWithCleanup(t *testing.T) {
	cleanup := Setup(t)
	t.Cleanup(cleanup)
}

// SeedUser cria um usuário de teste e retorna o modelo
func SeedUser(t *testing.T, name, email, password, role string) *models.User {
	t.Helper()

	if database.DB == nil {
		t.Fatal("database não inicializado - execute Setup() primeiro")
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: password, // Assumindo que já vem hasheado
		Role:     role,
	}

	if err := database.DB.Create(user).Error; err != nil {
		t.Fatalf("falha ao criar usuário de teste: %v", err)
	}

	return user
}

// SeedRecipe cria uma receita de teste e retorna o modelo
func SeedRecipe(t *testing.T, title, description string, userID uint, isGeneral bool) *models.Recipe {
	t.Helper()

	if database.DB == nil {
		t.Fatal("database não inicializado - execute Setup() primeiro")
	}

	var recipeUserID *uint
	if !isGeneral {
		recipeUserID = &userID
	}

	recipe := &models.Recipe{
		Title:       title,
		Description: description,
		PrepTime:    30,
		Servings:    4,
		Difficulty:  "média",
		UserID:      recipeUserID,
	}

	if err := database.DB.Create(recipe).Error; err != nil {
		t.Fatalf("falha ao criar receita de teste: %v", err)
	}

	return recipe
}

// SeedIngredient cria um ingrediente de teste e retorna o modelo
func SeedIngredient(t *testing.T, name, category string, calories float64) *models.Ingredient {
	t.Helper()

	if database.DB == nil {
		t.Fatal("database não inicializado - execute Setup() primeiro")
	}

	ingredient := &models.Ingredient{
		Name:     name,
		Category: category,
		Calories: calories,
		Protein:  1.0,
		Carbs:    2.0,
		Fat:      0.5,
		Source:   "test",
	}

	if err := database.DB.Create(ingredient).Error; err != nil {
		t.Fatalf("falha ao criar ingrediente de teste: %v", err)
	}

	return ingredient
}

// CleanTable limpa uma tabela específica
func CleanTable(t *testing.T, tableName string) {
	t.Helper()

	if database.DB == nil {
		t.Fatal("database não inicializado - execute Setup() primeiro")
	}

	if err := database.DB.Exec("DELETE FROM " + tableName).Error; err != nil {
		t.Fatalf("falha ao limpar tabela %s: %v", tableName, err)
	}
}

// AddChiURLParam adiciona um parâmetro de URL do chi ao contexto
func AddChiURLParam(req *http.Request, key, value string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
}

