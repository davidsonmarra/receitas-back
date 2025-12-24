package test

import (
	"context"
	"testing"

	"github.com/davidsonmarra/receitas-app/pkg/log"
)

func TestLoggerInit(t *testing.T) {
	tests := []struct {
		name   string
		config log.Config
		want   error
	}{
		{
			name: "development mode with debug level",
			config: log.Config{
				Level:       "debug",
				Development: true,
			},
			want: nil,
		},
		{
			name: "production mode with info level",
			config: log.Config{
				Level:       "info",
				Development: false,
			},
			want: nil,
		},
		{
			name: "warn level",
			config: log.Config{
				Level:       "warn",
				Development: false,
			},
			want: nil,
		},
		{
			name: "error level",
			config: log.Config{
				Level:       "error",
				Development: false,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := log.Init(tt.config)
			if err != tt.want {
				t.Errorf("Init() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestLoggerFunctions(t *testing.T) {
	// Inicializar logger para os testes
	config := log.Config{
		Level:       "debug",
		Development: true,
	}

	if err := log.Init(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Testar funções básicas de log (não devem causar panic)
	t.Run("Debug", func(t *testing.T) {
		log.Debug("test debug message", "key", "value")
	})

	t.Run("Info", func(t *testing.T) {
		log.Info("test info message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		log.Warn("test warn message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		log.Error("test error message", "key", "value")
	})
}

func TestLoggerWithContext(t *testing.T) {
	// Inicializar logger
	config := log.Config{
		Level:       "debug",
		Development: true,
	}

	if err := log.Init(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	ctx := context.Background()

	t.Run("context without request ID", func(t *testing.T) {
		log.DebugCtx(ctx, "test message without request ID")
		log.InfoCtx(ctx, "test message without request ID")
		log.WarnCtx(ctx, "test message without request ID")
		log.ErrorCtx(ctx, "test message without request ID")
	})

	t.Run("context with request ID", func(t *testing.T) {
		requestID := "test-request-id-123"
		ctxWithID := log.WithRequestID(ctx, requestID)

		// Verificar que o request ID foi adicionado
		retrievedID := log.GetRequestID(ctxWithID)
		if retrievedID != requestID {
			t.Errorf("GetRequestID() = %v, want %v", retrievedID, requestID)
		}

		// Testar funções de log com contexto
		log.DebugCtx(ctxWithID, "test message with request ID")
		log.InfoCtx(ctxWithID, "test message with request ID")
		log.WarnCtx(ctxWithID, "test message with request ID")
		log.ErrorCtx(ctxWithID, "test message with request ID")
	})
}

func TestRequestIDContextOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("add and retrieve request ID", func(t *testing.T) {
		requestID := "abc-123"
		ctxWithID := log.WithRequestID(ctx, requestID)

		retrievedID := log.GetRequestID(ctxWithID)
		if retrievedID != requestID {
			t.Errorf("GetRequestID() = %v, want %v", retrievedID, requestID)
		}
	})

	t.Run("retrieve from context without request ID", func(t *testing.T) {
		retrievedID := log.GetRequestID(ctx)
		if retrievedID != "" {
			t.Errorf("GetRequestID() = %v, want empty string", retrievedID)
		}
	})
}

func TestNewConfig(t *testing.T) {
	config := log.NewConfig()

	if config.Level != "info" {
		t.Errorf("NewConfig().Level = %v, want 'info'", config.Level)
	}

	if config.Development != false {
		t.Errorf("NewConfig().Development = %v, want false", config.Development)
	}
}
