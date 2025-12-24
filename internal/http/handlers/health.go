package handlers

import (
	"net/http"
	"time"

	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// HealthHandler retorna o status de saúde da aplicação
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}
