package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/davidsonmarra/receitas-app/pkg/storage"
)

const (
	maxImageSizeMB = 5         // 5MB
	maxImageWidth  = 2000      // pixels
	maxImageHeight = 2000      // pixels
	imageFolder    = "recipes" // pasta no Cloudinary
)

// GenerateUploadURL gera URL e assinatura para upload direto ao Cloudinary
func GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Receita não encontrada")
		return
	}

	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "Você não tem permissão para modificar esta receita")
		return
	}

	cloudinaryService, err := storage.NewCloudinaryService()
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to initialize cloudinary", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao configurar serviço de imagens")
		return
	}

	publicID := fmt.Sprintf("recipe_%s_%d", recipeID, time.Now().Unix())

	uploadSig, err := cloudinaryService.GenerateUploadSignature(publicID, imageFolder)
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to generate upload signature", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar URL de upload")
		return
	}

	log.InfoCtx(r.Context(), "upload signature generated",
		"recipe_id", recipeID,
		"public_id", publicID,
		"user_id", userID)

	response.JSON(w, http.StatusOK, uploadSig)
}

// ConfirmImageUploadRequest request de confirmação de upload
type ConfirmImageUploadRequest struct {
	PublicID  string `json:"public_id" validate:"required"`
	SecureURL string `json:"secure_url" validate:"required,url"`
	Width     int    `json:"width" validate:"required,min=1"`
	Height    int    `json:"height" validate:"required,min=1"`
	Format    string `json:"format" validate:"required"`
	Bytes     int    `json:"bytes" validate:"required,min=1"`
}

// ConfirmImageUpload confirma e salva metadados após upload direto ao Cloudinary
func ConfirmImageUpload(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	var req ConfirmImageUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, "Dados inválidos")
		return
	}

	if err := validator.New().Struct(req); err != nil {
		response.ValidationError(w, "Dados de confirmação incompletos")
		return
	}

	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Receita não encontrada")
		return
	}

	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "Você não tem permissão para modificar esta receita")
		return
	}

	// Se já tinha imagem antiga, tentar deletar (best effort)
	if recipe.ImagePublicID != "" {
		cloudinaryService, err := storage.NewCloudinaryService()
		if err == nil {
			cloudinaryService.DeleteImage(r.Context(), recipe.ImagePublicID)
		}
	}

	recipe.ImageURL = req.SecureURL
	recipe.ImagePublicID = req.PublicID

	if err := database.DB.Save(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update recipe with image", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao atualizar receita")
		return
	}

	log.InfoCtx(r.Context(), "recipe image confirmed",
		"recipe_id", recipe.ID,
		"user_id", userID,
		"public_id", req.PublicID)

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"message": "Imagem confirmada com sucesso",
		"recipe":  recipe,
	})
}


// DeleteRecipeImage remove a imagem de uma receita
func DeleteRecipeImage(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Obter userID do contexto
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Buscar receita existente
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Receita não encontrada")
		return
	}

	// Verificar autorização
	if !canModifyRecipe(&recipe, userID) {
		response.Error(w, http.StatusForbidden, "Você não tem permissão para modificar esta receita")
		return
	}

	// Verificar se tem imagem
	if recipe.ImagePublicID == "" {
		response.Error(w, http.StatusNotFound, "Esta receita não possui imagem")
		return
	}

	// Inicializar serviço Cloudinary
	cloudinaryService, err := storage.NewCloudinaryService()
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to initialize cloudinary", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao configurar serviço de imagens")
		return
	}

	// Deletar imagem do Cloudinary
	if err := cloudinaryService.DeleteImage(r.Context(), recipe.ImagePublicID); err != nil {
		log.ErrorCtx(r.Context(), "failed to delete image", "public_id", recipe.ImagePublicID, "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao deletar imagem")
		return
	}

	// Limpar campos de imagem no banco
	recipe.ImageURL = ""
	recipe.ImagePublicID = ""

	if err := database.DB.Save(&recipe).Error; err != nil {
		log.ErrorCtx(r.Context(), "failed to update recipe", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao atualizar receita")
		return
	}

	log.InfoCtx(r.Context(), "recipe image deleted", "recipe_id", recipe.ID, "user_id", userID)
	response.JSON(w, http.StatusOK, map[string]string{"message": "Imagem removida com sucesso"})
}

// GetRecipeImageVariants retorna URLs otimizadas da imagem em diferentes tamanhos
func GetRecipeImageVariants(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Buscar receita
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Receita não encontrada")
		return
	}

	// Verificar se tem imagem
	if recipe.ImagePublicID == "" {
		response.Error(w, http.StatusNotFound, "Esta receita não possui imagem")
		return
	}

	// Inicializar serviço Cloudinary
	cloudinaryService, err := storage.NewCloudinaryService()
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to initialize cloudinary", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao configurar serviço de imagens")
		return
	}

	// Gerar URLs otimizadas em diferentes tamanhos
	variants := make(map[string]interface{})

	// Thumbnail (pequeno para listagem)
	thumbnail, _ := cloudinaryService.GetOptimizedURL(recipe.ImagePublicID, 300, 300, "auto")
	variants["thumbnail"] = map[string]interface{}{
		"url":    thumbnail,
		"width":  300,
		"height": 300,
	}

	// Medium (para cards)
	medium, _ := cloudinaryService.GetOptimizedURL(recipe.ImagePublicID, 600, 600, "auto")
	variants["medium"] = map[string]interface{}{
		"url":    medium,
		"width":  600,
		"height": 600,
	}

	// Large (para visualização completa)
	large, _ := cloudinaryService.GetOptimizedURL(recipe.ImagePublicID, 1200, 1200, "auto")
	variants["large"] = map[string]interface{}{
		"url":    large,
		"width":  1200,
		"height": 1200,
	}

	// Original
	variants["original"] = map[string]interface{}{
		"url": recipe.ImageURL,
	}

	response.JSON(w, http.StatusOK, variants)
}

// GetOptimizedRecipeImage retorna URL otimizada customizada
func GetOptimizedRecipeImage(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")

	// Parse query params
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")
	quality := r.URL.Query().Get("quality")

	// Defaults
	if quality == "" {
		quality = "auto"
	}

	width := 800
	height := 800

	if widthStr != "" {
		if w, err := strconv.Atoi(widthStr); err == nil && w > 0 && w <= 2000 {
			width = w
		}
	}

	if heightStr != "" {
		if h, err := strconv.Atoi(heightStr); err == nil && h > 0 && h <= 2000 {
			height = h
		}
	}

	// Buscar receita
	var recipe models.Recipe
	if err := database.DB.First(&recipe, recipeID).Error; err != nil {
		response.Error(w, http.StatusNotFound, "Receita não encontrada")
		return
	}

	// Verificar se tem imagem
	if recipe.ImagePublicID == "" {
		response.Error(w, http.StatusNotFound, "Esta receita não possui imagem")
		return
	}

	// Inicializar serviço de imagens
	imageService, err := storage.ServiceFactory()
	if err != nil {
		log.ErrorCtx(r.Context(), "failed to initialize image service", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao configurar serviço de imagens")
		return
	}

	// Gerar URL otimizada
	optimizedURL, err := imageService.GetOptimizedURL(recipe.ImagePublicID, width, height, quality)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Erro ao gerar URL otimizada")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"url":     optimizedURL,
		"width":   width,
		"height":  height,
		"quality": quality,
	})
}
