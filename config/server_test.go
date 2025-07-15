package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/config"
)

func TestServerConfig_GetCORSOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Default configuration",
			input:    "http://localhost:3000,http://localhost:8080",
			expected: []string{"http://localhost:3000", "http://localhost:8080"},
		},
		{
			name:     "Single origin",
			input:    "https://example.com",
			expected: []string{"https://example.com"},
		},
		{
			name:     "Multiple origins with spaces",
			input:    "https://app.example.com, http://localhost:3000,  https://staging.example.com",
			expected: []string{"https://app.example.com", "http://localhost:3000", "https://staging.example.com"},
		},
		{
			name:     "Empty configuration",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				CORSAllowedOrigins: tt.input,
			}

			result := cfg.GetCORSOrigins()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServerConfig_Load_CORS_Validation(t *testing.T) {
	tests := []struct {
		name        string
		origins     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid origins",
			origins:     "https://example.com,http://localhost:3000",
			expectError: false,
		},
		{
			name:        "Wildcard not allowed",
			origins:     "*",
			expectError: true,
			errorMsg:    "wildcard '*' is not allowed in CORS origins",
		},
		{
			name:        "Wildcard mixed with valid origins",
			origins:     "https://example.com,*",
			expectError: true,
			errorMsg:    "wildcard '*' is not allowed in CORS origins",
		},
		{
			name:        "Missing protocol",
			origins:     "example.com",
			expectError: true,
			errorMsg:    "CORS origins must include protocol",
		},
		{
			name:        "Mixed valid and invalid",
			origins:     "https://example.com,example.com",
			expectError: true,
			errorMsg:    "CORS origins must include protocol",
		},
		{
			name:        "Default configuration should be valid",
			origins:     "http://localhost:3000,http://localhost:8080",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("SERVER_CORS_ALLOWED_ORIGINS", tt.origins)
			defer os.Unsetenv("SERVER_CORS_ALLOWED_ORIGINS")

			cfg := &config.Server{}
			config.Register(cfg)

			err := config.Load()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerConfig_Load_DefaultValues(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("SERVER_CORS_ALLOWED_ORIGINS")
	os.Unsetenv("SERVER_MODE")
	os.Unsetenv("SERVER_DOMAIN")
	os.Unsetenv("SERVER_PORT")

	cfg := &config.Server{}
	config.Register(cfg)

	err := config.Load()
	require.NoError(t, err)

	// Check default values
	assert.Equal(t, "http", cfg.Mode)
	assert.Equal(t, "example.com", cfg.Domain)
	assert.Equal(t, uint(8080), cfg.Port)
	assert.Equal(t, "http://localhost:3000,http://localhost:8080", cfg.CORSAllowedOrigins)

	// Check parsed CORS origins
	expectedOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	assert.Equal(t, expectedOrigins, cfg.GetCORSOrigins())
}

func TestServerConfig_TLS_Configuration(t *testing.T) {
	tests := []struct {
		name               string
		tlsMinVersion      string
		expectedMinVersion uint16
	}{
		{
			name:               "TLS 1.2",
			tlsMinVersion:      "1.2",
			expectedMinVersion: 0x0303, // tls.VersionTLS12
		},
		{
			name:               "TLS 1.3",
			tlsMinVersion:      "1.3",
			expectedMinVersion: 0x0304, // tls.VersionTLS13
		},
		{
			name:               "Default fallback",
			tlsMinVersion:      "invalid",
			expectedMinVersion: 0x0303, // tls.VersionTLS12 (default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				TLSMinVersion: tt.tlsMinVersion,
			}

			minVersion := cfg.GetTLSMinVersion()
			assert.Equal(t, tt.expectedMinVersion, minVersion)

			// Test secure TLS config generation
			tlsConfig := cfg.GetSecureTLSConfig()
			assert.NotNil(t, tlsConfig)
			assert.Equal(t, tt.expectedMinVersion, tlsConfig.MinVersion)
			// Note: PreferServerCipherSuites is deprecated in Go 1.18+ and ignored
			assert.NotEmpty(t, tlsConfig.CipherSuites)
			assert.NotEmpty(t, tlsConfig.CurvePreferences)
		})
	}
}

