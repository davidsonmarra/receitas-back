package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/davidsonmarra/receitas-app/internal/http/routes"
	"github.com/davidsonmarra/receitas-app/pkg/log"
)

// Server representa o servidor HTTP da aplicação
type Server struct {
	Port       int
	httpServer *http.Server
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

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("server starting",
		"port", s.Port,
		"address", addr,
	)

	return s.httpServer.ListenAndServe()
}

// Shutdown realiza o shutdown graceful do servidor
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		log.Info("shutting down HTTP server")
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
