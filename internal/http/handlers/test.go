package handlers

import (
	"log"
	"net/http"

	"github.com/davidsonmarra/receitas-app/pkg/response"
)

// TestHandler retorna uma mensagem "hello world" em JSON
func TestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição recebida: GET /test")

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "hello world",
	})
}
