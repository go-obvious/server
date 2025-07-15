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
