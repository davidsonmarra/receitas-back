package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/pagination"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// ListIngredients lista todos ingredientes com filtros e paginação
func ListIngredients(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r)

	query := database.DB.Model(&models.Ingredient{})

	// Filtro por nome
	if search := r.URL.Query().Get("search"); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filtro por categoria
	if category := r.URL.Query().Get("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to count ingredients", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to count ingredients")
		return
	}

	var ingredients []models.Ingredient
	offset := pagination.CalculateOffset(params)

	if err := query.Limit(params.Limit).Offset(offset).
		Order("name ASC").
		Find(&ingredients).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list ingredients", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list ingredients")
		return
	}

	log.InfoCtx(r.Context(), "ingredients listed", "total", total, "returned", len(ingredients))
	paginatedResponse := pagination.BuildResponse(ingredients, params, total)
	response.JSON(w, http.StatusOK, paginatedResponse)
}

// GetIngredient retorna um ingrediente por ID
func GetIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var ingredient models.Ingredient
	if err := database.DB.First(&ingredient, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	response.JSON(w, http.StatusOK, ingredient)
}

// CreateIngredient cria um novo ingrediente (admin only)
func CreateIngredient(w http.ResponseWriter, r *http.Request) {
	var ingredient models.Ingredient

	if err := json.NewDecoder(r.Body).Decode(&ingredient); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	if errs := validation.ValidateStruct(ingredient); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Normalizar categoria para lowercase
	ingredient.Category = strings.ToLower(ingredient.Category)
	if ingredient.Source == "" {
		ingredient.Source = "manual"
	}

	if err := database.DB.Create(&ingredient).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient created", "id", ingredient.ID, "name", ingredient.Name)
	response.JSON(w, http.StatusCreated, ingredient)
}

// UpdateIngredient atualiza um ingrediente (admin only)
func UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var ingredient models.Ingredient
	if err := database.DB.First(&ingredient, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Ingredient not found")
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Normalizar categoria se presente
	if category, ok := updateData["category"].(string); ok {
		updateData["category"] = strings.ToLower(category)
	}

	if err := database.DB.Model(&ingredient).Updates(updateData).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient updated", "id", ingredient.ID)
	response.JSON(w, http.StatusOK, ingredient)
}

// DeleteIngredient remove um ingrediente (admin only)
func DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verificar se ingrediente está em uso em alguma receita
	var count int64
	database.DB.Model(&models.RecipeIngredient{}).Where("ingredient_id = ?", id).Count(&count)

	if count > 0 {
		response.ValidationError(w, "Não é possível deletar ingrediente em uso em receitas.")
		return
	}

	if err := database.DB.Delete(&models.Ingredient{}, id).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete ingredient", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete ingredient")
		return
	}

	log.InfoCtx(r.Context(), "ingredient deleted", "id", id)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Ingredient deleted"})
}

// GetCategories retorna lista de categorias disponíveis
func GetCategories(w http.ResponseWriter, r *http.Request) {
	var categories []string

	database.DB.Model(&models.Ingredient{}).
		Distinct("category").
		Where("category != ''").
		Order("category ASC").
		Pluck("category", &categories)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

