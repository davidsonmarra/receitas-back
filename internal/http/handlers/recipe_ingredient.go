package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// AddRecipeIngredient adiciona ingrediente à receita
func AddRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	// Verificar ownership da receita
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	var recipeIng models.RecipeIngredient
	if err := json.NewDecoder(r.Body).Decode(&recipeIng); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	recipeIng.RecipeID = recipe.ID

	if errs := validation.ValidateStruct(recipeIng); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Verificar se ingrediente existe
	var ingredient models.Ingredient
	if err := database.DB.First(&ingredient, recipeIng.IngredientID).Error; err != nil {
		response.ValidationError(w, "Ingrediente não encontrado.")
		return
	}

	if err := database.DB.Create(&recipeIng).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to add ingredient to recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to add ingredient")
		return
	}

	// Recarregar com dados do ingrediente
	database.DB.Preload("Ingredient").First(&recipeIng, recipeIng.ID)

	log.InfoCtx(r.Context(), "ingredient added to recipe", "recipe_id", recipeID, "ingredient_id", recipeIng.IngredientID)
	response.JSON(w, http.StatusCreated, recipeIng)
}

// ListRecipeIngredients lista ingredientes de uma receita
func ListRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Verificar se receita existe
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	var recipeIngredients []models.RecipeIngredient
	if err := database.DB.
		Preload("Ingredient").
		Where("recipe_id = ?", recipeID).
		Order("\"order\" ASC, id ASC").
		Find(&recipeIngredients).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list recipe ingredients", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list ingredients")
		return
	}

	response.JSON(w, http.StatusOK, recipeIngredients)
}

// UpdateRecipeIngredient atualiza quantidade/unidade de ingrediente
func UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	ingredientID := chi.URLParam(r, "ingredient_id")
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	// Verificar ownership
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	var recipeIng models.RecipeIngredient
	if err := database.DB.Where("recipe_id = ? AND id = ?", recipeID, ingredientID).
		First(&recipeIng).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe ingredient not found")
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	if err := database.DB.Model(&recipeIng).Updates(updateData).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update recipe ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update ingredient")
		return
	}

	database.DB.Preload("Ingredient").First(&recipeIng, recipeIng.ID)
	log.InfoCtx(r.Context(), "recipe ingredient updated", "recipe_id", recipeID, "ingredient_id", ingredientID)
	response.JSON(w, http.StatusOK, recipeIng)
}

// DeleteRecipeIngredient remove ingrediente da receita
func DeleteRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	ingredientID := chi.URLParam(r, "ingredient_id")
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "You don't have permission to modify this recipe")
		return
	}

	result := database.DB.Where("recipe_id = ? AND id = ?", recipeID, ingredientID).
		Delete(&models.RecipeIngredient{})

	if result.Error != nil {
		log.ErrorCtx(r.Context(), "failed to delete recipe ingredient", "error", result.Error)
		response.Error(w, http.StatusInternalServerError, "Failed to delete ingredient")
		return
	}

	if result.RowsAffected == 0 {
		response.Error(w, http.StatusNotFound, "Recipe ingredient not found")
		return
	}

	log.InfoCtx(r.Context(), "recipe ingredient deleted", "recipe_id", recipeID, "ingredient_id", ingredientID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Ingredient removed from recipe"})
}

// GetRecipeNutrition calcula informação nutricional da receita
func GetRecipeNutrition(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Verificar se receita existe
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	var recipeIngredients []models.RecipeIngredient
	if err := database.DB.Preload("Ingredient").
		Where("recipe_id = ?", recipeID).
		Find(&recipeIngredients).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to calculate nutrition", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to calculate nutrition")
		return
	}

	totalCalories := 0.0
	totalProtein := 0.0
	totalCarbs := 0.0
	totalFat := 0.0
	totalFiber := 0.0

	for _, ri := range recipeIngredients {
		// Valores nutricionais são por 100g
		// Calcular proporção baseada na quantidade
		factor := ri.Quantity / 100.0

		totalCalories += ri.Ingredient.Calories * factor
		totalProtein += ri.Ingredient.Protein * factor
		totalCarbs += ri.Ingredient.Carbs * factor
		totalFat += ri.Ingredient.Fat * factor
		totalFiber += ri.Ingredient.Fiber * factor
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"total": map[string]float64{
			"calories": totalCalories,
			"protein":  totalProtein,
			"carbs":    totalCarbs,
			"fat":      totalFat,
			"fiber":    totalFiber,
		},
		"per_serving": map[string]float64{
			"calories": totalCalories / float64(recipe.Servings),
			"protein":  totalProtein / float64(recipe.Servings),
			"carbs":    totalCarbs / float64(recipe.Servings),
			"fat":      totalFat / float64(recipe.Servings),
			"fiber":    totalFiber / float64(recipe.Servings),
		},
		"servings": recipe.Servings,
	})
}

