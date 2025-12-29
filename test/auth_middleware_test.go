package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/middleware"
	"github.com/davidsonmarra/receitas-app/pkg/auth"
)

func TestRequireAuth_NoToken(t *testing.T) {
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestRequireAuth_InvalidFormat(t *testing.T) {
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer token.invalido.xyz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401, obteve %d", rec.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	// Gerar um token válido
	token, err := auth.GenerateToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	var contextUserID uint
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair userID do contexto
		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			t.Error("userID não encontrado no contexto")
		}
		contextUserID = userID
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado status 200, obteve %d", rec.Code)
	}

	if contextUserID != 123 {
		t.Errorf("esperado userID 123 no contexto, obteve %d", contextUserID)
	}
}

func TestRequireAuth_BlacklistedToken(t *testing.T) {
	// Gerar um token válido
	token, err := auth.GenerateToken(456, "blacklisted@example.com")
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	// Validar para obter claims
	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}

	// Adicionar à blacklist
	auth.AddToBlacklist(token, claims.ExpiresAt.Time)

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("esperado status 401 para token blacklisted, obteve %d", rec.Code)
	}
}

func TestGetUserEmailFromContext(t *testing.T) {
	// Gerar um token válido
	email := "context@example.com"
	token, err := auth.GenerateToken(789, email)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	var contextEmail string
	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair email do contexto
		e, ok := middleware.GetUserEmailFromContext(r.Context())
		if !ok {
			t.Error("email não encontrado no contexto")
		}
		contextEmail = e
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if contextEmail != email {
		t.Errorf("esperado email %s no contexto, obteve %s", email, contextEmail)
	}
}
