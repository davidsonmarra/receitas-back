package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davidsonmarra/receitas-app/internal/http/routes"
)

// Server representa o servidor HTTP da aplicação
type Server struct {
	Port int
}

// New cria uma nova instância do servidor
func New(port int) *Server {
	return &Server{
		Port: port,
	}
}

// Start inicia o servidor HTTP
func (s *Server) Start() error {
	router := routes.Setup()

	addr := fmt.Sprintf(":%d", s.Port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Servidor iniciado na porta %d", s.Port)
	log.Printf("Acesse: http://localhost%s/test", addr)

	return srv.ListenAndServe()
}
