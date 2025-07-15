# Obvious Service Framework

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE-OF-CONDUCT.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
![GitHub release](https://img.shields.io/github/release/go-obvious/server.svg)

A _**simple library**_ for quickly developing **secure web services**. Supports HTTP, HTTPS, AWS Gateway Lambda, and AWS Lambda with enterprise-grade security built-in.

## ✨ What Makes It Obvious

**🔒 Security First** - CORS protection, TLS 1.2+, security headers, and certificate validation out of the box

**⚡ Zero Config** - Sensible defaults for development, configurable for production

**🛡️ Production Ready** - Timeouts, health checks, structured logging, panic recovery, and request rate limiting

**🚀 Developer Friendly** - Fluent API, environment-based config, and community examples

The goal is simple: enable development of **secure** Service APIs - not the scaffolding.


## How to Use

```sh
go get github.com/go-obvious/server
```

### Configuration

The server can be configured using environment variables:

**Server Settings:**
- `SERVER_MODE` - Server mode: `http`, `https`, `aws-gateway-v1`, `aws-gateway-v2` (default: `http`)
- `SERVER_DOMAIN` - Server domain (default: `example.com`)
- `SERVER_PORT` - Server port (default: `8080`)

**Security Configuration:**
- `SERVER_CORS_ALLOWED_ORIGINS` - Comma-separated CORS origins (default: `http://localhost:3000,http://localhost:8080`)
- `SERVER_TLS_MIN_VERSION` - Minimum TLS version: `1.2` or `1.3` (default: `1.2`)
- `SERVER_SECURITY_HEADERS_ENABLED` - Enable automatic security headers (default: `true`)
- `SERVER_HSTS_MAX_AGE` - HSTS max age in seconds (default: `31536000`)

**Timeout Configuration:**
- `SERVER_READ_TIMEOUT` - HTTP read timeout (default: `30s`)
- `SERVER_WRITE_TIMEOUT` - HTTP write timeout (default: `30s`)
- `SERVER_IDLE_TIMEOUT` - HTTP idle timeout (default: `120s`)

**Rate Limiting Configuration:**
- `SERVER_RATE_LIMIT_ENABLED` - Enable request rate limiting (default: `false`)
- `SERVER_RATE_LIMIT_REQUESTS` - Max requests per window (default: `100`)
- `SERVER_RATE_LIMIT_WINDOW` - Time window duration (default: `1m`)
- `SERVER_RATE_LIMIT_BURST` - Burst capacity for token bucket (default: `10`)
- `SERVER_RATE_LIMIT_ALGORITHM` - Algorithm: `token_bucket`, `sliding_window`, `fixed_window` (default: `token_bucket`)
- `SERVER_RATE_LIMIT_KEY_EXTRACTOR` - Key extractor: `ip`, `header`, `custom` (default: `ip`)
- `SERVER_RATE_LIMIT_HEADER` - Header name for header extractor (default: `X-API-Key`)

**Security Features:**
- CORS wildcard `*` origins are blocked for security
- TLS 1.2+ with secure cipher suites and curves
- Automatic security headers: HSTS, CSP, X-Frame-Options, etc.
- Certificate validation at startup
- Configurable rate limiting with multiple algorithms and key extractors

## 🚀 Quick Start

### Basic HTTP Server
```go
package main

import (
    "context"
    "net/http"
    
    chi "github.com/go-chi/chi/v5" 
    "github.com/go-chi/render"
    "github.com/go-obvious/server"
)

type API struct{}

func (api *API) Name() string { return "my-api" }

func (api *API) Register(app server.Server) error {
    router := app.Router().(*chi.Mux)
    router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
        render.JSON(w, r, map[string]string{"message": "Hello, secure world!"})
    })
    return nil
}

func main() {
    version := &server.ServerVersion{Revision: "v1.0.0"}
    srv := server.New(version).WithAPIs(&API{})
    srv.Run(context.Background()) // Runs on :8080 with security enabled
}
```

### 🔒 Secure HTTPS Server  
```bash
# Set environment variables for HTTPS
export SERVER_MODE=https
export SERVER_CERTIFICATE_CERT_FILE=server.crt
export SERVER_CERTIFICATE_KEY_FILE=server.key
export SERVER_TLS_MIN_VERSION=1.3

go run main.go  # Now runs with TLS 1.3, HSTS, and security headers
```

### 🛡️ Rate Limited API Server
```bash
# Enable rate limiting with token bucket algorithm
export SERVER_RATE_LIMIT_ENABLED=true
export SERVER_RATE_LIMIT_REQUESTS=50
export SERVER_RATE_LIMIT_WINDOW=1m
export SERVER_RATE_LIMIT_ALGORITHM=token_bucket

go run main.go  # Now limits to 50 requests per minute per IP
```

## 📚 Examples & Documentation

### Examples
- **[examples/](./examples/)** - Community examples repository (git submodule)
- **[github.com/go-obvious/server-example](https://github.com/go-obvious/server-example)** - Additional examples and tutorials

All examples including basic HTTP servers, HTTPS configurations, rate limiting, and real-world deployment configurations are maintained in the community examples repository.
