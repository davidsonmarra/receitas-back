package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/cors"
)

// SetupCORS configura o middleware de CORS baseado no ambiente
func SetupCORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   getAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300, // 5 minutos
	})
}

// getAllowedOrigins retorna as origens permitidas baseado no ambiente
func getAllowedOrigins() []string {
	env := os.Getenv("ENV")

	// Em produção, usar origens específicas ou variável de ambiente
	if env == "production" {
		// Verificar se há CORS_ORIGINS definido
		corsOrigins := os.Getenv("CORS_ORIGINS")
		if corsOrigins != "" {
			// Exemplo: CORS_ORIGINS="https://app.com,https://admin.app.com"
			origins := strings.Split(corsOrigins, ",")
			// Limpar espaços em branco
			for i, origin := range origins {
				origins[i] = strings.TrimSpace(origin)
			}
			return origins
		}

		// Default para produção: permitir qualquer origem
		// Nota: Para React Native nativo, CORS não se aplica
		// Para web, configure CORS_ORIGINS com os domínios específicos
		return []string{"https://*"}
	}

	// Em desenvolvimento, permitir localhost em qualquer porta
	return []string{
		"http://localhost:*",
		"http://127.0.0.1:*",
		"http://[::1]:*",
	}
}

