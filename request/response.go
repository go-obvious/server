package request

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
)

const (
	MaxGzipSize           = 1024 * 1025 * 5
	ContentTypeJSON       = "application/json"
	ContentTypeGzip       = "gzip"
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
)

// SingleResponse simple class to make standard response objects for single element gets
type SingleResponse[DataType any] struct {
	Status Result   `json:"status"`
	Data   DataType `json:"data"`
}

// ListResponse simple class to make standard response objects for list of elements.
type ListResponse[DataType any] struct {
	Status Result     `json:"status"`
	Cursor Cursor     `json:"cursor"`
	Count  int        `json:"count"`
	Data   []DataType `json:"data"`
}

type Result struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// NewResult creates a new successful Result.
func NewResult() Result {
	return Result{
		Success: true,
	}
}

// Reply sends a JSON response with the given data and status code.
func Reply(r *http.Request, w http.ResponseWriter, data interface{}, statusCode int) {
	reply(r, w, data, statusCode, false)
}

// ReplyGzip sends a gzipped JSON response with the given data and status code.
func ReplyGzip(r *http.Request, w http.ResponseWriter, data interface{}, statusCode int, pretty bool) {
	replyCompressed(r, w, data, statusCode, pretty, true)
}

// ReplyErr sends an error response with the given error.
func ReplyErr(w http.ResponseWriter, r *http.Request, err error) {
	res := Result{Success: false}
	if err != nil {
		res.Error = err.Error()
	} else {
		res.Error = "unexpected server error"
	}

	if hec, ok := err.(HTTPErrorCoder); ok {
		reply(r, w, res, hec.HTTPCode(), false)
		return
	}
	reply(r, w, res, http.StatusInternalServerError, false)
}

// ReplyRaw sends a raw response with the given reader and status code.
func ReplyRaw(r *http.Request, w http.ResponseWriter, src io.Reader, statusCode int, contentType string) {
	if contentType != "" {
		w.Header().Set(HeaderContentType, contentType)
	}

	w.WriteHeader(statusCode)
	writeResponse(w, src)
}

// ReplyBytes sends a response with the given byte data and status code.
func ReplyBytes(r *http.Request, w http.ResponseWriter, data []byte, statusCode int, contentType string) {
	ReplyRaw(r, w, bytes.NewReader(data), statusCode, contentType)
}

// ReplyBytesGzip sends a gzipped response with the given byte data and status code.
func ReplyBytesGzip(r *http.Request, w http.ResponseWriter, data []byte, statusCode int, contentType string) {
	var gzipBuffer bytes.Buffer
	if err := compressGzip(&gzipBuffer, data); err != nil {
		writeError(w, `{"error": "Unable to encode a response"}`, http.StatusInternalServerError)
		return
	}

	if gzipBuffer.Len() > MaxGzipSize {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	w.Header().Set(HeaderContentEncoding, ContentTypeGzip)
	ReplyRaw(r, w, &gzipBuffer, statusCode, contentType)
}

// SetResponseHeaders sets the given headers on the response.
func SetResponseHeaders(w http.ResponseWriter, headers map[string]string) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
}

// Helper functions

func reply(r *http.Request, w http.ResponseWriter, data interface{}, statusCode int, pretty bool) {
	if statusCode == http.StatusNoContent || data == nil {
		w.WriteHeader(statusCode)
		return
	}

	var buffer bytes.Buffer
	if err := encodeJSON(&buffer, data, pretty); err != nil {
		writeError(w, `{"error": "Unable to encode a response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)
	writeResponse(w, &buffer)
}

func replyCompressed(r *http.Request, w http.ResponseWriter, data interface{}, statusCode int, pretty bool, gzipEnabled bool) {
	if statusCode == http.StatusNoContent || data == nil {
		w.WriteHeader(statusCode)
		return
	}

	var jsonBuffer bytes.Buffer
	if err := encodeJSON(&jsonBuffer, data, pretty); err != nil {
		writeError(w, `{"error": "Unable to encode a response"}`, http.StatusInternalServerError)
		return
	}

	if gzipEnabled {
		var gzipBuffer bytes.Buffer
		if err := compressGzip(&gzipBuffer, jsonBuffer.Bytes()); err != nil {
			writeError(w, `{"error": "Unable to encode a response"}`, http.StatusInternalServerError)
			return
		}

		if gzipBuffer.Len() > MaxGzipSize {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		w.Header().Set(HeaderContentEncoding, ContentTypeGzip)
		writeResponse(w, &gzipBuffer)
	} else {
		writeResponse(w, &jsonBuffer)
	}
}

func encodeJSON(buffer *bytes.Buffer, data interface{}, pretty bool) error {
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

func compressGzip(buffer *bytes.Buffer, data []byte) error {
	gw := gzip.NewWriter(buffer)
	if _, err := gw.Write(data); err != nil {
		return err
	}
	return gw.Close()
}

func writeResponse(w http.ResponseWriter, src io.Reader) {
	if _, err := io.Copy(w, src); err != nil {
		writeError(w, `{"error": "Unable to write a response"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
