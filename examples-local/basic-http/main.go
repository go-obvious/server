package main

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"

	"github.com/go-obvious/server"
)

// ExampleAPI demonstrates basic HTTP server functionality
type ExampleAPI struct{}

func (api *ExampleAPI) Name() string {
	return "example-api"
}

func (api *ExampleAPI) Register(app server.Server) error {
	router := app.Router().(*chi.Mux)
	
	router.Route("/api", func(r chi.Router) {
		r.Get("/hello", api.handleHello)
		r.Get("/secure", api.handleSecure)
		r.Post("/echo", api.handleEcho)
	})
	
	return nil
}

func (api *ExampleAPI) handleHello(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Hello endpoint called")
	
	response := map[string]interface{}{
		"message": "Hello from go-obvious/server!",
		"features": []string{
			"CORS Security",
			"Security Headers", 
			"Request ID Tracking",
			"Panic Recovery",
		},
	}
	
	render.JSON(w, r, response)
}

func (api *ExampleAPI) handleSecure(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Secure endpoint called")
	
	response := map[string]interface{}{
		"message": "This endpoint demonstrates security features",
		"security_headers": []string{
			"X-Content-Type-Options: nosniff",
			"X-Frame-Options: DENY", 
			"X-XSS-Protection: 1; mode=block",
			"Referrer-Policy: strict-origin-when-cross-origin",
			"Content-Security-Policy: default-src 'self'",
		},
		"cors": "Configured origins only",
		"request_id": r.Header.Get("X-Request-ID"),
	}
	
	render.JSON(w, r, response)
}

func (api *ExampleAPI) handleEcho(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Echo endpoint called")
	
	var body map[string]interface{}
	if err := render.DecodeJSON(r.Body, &body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	response := map[string]interface{}{
		"echo": body,
		"method": r.Method,
		"headers": r.Header,
	}
	
	render.JSON(w, r, response)
}

func main() {
	log.Info().Msg("Starting basic HTTP server example")
	
	// Create server with version info
	version := &server.ServerVersion{
		Revision: "v1.0.0",
		Tag:      "basic-http-example",
		Time:     "2024-01-01T00:00:00Z",
	}
	
	// Create and configure API
	api := &ExampleAPI{}
	
	// Start server with API
	srv := server.New(version).WithAPIs(api)
	
	log.Info().Msg("Server listening on http://localhost:8080")
	log.Info().Msg("Try: curl http://localhost:8080/api/hello")
	
	srv.Run(context.Background())
}