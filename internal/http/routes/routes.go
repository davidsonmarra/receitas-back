package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
	customMiddleware "github.com/davidsonmarra/receitas-app/internal/http/middleware"
)

// Setup configura e retorna o router com todas as rotas registradas
func Setup() *chi.Mux {
	r := chi.NewRouter()

	// Carregar configuração de rate limiting
	rateLimitConfig := customMiddleware.LoadRateLimitConfig()

	// Middlewares globais
	r.Use(customMiddleware.SetupCORS())      // CORS - deve ser o primeiro
	r.Use(customMiddleware.SecurityHeaders)  // Security headers
	r.Use(customMiddleware.RequestID)        // Adiciona Request ID a cada requisição
	
	// Rate limit global (se habilitado)
	if rateLimitConfig.Enabled {
		r.Use(customMiddleware.RateLimitGlobal(rateLimitConfig.Global))
	}
	
	r.Use(customMiddleware.RequestSizeLimit) // Limita tamanho do body da requisição
	r.Use(middleware.Recoverer)              // Recupera de panics

	// Rotas sem rate limit específico (apenas global)
	r.Get("/health", handlers.HealthHandler) // Health check endpoint
	r.Get("/test", handlers.TestHandler)

	// Rotas de usuários
	r.Route("/users", func(r chi.Router) {
		// POST /users/register - rate limit de escrita
		r.With(customMiddleware.RateLimitWrite(rateLimitConfig)).Post("/register", handlers.Register)
		
		// POST /users/login - rate limit de escrita
		r.With(customMiddleware.RateLimitWrite(rateLimitConfig)).Post("/login", handlers.Login)
		
		// POST /users/logout - requer autenticação
		r.With(customMiddleware.RequireAuth).Post("/logout", handlers.Logout)
	})

	// Rotas de receitas com rate limiting específico
	r.Route("/recipes", func(r chi.Router) {
		// Rotas públicas (sem autenticação)
		// GET /recipes - rate limit de leitura
		r.With(customMiddleware.RateLimitRead(rateLimitConfig)).Get("/", handlers.ListRecipes)
		
		// GET /recipes/{id} - rate limit de leitura
		r.With(customMiddleware.RateLimitRead(rateLimitConfig)).Get("/{id}", handlers.GetRecipe)
		
		// Rotas protegidas (requer autenticação)
		// POST /recipes - requer auth + rate limit de escrita
		r.With(customMiddleware.RequireAuth, customMiddleware.RateLimitWrite(rateLimitConfig)).Post("/", handlers.CreateRecipe)
		
		// PUT /recipes/{id} - requer auth + rate limit de escrita
		r.With(customMiddleware.RequireAuth, customMiddleware.RateLimitWrite(rateLimitConfig)).Put("/{id}", handlers.UpdateRecipe)
		
		// DELETE /recipes/{id} - requer auth + rate limit de escrita
		r.With(customMiddleware.RequireAuth, customMiddleware.RateLimitWrite(rateLimitConfig)).Delete("/{id}", handlers.DeleteRecipe)
	})

	return r
}
