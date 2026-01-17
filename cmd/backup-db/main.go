package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run cmd/backup-db/main.go <diretorio_destino>")
		os.Exit(1)
	}

	outputDir := os.Args[1]

	// Conectar ao banco
	fmt.Println("📡 Conectando ao banco de dados...")
	if err := database.Connect(); err != nil {
		fmt.Printf("❌ Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("✅ Conectado com sucesso!")

	// Criar diretório se não existir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("❌ Erro ao criar diretório: %v\n", err)
		os.Exit(1)
	}

	// Exportar cada tabela
	fmt.Println("\n📦 Exportando tabelas...")

	if err := exportUsers(outputDir); err != nil {
		fmt.Printf("❌ Erro ao exportar users: %v\n", err)
		os.Exit(1)
	}

	if err := exportIngredients(outputDir); err != nil {
		fmt.Printf("❌ Erro ao exportar ingredients: %v\n", err)
		os.Exit(1)
	}

	if err := exportRecipes(outputDir); err != nil {
		fmt.Printf("❌ Erro ao exportar recipes: %v\n", err)
		os.Exit(1)
	}

	if err := exportRecipeIngredients(outputDir); err != nil {
		fmt.Printf("❌ Erro ao exportar recipe_ingredients: %v\n", err)
		os.Exit(1)
	}

	if err := exportRatings(outputDir); err != nil {
		fmt.Printf("❌ Erro ao exportar ratings: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ Backup concluído com sucesso!")
	fmt.Printf("📁 Arquivos salvos em: %s\n", outputDir)
}

func exportUsers(outputDir string) error {
	var users []models.User
	if err := database.DB.Unscoped().Find(&users).Error; err != nil {
		return err
	}

	count := len(users)
	fmt.Printf("  👤 Users: %d registros\n", count)

	return saveToJSON(filepath.Join(outputDir, "users.json"), users)
}

func exportIngredients(outputDir string) error {
	var ingredients []models.Ingredient
	if err := database.DB.Find(&ingredients).Error; err != nil {
		return err
	}

	count := len(ingredients)
	fmt.Printf("  🥕 Ingredients: %d registros\n", count)

	return saveToJSON(filepath.Join(outputDir, "ingredients.json"), ingredients)
}

func exportRecipes(outputDir string) error {
	var recipes []models.Recipe
	if err := database.DB.Unscoped().Find(&recipes).Error; err != nil {
		return err
	}

	count := len(recipes)
	fmt.Printf("  📖 Recipes: %d registros\n", count)

	return saveToJSON(filepath.Join(outputDir, "recipes.json"), recipes)
}

func exportRecipeIngredients(outputDir string) error {
	var recipeIngredients []models.RecipeIngredient
	if err := database.DB.Find(&recipeIngredients).Error; err != nil {
		return err
	}

	count := len(recipeIngredients)
	fmt.Printf("  🔗 Recipe Ingredients: %d registros\n", count)

	return saveToJSON(filepath.Join(outputDir, "recipe_ingredients.json"), recipeIngredients)
}

func exportRatings(outputDir string) error {
	var ratings []models.Rating
	if err := database.DB.Unscoped().Find(&ratings).Error; err != nil {
		return err
	}

	count := len(ratings)
	fmt.Printf("  ⭐ Ratings: %d registros\n", count)

	return saveToJSON(filepath.Join(outputDir, "ratings.json"), ratings)
}

func saveToJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

