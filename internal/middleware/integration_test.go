package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	panicmw "github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
	"github.com/go-obvious/server/request"
)

// TestErrorContextPreservation tests that correlation context flows through
// the complete middleware chain and is preserved in error responses
func TestErrorContextPreservation(t *testing.T) {
	tests := []struct {
		name                  string
		correlationIDHeader   string
		traceIDHeader        string
		handlerBehavior      string
		expectedStatus       int
		expectCorrelationID  bool
		expectPanicRecovery  bool
	}{
		{
			name:                "Normal request with correlation context",
			correlationIDHeader: "test-correlation-123",
			traceIDHeader:       "test-trace-456",
			handlerBehavior:     "success",
			expectedStatus:      http.StatusOK,
			expectCorrelationID: true,
			expectPanicRecovery: false,
		},
		{
			name:                "Application error with correlation context",
			correlationIDHeader: "test-correlation-789",
			traceIDHeader:       "test-trace-012",
			handlerBehavior:     "error",
			expectedStatus:      http.StatusBadRequest,
			expectCorrelationID: true,
			expectPanicRecovery: false,
		},
		{
			name:                "Panic with correlation context",
			correlationIDHeader: "test-correlation-panic",
			traceIDHeader:       "test-trace-panic",
			handlerBehavior:     "panic",
			expectedStatus:      http.StatusInternalServerError,
			expectCorrelationID: true,
			expectPanicRecovery: true,
		},
		{
			name:                "Generated correlation ID on panic",
			handlerBehavior:     "panic",
			expectedStatus:      http.StatusInternalServerError,
			expectCorrelationID: true,
			expectPanicRecovery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler that simulates different behaviors
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch tt.handlerBehavior {
				case "success":
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]string{
						"status": "success",
					})
				case "error":
					// Simulate application error with correlation context
					err := request.ErrInvalidRequestWithContext(r, 
						&testError{message: "invalid input"})
					request.WrapRenderWithContext(w, r, err)
				case "panic":
					panic("test panic for integration testing")
				}
			})

			// Build middleware chain: requestid -> panic -> handler
			middlewareChain := requestid.Middleware(
				panicmw.Middleware(testHandler),
			)

			// Create request with correlation headers
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)

			if tt.correlationIDHeader != "" {
				req.Header.Set("X-Correlation-ID", tt.correlationIDHeader)
			}
			if tt.traceIDHeader != "" {
				req.Header.Set("X-Trace-ID", tt.traceIDHeader)
			}

			// Execute request
			rr := httptest.NewRecorder()
			middlewareChain.ServeHTTP(rr, req)

			// Verify response status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Verify correlation headers are present in response
			if tt.expectCorrelationID {
				correlationID := rr.Header().Get("X-Correlation-ID")
				assert.NotEmpty(t, correlationID)
				
				if tt.correlationIDHeader != "" {
					assert.Equal(t, tt.correlationIDHeader, correlationID)
				}

				requestID := rr.Header().Get("X-Request-ID")
				assert.NotEmpty(t, requestID)

				if tt.traceIDHeader != "" {
					assert.Equal(t, tt.traceIDHeader, rr.Header().Get("X-Trace-ID"))
				}
			}

			// Verify JSON error response structure
			if tt.expectedStatus != http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				
				var errorResp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
				require.NoError(t, err)

				// Verify error response contains correlation context
				if tt.expectCorrelationID {
					if tt.correlationIDHeader != "" {
						assert.Equal(t, tt.correlationIDHeader, errorResp["correlation_id"])
					} else {
						assert.NotEmpty(t, errorResp["correlation_id"])
					}
					assert.NotEmpty(t, errorResp["request_id"])
				}

				// Verify error message structure
				assert.NotEmpty(t, errorResp["error"])
				
				if tt.expectPanicRecovery {
					assert.Equal(t, "Internal server error", errorResp["error"])
				}
			}
		})
	}
}

// TestMiddlewareOrder tests that middleware order doesn't affect correlation context
func TestMiddlewareOrder(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Test different middleware orders
	middlewareOrders := []struct {
		name    string
		handler http.Handler
	}{
		{
			name:    "requestid -> panic",
			handler: requestid.Middleware(panicmw.Middleware(panicHandler)),
		},
		{
			name:    "panic -> requestid",
			handler: panicmw.Middleware(requestid.Middleware(panicHandler)),
		},
	}

	for _, order := range middlewareOrders {
		t.Run(order.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			req.Header.Set("X-Correlation-ID", "test-order-123")

			rr := httptest.NewRecorder()
			order.handler.ServeHTTP(rr, req)

			// Both orders should result in proper error handling
			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			
			// Should have correlation ID in response headers
			assert.NotEmpty(t, rr.Header().Get("X-Correlation-ID"))
		})
	}
}

// testError implements error interface for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}