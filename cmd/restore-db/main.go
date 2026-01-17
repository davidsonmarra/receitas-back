package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run cmd/restore-db/main.go <diretorio_backup>")
		os.Exit(1)
	}

	backupDir := os.Args[1]

	// Conectar ao banco
	fmt.Println("📡 Conectando ao banco de dados...")
	if err := database.Connect(); err != nil {
		fmt.Printf("❌ Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("✅ Conectado com sucesso!")

	// Verificar se o diretório existe
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		fmt.Printf("❌ Diretório não encontrado: %s\n", backupDir)
		os.Exit(1)
	}

	fmt.Println("\n⚠️  AVISO: Esta operação irá limpar e recriar todas as tabelas!")
	fmt.Println("Pressione Ctrl+C para cancelar ou Enter para continuar...")
	fmt.Scanln()

	// Recriar schema (drop e create tables)
	fmt.Println("\n🔄 Recriando schema do banco...")
	if err := recreateSchema(); err != nil {
		fmt.Printf("❌ Erro ao recriar schema: %v\n", err)
		os.Exit(1)
	}

	// Importar cada tabela na ordem correta (respeitando foreign keys)
	fmt.Println("\n📦 Importando dados...")

	if err := importUsers(backupDir); err != nil {
		fmt.Printf("❌ Erro ao importar users: %v\n", err)
		os.Exit(1)
	}

	if err := importIngredients(backupDir); err != nil {
		fmt.Printf("❌ Erro ao importar ingredients: %v\n", err)
		os.Exit(1)
	}

	if err := importRecipes(backupDir); err != nil {
		fmt.Printf("❌ Erro ao importar recipes: %v\n", err)
		os.Exit(1)
	}

	if err := importRecipeIngredients(backupDir); err != nil {
		fmt.Printf("❌ Erro ao importar recipe_ingredients: %v\n", err)
		os.Exit(1)
	}

	if err := importRatings(backupDir); err != nil {
		fmt.Printf("❌ Erro ao importar ratings: %v\n", err)
		os.Exit(1)
	}

	// Atualizar sequences (auto-increment)
	fmt.Println("\n🔢 Atualizando sequences...")
	if err := updateSequences(); err != nil {
		fmt.Printf("❌ Erro ao atualizar sequences: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Restauração concluída com sucesso!")
}

func recreateSchema() error {
	// Drop tables na ordem inversa (respeitando foreign keys)
	tables := []interface{}{
		&models.Rating{},
		&models.RecipeIngredient{},
		&models.Recipe{},
		&models.Ingredient{},
		&models.User{},
	}

	for _, table := range tables {
		if err := database.DB.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("erro ao dropar tabela: %w", err)
		}
	}

	// Criar tables
	for i := len(tables) - 1; i >= 0; i-- {
		if err := database.DB.AutoMigrate(tables[i]); err != nil {
			return fmt.Errorf("erro ao criar tabela: %w", err)
		}
	}

	return nil
}

func importUsers(backupDir string) error {
	var users []models.User
	if err := loadFromJSON(filepath.Join(backupDir, "users.json"), &users); err != nil {
		return err
	}

	for _, user := range users {
		// Usar Create com todos os campos, incluindo ID e timestamps
		if err := database.DB.Session(&gorm.Session{CreateBatchSize: 100}).Create(&user).Error; err != nil {
			return fmt.Errorf("erro ao inserir user ID %d: %w", user.ID, err)
		}
	}

	fmt.Printf("  👤 Users: %d registros importados\n", len(users))
	return nil
}

func importIngredients(backupDir string) error {
	var ingredients []models.Ingredient
	if err := loadFromJSON(filepath.Join(backupDir, "ingredients.json"), &ingredients); err != nil {
		return err
	}

	for _, ingredient := range ingredients {
		if err := database.DB.Session(&gorm.Session{CreateBatchSize: 100}).Create(&ingredient).Error; err != nil {
			return fmt.Errorf("erro ao inserir ingredient ID %d: %w", ingredient.ID, err)
		}
	}

	fmt.Printf("  🥕 Ingredients: %d registros importados\n", len(ingredients))
	return nil
}

func importRecipes(backupDir string) error {
	var recipes []models.Recipe
	if err := loadFromJSON(filepath.Join(backupDir, "recipes.json"), &recipes); err != nil {
		return err
	}

	for _, recipe := range recipes {
		if err := database.DB.Session(&gorm.Session{CreateBatchSize: 100}).Create(&recipe).Error; err != nil {
			return fmt.Errorf("erro ao inserir recipe ID %d: %w", recipe.ID, err)
		}
	}

	fmt.Printf("  📖 Recipes: %d registros importados\n", len(recipes))
	return nil
}

func importRecipeIngredients(backupDir string) error {
	var recipeIngredients []models.RecipeIngredient
	if err := loadFromJSON(filepath.Join(backupDir, "recipe_ingredients.json"), &recipeIngredients); err != nil {
		return err
	}

	for _, ri := range recipeIngredients {
		if err := database.DB.Session(&gorm.Session{CreateBatchSize: 100}).Create(&ri).Error; err != nil {
			return fmt.Errorf("erro ao inserir recipe_ingredient ID %d: %w", ri.ID, err)
		}
	}

	fmt.Printf("  🔗 Recipe Ingredients: %d registros importados\n", len(recipeIngredients))
	return nil
}

func importRatings(backupDir string) error {
	var ratings []models.Rating
	if err := loadFromJSON(filepath.Join(backupDir, "ratings.json"), &ratings); err != nil {
		return err
	}

	for _, rating := range ratings {
		if err := database.DB.Session(&gorm.Session{CreateBatchSize: 100}).Create(&rating).Error; err != nil {
			return fmt.Errorf("erro ao inserir rating ID %d: %w", rating.ID, err)
		}
	}

	fmt.Printf("  ⭐ Ratings: %d registros importados\n", len(ratings))
	return nil
}

func updateSequences() error {
	// Atualizar sequence de cada tabela para o próximo ID disponível
	sequences := map[string]string{
		"users":               "users_id_seq",
		"ingredients":         "ingredients_id_seq",
		"recipes":             "recipes_id_seq",
		"recipe_ingredients":  "recipe_ingredients_id_seq",
		"ratings":             "ratings_id_seq",
	}

	for table, sequence := range sequences {
		query := fmt.Sprintf("SELECT setval('%s', (SELECT COALESCE(MAX(id), 1) FROM %s), true)", sequence, table)
		if err := database.DB.Exec(query).Error; err != nil {
			return fmt.Errorf("erro ao atualizar sequence %s: %w", sequence, err)
		}
		fmt.Printf("  ✓ Sequence %s atualizada\n", sequence)
	}

	return nil
}

func loadFromJSON(filename string, target interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", filename, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(target)
}

