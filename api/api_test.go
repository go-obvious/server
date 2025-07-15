package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/api"
)

// Mock server that implements the Server interface
type mockServer struct {
	router *chi.Mux
}

func (m *mockServer) Router() interface{} {
	return m.router
}

func TestService_Name(t *testing.T) {
	service := &api.Service{
		APIName: "test-api",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	assert.Equal(t, "test-api", service.Name())
}

func TestService_Register_Success(t *testing.T) {
	// Create a mock server with a chi router
	mainRouter := chi.NewRouter()
	server := &mockServer{router: mainRouter}

	// Create a service with mount points
	testRouter := chi.NewRouter()
	testRouter.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from test API"))
	})

	service := &api.Service{
		APIName: "test-api",
		Router:  nil,
		Mounts: map[string]*chi.Mux{
			"/api/v1": testRouter,
		},
	}

	// Register the service
	err := service.Register(server)
	require.NoError(t, err)

	// Verify the router was set
	assert.Equal(t, mainRouter, service.Router)

	// Test that the routes were mounted correctly
	req, err := http.NewRequest("GET", "/api/v1/hello", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	mainRouter.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Hello from test API", rr.Body.String())
}

func TestService_Register_MultipleMounts(t *testing.T) {
	// Create a mock server with a chi router
	mainRouter := chi.NewRouter()
	server := &mockServer{router: mainRouter}

	// Create multiple routers for different API versions
	v1Router := chi.NewRouter()
	v1Router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Users V1"))
	})

	v2Router := chi.NewRouter()
	v2Router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Users V2"))
	})

	adminRouter := chi.NewRouter()
	adminRouter.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin Stats"))
	})

	service := &api.Service{
		APIName: "multi-api",
		Router:  nil,
		Mounts: map[string]*chi.Mux{
			"/api/v1": v1Router,
			"/api/v2": v2Router,
			"/admin":  adminRouter,
		},
	}

	// Register the service
	err := service.Register(server)
	require.NoError(t, err)

	// Test all mounted routes
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users", "Users V1"},
		{"/api/v2/users", "Users V2"},
		{"/admin/stats", "Admin Stats"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			mainRouter.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expected, rr.Body.String())
		})
	}
}

func TestService_Register_EmptyMounts(t *testing.T) {
	// Create a mock server with a chi router
	mainRouter := chi.NewRouter()
	server := &mockServer{router: mainRouter}

	service := &api.Service{
		APIName: "empty-api",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	// Register the service with empty mounts
	err := service.Register(server)
	require.NoError(t, err)

	// Verify the router was set
	assert.Equal(t, mainRouter, service.Router)
}

func TestService_Register_NilMounts(t *testing.T) {
	// Create a mock server with a chi router
	mainRouter := chi.NewRouter()
	server := &mockServer{router: mainRouter}

	service := &api.Service{
		APIName: "nil-mounts-api",
		Router:  nil,
		Mounts:  nil,
	}

	// Register the service with nil mounts
	err := service.Register(server)
	require.NoError(t, err)

	// Verify the router was set
	assert.Equal(t, mainRouter, service.Router)
}

func TestService_Register_BadRouter_NotChiMux(t *testing.T) {
	// Create a mock server with a non-chi router
	server := &mockServer{router: nil}

	service := &api.Service{
		APIName: "test-api",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	// Register should fail with bad router
	err := service.Register(server)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad router")
}

func TestService_Register_BadRouter_WrongType(t *testing.T) {
	// Create a mock server that returns wrong router type
	server := &mockBadServer{router: "not a chi router"}

	service := &api.Service{
		APIName: "test-api",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	// Register should fail with bad router
	err := service.Register(server)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad router")
}

func TestOnNotImplemented(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/not-implemented", nil)
	require.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	api.OnNotImplemented(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse JSON response
	expected := `{"success":true}`
	assert.JSONEq(t, expected, rr.Body.String())
}

func TestOnNotImplemented_WithDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/not-implemented", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			api.OnNotImplemented(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			expected := `{"success":true}`
			assert.JSONEq(t, expected, rr.Body.String())
		})
	}
}

func TestService_IntegrationWithChiRouter(t *testing.T) {
	// Create a full integration test with chi router
	mainRouter := chi.NewRouter()
	server := &mockServer{router: mainRouter}

	// Create API router with middleware
	apiRouter := chi.NewRouter()
	apiRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-API-Version", "v1")
			next.ServeHTTP(w, r)
		})
	})

	// Add routes
	apiRouter.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is running"))
	})

	apiRouter.Get("/not-implemented", api.OnNotImplemented)

	service := &api.Service{
		APIName: "integration-test-api",
		Router:  nil,
		Mounts: map[string]*chi.Mux{
			"/api": apiRouter,
		},
	}

	// Register the service
	err := service.Register(server)
	require.NoError(t, err)

	// Test the status endpoint
	t.Run("Status endpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/status", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mainRouter.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "API is running", rr.Body.String())
		assert.Equal(t, "v1", rr.Header().Get("X-API-Version"))
	})

	// Test the not-implemented endpoint
	t.Run("Not implemented endpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/not-implemented", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		mainRouter.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, "v1", rr.Header().Get("X-API-Version"))

		expected := `{"success":true}`
		assert.JSONEq(t, expected, rr.Body.String())
	})
}

// Mock server that returns wrong router type
type mockBadServer struct {
	router interface{}
}

func (m *mockBadServer) Router() interface{} {
	return m.router
}

func TestService_NilRouter(t *testing.T) {
	// Test with nil router
	service := &api.Service{
		APIName: "nil-router-test",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	// This should return the router name
	assert.Equal(t, "nil-router-test", service.Name())
}

func TestService_EmptyAPIName(t *testing.T) {
	// Test with empty API name
	service := &api.Service{
		APIName: "",
		Router:  nil,
		Mounts:  make(map[string]*chi.Mux),
	}

	// This should return empty string
	assert.Equal(t, "", service.Name())
}
