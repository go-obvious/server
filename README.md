# Obvious Service Framework

A _**simple library**_ for quickly developing web services. Supports HTTP, HTTPS, AWS Gateway Lambda, and AWS Lambda.

The goal is simple, enable a development of Service APIs - not the scaffolding.

## How to Use


```sh
go get github.com/go-obvious/server
```

### Example Usage

**main.go**
```go
package main

import (
    "context"
    "flag"

    "github.com/go-obvious/server"
    "github.com/myapp/hello"
)

func parseFlags() (mode *string, port *uint, domain *string) {
    mode = flag.String("mode", "http", "Mode to run the server in (http or lambda)")
    port = flag.Uint("port", 8080, "Port to run the server on")
    domain = flag.String("domain", "example.com", "Domain for the server")
    flag.Parse()
    return
}

func main() {
    mode, port, domain := parseFlags()

    server.New(
        &server.Config{
            Domain: *domain,
            Port:   *port,
            Mode:   *mode,
        },
        hello.NewService("/"),
    ).Run(context.Background())
}
```

**hello.go service**
```go
import (
    "net/http"

    "github.com/go-chi/chi"

    "github.com/go-obvious/pkg/httpapi"
    "github.com/go-obvious/pkg/server"
    "github.com/go-obvious/pkg/server/api"
)

type API struct {
    api.Service
}

func NewService(base string) *API {
    a := &API{}
    a.APIName = "hello"
    a.Mounts = map[string]*chi.Mux{}
    a.Mounts[base] = a.Routes()
    return a
}

func (a *API) Register(app server.Server) error {
    if err := a.Service.Register(app); err != nil {
        return err
    }
    return nil
}

func (a *API) Routes() *chi.Mux {
    r := chi.NewRouter()
    r.Get("/", a.Handler)
    return r
}

type Response struct {
    Message string `json:"message"`
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
    httpapi.Reply(r, w, &Response{Message: "hello"}, http.StatusOK)
}
```