package panic_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	middleware "github.com/go-obvious/server/internal/middleware/panic"
)

func TestMiddleware(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logrus.SetOutput(logger.Writer())

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

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectPanicLog {
				assert.NotEmpty(t, hook.Entries)
				assert.Contains(t, hook.LastEntry().Message, "panicked!")
			} else {
				assert.Empty(t, hook.Entries)
			}

			hook.Reset()
		})
	}
}
