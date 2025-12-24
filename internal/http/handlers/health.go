package handlers

import (
	"net/http"
	"time"

	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// HealthHandler retorna o status de saúde da aplicação
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar conexão com database
	dbStatus := "disconnected"
	if err := database.Ping(); err == nil {
		dbStatus = "connected"
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"database":  dbStatus,
	})
}
