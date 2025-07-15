# go-obvious/server Examples

This directory contains example applications demonstrating the features and capabilities of the go-obvious/server framework.

## Available Examples

### 1. [Basic HTTP Server](./basic-http/)
**What it demonstrates:**
- Simple HTTP API endpoints
- CORS security configuration
- Automatic security headers
- Request ID tracking and panic recovery
- JSON request/response handling

**Best for:** Getting started, development servers, simple APIs

```bash
cd basic-http
go run main.go
# Server runs on http://localhost:8080
```

### 2. [Secure HTTPS Server](./secure-https/)
**What it demonstrates:**
- HTTPS/TLS with modern security
- TLS 1.2+ enforcement with secure cipher suites
- HSTS and advanced security headers
- Self-signed certificate generation
- TLS configuration and validation

**Best for:** Production deployments, security-sensitive applications

```bash
cd secure-https
go run main.go
# Server runs on https://localhost:8443
```

## Quick Start Guide

1. **Choose an example** that matches your needs
2. **Navigate to the directory** and read its README
3. **Run the example** with `go run main.go`
4. **Test the endpoints** using the provided curl commands
5. **Experiment with configuration** using environment variables

## Common Configuration

All examples support these environment variables:

### Server Settings
```bash
SERVER_PORT=8080                    # Server port
SERVER_DOMAIN=example.com           # Server domain
SERVER_MODE=http                    # http, https, aws-gateway-v1, aws-gateway-v2
```

### Security Configuration
```bash
SERVER_CORS_ALLOWED_ORIGINS=http://localhost:3000,https://myapp.com
SERVER_TLS_MIN_VERSION=1.2          # 1.2 or 1.3
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=31536000        # 1 year in seconds
```

### Timeout Configuration
```bash
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
```

## Built-in Endpoints

Every server automatically includes:

- `GET /version` - Server version and build information
- `GET /healthz` - Health check endpoint for load balancers

## Testing Examples

### Basic Testing
```bash
# Health check
curl http://localhost:8080/healthz

# Version info
curl http://localhost:8080/version

# Custom API
curl http://localhost:8080/api/hello
```

### Security Testing
```bash
# Test CORS (should be rejected)
curl -H "Origin: https://evil.com" http://localhost:8080/api/hello

# Test security headers
curl -v https://localhost:8443/api/secure

# Test HTTPS enforcement
curl -k https://localhost:8443/api/tls-info
```

## Learning Path

1. **Start with basic-http** to understand core concepts
2. **Move to secure-https** to learn TLS and security features
3. **Experiment with configuration** to see how environment variables work
4. **Build your own API** using the patterns from these examples

## Production Considerations

When moving from examples to production:

### Security
- Use real TLS certificates from a trusted CA
- Configure specific CORS origins (never use `*`)
- Set appropriate HSTS max-age values
- Review and customize security headers for your needs

### Performance  
- Tune timeout values for your traffic patterns
- Monitor TLS handshake performance
- Consider HTTP/2 support for modern clients

### Operational
- Implement proper logging and monitoring
- Set up health check endpoints for load balancers
- Plan certificate renewal and rotation
- Test disaster recovery scenarios

## Framework Features

These examples showcase the following go-obvious/server features:

✅ **Security First**
- CORS protection with validation
- Modern TLS with secure defaults
- Comprehensive security headers
- Certificate validation

✅ **Developer Friendly**
- Fluent API design
- Environment-based configuration
- Automatic panic recovery
- Request ID tracking

✅ **Production Ready**
- Configurable timeouts
- Health check endpoints
- Structured logging support
- Multiple deployment modes

✅ **Minimal Code**
- Simple API interface
- Automatic middleware setup
- Built-in best practices
- Zero boilerplate

## Getting Help

- Check the main [README](../README.md) for complete configuration reference
- Review [CLAUDE.md](../CLAUDE.md) for architectural details
- Look at the framework source code for advanced usage
- Each example includes detailed comments and explanations

Start with the basic example and work your way up to more advanced features!