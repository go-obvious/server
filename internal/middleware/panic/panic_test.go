package panic_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	middleware "github.com/go-obvious/server/internal/middleware/panic"
)

func TestMiddleware(t *testing.T) {
	// Set up zerolog to log to a writer for testing
	var logBuffer bytes.Buffer
	writer := zerolog.ConsoleWriter{
		Out: &logBuffer,
	}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()

	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectPanicLog bool
	}{
		{
			name: "No Panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectPanicLog: false,
		},
		{
			name: "With Panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectPanicLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com/foo", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := middleware.Middleware(http.HandlerFunc(tt.handler))
			// Logs are captured in the logs slice
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			logOutput := logBuffer.String()
			if tt.expectPanicLog {
				// Check if the log contains the expected panic message
				assert.Contains(t, logOutput, "Request panicked - internal server error")
			} else {
				assert.NotContains(t, logOutput, "Request panicked - internal server error")
			}
		})
	}
}

func TestMiddleware_CorrelationContext(t *testing.T) {
	// Set up zerolog to log to a writer for testing
	var logBuffer bytes.Buffer
	writer := zerolog.ConsoleWriter{
		Out: &logBuffer,
	}
	log.Logger = zerolog.New(writer).With().Timestamp().Logger()

	// Test handler that panics
	panicHandler := func(w http.ResponseWriter, r *http.Request) {
		panic("test panic with correlation")
	}

	// Create request with correlation headers
	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	assert.NoError(t, err)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")
	req.Header.Set("X-Trace-ID", "test-trace-456")

	rr := httptest.NewRecorder()

	// Use requestid middleware first, then panic middleware
	requestidMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate requestid middleware behavior
			ctx := r.Context()
			// This would normally be set by the requestid middleware
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	handler := requestidMiddleware(middleware.Middleware(http.HandlerFunc(panicHandler)))
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Verify structured error response
	var errorResp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, "Internal server error", errorResp["error"])

	// Verify log contains structured context
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Request panicked - internal server error")
	assert.Contains(t, logOutput, "test panic with correlation")
}
