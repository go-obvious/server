package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/go-obvious/server/internal/middleware/requestid"
)

type HTTPErrorCoder interface {
	HTTPCode() int
}

var _ HTTPErrorCoder = (*ResponseError)(nil)

type ResponseError struct {
	CallerInfo     string `json:"-"` // typically "$file:$line"
	Err            error  `json:"-"` // low-level runtime error
	HTTPStatusCode int    `json:"-"` // http response status code

	StatusText    string `json:"status"`                   // user-level status message
	AppCode       *int64 `json:"code,omitempty"`           // application-specific error code
	ErrorText     string `json:"error,omitempty"`          // application-level error message, for debugging
	CorrelationID string `json:"correlation_id,omitempty"` // correlation ID for tracing
	RequestID     string `json:"request_id,omitempty"`     // request ID for debugging
	TraceID       string `json:"trace_id,omitempty"`       // trace ID for distributed tracing
}

// NewHTTPError creates a new ResponseError with the given error and HTTP status code.
func NewHTTPError(err error, code int) error {
	return &ResponseError{
		Err:            err,
		HTTPStatusCode: code,
	}
}

func (e *ResponseError) HTTPCode() int { return e.HTTPStatusCode }

// Error returns a string representation of the ResponseError.
func (e *ResponseError) Error() string {
	switch {
	case e.CallerInfo != "" && e.Err != nil:
		return fmt.Sprintf("%s: %s", e.CallerInfo, e.Err.Error())
	case e.CallerInfo != "":
		return e.CallerInfo
	case e.Err != nil:
		return e.Err.Error()
	default:
		v, _ := json.Marshal(e.AsFields())
		return string(v)
	}
}

// AsFields returns the ResponseError as zerolog.Fields for structured logging.
func (e *ResponseError) AsFields() map[string]interface{} {
	fields := map[string]interface{}{}

	if e.CallerInfo != "" {
		fields["caller_info"] = e.CallerInfo
	}
	if e.Err != nil {
		fields["error"] = e.Err.Error() // Ensure we get the error message as a string
	}
	if e.HTTPStatusCode != 0 {
		fields["status_code"] = e.HTTPStatusCode
	}
	if e.StatusText != "" {
		fields["status_text"] = e.StatusText
	}
	if e.AppCode != nil {
		fields["app_code"] = *e.AppCode
	}
	if e.ErrorText != "" {
		fields["error_text"] = e.ErrorText
	}
	if e.CorrelationID != "" {
		fields["correlation_id"] = e.CorrelationID
	}
	if e.RequestID != "" {
		fields["request_id"] = e.RequestID
	}
	if e.TraceID != "" {
		fields["trace_id"] = e.TraceID
	}

	return fields
}

// Render sets the HTTP status code for the response and includes correlation headers.
func (e *ResponseError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)

	// Set correlation headers for tracing
	if e.CorrelationID != "" {
		w.Header().Set(requestid.CorrelationIDHeader, e.CorrelationID)
	}
	if e.RequestID != "" {
		w.Header().Set(requestid.RequestIDHeader, e.RequestID)
	}
	if e.TraceID != "" {
		w.Header().Set(requestid.TraceIDHeader, e.TraceID)
	}

	return nil
}

// ErrInvalidRequest creates a ResponseError for invalid requests.
func ErrInvalidRequest(err error) render.Renderer {
	return &ResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "invalid request",
		ErrorText:      err.Error(),
	}
}

// ErrRender creates a ResponseError for rendering errors.
func ErrRender(err error) render.Renderer {
	return &ResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "unable to process response",
		ErrorText:      err.Error(),
	}
}

// WrapRender wraps the render.Render function and handles errors.
func WrapRender(w http.ResponseWriter, r *http.Request, v render.Renderer) {
	if err := render.Render(w, r, v); err != nil {
		if rerr := render.Render(w, r, ErrRender(err)); rerr != nil {
			panic(rerr)
		}
	}
}

// NewErrNotFound creates a ResponseError for resource not found errors.
func NewErrNotFound() *ResponseError {
	return &ResponseError{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "resource not found",
	}
}

// NewErrServer creates a ResponseError for internal server errors.
func NewErrServer() *ResponseError {
	return &ResponseError{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "error processing request",
	}
}

// GetResponseError extracts a ResponseError from a given error.
func GetResponseError(err error) (re *ResponseError, ok bool) {
	ok = errors.As(err, &re)
	return re, ok
}

// HasCode checks if the given error has the specified HTTP status code.
func HasCode(err error, code int) bool {
	if re, isResponseErr := GetResponseError(err); isResponseErr {
		return re.HTTPStatusCode == code
	}
	return false
}

// WithCorrelationContext enriches a ResponseError with correlation context from the request.
func WithCorrelationContext(r *http.Request, responseErr *ResponseError) *ResponseError {
	if reqCtx := requestid.GetContext(r.Context()); reqCtx != nil {
		responseErr.CorrelationID = reqCtx.CorrelationID
		responseErr.RequestID = reqCtx.RequestID
		responseErr.TraceID = reqCtx.TraceID
	}
	return responseErr
}

// NewContextAwareError creates a new ResponseError with correlation context from the request.
func NewContextAwareError(r *http.Request, err error, code int, statusText string) *ResponseError {
	responseErr := &ResponseError{
		Err:            err,
		HTTPStatusCode: code,
		StatusText:     statusText,
		ErrorText:      err.Error(),
	}
	return WithCorrelationContext(r, responseErr)
}

// ErrInvalidRequestWithContext creates a context-aware ResponseError for invalid requests.
func ErrInvalidRequestWithContext(r *http.Request, err error) render.Renderer {
	return NewContextAwareError(r, err, http.StatusBadRequest, "invalid request")
}

// ErrRenderWithContext creates a context-aware ResponseError for rendering errors.
func ErrRenderWithContext(r *http.Request, err error) render.Renderer {
	return NewContextAwareError(r, err, http.StatusInternalServerError, "unable to process response")
}

// NewErrNotFoundWithContext creates a context-aware ResponseError for resource not found errors.
func NewErrNotFoundWithContext(r *http.Request) *ResponseError {
	responseErr := &ResponseError{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "resource not found",
	}
	return WithCorrelationContext(r, responseErr)
}

// NewErrServerWithContext creates a context-aware ResponseError for internal server errors.
func NewErrServerWithContext(r *http.Request) *ResponseError {
	responseErr := &ResponseError{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "error processing request",
	}
	return WithCorrelationContext(r, responseErr)
}

// WrapRenderWithContext wraps the render.Render function with context-aware error handling.
func WrapRenderWithContext(w http.ResponseWriter, r *http.Request, v render.Renderer) {
	if err := render.Render(w, r, v); err != nil {
		if rerr := render.Render(w, r, ErrRenderWithContext(r, err)); rerr != nil {
			panic(rerr)
		}
	}
}
