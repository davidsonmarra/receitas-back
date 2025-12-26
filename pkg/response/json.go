package response

import (
	"encoding/json"
	"net/http"

	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/pagination"
)

// JSON escreve uma resposta JSON com o status code especificado
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("failed to encode JSON response", "error", err)
	}
}

// Error escreve uma resposta de erro em JSON
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{
		"error": message,
	})
}

// ValidationError escreve uma resposta de erro de validação no formato amigável
func ValidationError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, map[string]interface{}{
		"error": map[string]string{
			"title":   "Ops, algo deu errado!",
			"message": message,
		},
	})
}

// Paginated escreve uma resposta JSON paginada
func Paginated(w http.ResponseWriter, statusCode int, data interface{}, params pagination.Params, total int64) {
	response := pagination.BuildResponse(data, params, total)
	JSON(w, statusCode, response)
}
