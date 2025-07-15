# Basic HTTP Server Example

This example demonstrates the core features of the go-obvious/server framework with HTTP.

## Features Demonstrated

- **Basic HTTP Server**: Simple API endpoints
- **CORS Security**: Configurable origin restrictions  
- **Security Headers**: Automatic security headers
- **Request ID Tracking**: Distributed tracing support
- **Panic Recovery**: Graceful error handling
- **JSON API**: RESTful endpoint patterns

## Quick Start

```bash
cd examples-local/basic-http
go run main.go
```

The server will start on `http://localhost:8080`

## Test the API

### Hello Endpoint
```bash
curl http://localhost:8080/api/hello
```

Response:
```json
{
  "message": "Hello from go-obvious/server!",
  "features": ["CORS Security", "Security Headers", "Request ID Tracking", "Panic Recovery"]
}
```

### Security Features Endpoint
```bash
curl -v http://localhost:8080/api/secure
```

Notice the security headers in the response:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

### Echo Endpoint
```bash
curl -X POST http://localhost:8080/api/echo \
  -H "Content-Type: application/json" \
  -d '{"test": "data", "number": 42}'
```

## Configuration

The server uses default configuration suitable for development:

- **Port**: 8080
- **CORS Origins**: `http://localhost:3000,http://localhost:8080`
- **Security Headers**: Enabled
- **Timeouts**: 30s read/write, 120s idle

## Built-in Endpoints

The framework automatically provides:

- `GET /version` - Server version information
- `GET /healthz` - Health check endpoint

## Environment Variables

You can customize the server behavior:

```bash
# Change port
SERVER_PORT=3000 go run main.go

# Configure CORS for your frontend
SERVER_CORS_ALLOWED_ORIGINS=http://localhost:3000,https://myapp.com go run main.go

# Disable security headers (not recommended)
SERVER_SECURITY_HEADERS_ENABLED=false go run main.go
```

## Code Structure

- `ExampleAPI` - Implements the `server.API` interface
- `handleHello` - Simple JSON response
- `handleSecure` - Demonstrates security features
- `handleEcho` - POST endpoint with JSON body parsing

This example shows the minimal code needed to create a production-ready HTTP API with security best practices built-in.