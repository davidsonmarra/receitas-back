package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON escreve uma resposta JSON com o status code especificado
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
	}
}

// Error escreve uma resposta de erro em JSON
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{
		"error": message,
	})
}
