package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"

	"github.com/go-obvious/server/config"
	"github.com/go-obvious/server/internal/healthz"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/ratelimit"
	"github.com/go-obvious/server/internal/middleware/requestid"
	"github.com/go-obvious/server/internal/middleware/security"
	"github.com/go-obvious/server/internal/version"
)

// Server represents a configurable HTTP server interface.
// It provides methods to set up the server's address, listener, middleware, APIs, and to run the server.
type Server interface {
	// Router returns the underlying router instance used by the server.
	Router() interface{}

	// WithAddress sets the server's address and returns the updated Server instance.
	WithAddress(addr string) Server

	// WithListener sets a custom listener function for the server and returns the updated Server instance.
	WithListener(l ListenAndServeFunc) Server

	// WithMiddleware adds middleware to the server and returns the updated Server instance.
	// NOTE: Middlware must be added before APIs
	WithMiddleware(m ...Middleware) Server

	// WithAPIs registers APIs with the server and returns the updated Server instance.
	WithAPIs(apis ...API) Server

	// Run starts the server and blocks until the context is canceled.
	Run(ctx context.Context)
}

// Expose the Version struct
type ServerVersion = version.ServerVersion

// Middleware abstraction
type Middleware func(next http.Handler) http.Handler

type API interface {
	Name() string
	Register(app Server) error
}

// LifecycleAPI extends the API interface with optional lifecycle hooks.
// APIs can implement this interface to receive startup and shutdown notifications.
type LifecycleAPI interface {
	API
	// Start is called after the API is registered but before the server starts accepting requests.
	// It can be used for initialization that requires the server context.
	Start(ctx context.Context) error
	// Stop is called during graceful shutdown to allow the API to clean up resources.
	// The context will have a timeout for the shutdown process.
	Stop(ctx context.Context) error
}

func New(
	ver *ServerVersion,
) Server {
	cfg := config.Server{}
	config.Register(&cfg)

	// This will load all configurations which have been registered
	if err := config.Load(); err != nil {
		log.Fatal().Err(err).Msg("error while loading configuration")
	}

	// Registers the callers version
	version.SetVersion(ver)

	app := server{
		addr:            fmt.Sprintf(":%d", cfg.Port),
		router:          chi.NewRouter(),
		serve:           HTTPListener(),
		apis:            make([]API, 0),
		shutdownTimeout: 30 * time.Second, // Default shutdown timeout
		isHTTPListener:  true,             // Default to HTTP listener
	}

	//app.router.Use(middleware.Logger)
	app.router.Use(panic.Middleware)

	// Rate limiting middleware (before other middleware to protect the entire pipeline)
	rateLimitConfig := ratelimit.Config{
		Enabled:      cfg.RateLimitEnabled,
		Requests:     cfg.RateLimitRequests,
		Window:       cfg.RateLimitWindow,
		Burst:        cfg.RateLimitBurst,
		Algorithm:    ratelimit.Algorithm(cfg.GetRateLimitAlgorithm()),
		KeyExtractor: ratelimit.KeyExtractor(cfg.GetRateLimitKeyExtractor()),
		HeaderName:   cfg.RateLimitHeader,
	}
	app.router.Use(ratelimit.Middleware(rateLimitConfig))

	// Security headers middleware
	securityConfig := security.Config{
		Enabled:    cfg.SecurityHeaders,
		HSTSMaxAge: cfg.HSTSMaxAge,
	}
	app.router.Use(security.Middleware(securityConfig))

	cors := cors.New(cors.Options{
		AllowedOrigins: cfg.GetCORSOrigins(),
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Api-Key",
			"User-Agent",
			"Referer",
			"Accept-Encoding",
			"Accept-Language",
			"Sec-Fetch-Dest",
			"Sec-Fetch-Mode",
			"Sec-Fetch-Site",
		},
		MaxAge: 0,
	})
	app.router.Use(cors.Handler)
	app.router.Use(apicaller.Middleware)
	app.router.Use(requestid.Middleware)

	return &app
}

type server struct {
	addr         string
	router       *chi.Mux
	serve        ListenAndServeFunc
	apis         []API
	shutdownTimeout time.Duration
	isHTTPListener bool // Track if using basic HTTP listener
}

