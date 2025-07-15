package config

import (
	"crypto/tls"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Server struct {
	Mode               string        `envconfig:"SERVER_MODE" default:"http"`
	Domain             string        `envconfig:"SERVER_DOMAIN" default:"example.com"`
	Port               uint          `envconfig:"SERVER_PORT" default:"8080"`
	CORSAllowedOrigins string        `envconfig:"SERVER_CORS_ALLOWED_ORIGINS" default:"http://localhost:3000,http://localhost:8080"`
	ReadTimeout        time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout       time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout        time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"120s"`
	TLSMinVersion      string        `envconfig:"SERVER_TLS_MIN_VERSION" default:"1.2"`
	SecurityHeaders    bool          `envconfig:"SERVER_SECURITY_HEADERS_ENABLED" default:"true"`
	HSTSMaxAge         int           `envconfig:"SERVER_HSTS_MAX_AGE" default:"31536000"`

	// Rate Limiting Configuration
	RateLimitEnabled   bool          `envconfig:"SERVER_RATE_LIMIT_ENABLED" default:"false"`
	RateLimitRequests  int           `envconfig:"SERVER_RATE_LIMIT_REQUESTS" default:"100"`
	RateLimitWindow    time.Duration `envconfig:"SERVER_RATE_LIMIT_WINDOW" default:"1m"`
	RateLimitBurst     int           `envconfig:"SERVER_RATE_LIMIT_BURST" default:"10"`
	RateLimitAlgorithm string        `envconfig:"SERVER_RATE_LIMIT_ALGORITHM" default:"token_bucket"`
	RateLimitExtractor string        `envconfig:"SERVER_RATE_LIMIT_KEY_EXTRACTOR" default:"ip"`
	RateLimitHeader    string        `envconfig:"SERVER_RATE_LIMIT_HEADER" default:"X-API-Key"`

	*Certificate
}

type Certificate struct {
	CertFile string `envconfig:"SERVER_CERTIFICATE_CERT_FILE"`
	KeyFile  string `envconfig:"SERVER_CERTIFICATE_KEY_FILE"`
	CAFile   string `envconfig:"SERVER_CERTIFICATE_CA_FILE"`
}

func (c *Server) Load() error {
	if err := envconfig.Process("server", c); err != nil {
		return err
	}

	// Validate CORS configuration
	if err := c.validateCORS(); err != nil {
		return err
	}

	// Validate TLS configuration
	if err := c.validateTLS(); err != nil {
		return err
	}

	// Validate rate limiting configuration
	return c.validateRateLimit()
}

// GetCORSOrigins returns the CORS allowed origins as a slice
func (c *Server) GetCORSOrigins() []string {
	if c.CORSAllowedOrigins == "" {
		return []string{}
	}

	origins := strings.Split(c.CORSAllowedOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

// validateCORS validates the CORS configuration
func (c *Server) validateCORS() error {
	origins := c.GetCORSOrigins()

	for _, origin := range origins {
		if origin == "*" {
			return errors.New("wildcard '*' is not allowed in CORS origins for security reasons; use specific domains instead")
		}

		if origin != "" && !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
			return errors.New("CORS origins must include protocol (http:// or https://)")
		}
	}

	return nil
}

// GetTLSMinVersion returns the TLS minimum version as a uint16
func (c *Server) GetTLSMinVersion() uint16 {
	switch c.TLSMinVersion {
	case "1.3":
		return tls.VersionTLS13
	case "1.2":
		return tls.VersionTLS12
	default:
		return tls.VersionTLS12 // Safe default
	}
}

// GetSecureTLSConfig returns a secure TLS configuration
func (c *Server) GetSecureTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: c.GetTLSMinVersion(),
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites (Go will automatically use these for TLS 1.3)
			// TLS 1.2 secure cipher suites
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}
}

// validateTLS validates the TLS configuration
func (c *Server) validateTLS() error {
	// Validate TLS min version
	if c.TLSMinVersion != "1.2" && c.TLSMinVersion != "1.3" {
		return errors.New("TLS_MIN_VERSION must be either '1.2' or '1.3'")
	}

	// If HTTPS mode is enabled, validate certificate files
	if c.Mode == "https" && c.Certificate != nil {
		if c.CertFile != "" {
			if _, err := os.Stat(c.CertFile); os.IsNotExist(err) {
				return errors.New("TLS certificate file does not exist: " + c.CertFile)
			}
		}
		if c.KeyFile != "" {
			if _, err := os.Stat(c.KeyFile); os.IsNotExist(err) {
				return errors.New("TLS key file does not exist: " + c.KeyFile)
			}
		}
	}

	// Validate timeout values
	if c.ReadTimeout < 0 || c.WriteTimeout < 0 || c.IdleTimeout < 0 {
		return errors.New("timeout values must be non-negative")
	}

	// Validate HSTS max age
	if c.HSTSMaxAge < 0 {
		return errors.New("HSTS max age must be non-negative")
	}

	return nil
}

// GetRateLimitAlgorithm returns the rate limiting algorithm
func (c *Server) GetRateLimitAlgorithm() string {
	algorithm := strings.ToLower(c.RateLimitAlgorithm)
	switch algorithm {
	case "token_bucket", "sliding_window", "fixed_window":
		return algorithm
	default:
		return "token_bucket" // Safe default
	}
}

// GetRateLimitKeyExtractor returns the rate limiting key extractor
func (c *Server) GetRateLimitKeyExtractor() string {
	extractor := strings.ToLower(c.RateLimitExtractor)
	switch extractor {
	case "ip", "header", "custom":
		return extractor
	default:
		return "ip" // Safe default
	}
}

// validateRateLimit validates the rate limiting configuration
func (c *Server) validateRateLimit() error {
	if !c.RateLimitEnabled {
		return nil // Skip validation if rate limiting is disabled
	}

	// Validate requests per window
	if c.RateLimitRequests <= 0 {
		return errors.New("rate limit requests must be greater than 0")
	}

	// Validate window duration
	if c.RateLimitWindow <= 0 {
		return errors.New("rate limit window must be greater than 0")
	}

	// Validate burst size
	if c.RateLimitBurst < 0 {
		return errors.New("rate limit burst must be non-negative")
	}

	// Validate algorithm
	algorithm := strings.ToLower(c.RateLimitAlgorithm)
	if algorithm != "token_bucket" && algorithm != "sliding_window" && algorithm != "fixed_window" {
		return errors.New("rate limit algorithm must be one of: token_bucket, sliding_window, fixed_window")
	}

	// Validate key extractor
	extractor := strings.ToLower(c.RateLimitExtractor)
	if extractor != "ip" && extractor != "header" && extractor != "custom" {
		return errors.New("rate limit key extractor must be one of: ip, header, custom")
	}

	// Validate header name if using header extractor
	if extractor == "header" && c.RateLimitHeader == "" {
		return errors.New("rate limit header name is required when using header key extractor")
	}

	return nil
}
