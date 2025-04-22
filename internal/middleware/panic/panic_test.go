package panic_test

import (
	"bytes"
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
				assert.Contains(t, logOutput, "panicked!")
			} else {
				assert.NotContains(t, logOutput, "panicked!")
			}
		})
	}
}
