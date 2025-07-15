package security_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server/internal/middleware/security"
)

func TestSecurityMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		config             security.Config
		useTLS             bool
		expectedHeaders    map[string]string
		notExpectedHeaders []string
	}{
		{
			name: "Security headers enabled",
			config: security.Config{
				Enabled:    true,
				HSTSMaxAge: 31536000,
			},
			useTLS: false,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options":  "nosniff",
				"X-Frame-Options":         "DENY",
				"X-XSS-Protection":        "1; mode=block",
				"Referrer-Policy":         "strict-origin-when-cross-origin",
				"Content-Security-Policy": "default-src 'self'",
			},
			notExpectedHeaders: []string{"Strict-Transport-Security"}, // No HSTS without TLS
		},
		{
			name: "Security headers enabled with TLS",
			config: security.Config{
				Enabled:    true,
				HSTSMaxAge: 31536000,
			},
			useTLS: true,
			expectedHeaders: map[string]string{
				"X-Content-Type-Options":    "nosniff",
				"X-Frame-Options":           "DENY",
				"X-XSS-Protection":          "1; mode=block",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Content-Security-Policy":   "default-src 'self'",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			},
		},
		{
			name: "Security headers disabled",
			config: security.Config{
				Enabled:    false,
				HSTSMaxAge: 31536000,
			},
			useTLS: false,
			notExpectedHeaders: []string{
				"X-Content-Type-Options",
				"X-Frame-Options",
				"X-XSS-Protection",
				"Referrer-Policy",
				"Content-Security-Policy",
				"Strict-Transport-Security",
			},
		},
		{
			name: "Custom HSTS max age",
			config: security.Config{
				Enabled:    true,
				HSTSMaxAge: 3600, // 1 hour
			},
			useTLS: true,
			expectedHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=3600; includeSubDomains",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply security middleware
			secureHandler := security.Middleware(tt.config)(handler)

			// Create test request
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			// Add TLS connection if needed
			if tt.useTLS {
				req.TLS = &tls.ConnectionState{
					Version: tls.VersionTLS12,
				}
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			secureHandler.ServeHTTP(rr, req)

			// Check expected headers
			for header, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(header),
					"Expected header %s to be %s", header, expectedValue)
			}

			// Check headers that should not be present
			for _, header := range tt.notExpectedHeaders {
				assert.Empty(t, rr.Header().Get(header),
					"Expected header %s to not be present", header)
			}

			// Ensure the response is OK
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}
