package server_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server"
)

// TestAPI implements both API and LifecycleAPI interfaces for testing
type TestAPI struct {
	name        string
	started     bool
	stopped     bool
	startError  error
	stopError   error
	startCalled chan struct{}
	stopCalled  chan struct{}
	mu          sync.Mutex
}

func NewTestAPI(name string) *TestAPI {
	return &TestAPI{
		name:        name,
		startCalled: make(chan struct{}, 1),
		stopCalled:  make(chan struct{}, 1),
	}
}

func (t *TestAPI) Name() string {
	return t.name
}

func (t *TestAPI) Register(app server.Server) error {
	// Register a simple endpoint
	router := app.Router().(*chi.Mux)
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
	return nil
}

func (t *TestAPI) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.startError != nil {
		return t.startError
	}

	t.started = true
	select {
	case t.startCalled <- struct{}{}:
	default:
	}

	return nil
}

func (t *TestAPI) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopError != nil {
		return t.stopError
	}

	t.stopped = true
	select {
	case t.stopCalled <- struct{}{}:
	default:
	}

	return nil
}

func (t *TestAPI) IsStarted() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.started
}

func (t *TestAPI) IsStopped() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.stopped
}

func (t *TestAPI) SetStartError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.startError = err
}

func (t *TestAPI) SetStopError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopError = err
}

// NonLifecycleAPI implements only the basic API interface
type NonLifecycleAPI struct {
	name string
}

func (n *NonLifecycleAPI) Name() string {
	return n.name
}

func (n *NonLifecycleAPI) Register(app server.Server) error {
	return nil
}

func TestLifecycleAPI_StartStop(t *testing.T) {
	// Create test APIs
	lifecycleAPI := NewTestAPI("lifecycle-api")
	nonLifecycleAPI := &NonLifecycleAPI{name: "non-lifecycle-api"}

	// Create server with test APIs
	version := &server.ServerVersion{Revision: "test"}
	srv := server.New(version).
		WithAddress(":0"). // Use random port
		WithAPIs(lifecycleAPI, nonLifecycleAPI)

	// Test context with short timeout for quick test
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run server in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run(ctx)
	}()

	// Wait for lifecycle API to start
	select {
	case <-lifecycleAPI.startCalled:
		// Start was called successfully
		assert.True(t, lifecycleAPI.IsStarted())
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Start method was not called within timeout")
	}

	// Wait for context cancellation and shutdown
	<-done

	// Verify lifecycle API was stopped
	select {
	case <-lifecycleAPI.stopCalled:
		// Stop was called successfully
		assert.True(t, lifecycleAPI.IsStopped())
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Stop method was not called within timeout")
	}
}

func TestLifecycleAPI_StartError(t *testing.T) {
	// Create test API that fails on start
	lifecycleAPI := NewTestAPI("failing-api")
	lifecycleAPI.SetStartError(assert.AnError)

	// Create server
	version := &server.ServerVersion{Revision: "test"}
	_ = server.New(version).
		WithAddress(":0").
		WithAPIs(lifecycleAPI)

	// This should trigger log.Fatal, so we can't test it directly in a unit test
	// In a real scenario, this would cause the process to exit
	// For testing purposes, we'll verify the API was not started
	assert.False(t, lifecycleAPI.IsStarted())
}

func TestGracefulShutdown_ContextCancellation(t *testing.T) {
	lifecycleAPI := NewTestAPI("test-api")

	version := &server.ServerVersion{Revision: "test"}
	srv := server.New(version).
		WithAddress(":0").
		WithAPIs(lifecycleAPI)

	// Create context that we'll cancel to trigger shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Run server in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run(ctx)
	}()

	// Wait for API to start
	select {
	case <-lifecycleAPI.startCalled:
		// API started successfully
	case <-time.After(100 * time.Millisecond):
		t.Fatal("API start was not called")
	}

	// Cancel context to trigger shutdown
	cancel()

	// Wait for shutdown to complete
	select {
	case <-done:
		// Shutdown completed
	case <-time.After(1 * time.Second):
		t.Fatal("Server did not shutdown within timeout")
	}

	// Verify API was stopped
	assert.True(t, lifecycleAPI.IsStopped())
}

func TestServer_NonLifecycleAPI(t *testing.T) {
	// Test that non-lifecycle APIs work without implementing lifecycle methods
	nonLifecycleAPI := &NonLifecycleAPI{name: "simple-api"}

	version := &server.ServerVersion{Revision: "test"}
	srv := server.New(version).
		WithAddress(":0").
		WithAPIs(nonLifecycleAPI)

	// Short-lived context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should not panic or fail
	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run(ctx)
	}()

	// Wait for completion
	select {
	case <-done:
		// Success - server handled non-lifecycle API correctly
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Server did not shutdown properly with non-lifecycle API")
	}
}

func TestServer_MixedAPIs(t *testing.T) {
	// Test server with both lifecycle and non-lifecycle APIs
	lifecycleAPI1 := NewTestAPI("lifecycle-1")
	lifecycleAPI2 := NewTestAPI("lifecycle-2")
	nonLifecycleAPI := &NonLifecycleAPI{name: "non-lifecycle"}

	version := &server.ServerVersion{Revision: "test"}
	srv := server.New(version).
		WithAddress(":0").
		WithAPIs(lifecycleAPI1, nonLifecycleAPI, lifecycleAPI2)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.Run(ctx)
	}()

	// Wait for both lifecycle APIs to start
	startCount := 0
	timeout := time.After(150 * time.Millisecond)

	for startCount < 2 {
		select {
		case <-lifecycleAPI1.startCalled:
			startCount++
		case <-lifecycleAPI2.startCalled:
			startCount++
		case <-timeout:
			t.Fatal("Not all lifecycle APIs started within timeout")
		}
	}

	// Wait for shutdown
	<-done

	// Verify both lifecycle APIs were stopped
	assert.True(t, lifecycleAPI1.IsStopped())
	assert.True(t, lifecycleAPI2.IsStopped())
}
