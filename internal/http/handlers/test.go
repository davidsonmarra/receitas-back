package handlers

import (
	"net/http"

	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// TestHandler retorna uma mensagem "hello world" em JSON
func TestHandler(w http.ResponseWriter, r *http.Request) {
	log.DebugCtx(r.Context(), "handling test request",
		"method", r.Method,
		"path", r.URL.Path,
	)

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "hello world",
	})
}
