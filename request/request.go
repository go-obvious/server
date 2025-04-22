package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	MaxBodySize = 1048576 // 1MB
)

// Param returns the URL parameter from the request.
func Param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// QS returns the query string parameter from the request.
func QS(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// QSAll returns all values for the query string parameter from the request.
func QSAll(r *http.Request, name string) []string {
	return r.URL.Query()[name]
}

// QSDefault returns the query string parameter or a default value if not present.
func QSDefault(r *http.Request, name string, defaultValue string) string {
	if val := r.URL.Query().Get(name); val != "" {
		return val
	}
	return defaultValue
}

// GetBody deserializes the request body into the provided record or returns an error.
func GetBody(w http.ResponseWriter, r *http.Request, record interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodySize)
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(record); err != nil {
		return handleJSONDecodeError(err)
	}
	return nil
}

// handleJSONDecodeError handles JSON decoding errors and returns a formatted error message.
func handleJSONDecodeError(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError

	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf("request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("request body contains badly-formed JSON")
	case errors.As(err, &unmarshalTypeError):
		return fmt.Errorf("request body contains an invalid value for the %q field (at position %d)",
			unmarshalTypeError.Field, unmarshalTypeError.Offset)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return fmt.Errorf("request body contains unknown field %s", fieldName)
	case errors.Is(err, io.EOF):
		return errors.New("request body must not be empty")
	case err.Error() == "http: request body too large":
		return errors.New("request body must not be larger than 1MB")
	default:
		return err
	}
}
