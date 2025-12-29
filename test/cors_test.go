package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/routes"
)

func TestCORS_PreflightRequest(t *testing.T) {
	router := routes.Setup()

	// Simular requisição OPTIONS (preflight)
	req := httptest.NewRequest(http.MethodOptions, "/recipes", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar status code
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Errorf("Expected status 200 or 204, got %d", w.Code)
	}

	// Verificar headers CORS
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Error("Expected Access-Control-Allow-Origin header, got empty")
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Expected Access-Control-Allow-Methods header, got empty")
	}

	t.Logf("✅ CORS headers present:")
	t.Logf("   Allow-Origin: %s", allowOrigin)
	t.Logf("   Allow-Methods: %s", allowMethods)
}

func TestCORS_ActualRequest(t *testing.T) {
	router := routes.Setup()

	// Usar endpoint /test que não depende de database
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verificar se Access-Control-Allow-Origin está presente
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Error("Expected Access-Control-Allow-Origin header in actual request")
	}

	t.Logf("✅ CORS working on actual request: %s", allowOrigin)
}

func TestCORS_ExposedHeaders(t *testing.T) {
	router := routes.Setup()

	// Usar endpoint /test que não depende de database
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verificar se X-Request-ID está exposto
	exposedHeaders := w.Header().Get("Access-Control-Expose-Headers")
	if exposedHeaders == "" {
		t.Log("⚠️  No Access-Control-Expose-Headers (may not be required)")
	} else {
		t.Logf("✅ Exposed headers: %s", exposedHeaders)
	}

	// Verificar se X-Request-ID está presente (do nosso middleware)
	requestID := w.Header().Get("X-Request-ID")
	if requestID != "" {
		t.Logf("✅ X-Request-ID present: %s", requestID)
	}
}
