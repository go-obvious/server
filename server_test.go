package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-obvious/server"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
	"github.com/stretchr/testify/assert"
)

type mockAPI struct {
	name string
}

func (m *mockAPI) Name() string {
	return m.name
}

func (m *mockAPI) Register(app server.Server) error {
	r := app.Router().(*chi.Mux)
	r.Get("/mock", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return nil
}

func TestNew(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test custom API route
	req, _ = http.NewRequest("GET", "/mock", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRun(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		srv.Run(ctx)
	}()

	// Test if the server is running by making a request to a known route
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	srv.Router().(*chi.Mux).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewWithNilMiddleware(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	var middleware []server.Middleware

	apis := []server.API{
		&mockAPI{name: "mockAPI"},
	}

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test custom API route
	req, _ = http.NewRequest("GET", "/mock", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestNewWithNilAPI(t *testing.T) {
	version := &server.ServerVersion{
		Revision: "1.0.0",
	}

	middleware := []server.Middleware{
		apicaller.Middleware,
		panic.Middleware,
		requestid.Middleware,
	}

	var apis []server.API

	srv := server.New(version, middleware, apis...)

	assert.NotNil(t, srv)
	assert.IsType(t, &server.ServerVersion{}, version)

	router := srv.Router().(*chi.Mux)
	assert.NotNil(t, router)

	// Test built-in routes
	req, _ := http.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req, _ = http.NewRequest("GET", "/healthz", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
