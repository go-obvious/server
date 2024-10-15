package apicaller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server/internal/middleware/apicaller"
)

func TestMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		apiCtx := apicaller.GetContext(ctx)
		assert.NotNil(t, apiCtx)
		assert.Equal(t, "test-agent", apiCtx.UserAgent)
		assert.Equal(t, "v1", apiCtx.APIVersion)
		w.WriteHeader(http.StatusOK)
	})

	middleware := apicaller.Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set(apicaller.APIVersionHdr, "v1")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetContext(t *testing.T) {
	ctx := context.Background()
	apiCtx := &apicaller.Context{
		UserAgent:  "test-agent",
		APIVersion: "v1",
	}
	ctx = apicaller.SaveContext(ctx, apiCtx)

	retrievedCtx := apicaller.GetContext(ctx)
	assert.NotNil(t, retrievedCtx)
	assert.Equal(t, apiCtx, retrievedCtx)
}

func TestSaveContext(t *testing.T) {
	ctx := context.Background()
	apiCtx := &apicaller.Context{
		UserAgent:  "test-agent",
		APIVersion: "v1",
	}

	newCtx := apicaller.SaveContext(ctx, apiCtx)
	retrievedCtx := apicaller.GetContext(newCtx)
	assert.NotNil(t, retrievedCtx)
	assert.Equal(t, apiCtx, retrievedCtx)
}
