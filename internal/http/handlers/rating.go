package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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

// CreateOrUpdateRatingRequest representa os dados para criar/atualizar uma avaliação
type CreateOrUpdateRatingRequest struct {
	Score   int    `json:"score" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"omitempty,max=1000"`
}

// CreateOrUpdateRating cria ou atualiza uma avaliação (upsert)
func CreateOrUpdateRating(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Verificar se a receita existe
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Receita não encontrada")
			return
		}
		log.ErrorCtx(r.Context(), "failed to find recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar receita")
		return
	}

	// Decodificar request
	var req CreateOrUpdateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Formato de dados inválido.")
		return
	}

	// Validar os dados
	if errs := validation.ValidateStruct(req); errs != nil {
		message := validation.FormatErrors(errs)
		response.ValidationError(w, message)
		return
	}

	// Converter recipeID para uint
	recipeIDUint, err := strconv.ParseUint(recipeID, 10, 32)
	if err != nil {
		response.ValidationError(w, "ID da receita inválido")
		return
	}

	// Verificar se já existe uma avaliação deste usuário para esta receita
	var existingRating models.Rating
	err = database.DB.Where("recipe_id = ? AND user_id = ?", recipeIDUint, userID).
		First(&existingRating).Error

	if err == nil {
		// Avaliação existe, vamos atualizar
		existingRating.Score = req.Score
		existingRating.Comment = req.Comment

		if err := database.DB.Save(&existingRating).Error; err != nil {
			log.ErrorCtx(r.Context(), "failed to update rating", "error", err)
			response.Error(w, http.StatusInternalServerError, "Erro ao atualizar avaliação")
			return
		}

		// Carregar dados do usuário para retorno
		database.DB.Preload("User").First(&existingRating, existingRating.ID)

		log.InfoCtx(r.Context(), "rating updated", "rating_id", existingRating.ID, "user_id", userID, "recipe_id", recipeID)
		response.JSON(w, http.StatusOK, existingRating)
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Erro diferente de "não encontrado"
		log.ErrorCtx(r.Context(), "failed to check existing rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao verificar avaliação existente")
		return
	}

	// Criar nova avaliação
	rating := models.Rating{
		RecipeID: uint(recipeIDUint),
		UserID:   userID,
		Score:    req.Score,
		Comment:  req.Comment,
	}

	if err := database.DB.Create(&rating).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to create rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao criar avaliação")
		return
	}

	// Carregar dados do usuário para retorno
	database.DB.Preload("User").First(&rating, rating.ID)

	log.InfoCtx(r.Context(), "rating created", "rating_id", rating.ID, "user_id", userID, "recipe_id", recipeID)
	response.JSON(w, http.StatusCreated, rating)
}

// GetMyRating obtém a avaliação do usuário logado para uma receita específica
func GetMyRating(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	var rating models.Rating
	err := database.DB.
		Preload("User").
		Where("recipe_id = ? AND user_id = ?", recipeID, userID).
		First(&rating).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Você ainda não avaliou esta receita")
			return
		}
		log.ErrorCtx(r.Context(), "failed to get rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar avaliação")
		return
	}

	response.JSON(w, http.StatusOK, rating)
}

// DeleteMyRating deleta a avaliação do usuário logado (soft delete)
func DeleteMyRating(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Buscar a avaliação
	var rating models.Rating
	err := database.DB.Where("recipe_id = ? AND user_id = ?", recipeID, userID).
		First(&rating).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Avaliação não encontrada")
			return
		}
		log.ErrorCtx(r.Context(), "failed to find rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar avaliação")
		return
	}

	// Deletar (soft delete)
	if err := database.DB.Delete(&rating).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao deletar avaliação")
		return
	}

	log.InfoCtx(r.Context(), "rating deleted", "rating_id", rating.ID, "user_id", userID, "recipe_id", recipeID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Avaliação deletada com sucesso"})
}

