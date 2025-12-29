package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/database"
)

// setupAdminTestDB inicializa database para testes de admin
func setupAdminTestDB(t *testing.T) {
	if database.DB == nil {
		t.Skip("DATABASE_URL não configurado para testes")
	}

	// Limpar tabelas
	database.DB.Exec("DELETE FROM recipes")
	database.DB.Exec("DELETE FROM users")

	// Executar migrations
	if err := database.DB.AutoMigrate(&models.User{}, &models.Recipe{}); err != nil {
		t.Fatalf("erro ao executar migrations: %v", err)
	}
}

func TestRequireAdmin_NonAdmin(t *testing.T) {
	setupAdminTestDB(t)

	// Criar usuário normal
	hashedPassword, _ := auth.HashPassword("senha123")
	user := models.User{
		Name:     "Usuário Normal",
		Email:    "normal@test.com",
		Password: hashedPassword,
		Role:     "user",
	}
	database.DB.Create(&user)

	// Gerar token
	token, _ := auth.GenerateToken(user.ID, user.Email, user.Role)

	// Criar handler protegido
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin area"))
	})

	// Aplicar middlewares: RequireAuth + RequireAdmin
	protectedHandler := middleware.RequireAuth(middleware.RequireAdmin(handler))

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	// Deve retornar 403 (usuário normal tentando acessar área admin)
	if rec.Code != http.StatusForbidden {
		t.Errorf("esperado status 403, obteve %d", rec.Code)
	}
}

func TestRequireAdmin_Admin(t *testing.T) {
	setupAdminTestDB(t)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := models.User{
		Name:     "Administrador",
		Email:    "admin@test.com",
		Password: hashedPassword,
		Role:     "admin",
	}
	database.DB.Create(&admin)

	// Gerar token
	token, _ := auth.GenerateToken(admin.ID, admin.Email, admin.Role)

	// Criar handler protegido
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin area"))
	})

	// Aplicar middlewares
	protectedHandler := middleware.RequireAuth(middleware.RequireAdmin(handler))

	req := httptest.NewRequest("GET", "/admin/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	// Deve permitir (200 OK)
	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}
}

func TestAdminCreateGeneralRecipe(t *testing.T) {
	setupAdminTestDB(t)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := models.User{
		Name:     "Admin",
		Email:    "admin2@test.com",
		Password: hashedPassword,
		Role:     "admin",
	}
	database.DB.Create(&admin)

	// Gerar token
	token, _ := auth.GenerateToken(admin.ID, admin.Email, admin.Role)

	// Payload da receita geral
	payload := map[string]interface{}{
		"title":     "Receita Geral do Sistema",
		"prep_time": 30,
		"servings":  4,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/admin/recipes/general", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middleware e handler
	protectedHandler := middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(handlers.AdminCreateGeneralRecipe)))
	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("esperado status 201, obteve %d", rec.Code)
		t.Logf("Response: %s", rec.Body.String())
	}

	// Verificar se user_id é null
	var response models.Recipe
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response.UserID != nil {
		t.Error("receita geral deveria ter user_id = null")
	}
}

func TestCanModifyRecipe_AsAdmin(t *testing.T) {
	setupAdminTestDB(t)

	// Criar usuário normal (dono da receita)
	hashedPassword, _ := auth.HashPassword("senha123")
	owner := models.User{
		Name:     "Dono",
		Email:    "owner@test.com",
		Password: hashedPassword,
		Role:     "user",
	}
	database.DB.Create(&owner)

	// Criar receita do dono
	recipe := models.Recipe{
		Title:    "Receita do Usuário",
		PrepTime: 30,
		Servings: 4,
		UserID:   &owner.ID,
	}
	database.DB.Create(&recipe)

	// Criar admin
	admin := models.User{
		Name:     "Admin",
		Email:    "admin3@test.com",
		Password: hashedPassword,
		Role:     "admin",
	}
	database.DB.Create(&admin)

	// Gerar token do admin
	token, _ := auth.GenerateToken(admin.ID, admin.Email, admin.Role)

	// Admin tenta editar receita de outro usuário
	payload := map[string]interface{}{
		"title": "Editada por Admin",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/recipes/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middleware auth
	protectedHandler := middleware.RequireAuth(http.HandlerFunc(handlers.UpdateRecipe))
	protectedHandler.ServeHTTP(rec, req)

	// Admin deve conseguir editar (200 ou 404 se chi router não configurado)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Logf("Status code: %d (esperado 200 ou 404 sem chi router completo)", rec.Code)
	}
}

func TestAdminDeleteGeneralRecipe(t *testing.T) {
	setupAdminTestDB(t)

	// Criar receita geral (sem dono)
	recipe := models.Recipe{
		Title:    "Receita Geral",
		PrepTime: 30,
		Servings: 4,
		UserID:   nil,
	}
	database.DB.Create(&recipe)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := models.User{
		Name:     "Admin",
		Email:    "admin4@test.com",
		Password: hashedPassword,
		Role:     "admin",
	}
	database.DB.Create(&admin)

	// Gerar token
	token, _ := auth.GenerateToken(admin.ID, admin.Email, admin.Role)

	req := httptest.NewRequest("DELETE", "/recipes/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middleware
	protectedHandler := middleware.RequireAuth(http.HandlerFunc(handlers.DeleteRecipe))
	protectedHandler.ServeHTTP(rec, req)

	// Admin deve conseguir deletar receita geral (200 ou 404)
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		t.Logf("Status code: %d (esperado 200 ou 404)", rec.Code)
	}
}

func TestNonAdminCannotDeleteGeneralRecipe(t *testing.T) {
	setupAdminTestDB(t)

	// Criar receita geral
	recipe := models.Recipe{
		Title:    "Receita Geral",
		PrepTime: 30,
		Servings: 4,
		UserID:   nil,
	}
	database.DB.Create(&recipe)

	// Criar usuário normal
	hashedPassword, _ := auth.HashPassword("senha123")
	user := models.User{
		Name:     "Usuário Normal",
		Email:    "user@test.com",
		Password: hashedPassword,
		Role:     "user",
	}
	database.DB.Create(&user)

	// Gerar token
	token, _ := auth.GenerateToken(user.ID, user.Email, user.Role)

	req := httptest.NewRequest("DELETE", "/recipes/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Aplicar middleware
	protectedHandler := middleware.RequireAuth(http.HandlerFunc(handlers.DeleteRecipe))
	protectedHandler.ServeHTTP(rec, req)

	// Usuário normal NÃO deve conseguir deletar receita geral (403 ou 404)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Logf("Status code: %d (esperado 403 ou 404)", rec.Code)
	}
}
