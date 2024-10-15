package api

//Common API data, interfaces, helpers and handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"

	"github.com/go-obvious/server/request"
)

type Server interface {
	Router() interface{}
}

type Service struct {
	APIName string
	Router  *chi.Mux
	Mounts  map[string]*chi.Mux
}

func (a *Service) Name() string {
	return a.APIName
}

func (a *Service) Register(app Server) error {
	router, ok := app.Router().(*chi.Mux)
	if !ok || router == nil {
		return fmt.Errorf("bad router")
	}
	for apiBase, routes := range a.Mounts {
		router.Mount(apiBase, routes)
	}
	a.Router = router
	return nil
}

// Common Placeholder...
func OnNotImplemented(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("method", "api.OnNotImplemented").Trace("http.call")

	status := http.StatusOK
	result := request.Result{Success: true}

	request.Reply(r, w, result, status)
}