// ListRecipeRatings lista todas as avaliações de uma receita com paginação
func ListRecipeRatings(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Verificar se a receita existe
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Receita não encontrada")
			return
		}
		log.ErrorCtx(r.Context(), "failed to find recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar receita")
		return
	}

	// Extrair parâmetros de paginação
	params := pagination.ExtractParams(r)

	// Extrair parâmetro de ordenação
	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "newest"
	}

	// Determinar ordem
	var orderBy string
	switch sortBy {
	case "highest":
		orderBy = "score DESC, created_at DESC"
	case "lowest":
		orderBy = "score ASC, created_at DESC"
	case "newest":
		orderBy = "created_at DESC"
	case "oldest":
		orderBy = "created_at ASC"
	default:
		orderBy = "created_at DESC"
	}

	// Count total de avaliações
	var total int64
	if err := database.DB.Model(&models.Rating{}).Where("recipe_id = ?", recipeID).Count(&total).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to count ratings", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao contar avaliações")
		return
	}

	// Buscar avaliações paginadas
	var ratings []models.Rating
	offset := pagination.CalculateOffset(params)
	if err := database.DB.
		Preload("User").
		Where("recipe_id = ?", recipeID).
		Order(orderBy).
		Limit(params.Limit).
		Offset(offset).
		Find(&ratings).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to list ratings", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao listar avaliações")
		return
	}

	// Montar resposta paginada
	paginatedResponse := pagination.BuildResponse(ratings, params, total)
	response.JSON(w, http.StatusOK, paginatedResponse)
}

// RatingStatsResponse representa as estatísticas de avaliações de uma receita
type RatingStatsResponse struct {
	AverageRating float64         `json:"average_rating"`
	TotalRatings  int64           `json:"total_ratings"`
	Distribution  map[string]int64 `json:"distribution"`
}

// GetRatingStats retorna estatísticas de avaliação de uma receita
func GetRatingStats(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Verificar se a receita existe
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Receita não encontrada")
			return
		}
		log.ErrorCtx(r.Context(), "failed to find recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar receita")
		return
	}

	// Calcular estatísticas
	stats := RatingStatsResponse{
		Distribution: make(map[string]int64),
	}

	// Calcular média e total
	var result struct {
		Average float64
		Total   int64
	}

	err := database.DB.Model(&models.Rating{}).
		Select("AVG(score) as average, COUNT(*) as total").
		Where("recipe_id = ?", recipeID).
		Scan(&result).Error

	if err != nil {
		log.ErrorCtx(r.Context(), "failed to calculate rating stats", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao calcular estatísticas")
		return
	}

	stats.AverageRating = result.Average
	stats.TotalRatings = result.Total

	// Calcular distribuição
	var distributions []struct {
		Score int
		Count int64
	}

	err = database.DB.Model(&models.Rating{}).
		Select("score, COUNT(*) as count").
		Where("recipe_id = ?", recipeID).
		Group("score").
		Scan(&distributions).Error

	if err != nil {
		log.ErrorCtx(r.Context(), "failed to calculate rating distribution", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao calcular distribuição")
		return
	}

	// Preencher distribuição (garantir que todas as notas de 1-5 estejam presentes)
	for i := 1; i <= 5; i++ {
		stats.Distribution[strconv.Itoa(i)] = 0
	}
	for _, dist := range distributions {
		stats.Distribution[strconv.Itoa(dist.Score)] = dist.Count
	}

	response.JSON(w, http.StatusOK, stats)
}

// AdminDeleteRating permite que admins deletem qualquer avaliação (moderação)
func AdminDeleteRating(w http.ResponseWriter, r *http.Request) {
	ratingID := chi.URLParam(r, "rating_id")

	// Obter userID do contexto (para log)
	userID, _ := middleware.GetUserIDFromContext(r.Context())

	// Buscar a avaliação
	var rating models.Rating
	if err := database.DB.First(&rating, ratingID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Avaliação não encontrada")
			return
		}
		log.ErrorCtx(r.Context(), "failed to find rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar avaliação")
		return
	}

	// Deletar (soft delete)
	if err := database.DB.Delete(&rating).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to delete rating", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao deletar avaliação")
		return
	}

	log.InfoCtx(r.Context(), "rating deleted by admin", 
		"rating_id", rating.ID, 
		"rating_user_id", rating.UserID, 
		"admin_user_id", userID,
		"recipe_id", rating.RecipeID)
	
	response.JSON(w, http.StatusOK, map[string]string{"message": "Avaliação deletada com sucesso"})
}

// calculateRatingStats calcula as estatísticas de avaliação de uma receita
// Retorna a média e o total de avaliações
func calculateRatingStats(db *gorm.DB, recipeID uint) (avgRating float64, count int64) {
	var result struct {
		Average float64
		Total   int64
	}

	err := db.Model(&models.Rating{}).
		Select("COALESCE(AVG(score), 0) as average, COUNT(*) as total").
		Where("recipe_id = ?", recipeID).
		Scan(&result).Error

	if err != nil {
		return 0, 0
	}

	return result.Average, result.Total
}

