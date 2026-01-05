package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/pagination"
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

	// Obter userID do contexto (adicionado pelo middleware RequireAuth)
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Atribuir criador à receita
	recipe.UserID = &userID

	if err := database.DB.Create(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	log.InfoCtx(r.Context(), "recipe created", "id", recipe.ID, "user_id", userID)
	response.JSON(w, http.StatusCreated, recipe)
}

// ListRecipes lista todas as receitas com paginação
func ListRecipes(w http.ResponseWriter, r *http.Request) {
	// Extrair parâmetros de paginação
	params := pagination.ExtractParams(r)

	// Extrair parâmetro de ordenação
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "newest"
	}

	// Count total de receitas
	var total int64
	if err := database.DB.Model(&models.Recipe{}).Count(&total).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to count recipes", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to count recipes")
		return
	}

	// Buscar receitas paginadas
	var recipes []models.Recipe
	offset := pagination.CalculateOffset(params)
	
	query := database.DB.Limit(params.Limit).Offset(offset)

	// Aplicar ordenação
	if sortBy == "rating" {
		// Ordenar por rating (média de avaliações)
		// Usa subquery para calcular a média e ordenar
		query = query.
			Joins("LEFT JOIN (SELECT recipe_id, AVG(score) as avg_score, COUNT(*) as rating_count FROM ratings WHERE deleted_at IS NULL GROUP BY recipe_id) r ON r.recipe_id = recipes.id").
			Order("COALESCE(r.avg_score, 0) DESC, r.rating_count DESC, recipes.created_at DESC")
	} else {
		// Ordenação padrão por data de criação
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&recipes).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list recipes", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list recipes")
		return
	}

	// Calcular estatísticas de avaliação para cada receita
	for i := range recipes {
		recipes[i].AverageRating, recipes[i].RatingCount = calculateRatingStats(database.DB, recipes[i].ID)
	}

	// Montar resposta paginada
	paginatedResponse := pagination.BuildResponse(recipes, params, total)
	response.JSON(w, http.StatusOK, paginatedResponse)
}

// GetRecipe busca uma receita por ID
func GetRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var recipe models.Recipe
	if err := database.DB.
		Preload("User").
		Preload("Ingredients", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC, id ASC")
		}).
		Preload("Ingredients.Ingredient").
		First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Calcular estatísticas de avaliação
	recipe.AverageRating, recipe.RatingCount = calculateRatingStats(database.DB, recipe.ID)

	response.JSON(w, http.StatusOK, recipe)
}

// UpdateRecipeRequest representa os dados permitidos para atualização
type UpdateRecipeRequest struct {
	Title        *string `json:"title" validate:"omitempty,min=3,max=200"`
	Description  *string `json:"description"`
	Instructions *string `json:"instructions" validate:"omitempty,min=10,max=10000"`
	PrepTime     *int    `json:"prep_time" validate:"omitempty,min=1"`
	Servings     *int    `json:"servings" validate:"omitempty,min=1"`
	Difficulty   *string `json:"difficulty" validate:"omitempty,oneof=fácil média difícil"`
}

// UpdateRecipe atualiza uma receita
func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Buscar receita existente
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Verificar autorização
	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "Você não tem permissão para modificar esta receita")
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
	if updateReq.Instructions != nil {
		recipe.Instructions = *updateReq.Instructions
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

	log.InfoCtx(r.Context(), "recipe updated", "id", recipe.ID, "user_id", userID)
	response.JSON(w, http.StatusOK, recipe)
}

// DeleteRecipe remove uma receita (soft delete)
func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Buscar receita existente antes de deletar
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Verificar autorização
	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "Você não tem permissão para deletar esta receita")
		return
	}

	if err := database.DB.Delete(&recipe, id).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete recipe")
		return
	}

	log.InfoCtx(r.Context(), "recipe deleted", "id", id, "user_id", userID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Recipe deleted"})
}

// canModifyRecipe verifica se o usuário pode modificar a receita
func canModifyRecipe(recipe *models.Recipe, userID uint) bool {
	// Verificar se usuário é admin (admin pode modificar qualquer receita)
	if isAdmin(userID) {
		return true
	}

	// Se não é admin, verificar ownership
	if recipe.UserID != nil {
		// Apenas o criador pode modificar
		return *recipe.UserID == userID
	}

	// Receita geral (sem dono) - apenas admin pode modificar
	return false
}
