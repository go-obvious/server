package request_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/internal/middleware/requestid"
	"github.com/go-obvious/server/request"
)

func TestResponseError_BasicFunctionality(t *testing.T) {
	err := errors.New("test error")
	responseErr := &request.ResponseError{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "bad request",
		ErrorText:      "invalid input",
	}

	// Test Error method
	assert.Contains(t, responseErr.Error(), "test error")

	// Test HTTPCode method
	assert.Equal(t, http.StatusBadRequest, responseErr.HTTPCode())

	// Test AsFields method
	fields := responseErr.AsFields()
	assert.Equal(t, err.Error(), fields["error"])
	assert.Equal(t, http.StatusBadRequest, fields["status_code"])
	assert.Equal(t, "bad request", fields["status_text"])
	assert.Equal(t, "invalid input", fields["error_text"])
}

func TestResponseError_WithCorrelationContext(t *testing.T) {
	// Create request with correlation context
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Create correlation context
	reqCtx := &requestid.Context{
		RequestID:     "test-request-123",
		CorrelationID: "test-correlation-456",
		TraceID:       "test-trace-789",
	}
	ctx := requestid.SaveContext(context.Background(), reqCtx)
	req = req.WithContext(ctx)

	// Create ResponseError and enrich with correlation context
	responseErr := &request.ResponseError{
		Err:            errors.New("test error"),
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "internal error",
		ErrorText:      "something went wrong",
	}

	enrichedErr := request.WithCorrelationContext(req, responseErr)

	// Verify correlation context is included
	assert.Equal(t, "test-correlation-456", enrichedErr.CorrelationID)
	assert.Equal(t, "test-request-123", enrichedErr.RequestID)
	assert.Equal(t, "test-trace-789", enrichedErr.TraceID)

	// Test AsFields includes correlation context
	fields := enrichedErr.AsFields()
	assert.Equal(t, "test-correlation-456", fields["correlation_id"])
	assert.Equal(t, "test-request-123", fields["request_id"])
	assert.Equal(t, "test-trace-789", fields["trace_id"])
}

func TestResponseError_Render(t *testing.T) {
	responseErr := &request.ResponseError{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "not found",
		CorrelationID:  "test-correlation-123",
		RequestID:      "test-request-456",
		TraceID:        "test-trace-789",
	}

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	err = responseErr.Render(rr, req)
	require.NoError(t, err)

	// Verify correlation headers are set
	assert.Equal(t, "test-correlation-123", rr.Header().Get("X-Correlation-ID"))
	assert.Equal(t, "test-request-456", rr.Header().Get("X-Request-ID"))
	assert.Equal(t, "test-trace-789", rr.Header().Get("X-Trace-ID"))
}

func TestNewContextAwareError(t *testing.T) {
	// Create request with correlation context
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	reqCtx := &requestid.Context{
		RequestID:     "test-request-123",
		CorrelationID: "test-correlation-456",
		TraceID:       "test-trace-789",
	}
	ctx := requestid.SaveContext(context.Background(), reqCtx)
	req = req.WithContext(ctx)

	// Create context-aware error
	testErr := errors.New("test error")
	responseErr := request.NewContextAwareError(req, testErr, http.StatusBadRequest, "bad request")

	// Verify error details
	assert.Equal(t, testErr, responseErr.Err)
	assert.Equal(t, http.StatusBadRequest, responseErr.HTTPStatusCode)
	assert.Equal(t, "bad request", responseErr.StatusText)
	assert.Equal(t, "test error", responseErr.ErrorText)

	// Verify correlation context
	assert.Equal(t, "test-correlation-456", responseErr.CorrelationID)
	assert.Equal(t, "test-request-123", responseErr.RequestID)
	assert.Equal(t, "test-trace-789", responseErr.TraceID)
}

