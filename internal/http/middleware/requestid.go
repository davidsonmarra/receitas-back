package middleware

import (
	"net/http"

	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestID é um middleware que adiciona um ID único para cada requisição
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Gerar um novo UUID para a requisição
		requestID := uuid.New().String()

		// Adicionar ao contexto
		ctx := log.WithRequestID(r.Context(), requestID)

		// Adicionar ao header de resposta
		w.Header().Set(requestIDHeader, requestID)

		// Continuar com o contexto atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
