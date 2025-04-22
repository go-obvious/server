package version_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/internal/version"
)

func TestEndpoint(t *testing.T) {
	version.SetVersion(&version.ServerVersion{
		Revision: "test",
		Tag:      "test",
		Time:     "test",
	})

	handler := version.Endpoint()
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"revision":"test","tag":"test","time":"test"}`, rr.Body.String())
}
