package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
)

type HTTPErrorCoder interface {
	HTTPCode() int
}

var _ HTTPErrorCoder = (*ResponseError)(nil)

type ResponseError struct {
	CallerInfo     string `json:"-"` // typically "$file:$line"
	Err            error  `json:"-"` // low-level runtime error
	HTTPStatusCode int    `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    *int64 `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
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

// AsFields returns the ResponseError as logrus.Fields for structured logging.
func (e *ResponseError) AsFields() logrus.Fields {
	fields := logrus.Fields{}

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

	return fields
}

// Render sets the HTTP status code for the response.
func (e *ResponseError) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
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
