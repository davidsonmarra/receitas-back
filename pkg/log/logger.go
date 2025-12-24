package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

// contextKey é o tipo usado para chaves no contexto
type contextKey string

const requestIDKey contextKey = "request_id"

// Init inicializa o logger global com a configuração fornecida
func Init(config Config) error {
	var zapConfig zap.Config

	if config.Development {
		// Desenvolvimento: logs mais verbosos e legíveis
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Produção: JSON estruturado
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Configurar nível de log
	level, err := parseLevel(config.Level)
	if err != nil {
		return err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Criar logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	logger = zapLogger.Sugar()
	return nil
}

// Sync faz flush dos buffers do logger
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// Debug loga uma mensagem de debug com campos opcionais
func Debug(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Debugw(msg, keysAndValues...)
	}
}

// Info loga uma mensagem informativa com campos opcionais
func Info(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Infow(msg, keysAndValues...)
	}
}

// Warn loga uma mensagem de aviso com campos opcionais
func Warn(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Warnw(msg, keysAndValues...)
	}
}

// Error loga uma mensagem de erro com campos opcionais
func Error(msg string, keysAndValues ...interface{}) {
	if logger != nil {
		logger.Errorw(msg, keysAndValues...)
	}
}

// DebugCtx loga uma mensagem de debug com contexto (inclui request_id se disponível)
func DebugCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if logger != nil {
		keysAndValues = appendRequestID(ctx, keysAndValues)
		logger.Debugw(msg, keysAndValues...)
	}
}

// InfoCtx loga uma mensagem informativa com contexto (inclui request_id se disponível)
func InfoCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if logger != nil {
		keysAndValues = appendRequestID(ctx, keysAndValues)
		logger.Infow(msg, keysAndValues...)
	}
}

// WarnCtx loga uma mensagem de aviso com contexto (inclui request_id se disponível)
func WarnCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if logger != nil {
		keysAndValues = appendRequestID(ctx, keysAndValues)
		logger.Warnw(msg, keysAndValues...)
	}
}

// ErrorCtx loga uma mensagem de erro com contexto (inclui request_id se disponível)
func ErrorCtx(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if logger != nil {
		keysAndValues = appendRequestID(ctx, keysAndValues)
		logger.Errorw(msg, keysAndValues...)
	}
}

// WithRequestID adiciona um request ID ao contexto
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID recupera o request ID do contexto
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// appendRequestID adiciona o request_id aos campos de log se existir no contexto
func appendRequestID(ctx context.Context, keysAndValues []interface{}) []interface{} {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		return append(keysAndValues, "request_id", requestID)
	}
	return keysAndValues
}

// parseLevel converte uma string de nível para zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	if level == "" {
		return zapcore.InfoLevel, nil // Padrão: info
	}

	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}
