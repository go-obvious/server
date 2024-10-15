package requestid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/middleware"
	"github.com/go-obvious/server/internal/middleware/requestid"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		requestID     string
		expectedReqID string
	}{
		{
			name:          "With Request ID",
			requestID:     "test-request-id",
			expectedReqID: "test-request-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := requestid.GetContext(r.Context())
				if ctx == nil {
					t.Fatal("Context is nil")
				}
				if ctx.RequestID != tt.expectedReqID {
					t.Errorf("Unexpected Request ID. Expected: %s, Got: %s", tt.expectedReqID, ctx.RequestID)
				}
			}))

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.requestID != "" {
				req.Header.Set(middleware.RequestIDHeader, tt.requestID)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		})
	}
}
