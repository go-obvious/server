package about_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/internal/about"
)

func TestEndpoint(t *testing.T) {
	about.SetVersion(&about.ServerVersion{
		Revision: "test",
		Tag:      "test",
		Time:     "test",
	})

	handler := about.Endpoint()
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"revision":"test","tag":"test","time":"test"}`, rr.Body.String())
}
