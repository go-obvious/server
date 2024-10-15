package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"

	"github.com/go-obvious/server/internal/about"
	"github.com/go-obvious/server/internal/healthz"
	"github.com/go-obvious/server/internal/listener"
	"github.com/go-obvious/server/internal/middleware/apicaller"
	"github.com/go-obvious/server/internal/middleware/panic"
	"github.com/go-obvious/server/internal/middleware/requestid"
)

type Config struct {
	Mode   string `envconfig:"SERVICE_MODE" default:"http"`
	Domain string `envconfig:"SERVICE_DOMAIN" default:""`
	Port   uint   `envconfig:"SERVICE_PORT" default:"4444"`
	*Certificate
}

type Certificate struct {
	Cert string `envconfig:"CERTIFICATE_CERT" default:""`
	Key  string `envconfig:"CERTIFICATE_KEY" default:""`
}

type Server interface {
	Config() interface{}
	Router() interface{}
	Run(ctx context.Context)
	Shutdown()
}

type ServerVersion = about.ServerVersion

type API interface {
	Name() string
	Register(app Server) error
}

func New(
	cfg *Config,
	version *ServerVersion,
	apis ...API,
) Server {
	about.SetVersion(version)

	app := server{
		cfg:    cfg,
		addr:   fmt.Sprintf(":%d", cfg.Port),
		router: chi.NewRouter(),
		serve:  listener.GetListener(cfg.Mode),
	}

	app.router.Use(middleware.Logger)
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
	cfg    *Config
	addr   string
	router *chi.Mux
	serve  listener.ListenAndServeFunc
}

func (a *server) Config() interface{} {
	return a.cfg
}

func (a *server) Router() interface{} {
	return a.router
}

func (a *server) Run(ctx context.Context) {
	logrus.Info("Running HTTP server")
	if err := a.serve(a.addr, a.router); err != nil {
		logrus.WithError(err).Fatal("error while running HTTP server")
	}
}

// Shutdown implements Server.
func (a *server) Shutdown() {
	logrus.Fatal("unimplemented")
}