func TestContextAwareErrorHelpers(t *testing.T) {
	// Create request with correlation context
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	reqCtx := &requestid.Context{
		RequestID:     "test-request-123",
		CorrelationID: "test-correlation-456",
	}
	ctx := requestid.SaveContext(context.Background(), reqCtx)
	req = req.WithContext(ctx)

	tests := []struct {
		name         string
		createError  func() interface{}
		expectedCode int
		expectedText string
	}{
		{
			name: "ErrInvalidRequestWithContext",
			createError: func() interface{} {
				return request.ErrInvalidRequestWithContext(req, errors.New("invalid input"))
			},
			expectedCode: http.StatusBadRequest,
			expectedText: "invalid request",
		},
		{
			name: "ErrRenderWithContext",
			createError: func() interface{} {
				return request.ErrRenderWithContext(req, errors.New("render failed"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedText: "unable to process response",
		},
		{
			name: "NewErrNotFoundWithContext",
			createError: func() interface{} {
				return request.NewErrNotFoundWithContext(req)
			},
			expectedCode: http.StatusNotFound,
			expectedText: "resource not found",
		},
		{
			name: "NewErrServerWithContext",
			createError: func() interface{} {
				return request.NewErrServerWithContext(req)
			},
			expectedCode: http.StatusInternalServerError,
			expectedText: "error processing request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorResult := tt.createError()
			
			// Extract ResponseError
			var responseErr *request.ResponseError
			if _, ok := errorResult.(interface{ Render(http.ResponseWriter, *http.Request) error }); ok {
				// For render.Renderer types, we need to extract the underlying ResponseError
				if re, ok := errorResult.(*request.ResponseError); ok {
					responseErr = re
				}
			} else if re, ok := errorResult.(*request.ResponseError); ok {
				responseErr = re
			}

			require.NotNil(t, responseErr)
			assert.Equal(t, tt.expectedCode, responseErr.HTTPStatusCode)
			assert.Equal(t, tt.expectedText, responseErr.StatusText)
			assert.Equal(t, "test-correlation-456", responseErr.CorrelationID)
			assert.Equal(t, "test-request-123", responseErr.RequestID)
		})
	}
}

func TestWrapRenderWithContext(t *testing.T) {
	// Create request with correlation context
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	reqCtx := &requestid.Context{
		RequestID:     "test-request-123",
		CorrelationID: "test-correlation-456",
	}
	ctx := requestid.SaveContext(context.Background(), reqCtx)
	req = req.WithContext(ctx)

	// Test successful render
	t.Run("Successful render", func(t *testing.T) {
		rr := httptest.NewRecorder()
		successRenderer := &mockRenderer{shouldFail: false}
		
		// Should not panic
		assert.NotPanics(t, func() {
			request.WrapRenderWithContext(rr, req, successRenderer)
		})
	})

	// Test failed render that triggers error handler
	t.Run("Failed render with error handler", func(t *testing.T) {
		rr := httptest.NewRecorder()
		failRenderer := &mockRenderer{shouldFail: true}
		
		// Should not panic (error should be handled)
		assert.NotPanics(t, func() {
			request.WrapRenderWithContext(rr, req, failRenderer)
		})
	})
}

func TestHasCode(t *testing.T) {
	responseErr := &request.ResponseError{
		HTTPStatusCode: http.StatusNotFound,
	}

	assert.True(t, request.HasCode(responseErr, http.StatusNotFound))
	assert.False(t, request.HasCode(responseErr, http.StatusBadRequest))
	assert.False(t, request.HasCode(errors.New("regular error"), http.StatusNotFound))
}

func TestGetResponseError(t *testing.T) {
	responseErr := &request.ResponseError{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "bad request",
	}

	// Test with ResponseError
	extracted, ok := request.GetResponseError(responseErr)
	assert.True(t, ok)
	assert.Equal(t, responseErr, extracted)

	// Test with regular error
	extracted, ok = request.GetResponseError(errors.New("regular error"))
	assert.False(t, ok)
	assert.Nil(t, extracted)
}

// mockRenderer for testing
type mockRenderer struct {
	shouldFail bool
}

func (m *mockRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	if m.shouldFail {
		return errors.New("render failed")
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	return nil
}