package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/davidsonmarra/receitas-app/internal/http/handlers"
)

// Setup configura e retorna o router com todas as rotas registradas
func Setup() *chi.Mux {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Rotas
	r.Get("/test", handlers.TestHandler)

	return r
}
