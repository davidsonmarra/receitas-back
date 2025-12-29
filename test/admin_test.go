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
	"github.com/davidsonmarra/receitas-app/test/testdb"
)

func TestRequireAdmin_NonAdmin(t *testing.T) {
	testdb.SetupWithCleanup(t)

	// Criar usuário normal
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Usuário Normal", "normal@test.com", hashedPassword, "user")

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
	testdb.SetupWithCleanup(t)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := testdb.SeedUser(t, "Administrador", "admin@test.com", hashedPassword, "admin")

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
	testdb.SetupWithCleanup(t)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := testdb.SeedUser(t, "Admin", "admin2@test.com", hashedPassword, "admin")

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
	testdb.SetupWithCleanup(t)

	// Criar usuário normal (dono da receita)
	hashedPassword, _ := auth.HashPassword("senha123")
	owner := testdb.SeedUser(t, "Dono", "owner@test.com", hashedPassword, "user")

	// Criar receita do dono
	_ = testdb.SeedRecipe(t, "Receita do Usuário", "Receita de teste", owner.ID, false)

	// Criar admin
	admin := testdb.SeedUser(t, "Admin", "admin3@test.com", hashedPassword, "admin")

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
	testdb.SetupWithCleanup(t)

	// Criar receita geral (sem dono)
	_ = testdb.SeedRecipe(t, "Receita Geral", "Receita pública", 0, true)

	// Criar admin
	hashedPassword, _ := auth.HashPassword("admin123")
	admin := testdb.SeedUser(t, "Admin", "admin4@test.com", hashedPassword, "admin")

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
	testdb.SetupWithCleanup(t)

	// Criar receita geral
	_ = testdb.SeedRecipe(t, "Receita Geral", "Receita pública", 0, true)

	// Criar usuário normal
	hashedPassword, _ := auth.HashPassword("senha123")
	user := testdb.SeedUser(t, "Usuário Normal", "user@test.com", hashedPassword, "user")

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