func TestServerConfig_TLS_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Server
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid TLS 1.2 configuration",
			setupConfig: func() *config.Server {
				return &config.Server{
					TLSMinVersion:   "1.2",
					ReadTimeout:     30000000000, // 30s in nanoseconds
					WriteTimeout:    30000000000,
					IdleTimeout:     120000000000,
					SecurityHeaders: true,
					HSTSMaxAge:      31536000,
				}
			},
			expectError: false,
		},
		{
			name: "Valid TLS 1.3 configuration",
			setupConfig: func() *config.Server {
				return &config.Server{
					TLSMinVersion:   "1.3",
					ReadTimeout:     30000000000,
					WriteTimeout:    30000000000,
					IdleTimeout:     120000000000,
					SecurityHeaders: true,
					HSTSMaxAge:      31536000,
				}
			},
			expectError: false,
		},
		{
			name: "Invalid TLS version",
			setupConfig: func() *config.Server {
				return &config.Server{
					TLSMinVersion: "1.1",
				}
			},
			expectError: true,
			errorMsg:    "TLS_MIN_VERSION must be either '1.2' or '1.3'",
		},
		{
			name: "Negative timeout values",
			setupConfig: func() *config.Server {
				return &config.Server{
					TLSMinVersion: "1.2",
					ReadTimeout:   -1,
				}
			},
			expectError: true,
			errorMsg:    "timeout values must be non-negative",
		},
		{
			name: "Negative HSTS max age",
			setupConfig: func() *config.Server {
				return &config.Server{
					TLSMinVersion: "1.2",
					HSTSMaxAge:    -1,
				}
			},
			expectError: true,
			errorMsg:    "HSTS max age must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			// Set environment variables for the config
			if cfg.TLSMinVersion != "" {
				os.Setenv("SERVER_TLS_MIN_VERSION", cfg.TLSMinVersion)
				defer os.Unsetenv("SERVER_TLS_MIN_VERSION")
			}
			if cfg.ReadTimeout != 0 {
				os.Setenv("SERVER_READ_TIMEOUT", cfg.ReadTimeout.String())
				defer os.Unsetenv("SERVER_READ_TIMEOUT")
			}
			if cfg.HSTSMaxAge != 0 {
				os.Setenv("SERVER_HSTS_MAX_AGE", fmt.Sprintf("%d", cfg.HSTSMaxAge))
				defer os.Unsetenv("SERVER_HSTS_MAX_AGE")
			}

			// Create fresh config and register it
			freshCfg := &config.Server{}
			config.Register(freshCfg)

			// Test Load which calls validateTLS internally
			err := config.Load()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerConfig_RateLimit_Configuration(t *testing.T) {
	tests := []struct {
		name              string
		algorithm         string
		extractor         string
		expectedAlgorithm string
		expectedExtractor string
	}{
		{
			name:              "Valid token_bucket algorithm",
			algorithm:         "token_bucket",
			extractor:         "ip",
			expectedAlgorithm: "token_bucket",
			expectedExtractor: "ip",
		},
		{
			name:              "Valid sliding_window algorithm",
			algorithm:         "sliding_window",
			extractor:         "header",
			expectedAlgorithm: "sliding_window",
			expectedExtractor: "header",
		},
		{
			name:              "Valid fixed_window algorithm",
			algorithm:         "fixed_window",
			extractor:         "custom",
			expectedAlgorithm: "fixed_window",
			expectedExtractor: "custom",
		},
		{
			name:              "Invalid algorithm defaults to token_bucket",
			algorithm:         "invalid_algorithm",
			extractor:         "ip",
			expectedAlgorithm: "token_bucket",
			expectedExtractor: "ip",
		},
		{
			name:              "Invalid extractor defaults to ip",
			algorithm:         "token_bucket",
			extractor:         "invalid_extractor",
			expectedAlgorithm: "token_bucket",
			expectedExtractor: "ip",
		},
		{
			name:              "Case insensitive algorithm",
			algorithm:         "TOKEN_BUCKET",
			extractor:         "IP",
			expectedAlgorithm: "token_bucket",
			expectedExtractor: "ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				RateLimitAlgorithm: tt.algorithm,
				RateLimitExtractor: tt.extractor,
			}

			algorithm := cfg.GetRateLimitAlgorithm()
			extractor := cfg.GetRateLimitKeyExtractor()

			assert.Equal(t, tt.expectedAlgorithm, algorithm)
			assert.Equal(t, tt.expectedExtractor, extractor)
		})
	}
}

