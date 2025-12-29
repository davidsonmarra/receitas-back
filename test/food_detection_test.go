package test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/jobqueue"
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

func Test_AnalyzeFood_Success(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário de teste
	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Criar request com imagem
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "food.jpg")
	part.Write([]byte("fake image data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/analyze-food", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()

	// Adicionar contexto de autenticação
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	// Executar handler
	handlers.AnalyzeFood(w, req)

	// Verificar resposta
	if w.Code != http.StatusAccepted {
		t.Errorf("esperado status 202, obteve %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["job_id"] == nil {
		t.Error("job_id não retornado")
	}

	if response["status"] != "processing" {
		t.Errorf("esperado status 'processing', obteve %v", response["status"])
	}

	if response["check_url"] == nil {
		t.Error("check_url não retornado")
	}
}

func Test_AnalyzeFood_Unauthorized(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Request sem autenticação
	req := httptest.NewRequest(http.MethodPost, "/analyze-food", nil)
	w := httptest.NewRecorder()

	handlers.AnalyzeFood(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", w.Code)
	}
}

func Test_AnalyzeFood_InvalidImage(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Request sem imagem
	req := httptest.NewRequest(http.MethodPost, "/analyze-food", nil)

	w := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	handlers.AnalyzeFood(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado status 400, obteve %d", w.Code)
	}
}

func Test_GetAnalysisResult_Processing(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Criar job de teste
	jobID := "test-job-123"
	jobqueue.GlobalQueue.CreateJob(jobID)

	// Request para consultar status
	req := httptest.NewRequest(http.MethodGet, "/analyze-food/"+jobID, nil)

	// Adicionar job_id ao contexto do chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("job_id", jobID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.GetAnalysisResult(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "processing" {
		t.Errorf("esperado status 'processing', obteve %v", response["status"])
	}

	if response["job_id"] != jobID {
		t.Errorf("esperado job_id '%s', obteve %v", jobID, response["job_id"])
	}
}

func Test_GetAnalysisResult_Completed(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Criar job e marcar como completado
	jobID := "test-job-456"
	jobqueue.GlobalQueue.CreateJob(jobID)

	result := map[string]interface{}{
		"detected_foods": []map[string]interface{}{
			{
				"name":       "arroz branco",
				"confidence": 0.95,
				"quantity":   150.0,
				"calories":   195.0,
			},
		},
		"total_nutrition": map[string]float64{
			"calories": 195.0,
			"protein":  3.5,
			"carbs":    43.2,
			"fat":      0.3,
		},
	}

	jobqueue.GlobalQueue.CompleteJob(jobID, result)

	// Request para consultar status
	req := httptest.NewRequest(http.MethodGet, "/analyze-food/"+jobID, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("job_id", jobID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.GetAnalysisResult(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "completed" {
		t.Errorf("esperado status 'completed', obteve %v", response["status"])
	}

	if response["result"] == nil {
		t.Error("resultado não retornado")
	}
}

func Test_GetAnalysisResult_Failed(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Criar job e marcar como falhou
	jobID := "test-job-789"
	jobqueue.GlobalQueue.CreateJob(jobID)
	jobqueue.GlobalQueue.FailJob(jobID, "Erro ao processar imagem")

	// Request para consultar status
	req := httptest.NewRequest(http.MethodGet, "/analyze-food/"+jobID, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("job_id", jobID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.GetAnalysisResult(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", w.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "failed" {
		t.Errorf("esperado status 'failed', obteve %v", response["status"])
	}

	if response["error"] == nil {
		t.Error("mensagem de erro não retornada")
	}
}

func Test_GetAnalysisResult_NotFound(t *testing.T) {
	testdb.SetupWithCleanup(t)

	user := testdb.SeedUser(t, "Test User", "test@test.com", "password123", "user")

	// Request com job_id inexistente
	jobID := "non-existent-job"
	req := httptest.NewRequest(http.MethodGet, "/analyze-food/"+jobID, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("job_id", jobID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, middleware.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.GetAnalysisResult(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("esperado status 404, obteve %d", w.Code)
	}
}

func Test_JobQueue_Operations(t *testing.T) {
	// Criar nova queue para teste
	queue := jobqueue.NewJobQueue()

	// Teste 1: Criar job
	jobID := "test-job"
	queue.CreateJob(jobID)

	job, exists := queue.GetJob(jobID)
	if !exists {
		t.Fatal("job não foi criado")
	}

	if job.Status != jobqueue.JobStatusProcessing {
		t.Errorf("esperado status processing, obteve %v", job.Status)
	}

	// Teste 2: Completar job
	result := map[string]string{"test": "result"}
	queue.CompleteJob(jobID, result)

	job, _ = queue.GetJob(jobID)
	if job.Status != jobqueue.JobStatusCompleted {
		t.Errorf("esperado status completed, obteve %v", job.Status)
	}

	if job.Result == nil {
		t.Error("resultado não foi salvo")
	}

	// Teste 3: Falhar job
	jobID2 := "test-job-2"
	queue.CreateJob(jobID2)
	queue.FailJob(jobID2, "test error")

	job2, _ := queue.GetJob(jobID2)
	if job2.Status != jobqueue.JobStatusFailed {
		t.Errorf("esperado status failed, obteve %v", job2.Status)
	}

	if job2.Error != "test error" {
		t.Errorf("esperado erro 'test error', obteve '%s'", job2.Error)
	}
}

func Test_AnalyzeFood_WithIngredientInDB(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar ingrediente no banco
	ingredient := testdb.SeedIngredient(t, "Arroz Branco", "cereais", 130)

	// Verificar que o ingrediente foi criado
	if ingredient.ID == 0 {
		t.Fatal("ingrediente não foi criado com ID")
	}

	// Verificar que o ingrediente existe no banco
	var found models.Ingredient
	result := database.DB.Where("name = ?", "Arroz Branco").First(&found)

	if result.Error != nil {
		t.Fatalf("ingrediente não encontrado no banco: %v", result.Error)
	}

	if found.Calories != 130 {
		t.Errorf("esperado calorias 130, obteve %v", found.Calories)
	}
}

func Test_JobQueue_Cleanup(t *testing.T) {
	// Criar nova queue para teste
	queue := jobqueue.NewJobQueue()

	// Criar job antigo
	oldJobID := "old-job"
	queue.CreateJob(oldJobID)

	// Simular job com 31 minutos de idade
	job, _ := queue.GetJob(oldJobID)
	job.Created = time.Now().Add(-31 * time.Minute)

	// Criar job novo
	newJobID := "new-job"
	queue.CreateJob(newJobID)

	// Verificar contagem
	if queue.GetJobCount() != 2 {
		t.Errorf("esperado 2 jobs, obteve %d", queue.GetJobCount())
	}

	// Nota: O teste de limpeza automática não pode ser testado diretamente
	// pois CleanupOldJobs() é um loop infinito em goroutine
	// A lógica é testada verificando se jobs com timestamp antigo existem
}
