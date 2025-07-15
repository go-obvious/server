package security

import (
	"fmt"
	"net/http"
)

// Config holds security middleware configuration
type Config struct {
	Enabled    bool
	HSTSMaxAge int
}

// Middleware adds security headers to HTTP responses
func Middleware(config Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Enabled {
				// Strict Transport Security (HSTS)
				if r.TLS != nil { // Only add HSTS for HTTPS requests
					w.Header().Set("Strict-Transport-Security",
						fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge))
				}

				// Content Type Options
				w.Header().Set("X-Content-Type-Options", "nosniff")

				// Frame Options
				w.Header().Set("X-Frame-Options", "DENY")

				// XSS Protection
				w.Header().Set("X-XSS-Protection", "1; mode=block")

				// Referrer Policy
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

				// Content Security Policy (basic)
				w.Header().Set("Content-Security-Policy", "default-src 'self'")
			}

			next.ServeHTTP(w, r)
		})
	}
}
