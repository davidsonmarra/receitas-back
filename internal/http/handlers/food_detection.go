package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/foodai"
	"github.com/davidsonmarra/receitas-app/pkg/jobqueue"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

const (
	maxFoodImageSize = 5 * 1024 * 1024 // 5MB
)

// AnalyzeFood inicia análise assíncrona de alimentos em uma imagem
func AnalyzeFood(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Obter userID do contexto (validado pelo middleware RequireAuth)
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	// Parse multipart form (max 10MB em memória)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.ErrorCtx(ctx, "erro ao parsear form", "error", err)
		response.ValidationError(w, "Erro ao processar formulário")
		return
	}

	// Obter arquivo de imagem
	file, _, err := r.FormFile("image")
	if err != nil {
		log.WarnCtx(ctx, "imagem não fornecida", "error", err)
		response.ValidationError(w, "Imagem é obrigatória")
		return
	}
	defer file.Close()

	// Ler dados da imagem
	imageData, err := io.ReadAll(file)
	if err != nil {
		log.ErrorCtx(ctx, "erro ao ler imagem", "error", err)
		response.Error(w, http.StatusInternalServerError, "Erro ao processar imagem")
		return
	}

	// Validar tamanho
	if len(imageData) > maxFoodImageSize {
		response.ValidationError(w, fmt.Sprintf("Imagem muito grande. Máximo: %dMB", maxFoodImageSize/(1024*1024)))
		return
	}

	if len(imageData) == 0 {
		response.ValidationError(w, "Imagem vazia")
		return
	}

	// Gerar job ID único
	jobID := uuid.New().String()

	// Criar job na queue com status "processing"
	jobqueue.GlobalQueue.CreateJob(jobID)

	// Processar imagem em goroutine (assíncrono)
	go processImageAsync(jobID, imageData, userID)

	// Retornar imediatamente com job_id
	log.InfoCtx(ctx, "análise de alimentos iniciada", "job_id", jobID, "user_id", userID)

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":    jobID,
		"status":    "processing",
		"check_url": fmt.Sprintf("/analyze-food/%s", jobID),
	})
}

// GetAnalysisResult consulta o status e resultado de uma análise
func GetAnalysisResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "job_id")

	if jobID == "" {
		response.ValidationError(w, "job_id é obrigatório")
		return
	}

	// Buscar job na queue
	job, exists := jobqueue.GlobalQueue.GetJob(jobID)
	if !exists {
		log.WarnCtx(ctx, "job não encontrado", "job_id", jobID)
		response.Error(w, http.StatusNotFound, "Análise não encontrada")
		return
	}

	// Retornar resultado completo
	response.JSON(w, http.StatusOK, job)
}

// processImageAsync processa a imagem em background
func processImageAsync(jobID string, imageData []byte, userID uint) {
	log.Info("processando imagem", "job_id", jobID, "size", len(imageData))

	// Criar cliente Gemini
	client := foodai.NewGeminiClient()

	// Analisar imagem com Gemini
	detected, err := client.AnalyzeFood(imageData)
	if err != nil {
		log.Error("erro ao analisar com Gemini", "job_id", jobID, "error", err)
		jobqueue.GlobalQueue.FailJob(jobID, fmt.Sprintf("Erro ao analisar imagem: %v", err))
		return
	}

	if len(detected.Foods) == 0 {
		log.Warn("nenhum alimento detectado", "job_id", jobID)
		jobqueue.GlobalQueue.FailJob(jobID, "Nenhum alimento detectado na imagem")
		return
	}

	// Processar cada alimento detectado
	var results []map[string]interface{}
	var totalCalories, totalProtein, totalCarbs, totalFat float64

	for _, food := range detected.Foods {
		// Buscar no banco de ingredientes
		var ingredient models.Ingredient
		result := database.DB.Where("name ILIKE ?", "%"+food.Name+"%").First(&ingredient)

		var calories, protein, carbs, fat float64
		foundInDB := result.Error == nil

		if foundInDB {
			// Calcular baseado na quantidade detectada
			factor := food.Quantity / 100.0 // DB tem valores por 100g
			calories = ingredient.Calories * factor
			protein = ingredient.Protein * factor
			carbs = ingredient.Carbs * factor
			fat = ingredient.Fat * factor

			log.Debug("ingrediente encontrado no DB", "name", food.Name, "db_name", ingredient.Name, "calories", calories)
		} else {
			// Usar valores médios estimados (150 kcal/100g)
			factor := food.Quantity / 100.0
			calories = 150 * factor
			protein = 5 * factor  // ~5g proteína/100g
			carbs = 20 * factor   // ~20g carbs/100g
			fat = 3 * factor      // ~3g gordura/100g

			log.Warn("ingrediente não encontrado no DB, usando valores médios", "name", food.Name)
		}

		results = append(results, map[string]interface{}{
			"name":        food.Name,
			"confidence":  food.Confidence,
			"quantity":    food.Quantity,
			"unit":        "g",
			"calories":    roundToOneDecimal(calories),
			"protein":     roundToOneDecimal(protein),
			"carbs":       roundToOneDecimal(carbs),
			"fat":         roundToOneDecimal(fat),
			"found_in_db": foundInDB,
		})

		totalCalories += calories
		totalProtein += protein
		totalCarbs += carbs
		totalFat += fat
	}

	// Montar resultado final
	finalResult := map[string]interface{}{
		"detected_foods": results,
		"total_nutrition": map[string]float64{
			"calories": roundToOneDecimal(totalCalories),
			"protein":  roundToOneDecimal(totalProtein),
			"carbs":    roundToOneDecimal(totalCarbs),
			"fat":      roundToOneDecimal(totalFat),
		},
	}

	// Marcar job como completado
	jobqueue.GlobalQueue.CompleteJob(jobID, finalResult)

	log.Info("análise completada", "job_id", jobID, "foods_detected", len(results), "total_calories", totalCalories)
}

// roundToOneDecimal arredonda para uma casa decimal
func roundToOneDecimal(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}

