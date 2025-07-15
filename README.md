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

**TLS Certificate Configuration:**

- `SERVER_CERTIFICATE_CERT_FILE` - Path to TLS certificate file (required for HTTPS)
- `SERVER_CERTIFICATE_KEY_FILE` - Path to TLS private key file (required for HTTPS)
- `SERVER_CERTIFICATE_CA_FILE` - Path to CA certificate file (optional)

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

## 🔧 Configuration Registry

The server provides an elegant configuration registry pattern that allows external services and libraries to register their own configurations. This enables a unified, fail-fast configuration system where all components validate their settings at startup.

### How It Works

The configuration registry uses a simple `Configurable` interface:

```go
type Configurable interface {
    Load() error
}
```

**The Pattern:**

1. **Self-Registration** - Components register themselves during initialization
2. **Centralized Loading** - Server calls `config.Load()` to process all configurations
3. **Fail-Fast Validation** - Invalid configuration prevents startup with clear error messages
4. **Environment-Driven** - Uses standard environment variables with sensible defaults

### Integrating Your Service Configuration

```go
package myservice

import (
    "github.com/kelseyhightower/envconfig"
    "github.com/go-obvious/server/config"
)

// Define your configuration struct
type Config struct {
    DatabaseURL    string `envconfig:"MY_SERVICE_DATABASE_URL" default:"localhost:5432"`
    APIKey         string `envconfig:"MY_SERVICE_API_KEY" required:"true"`
    MaxConnections int    `envconfig:"MY_SERVICE_MAX_CONNECTIONS" default:"10"`
    Debug          bool   `envconfig:"MY_SERVICE_DEBUG" default:"false"`
}

// Implement the Configurable interface
func (c *Config) Load() error {
    // Load environment variables
    if err := envconfig.Process("my_service", c); err != nil {
        return fmt.Errorf("failed to load MyService config: %w", err)
    }
    
    // Add custom validation
    if c.MaxConnections < 1 || c.MaxConnections > 100 {
        return fmt.Errorf("MY_SERVICE_MAX_CONNECTIONS must be between 1 and 100, got %d", c.MaxConnections)
    }
    
    return nil
}

// Service with self-registering configuration
type Service struct {
    config *Config
}

func NewService() *Service {
    cfg := &Config{}
    
    // Self-register with the configuration system
    config.Register(cfg)
    
    return &Service{config: cfg}
}

// Access validated configuration after server.Run()
func (s *Service) Connect() error {
    // config is guaranteed to be loaded and validated
    return connectToDatabase(s.config.DatabaseURL)
}
```

### Usage in Your Application

```go
func main() {
    // Components self-register their configurations
    myService := myservice.NewService()
    anotherService := another.NewService()
    
    // Server automatically loads and validates ALL registered configurations
    version := &server.ServerVersion{Revision: "v1.0.0"}
    srv := server.New(version).WithAPIs(myService, anotherService)
    
    // config.Load() is called automatically - fails fast if any config is invalid
    srv.Run(context.Background())
    
    // All configurations are guaranteed valid at this point
}
```

### Environment Variables

Set your service environment variables:

```bash
export MY_SERVICE_DATABASE_URL=postgres://user:pass@localhost:5432/mydb
export MY_SERVICE_API_KEY=your-secret-key
export MY_SERVICE_MAX_CONNECTIONS=20
export MY_SERVICE_DEBUG=true
```

### Benefits of This Pattern

**🎯 Centralized** - All configuration loading happens in one place  
**🚀 Fail-Fast** - Invalid configuration prevents startup with clear error messages  
**🔧 Self-Contained** - Each service manages its own configuration and validation  
**🌍 Environment-Driven** - Follows 12-factor app principles  
**🔒 Type-Safe** - Strongly-typed configuration with compile-time checks  
**🔄 Extensible** - Easy to add new configurable components

The configuration registry eliminates configuration chaos by providing a standard pattern that scales from simple services to complex microservice architectures.

## 🔄 API Lifecycle Management

The server provides optional lifecycle hooks for APIs that need startup initialization or graceful shutdown cleanup. This enables proper resource management in production environments.

### How It Works

APIs can implement the optional `LifecycleAPI` interface to receive lifecycle notifications:

```go
type LifecycleAPI interface {
    API
    // Start is called after registration but before the server accepts requests
    Start(ctx context.Context) error
    // Stop is called during graceful shutdown with a timeout context
    Stop(ctx context.Context) error
}
```

**Lifecycle Flow:**
1. **Registration** - APIs register routes with `Register()`
2. **Startup** - `Start()` called for lifecycle-aware APIs  
3. **Runtime** - Server handles requests normally
4. **Shutdown Signal** - SIGTERM/SIGINT or context cancellation
5. **Graceful Shutdown** - Server stops accepting new requests
6. **API Cleanup** - `Stop()` called for resource cleanup
7. **Connection Draining** - Existing requests complete with timeout

### Database Service Example

```go
package database

import (
    "context"
    "database/sql"
    "time"
    
    "github.com/go-obvious/server"
)

type DatabaseService struct {
    db     *sql.DB
    config *Config
}

func NewDatabaseService() *DatabaseService {
    return &DatabaseService{
        config: &Config{}, // Your database config
    }
}

// Implement the required API interface
func (d *DatabaseService) Name() string { return "database" }

func (d *DatabaseService) Register(app server.Server) error {
    router := app.Router().(*chi.Mux)
    router.Get("/users", d.getUsers)
    router.Post("/users", d.createUser)
    return nil
}

// Implement optional lifecycle hooks
func (d *DatabaseService) Start(ctx context.Context) error {
    log.Info().Msg("Connecting to database")
    
    db, err := sql.Open("postgres", d.config.DatabaseURL)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    
    // Test connection
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }
    
    d.db = db
    log.Info().Msg("Database connection established")
    return nil
}

func (d *DatabaseService) Stop(ctx context.Context) error {
    log.Info().Msg("Closing database connections")
    
    if d.db != nil {
        // Wait for active queries to complete or timeout
        if err := d.db.Close(); err != nil {
            log.Error().Err(err).Msg("Error closing database")
            return err
        }
    }
    
    log.Info().Msg("Database connections closed")
    return nil
}

// Your API handlers...
func (d *DatabaseService) getUsers(w http.ResponseWriter, r *http.Request) {
    // Use d.db for queries - guaranteed to be connected
}
```

