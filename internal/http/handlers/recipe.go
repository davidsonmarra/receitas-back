package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// CreateRecipe cria uma nova receita
func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe models.Recipe

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := database.DB.Create(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	log.InfoCtx(r.Context(), "recipe created", "id", recipe.ID)
	response.JSON(w, http.StatusCreated, recipe)
}

// ListRecipes lista todas as receitas
func ListRecipes(w http.ResponseWriter, r *http.Request) {
	var recipes []models.Recipe

	if err := database.DB.Find(&recipes).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list recipes", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list recipes")
		return
	}

	response.JSON(w, http.StatusOK, recipes)
}

// GetRecipe busca uma receita por ID
func GetRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	response.JSON(w, http.StatusOK, recipe)
}

// UpdateRecipe atualiza uma receita
func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := database.DB.Save(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	log.InfoCtx(r.Context(), "recipe updated", "id", recipe.ID)
	response.JSON(w, http.StatusOK, recipe)
}

// DeleteRecipe remove uma receita (soft delete)
func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := database.DB.Delete(&models.Recipe{}, id).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete recipe")
		return
	}

	log.InfoCtx(r.Context(), "recipe deleted", "id", id)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Recipe deleted"})
}
