package panic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/go-obvious/server/internal/middleware/requestid"
)

// ErrorResponse represents a structured error response with correlation context
type ErrorResponse struct {
	Error         string `json:"error"`
	CorrelationID string `json:"correlation_id,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	TraceID       string `json:"trace_id,omitempty"`
}

// This is another middleware that must stay on the top since
// we rely on it to convert business-logic-level panics into HTTP 500s.
func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rvr := recover()
			if rvr != nil && rvr != http.ErrAbortHandler {
				// Get correlation context
				reqCtx := requestid.GetContext(r.Context())
				correlationID := ""
				requestIDValue := ""
				traceID := ""
				
				if reqCtx != nil {
					correlationID = reqCtx.CorrelationID
					requestIDValue = reqCtx.RequestID
					traceID = reqCtx.TraceID
				}

				stack := string(debug.Stack())
				stackLines := strings.Split(stack, "\n")

				// Enhanced structured logging with correlation context
				logEvent := log.Error().
					Str("panic", fmt.Sprint(rvr)).
					Str("host", r.Host).
					Str("method", r.Method).
					Str("uri", r.RequestURI).
					Interface("url", r.URL).
					Str("remote", r.RemoteAddr).
					Strs("stack", stackLines)

				// Add correlation context to logs
				if correlationID != "" {
					logEvent = logEvent.Str("correlation_id", correlationID)
				}
				if requestIDValue != "" {
					logEvent = logEvent.Str("request_id", requestIDValue)
				}
				if traceID != "" {
					logEvent = logEvent.Str("trace_id", traceID)
				}

				// Add request headers for better debugging
				headers := make(map[string]string)
				for key, values := range r.Header {
					if len(values) > 0 {
						headers[key] = values[0] // Take first value
					}
				}
				logEvent = logEvent.Interface("headers", headers)

				logEvent.Msg("Request panicked - internal server error")

				// Create structured error response
				errorResp := ErrorResponse{
					Error:         "Internal server error",
					CorrelationID: correlationID,
					RequestID:     requestIDValue,
					TraceID:       traceID,
				}

				// Set response headers for correlation tracking
				w.Header().Set("Content-Type", "application/json")
				if correlationID != "" {
					w.Header().Set(requestid.CorrelationIDHeader, correlationID)
				}
				if requestIDValue != "" {
					w.Header().Set(requestid.RequestIDHeader, requestIDValue)
				}
				if traceID != "" {
					w.Header().Set(requestid.TraceIDHeader, traceID)
				}

				w.WriteHeader(http.StatusInternalServerError)

				// Send structured JSON error response
				if err := json.NewEncoder(w).Encode(errorResp); err != nil {
					// Fallback to plain text if JSON encoding fails
					fmt.Fprintf(w, `{"error":"Internal server error","correlation_id":"%s"}`, correlationID)
				}
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
