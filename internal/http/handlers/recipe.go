package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// CreateRecipe cria uma nova receita
func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe models.Recipe

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Validar os dados
	if errs := validation.ValidateStruct(recipe); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
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

// UpdateRecipeRequest representa os dados permitidos para atualização
type UpdateRecipeRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=200"`
	Description *string `json:"description"`
	PrepTime    *int    `json:"prep_time" validate:"omitempty,min=1"`
	Servings    *int    `json:"servings" validate:"omitempty,min=1"`
	Difficulty  *string `json:"difficulty" validate:"omitempty,oneof=fácil média difícil"`
}

// UpdateRecipe atualiza uma receita
func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Buscar receita existente
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Decodificar para struct de update (sem campos protegidos)
	var updateReq UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Validar os dados
	if errs := validation.ValidateStruct(updateReq); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Aplicar apenas os campos que foram enviados
	if updateReq.Title != nil {
		recipe.Title = *updateReq.Title
	}
	if updateReq.Description != nil {
		recipe.Description = *updateReq.Description
	}
	if updateReq.PrepTime != nil {
		recipe.PrepTime = *updateReq.PrepTime
	}
	if updateReq.Servings != nil {
		recipe.Servings = *updateReq.Servings
	}
	if updateReq.Difficulty != nil {
		recipe.Difficulty = *updateReq.Difficulty
	}

	// Salvar no banco
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