### Background Worker Example

```go
package worker

import (
    "context"
    "sync"
    "time"
)

type BackgroundWorker struct {
    stopCh chan struct{}
    wg     sync.WaitGroup
}

func NewBackgroundWorker() *BackgroundWorker {
    return &BackgroundWorker{
        stopCh: make(chan struct{}),
    }
}

func (w *BackgroundWorker) Name() string { return "background-worker" }

func (w *BackgroundWorker) Register(app server.Server) error {
    // Optional: register health check endpoint
    return nil
}

func (w *BackgroundWorker) Start(ctx context.Context) error {
    log.Info().Msg("Starting background worker")
    
    w.wg.Add(1)
    go func() {
        defer w.wg.Done()
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                w.doWork()
            case <-w.stopCh:
                log.Info().Msg("Background worker stopping")
                return
            }
        }
    }()
    
    return nil
}

func (w *BackgroundWorker) Stop(ctx context.Context) error {
    log.Info().Msg("Shutting down background worker")
    
    close(w.stopCh)
    
    // Wait for worker to finish with context timeout
    done := make(chan struct{})
    go func() {
        w.wg.Wait()
        close(done)
    }()
    
    select {
    case <-done:
        log.Info().Msg("Background worker stopped gracefully")
    case <-ctx.Done():
        log.Warn().Msg("Background worker shutdown timed out")
    }
    
    return nil
}
```

### Using Lifecycle APIs

```go
func main() {
    // Services with lifecycle management
    dbService := database.NewDatabaseService()
    worker := worker.NewBackgroundWorker()
    
    // Regular API without lifecycle hooks
    apiService := &MyAPI{}
    
    version := &server.ServerVersion{Revision: "v1.0.0"}
    srv := server.New(version).WithAPIs(dbService, worker, apiService)
    
    // Graceful shutdown with SIGTERM/SIGINT handling
    srv.Run(context.Background())
    
    // Lifecycle flow:
    // 1. dbService.Start() - connects to database
    // 2. worker.Start() - starts background processes  
    // 3. Server accepts requests
    // 4. On SIGTERM: stops accepting requests
    // 5. worker.Stop() - stops background processes
    // 6. dbService.Stop() - closes database connections
    // 7. Existing requests complete (up to 30s timeout)
}
```

### Lifecycle Benefits

**🔄 Resource Management** - Proper startup and cleanup of external resources  
**⚡ Fail-Fast** - Startup errors prevent server from accepting requests  
**🛡️ Graceful Shutdown** - Clean resource cleanup on SIGTERM/SIGINT  
**⏱️ Timeout Protection** - Configurable shutdown timeout prevents hanging  
**🔍 Observable** - Structured logging for lifecycle events  
**🧩 Optional** - Existing APIs work unchanged, opt-in for advanced features

The lifecycle management ensures production-ready resource handling while maintaining the simplicity of the basic API interface.

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
- **[examples/](./examples/)** - Comprehensive examples with basic HTTP servers, lifecycle management, and advanced features
- **[github.com/go-obvious/server-example](https://github.com/go-obvious/server-example)** - Additional examples and tutorials

All examples including basic HTTP servers, HTTPS configurations, rate limiting, configuration registry, lifecycle management, and real-world deployment configurations are maintained in the examples directory and community repositories.

### Documentation
- **[Configuration Reference](./CLAUDE.md)** - Complete configuration options and development commands
- **[Contributing Guide](./CONTRIBUTING.md)** - How to contribute to the project
- **[Code of Conduct](./CODE-OF-CONDUCT.md)** - Community guidelines

## 🧪 Testing & Coverage

The server framework includes comprehensive test coverage for all core functionality:

- **API Package**: 100% coverage - Service registration, routing, and integration
- **Request Package**: 75.5% coverage - HTTP request/response handling, JSON processing, error handling
- **Config Package**: 100% coverage - Configuration loading, validation, and registry
- **Middleware**: 90%+ coverage - Security headers, rate limiting, request tracking, panic recovery

Run tests with:
```sh
make test          # Run all tests
make test-coverage # Run tests with coverage report
make lint         # Run code quality checks
```

## 🚀 Production Ready Features

### Security
- **CORS Protection** - Configurable origin restrictions, wildcard blocking
- **TLS Hardening** - TLS 1.2+ minimum, secure cipher suites, certificate validation
- **Security Headers** - HSTS, CSP, X-Frame-Options, X-Content-Type-Options
- **Rate Limiting** - Multiple algorithms (token bucket, sliding window, fixed window)
- **Request Validation** - Body size limits, timeout protection

### Reliability  
- **Graceful Shutdown** - SIGTERM/SIGINT handling, connection draining
- **Health Checks** - Built-in `/healthz` endpoint
- **Panic Recovery** - Automatic recovery with error tracking
- **Structured Logging** - Request correlation, error context
- **Timeout Management** - Read, write, and idle timeouts

### Observability
- **Request Tracking** - Correlation IDs, request/response logging
- **Error Context** - Enhanced error reporting with stack traces
- **Metrics Ready** - Middleware hooks for metrics collection
- **Health Monitoring** - Service health and dependency checks
