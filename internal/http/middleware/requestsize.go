package middleware

import (
	"net/http"

	"github.com/davidsonmarra/receitas-app/pkg/response"
)

const (
	// MaxRequestSize é o tamanho máximo permitido para o body da requisição (1MB)
	MaxRequestSize = 1 << 20 // 1MB em bytes
)

// RequestSizeLimit limita o tamanho do body da requisição
func RequestSizeLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limitar o tamanho do body
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)

		// Continuar para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// handleRequestTooLarge é chamado quando o body excede o limite
// Esta função seria chamada automaticamente pelo MaxBytesReader quando o limite for excedido
func handleRequestTooLarge(w http.ResponseWriter) {
	response.ValidationError(w, "A requisição é muito grande. Limite: 1MB.")
}