func TestServerConfig_Load_EnvconfigError(t *testing.T) {
	config.Reset()

	// Test envconfig.Process() error handling (currently missing coverage)
	// This tests the error path in Server.Load() at lines 44-46

	// Set an invalid duration format that will cause envconfig.Process to fail
	os.Setenv("SERVER_READ_TIMEOUT", "invalid-duration-format")
	defer os.Unsetenv("SERVER_READ_TIMEOUT")

	cfg := &config.Server{}
	config.Register(cfg)

	err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "time: invalid duration")
}

func TestServerConfig_HTTPS_CertificateValidation(t *testing.T) {
	// Test HTTPS certificate file validation (currently missing coverage)
	// This tests the certificate validation paths in validateTLS() at lines 135-146

	tests := []struct {
		name        string
		setupFiles  func() (certFile, keyFile string, cleanup func())
		expectError bool
		errorMsg    string
	}{
		{
			name: "HTTPS mode with valid certificate files",
			setupFiles: func() (string, string, func()) {
				// Create temporary certificate files
				certFile := "/tmp/test_cert.pem"
				keyFile := "/tmp/test_key.pem"

				// Create empty files (content doesn't matter for validation)
				os.WriteFile(certFile, []byte("cert content"), 0644)
				os.WriteFile(keyFile, []byte("key content"), 0644)

				cleanup := func() {
					os.Remove(certFile)
					os.Remove(keyFile)
				}

				return certFile, keyFile, cleanup
			},
			expectError: false,
		},
		{
			name: "HTTPS mode with missing certificate file",
			setupFiles: func() (string, string, func()) {
				certFile := "/tmp/nonexistent_cert.pem"
				keyFile := "/tmp/test_key.pem"

				// Create only key file
				os.WriteFile(keyFile, []byte("key content"), 0644)

				cleanup := func() {
					os.Remove(keyFile)
				}

				return certFile, keyFile, cleanup
			},
			expectError: true,
			errorMsg:    "TLS certificate file does not exist",
		},
		{
			name: "HTTPS mode with missing key file",
			setupFiles: func() (string, string, func()) {
				certFile := "/tmp/test_cert.pem"
				keyFile := "/tmp/nonexistent_key.pem"

				// Create only cert file
				os.WriteFile(certFile, []byte("cert content"), 0644)

				cleanup := func() {
					os.Remove(certFile)
				}

				return certFile, keyFile, cleanup
			},
			expectError: true,
			errorMsg:    "TLS key file does not exist",
		},
		{
			name: "HTTP mode - no certificate validation",
			setupFiles: func() (string, string, func()) {
				// Use non-existent files but HTTP mode should skip validation
				return "/tmp/nonexistent_cert.pem", "/tmp/nonexistent_key.pem", func() {}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Reset()

			certFile, keyFile, cleanup := tt.setupFiles()
			defer cleanup()

			// Set environment variables for HTTPS mode with certificate
			if tt.name == "HTTP mode - no certificate validation" {
				os.Setenv("SERVER_MODE", "http")
			} else {
				os.Setenv("SERVER_MODE", "https")
			}
			defer os.Unsetenv("SERVER_MODE")

			os.Setenv("SERVER_CERTIFICATE_CERT_FILE", certFile)
			defer os.Unsetenv("SERVER_CERTIFICATE_CERT_FILE")

			os.Setenv("SERVER_CERTIFICATE_KEY_FILE", keyFile)
			defer os.Unsetenv("SERVER_CERTIFICATE_KEY_FILE")

			// Set TLS version to pass other validations
			os.Setenv("SERVER_TLS_MIN_VERSION", "1.2")
			defer os.Unsetenv("SERVER_TLS_MIN_VERSION")

			// Create fresh config and register it
			cfg := &config.Server{}
			config.Register(cfg)

			err := config.Load()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestServerConfig_RateLimit_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Server
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid rate limiting configuration",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:   true,
					RateLimitRequests:  100,
					RateLimitWindow:    60000000000, // 1 minute in nanoseconds
					RateLimitBurst:     10,
					RateLimitAlgorithm: "token_bucket",
					RateLimitExtractor: "ip",
					RateLimitHeader:    "X-API-Key",
				}
			},
			expectError: false,
		},
		{
			name: "Rate limiting disabled - no validation",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  false,
					RateLimitRequests: -1, // Invalid but should be ignored when disabled
				}
			},
			expectError: false,
		},
		{
			name: "Invalid requests - zero",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: 0,
					RateLimitWindow:   60000000000,
				}
			},
			expectError: true,
			errorMsg:    "rate limit requests must be greater than 0",
		},
		{
			name: "Invalid requests - negative",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: -10,
					RateLimitWindow:   60000000000,
				}
			},
			expectError: true,
			errorMsg:    "rate limit requests must be greater than 0",
		},
		{
			name: "Invalid window - zero",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: 100,
					RateLimitWindow:   0,
				}
			},
			expectError: true,
			errorMsg:    "rate limit window must be greater than 0",
		},
		{
			name: "Invalid window - negative",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: 100,
					RateLimitWindow:   -1000000000,
				}
			},
			expectError: true,
			errorMsg:    "rate limit window must be greater than 0",
		},
		{
			name: "Invalid burst - negative",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: 100,
					RateLimitWindow:   60000000000,
					RateLimitBurst:    -5,
				}
			},
			expectError: true,
			errorMsg:    "rate limit burst must be non-negative",
		},
		{
			name: "Valid burst - zero",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:  true,
					RateLimitRequests: 100,
					RateLimitWindow:   60000000000,
					RateLimitBurst:    0, // Valid - uses requests as capacity
				}
			},
			expectError: false,
		},
		{
			name: "Invalid algorithm",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:   true,
					RateLimitRequests:  100,
					RateLimitWindow:    60000000000,
					RateLimitAlgorithm: "invalid_algorithm",
				}
			},
			expectError: true,
			errorMsg:    "rate limit algorithm must be one of: token_bucket, sliding_window, fixed_window",
		},
		{
			name: "Invalid key extractor",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:   true,
					RateLimitRequests:  100,
					RateLimitWindow:    60000000000,
					RateLimitExtractor: "invalid_extractor",
				}
			},
			expectError: true,
			errorMsg:    "rate limit key extractor must be one of: ip, header, custom",
		},
		{
			name: "Header extractor missing header name",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:   true,
					RateLimitRequests:  100,
					RateLimitWindow:    60000000000,
					RateLimitExtractor: "header",
					RateLimitHeader:    "", // Missing header name
				}
			},
			expectError: true,
			errorMsg:    "rate limit header name is required when using header key extractor",
		},
		{
			name: "Header extractor with valid header name",
			setupConfig: func() *config.Server {
				return &config.Server{
					RateLimitEnabled:   true,
					RateLimitRequests:  100,
					RateLimitWindow:    60000000000,
					RateLimitExtractor: "header",
					RateLimitHeader:    "X-API-Key",
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			// Set environment variables for the config
			os.Setenv("SERVER_RATE_LIMIT_ENABLED", fmt.Sprintf("%t", cfg.RateLimitEnabled))
			defer os.Unsetenv("SERVER_RATE_LIMIT_ENABLED")

			os.Setenv("SERVER_RATE_LIMIT_REQUESTS", fmt.Sprintf("%d", cfg.RateLimitRequests))
			defer os.Unsetenv("SERVER_RATE_LIMIT_REQUESTS")

			os.Setenv("SERVER_RATE_LIMIT_WINDOW", cfg.RateLimitWindow.String())
			defer os.Unsetenv("SERVER_RATE_LIMIT_WINDOW")

			if cfg.RateLimitBurst != 0 {
				os.Setenv("SERVER_RATE_LIMIT_BURST", fmt.Sprintf("%d", cfg.RateLimitBurst))
				defer os.Unsetenv("SERVER_RATE_LIMIT_BURST")
			}

			if cfg.RateLimitAlgorithm != "" {
				os.Setenv("SERVER_RATE_LIMIT_ALGORITHM", cfg.RateLimitAlgorithm)
				defer os.Unsetenv("SERVER_RATE_LIMIT_ALGORITHM")
			}

			if cfg.RateLimitExtractor != "" {
				os.Setenv("SERVER_RATE_LIMIT_KEY_EXTRACTOR", cfg.RateLimitExtractor)
				defer os.Unsetenv("SERVER_RATE_LIMIT_KEY_EXTRACTOR")
			}

			// Always set the header name even if empty to test empty case
			os.Setenv("SERVER_RATE_LIMIT_HEADER", cfg.RateLimitHeader)
			defer os.Unsetenv("SERVER_RATE_LIMIT_HEADER")

			// Create fresh config and register it
			freshCfg := &config.Server{}
			config.Register(freshCfg)

			// Test Load which calls validateRateLimit internally
			err := config.Load()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