func (a *server) WithAPIs(apis ...API) Server {
	for _, api := range apis {
		if err := api.Register(a); err != nil {
			log.Fatal().Err(err).Msg("error while registering API")
		}
		// Store APIs for lifecycle management
		a.apis = append(a.apis, api)
	}
	// Finally add Built in routes
	a.router.Mount("/version", version.Endpoint())
	a.router.Mount("/healthz", healthz.Endpoint())
	return a
}

func (a *server) WithMiddleware(middlewares ...Middleware) Server {
	// Add custom middleware layers
	for _, m := range middlewares {
		if m != nil {
			a.router.Use(m)
		}
	}
	return a
}

func (a *server) WithAddress(addr string) Server {
	a.addr = addr
	return a
}

func (a *server) Router() interface{} {
	return a.router
}

func (a *server) WithListener(l ListenAndServeFunc) Server {
	a.serve = l
	a.isHTTPListener = false // Custom listener, disable HTTP-specific optimizations
	return a
}

func (a *server) Run(ctx context.Context) {
	// Create a context that cancels on OS signals
	ctx, cancel := a.createShutdownContext(ctx)
	defer cancel()

	// Start lifecycle APIs
	if err := a.startAPIs(ctx); err != nil {
		log.Fatal().Err(err).Msg("error starting APIs")
	}

	// Create HTTP server for graceful shutdown support
	srv := a.createHTTPServer()
	
	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", a.addr).Msg("Starting HTTP server")
		if err := a.serveWithServer(srv); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Info().Msg("Shutdown signal received, starting graceful shutdown")
		a.gracefulShutdown(srv)
	case err := <-serverErr:
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// createShutdownContext creates a context that cancels on OS signals or parent context cancellation
func (a *server) createShutdownContext(parentCtx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parentCtx)
	
	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		select {
		case sig := <-sigChan:
			log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
			cancel()
		case <-ctx.Done():
			// Parent context was cancelled
		}
	}()
	
	return ctx, cancel
}

// startAPIs calls Start() on any APIs that implement LifecycleAPI
func (a *server) startAPIs(ctx context.Context) error {
	for _, api := range a.apis {
		if lifecycleAPI, ok := api.(LifecycleAPI); ok {
			log.Debug().Str("api", api.Name()).Msg("Starting lifecycle API")
			if err := lifecycleAPI.Start(ctx); err != nil {
				return fmt.Errorf("failed to start API %s: %w", api.Name(), err)
			}
		}
	}
	return nil
}

// createHTTPServer creates an http.Server instance for graceful shutdown
func (a *server) createHTTPServer() *http.Server {
	return &http.Server{
		Addr:    a.addr,
		Handler: a.router,
	}
}

// serveWithServer starts the server using the appropriate listener
func (a *server) serveWithServer(srv *http.Server) error {
	// For basic HTTP listener, we can use the server directly for graceful shutdown
	if a.isHTTPListener {
		return srv.ListenAndServe()
	}
	
	// For custom listeners, we fall back to the original serve function
	// Note: This won't support graceful shutdown for custom listeners
	return a.serve(a.addr, a.router)
}

// gracefulShutdown performs graceful shutdown of the server and APIs
func (a *server) gracefulShutdown(srv *http.Server) {
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server gracefully
	log.Info().Dur("timeout", a.shutdownTimeout).Msg("Shutting down HTTP server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
		// Force close if graceful shutdown failed
		if closeErr := srv.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Error force closing server")
		}
	}

	// Stop lifecycle APIs
	a.stopAPIs(shutdownCtx)
	
	log.Info().Msg("Server shutdown complete")
}

// stopAPIs calls Stop() on any APIs that implement LifecycleAPI
func (a *server) stopAPIs(ctx context.Context) {
	for _, api := range a.apis {
		if lifecycleAPI, ok := api.(LifecycleAPI); ok {
			log.Debug().Str("api", api.Name()).Msg("Stopping lifecycle API")
			if err := lifecycleAPI.Stop(ctx); err != nil {
				log.Error().Err(err).Str("api", api.Name()).Msg("Error stopping API")
			}
		}
	}
}
