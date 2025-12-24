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

	// Middlewares
	r.Use(customMiddleware.RequestID) // Adiciona Request ID a cada requisição
	r.Use(middleware.Recoverer)       // Recupera de panics

	// Rotas
	r.Get("/health", handlers.HealthHandler) // Health check endpoint
	r.Get("/test", handlers.TestHandler)

	return r
}
