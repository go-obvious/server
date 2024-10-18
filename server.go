package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"

	"github.com/go-obvious/server/config"
	"github.com/go-obvious/server/internal/about"
	"github.com/go-obvious/server/internal/healthz"
	"github.com/go-obvious/server/internal/listener"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
)

type Server interface {
	Router() interface{}
	Run(ctx context.Context)
}

// Expose the Version struct
type ServerVersion = about.ServerVersion

type API interface {
	Name() string
	Register(app Server) error
}

func New(
	version *ServerVersion,
	apis ...API,
) Server {
	cfg := config.Server{}
	config.Register(&cfg)

	// This will load all configurations which have been registered
	if err := config.Load(); err != nil {
		logrus.WithError(err).Fatal("error while loading configuration")
	}

	// Registers the callers version
	about.SetVersion(version)

	app := server{
		addr:   fmt.Sprintf(":%d", cfg.Port),
		router: chi.NewRouter(),
		serve:  listener.GetListener(cfg.Mode),
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

	// Built in routes
	app.router.Mount("/about", about.Endpoint())
	app.router.Mount("/healthz", healthz.Endpoint())

	for _, api := range apis {
		if err := api.Register(&app); err != nil {
			logrus.Fatal(err)
		}
	}

	return &app
}

type server struct {
	addr   string
	router *chi.Mux
	serve  listener.ListenAndServeFunc
}

func (a *server) Router() interface{} {
	return a.router
}

func (a *server) Run(ctx context.Context) {
	logrus.Debug("Running HTTP server")
	if err := a.serve(a.addr, a.router); err != nil {
		logrus.WithError(err).Fatal("error while running HTTP server")
	}
}
