package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/pagination"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/validation"
)

// AdminListRecipes lista todas as receitas (incluindo com dono) para admin
// Diferente do endpoint público, este inclui informações do usuário criador
func AdminListRecipes(w http.ResponseWriter, r *http.Request) {
	params := pagination.ExtractParams(r)

	var total int64
	if err := database.DB.Model(&models.Recipe{}).Count(&total).Error; err != nil {
		log.ErrorCtx(r.Context(), "admin failed to count recipes", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to count recipes")
		return
	}

	var recipes []models.Recipe
	offset := pagination.CalculateOffset(params)

	// Admin vê todas receitas, incluindo relação com usuário (Preload)
	if err := database.DB.Preload("User").Limit(params.Limit).Offset(offset).
		Order("created_at DESC").
		Find(&recipes).Error; err != nil {
		log.ErrorCtx(r.Context(), "admin failed to list recipes", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to list recipes")
		return
	}

	userID, _ := middleware.GetUserIDFromContext(r.Context())
	log.InfoCtx(r.Context(), "admin listed recipes", "admin_id", userID, "total", total)

	paginatedResponse := pagination.BuildResponse(recipes, params, total)
	response.JSON(w, http.StatusOK, paginatedResponse)
}

// AdminUpdateRecipe atualiza qualquer receita (mesmo de outro usuário ou receita geral)
func AdminUpdateRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Buscar receita existente
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	// Decode e validação
	var updateReq UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	if errs := validation.ValidateStruct(updateReq); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Aplicar updates
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

	if err := database.DB.Save(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "admin failed to update recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to update recipe")
		return
	}

	userID, _ := middleware.GetUserIDFromContext(r.Context())
	log.InfoCtx(r.Context(), "admin updated recipe",
		"admin_id", userID,
		"recipe_id", recipe.ID,
		"recipe_owner", recipe.UserID)

	response.JSON(w, http.StatusOK, recipe)
}

// AdminDeleteRecipe deleta qualquer receita (admin override)
func AdminDeleteRecipe(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Buscar receita para log de auditoria
	var recipe models.Recipe
	if err := database.DB.First(&recipe, id).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Recipe not found")
		return
	}

	if err := database.DB.Delete(&recipe, id).Error; err != nil {
		log.ErrorCtx(r.Context(), "admin failed to delete recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to delete recipe")
		return
	}

	userID, _ := middleware.GetUserIDFromContext(r.Context())
	log.InfoCtx(r.Context(), "admin deleted recipe",
		"admin_id", userID,
		"recipe_id", id,
		"recipe_owner", recipe.UserID,
		"recipe_title", recipe.Title)

	response.JSON(w, http.StatusOK, map[string]string{"message": "Recipe deleted by admin"})
}

// AdminCreateGeneralRecipe cria receita geral (sem user_id)
// Receitas gerais são receitas do sistema, não atribuídas a nenhum usuário
func AdminCreateGeneralRecipe(w http.ResponseWriter, r *http.Request) {
	var recipe models.Recipe

	if err := json.NewDecoder(r.Body).Decode(&recipe); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	if errs := validation.ValidateStruct(recipe); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Receita geral: user_id = nil (forçar)
	recipe.UserID = nil

	if err := database.DB.Create(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "admin failed to create general recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Failed to create recipe")
		return
	}

	userID, _ := middleware.GetUserIDFromContext(r.Context())
	log.InfoCtx(r.Context(), "admin created general recipe",
		"admin_id", userID,
		"recipe_id", recipe.ID,
		"recipe_title", recipe.Title)

	response.JSON(w, http.StatusCreated, recipe)
}
