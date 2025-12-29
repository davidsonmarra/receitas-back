package middleware

import (
	"net/http"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// RequireAdmin verifica se o usuário autenticado é admin
// Middleware de segurança que implementa:
// - Fail secure (default deny)
// - Defense in depth (verifica banco, não confia apenas no JWT)
// - Audit trail (logs de todas tentativas)
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obter userID do contexto (já validado por RequireAuth)
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "Autenticação necessária")
			return
		}

		// Buscar usuário do banco para verificar role
		// Security: Sempre verifica banco (não confia apenas no JWT)
		var user models.User
		if err := database.DB.Select("role").First(&user, userID).Error; err != nil {
			log.ErrorCtx(r.Context(), "failed to find user for admin check", "user_id", userID, "error", err)
			response.Error(w, http.StatusForbidden, "Acesso negado")
			return
		}

		// Verificar se é admin (fail secure: qualquer coisa != "admin" nega)
		if user.Role != "admin" {
			log.WarnCtx(r.Context(), "non-admin attempted admin access",
				"user_id", userID,
				"role", user.Role,
				"path", r.URL.Path,
				"method", r.Method)
			response.Error(w, http.StatusForbidden, "Acesso restrito a administradores")
			return
		}

		// Log de acesso admin bem-sucedido (auditoria)
		log.InfoCtx(r.Context(), "admin access granted",
			"user_id", userID,
			"path", r.URL.Path,
			"method", r.Method)

		next.ServeHTTP(w, r)
	})
}
