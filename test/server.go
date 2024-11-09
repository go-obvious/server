package test

// API Service Helper methods for testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/go-obvious/server/api"
)

func InvokeService(api api.Service, route string, request http.Request) (*http.Response, error) {
	api.Router = chi.NewRouter()
	for apiBase, routes := range api.Mounts {
		api.Router.Mount(apiBase, routes)
	}
	s := httptest.NewServer(api.Router)
	defer s.Close()

	// setup the target URL
	u, err := url.Parse(s.URL + route)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s/%s: %v", s.URL, route, err)
	}

	// Preserve the original query parameters
	q := request.URL.Query()
	for key, values := range q {
		for _, value := range values {
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()
	request.URL = u

	// invoke the endpoint
	return s.Client().Do(&request)
}
