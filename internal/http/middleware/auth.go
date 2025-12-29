package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/davidsonmarra/receitas-app/pkg/auth"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// Chaves do contexto
type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

// RequireAuth é um middleware que valida o token JWT
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair token do header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "Token não fornecido")
			return
		}

		// Remover "Bearer " do início
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Error(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		// Verificar se o token está na blacklist
		if auth.IsBlacklisted(tokenString) {
			response.Error(w, http.StatusUnauthorized, "Token inválido")
			return
		}

		// Validar token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Token inválido")
			return
		}

		// Adicionar informações do usuário ao contexto
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// Continuar para o próximo handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extrai o ID do usuário do contexto
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint)
	return userID, ok
}

// GetUserEmailFromContext extrai o email do usuário do contexto
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}
