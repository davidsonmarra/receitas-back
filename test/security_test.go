package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsonmarra/receitas-app/internal/http/routes"
)

func TestSecurityHeaders(t *testing.T) {
	router := routes.Setup()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verificar headers de segurança
	tests := []struct {
		header   string
		expected string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), accelerometer=(), gyroscope=()"},
	}

	for _, tt := range tests {
		got := w.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("%s: got %q, want %q", tt.header, got, tt.expected)
		}
	}
}

func TestSecurityHeaders_HSTS(t *testing.T) {
	router := routes.Setup()

	// Simular requisição HTTPS (via proxy)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")
	expected := "max-age=31536000; includeSubDomains; preload"

	if hsts != expected {
		t.Errorf("HSTS header: got %q, want %q", hsts, expected)
	}
}

func TestSecurityHeaders_NoHSTSOnHTTP(t *testing.T) {
	router := routes.Setup()

	// Requisição HTTP (sem X-Forwarded-Proto)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")

	if hsts != "" {
		t.Errorf("HSTS header should not be set on HTTP: got %q", hsts)
	}
}
