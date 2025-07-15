package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type ctxKeyType int

const (
	CtxKey ctxKeyType = iota
)

// Common correlation ID headers
const (
	CorrelationIDHeader = "X-Correlation-ID"
	TraceIDHeader       = "X-Trace-ID"
	RequestIDHeader     = "X-Request-ID"
)

type Context struct {
	RequestID     string `json:"request_id"`
	CorrelationID string `json:"correlation_id"`
	TraceID       string `json:"trace_id,omitempty"`
}

func NewContext(r *http.Request) *Context {
	return &Context{
		RequestID:     middleware.GetReqID(r.Context()),
		CorrelationID: GetCorrelationID(r.Context()),
		TraceID:       GetTraceID(r.Context()),
	}
}

func GetContext(ctx context.Context) *Context {
	if ctx == nil {
		return nil
	}

	if thisCtx, ok := ctx.Value(CtxKey).(*Context); ok {
		return thisCtx
	}

	return nil
}

func GetCorrelationID(ctx context.Context) string {
	if reqCtx := GetContext(ctx); reqCtx != nil {
		return reqCtx.CorrelationID
	}
	return ""
}

func GetTraceID(ctx context.Context) string {
	if reqCtx := GetContext(ctx); reqCtx != nil {
		return reqCtx.TraceID
	}
	return ""
}

func GetRequestID(ctx context.Context) string {
	if reqCtx := GetContext(ctx); reqCtx != nil {
		return reqCtx.RequestID
	}
	return ""
}

func SaveContext(ctx context.Context, ref *Context) context.Context {
	return context.WithValue(ctx, CtxKey, ref)
}

// generateID creates a random hex string for correlation/trace IDs
func generateID() string {
	bytes := make([]byte, 16) // 128-bit ID
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simpler method if crypto/rand fails
		return hex.EncodeToString([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	}
	return hex.EncodeToString(bytes)
}

// extractCorrelationID attempts to extract correlation ID from various headers
func extractCorrelationID(r *http.Request) string {
	// Try correlation ID header first
	if correlationID := r.Header.Get(CorrelationIDHeader); correlationID != "" {
		return correlationID
	}
	
	// Try trace ID header
	if traceID := r.Header.Get(TraceIDHeader); traceID != "" {
		return traceID
	}
	
	// Try request ID header
	if requestID := r.Header.Get(RequestIDHeader); requestID != "" {
		return requestID
	}
	
	return ""
}

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		reqID := middleware.GetReqID(r.Context())
		if reqID == "" {
			reqID = generateID()
		}
		
		// Extract or generate correlation ID
		correlationID := extractCorrelationID(r)
		if correlationID == "" {
			correlationID = generateID()
		}
		
		// Extract trace ID if present
		traceID := r.Header.Get(TraceIDHeader)
		
		// Create request context
		reqCtx := &Context{
			RequestID:     reqID,
			CorrelationID: correlationID,
			TraceID:       traceID,
		}
		
		// Set response headers for correlation tracking
		w.Header().Set(CorrelationIDHeader, correlationID)
		w.Header().Set(RequestIDHeader, reqID)
		if traceID != "" {
			w.Header().Set(TraceIDHeader, traceID)
		}
		
		// Save context and continue
		ctx := SaveContext(r.Context(), reqCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return middleware.RequestID(http.HandlerFunc(fn))
}
