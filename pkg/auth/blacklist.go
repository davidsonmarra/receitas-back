package auth

import (
	"sync"
	"time"
)

// tokenBlacklist armazena tokens que foram invalidados (logout)
type tokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiration time
}

var blacklist *tokenBlacklist

func init() {
	blacklist = &tokenBlacklist{
		tokens: make(map[string]time.Time),
	}
	// Iniciar goroutine para limpeza automática
	go blacklist.cleanupExpired()
}

// AddToBlacklist adiciona um token à blacklist
func AddToBlacklist(token string, expiration time.Time) {
	blacklist.mu.Lock()
	defer blacklist.mu.Unlock()
	blacklist.tokens[token] = expiration
}

// IsBlacklisted verifica se um token está na blacklist
func IsBlacklisted(token string) bool {
	blacklist.mu.RLock()
	defer blacklist.mu.RUnlock()
	_, exists := blacklist.tokens[token]
	return exists
}

// cleanupExpired remove tokens expirados da blacklist periodicamente
func (b *tokenBlacklist) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		now := time.Now()
		for token, expiration := range b.tokens {
			if now.After(expiration) {
				delete(b.tokens, token)
			}
		}
		b.mu.Unlock()
	}
}

