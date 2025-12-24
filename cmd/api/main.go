package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/davidsonmarra/receitas-app/internal/models"
	"github.com/davidsonmarra/receitas-app/internal/server"
	"github.com/davidsonmarra/receitas-app/pkg/database"
	"github.com/davidsonmarra/receitas-app/pkg/log"
)

func main() {
	// Inicializar logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Padrão
	}

	env := os.Getenv("ENV")
	isDevelopment := env != "production"

	logConfig := log.Config{
		Level:       logLevel,
		Development: isDevelopment,
	}

	if err := log.Init(logConfig); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() // Flush buffers antes de sair

	// Conectar ao database
	log.Info("connecting to database")
	if err := database.Connect(); err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Auto migrate (criar/atualizar tabelas)
	log.Info("running database migrations")
	if err := database.DB.AutoMigrate(&models.Recipe{}); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	log.Info("database connected successfully")

	// Configuração da porta (lê de PORT env var ou usa 8080)
	port := getPort()

	log.Info("starting API server",
		"env", env,
		"log_level", logLevel,
		"port", port,
	)

	// Cria o servidor
	srv := server.New(port)

	// Canal para capturar sinais de shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Aguardar sinal de shutdown
	<-quit
	log.Info("shutting down server gracefully...")

	// Criar contexto com timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown graceful
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}

// getPort retorna a porta configurada via PORT env var ou 8080 como padrão
func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}
	return 8080 // Padrão para desenvolvimento
}
