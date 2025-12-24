package main

import (
	"log"
	"os"

	"github.com/davidsonmarra/receitas-app/internal/server"
)

func main() {
	// Configuração da porta (padrão: 8080)
	port := 8080

	// Cria e inicia o servidor
	srv := server.New(port)

	log.Println("Iniciando servidor API...")

	if err := srv.Start(); err != nil {
		log.Printf("Erro ao iniciar servidor: %v", err)
		os.Exit(1)
	}
}
