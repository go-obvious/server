package requestid_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/internal/middleware/requestid"
)

func TestMiddleware_CorrelationID(t *testing.T) {
	tests := []struct {
		name                    string
		correlationIDHeader     string
		traceIDHeader          string
		requestIDHeader        string
		expectedCorrelationID  string
		shouldGenerateCorrelationID bool
	}{
		{
			name:                  "With Correlation ID header",
			correlationIDHeader:   "test-correlation-id",
			expectedCorrelationID: "test-correlation-id",
		},
		{
			name:                  "With Trace ID header (fallback)",
			traceIDHeader:         "test-trace-id",
			expectedCorrelationID: "test-trace-id",
		},
		{
			name:                  "With Request ID header (fallback)",
			requestIDHeader:       "test-request-id",
			expectedCorrelationID: "test-request-id",
		},
		{
			name:                        "No headers - generate correlation ID",
			shouldGenerateCorrelationID: true,
		},
		{
			name:                  "Correlation ID takes precedence over others",
			correlationIDHeader:   "correlation-id",
			traceIDHeader:         "trace-id",
			requestIDHeader:       "request-id",
			expectedCorrelationID: "correlation-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := requestid.GetContext(r.Context())
				require.NotNil(t, ctx)

				if tt.shouldGenerateCorrelationID {
					assert.NotEmpty(t, ctx.CorrelationID)
					assert.Len(t, ctx.CorrelationID, 32) // 16 bytes hex-encoded
				} else {
					assert.Equal(t, tt.expectedCorrelationID, ctx.CorrelationID)
				}

				// Verify context accessor functions
				assert.Equal(t, ctx.CorrelationID, requestid.GetCorrelationID(r.Context()))
				assert.Equal(t, ctx.RequestID, requestid.GetRequestID(r.Context()))
				assert.Equal(t, ctx.TraceID, requestid.GetTraceID(r.Context()))
			}))

			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)

			if tt.correlationIDHeader != "" {
				req.Header.Set(requestid.CorrelationIDHeader, tt.correlationIDHeader)
			}
			if tt.traceIDHeader != "" {
				req.Header.Set(requestid.TraceIDHeader, tt.traceIDHeader)
			}
			if tt.requestIDHeader != "" {
				req.Header.Set(requestid.RequestIDHeader, tt.requestIDHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Verify response headers are set
			assert.NotEmpty(t, rr.Header().Get(requestid.CorrelationIDHeader))
			assert.NotEmpty(t, rr.Header().Get(requestid.RequestIDHeader))
		})
	}
}

func TestMiddleware_TraceID(t *testing.T) {
	tests := []struct {
		name            string
		traceIDHeader   string
		expectedTraceID string
	}{
		{
			name:            "With Trace ID header",
			traceIDHeader:   "test-trace-id",
			expectedTraceID: "test-trace-id",
		},
		{
			name:            "No Trace ID header",
			expectedTraceID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := requestid.GetContext(r.Context())
				require.NotNil(t, ctx)
				assert.Equal(t, tt.expectedTraceID, ctx.TraceID)
			}))

			req, err := http.NewRequest("GET", "/", nil)
			require.NoError(t, err)

			if tt.traceIDHeader != "" {
				req.Header.Set(requestid.TraceIDHeader, tt.traceIDHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Verify trace ID is included in response if present
			if tt.expectedTraceID != "" {
				assert.Equal(t, tt.expectedTraceID, rr.Header().Get(requestid.TraceIDHeader))
			}
		})
	}
}

func TestMiddleware_ResponseHeaders(t *testing.T) {
	handler := requestid.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	req.Header.Set(requestid.CorrelationIDHeader, "test-correlation")
	req.Header.Set(requestid.TraceIDHeader, "test-trace")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Verify all expected response headers are present
	assert.Equal(t, "test-correlation", rr.Header().Get(requestid.CorrelationIDHeader))
	assert.NotEmpty(t, rr.Header().Get(requestid.RequestIDHeader))
	assert.Equal(t, "test-trace", rr.Header().Get(requestid.TraceIDHeader))
}

func TestContext_Accessors(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() *requestid.Context
		testFunction  func(ctx *requestid.Context) string
		expectedValue string
	}{
		{
			name: "Get RequestID",
			setupContext: func() *requestid.Context {
				return &requestid.Context{RequestID: "test-request"}
			},
			testFunction: func(ctx *requestid.Context) string {
				return requestid.GetRequestID(requestid.SaveContext(context.Background(), ctx))
			},
			expectedValue: "test-request",
		},
		{
			name: "Get CorrelationID",
			setupContext: func() *requestid.Context {
				return &requestid.Context{CorrelationID: "test-correlation"}
			},
			testFunction: func(ctx *requestid.Context) string {
				return requestid.GetCorrelationID(requestid.SaveContext(context.Background(), ctx))
			},
			expectedValue: "test-correlation",
		},
		{
			name: "Get TraceID",
			setupContext: func() *requestid.Context {
				return &requestid.Context{TraceID: "test-trace"}
			},
			testFunction: func(ctx *requestid.Context) string {
				return requestid.GetTraceID(requestid.SaveContext(context.Background(), ctx))
			},
			expectedValue: "test-trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			result := tt.testFunction(ctx)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestContext_NilContext(t *testing.T) {
	// Test accessor functions with nil context
	assert.Empty(t, requestid.GetRequestID(nil))
	assert.Empty(t, requestid.GetCorrelationID(nil))
	assert.Empty(t, requestid.GetTraceID(nil))
	assert.Nil(t, requestid.GetContext(nil))
}
