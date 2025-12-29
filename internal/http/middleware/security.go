package middleware

import "net/http"

// SecurityHeaders adiciona headers de segurança em todas as respostas
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// X-Frame-Options: Previne clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Previne MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Proteção XSS para browsers antigos
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security: Force HTTPS (apenas em conexões HTTPS)
		// Railway/Heroku usam X-Forwarded-Proto para indicar HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content-Security-Policy: Previne XSS e injection attacks
		// Para API, apenas permitir 'self'
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Referrer-Policy: Controla informações de referrer
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Desabilita features não necessárias
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), accelerometer=(), gyroscope=()")

		next.ServeHTTP(w, r)
	})
}
