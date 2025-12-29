package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davidsonmarra/receitas-app/pkg/log"
	"github.com/davidsonmarra/receitas-app/pkg/response"
	"github.com/go-chi/httprate"
)

// RateLimitConfig armazena as configurações de rate limiting
type RateLimitConfig struct {
	Enabled bool
	Global  int
	Read    int
	Write   int
}

// LoadRateLimitConfig carrega as configurações de rate limiting das variáveis de ambiente
func LoadRateLimitConfig() RateLimitConfig {
	config := RateLimitConfig{
		Enabled: true,
		Global:  100,
		Read:    60,
		Write:   20,
	}

	// RATE_LIMIT_ENABLED
	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		config.Enabled = enabled == "true"
	}

	// RATE_LIMIT_GLOBAL
	if global := os.Getenv("RATE_LIMIT_GLOBAL"); global != "" {
		if val, err := strconv.Atoi(global); err == nil && val > 0 {
			config.Global = val
		}
	}

	// RATE_LIMIT_READ
	if read := os.Getenv("RATE_LIMIT_READ"); read != "" {
		if val, err := strconv.Atoi(read); err == nil && val > 0 {
			config.Read = val
		}
	}

	// RATE_LIMIT_WRITE
	if write := os.Getenv("RATE_LIMIT_WRITE"); write != "" {
		if val, err := strconv.Atoi(write); err == nil && val > 0 {
			config.Write = val
		}
	}

	log.Info("rate limit config loaded",
		"enabled", config.Enabled,
		"global", config.Global,
		"read", config.Read,
		"write", config.Write,
	)

	return config
}

// getClientIP extrai o IP real do cliente considerando proxies
func getClientIP(r *http.Request) string {
	// Tentar X-Forwarded-For (usado por muitos proxies/load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Pegar o primeiro IP da lista (cliente original)
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Tentar X-Real-IP (usado por nginx e outros)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback para RemoteAddr
	// RemoteAddr pode incluir porta, remover se presente
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// rateLimitResponse envia uma resposta 429 formatada
func rateLimitResponse(w http.ResponseWriter, r *http.Request) {
	log.WarnCtx(r.Context(), "rate limit exceeded")

	response.JSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"error": map[string]string{
			"title":   "Ops, muitas requisições!",
			"message": "Você excedeu o limite de requisições. Tente novamente em alguns segundos.",
		},
	})
}

// RateLimitGlobal cria um middleware de rate limit global por IP
func RateLimitGlobal(requestsPerMinute int) func(http.Handler) http.Handler {
	log.Info("creating global rate limiter", "limit", requestsPerMinute)

	return httprate.Limit(
		requestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(rateLimitResponse),
	)
}

// RateLimitEndpoint cria um middleware de rate limit por endpoint e IP
func RateLimitEndpoint(requestsPerMinute int) func(http.Handler) http.Handler {
	return httprate.Limit(
		requestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByEndpoint),
		httprate.WithLimitHandler(rateLimitResponse),
	)
}

// RateLimitRead é um atalho para criar rate limit de leitura
func RateLimitRead(config RateLimitConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return RateLimitEndpoint(config.Read)
}

// RateLimitWrite é um atalho para criar rate limit de escrita
func RateLimitWrite(config RateLimitConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return RateLimitEndpoint(config.Write)
}
