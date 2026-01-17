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
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
			response.Error(w, http.StatusUnauthorized, "Token não fornecido")
			return
		}

		// Remover "Bearer " do início
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
			response.Error(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		// Verificar se o token está na blacklist
		if auth.IsBlacklisted(tokenString) {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			response.ErrorWithCode(w, http.StatusUnauthorized, "Token inválido", "TOKEN_INVALID")
			return
		}

		// Validar token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			// Verificar se é erro de expiração
			if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "exp") {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				w.Header().Set("X-Token-Expired", "true")
				response.ErrorWithCode(w, http.StatusUnauthorized, "Token expirado", "TOKEN_EXPIRED")
				return
			}
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			response.ErrorWithCode(w, http.StatusUnauthorized, "Token inválido", "TOKEN_INVALID")
			return
		}

		// Validar que é um access token (não aceitar refresh tokens)
		if claims.TokenType != auth.TokenTypeAccess {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			response.ErrorWithCode(w, http.StatusUnauthorized, "Tipo de token inválido", "TOKEN_TYPE_INVALID")
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
