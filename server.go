package server

import (
	"context"
	"fmt"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"

	"github.com/go-obvious/server/config"
	"github.com/go-obvious/server/internal/healthz"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
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
		addr:   fmt.Sprintf(":%d", cfg.Port),
		router: chi.NewRouter(),
		serve:  HTTPListener(),
	}

	//app.router.Use(middleware.Logger)
	app.router.Use(panic.Middleware)
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
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
	addr   string
	router *chi.Mux
	serve  ListenAndServeFunc
}

func (a *server) WithAPIs(apis ...API) Server {
	for _, api := range apis {
		if err := api.Register(a); err != nil {
			log.Fatal().Err(err).Msg("error while registering API")
		}
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
	return a
}

func (a *server) Run(ctx context.Context) {
	log.Debug().Msg("Running HTTP server")
	if err := a.serve(a.addr, a.router); err != nil {
		log.Fatal().Err(err).Msg("error while running HTTP server")
	}
}
