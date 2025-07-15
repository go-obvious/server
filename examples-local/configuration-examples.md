# Configuration Examples

This document provides real-world configuration examples for different deployment scenarios.

## Development Environment

Basic configuration for local development:

```bash
# .env.development
SERVER_MODE=http
SERVER_PORT=8080
SERVER_DOMAIN=localhost
SERVER_CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080,http://localhost:5173
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
```

## Staging Environment

Secure configuration for staging with real certificates:

```bash
# .env.staging
SERVER_MODE=https
SERVER_PORT=443
SERVER_DOMAIN=staging.myapp.com
SERVER_CORS_ALLOWED_ORIGINS=https://staging-frontend.myapp.com,https://staging-admin.myapp.com
SERVER_TLS_MIN_VERSION=1.2
SERVER_CERTIFICATE_CERT_FILE=/etc/ssl/certs/staging.crt
SERVER_CERTIFICATE_KEY_FILE=/etc/ssl/private/staging.key
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=86400
SERVER_READ_TIMEOUT=60s
SERVER_WRITE_TIMEOUT=60s
SERVER_IDLE_TIMEOUT=300s
```

## Production Environment

Enterprise production configuration:

```bash
# .env.production
SERVER_MODE=https
SERVER_PORT=443
SERVER_DOMAIN=api.myapp.com
SERVER_CORS_ALLOWED_ORIGINS=https://myapp.com,https://admin.myapp.com,https://mobile.myapp.com
SERVER_TLS_MIN_VERSION=1.3
SERVER_CERTIFICATE_CERT_FILE=/etc/ssl/certs/production.crt
SERVER_CERTIFICATE_KEY_FILE=/etc/ssl/private/production.key
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=31536000
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
```

## AWS Lambda Gateway V1

Configuration for AWS API Gateway Lambda:

```bash
# .env.lambda
SERVER_MODE=aws-gateway-v1
SERVER_CORS_ALLOWED_ORIGINS=https://myapp.com
SERVER_SECURITY_HEADERS_ENABLED=true
```

## AWS Lambda Gateway V2

Configuration for AWS API Gateway V2 Lambda:

```bash
# .env.lambda-v2  
SERVER_MODE=aws-gateway-v2
SERVER_CORS_ALLOWED_ORIGINS=https://myapp.com
SERVER_SECURITY_HEADERS_ENABLED=true
```

## High-Traffic Production

Configuration for high-traffic scenarios:

```bash
# .env.high-traffic
SERVER_MODE=https
SERVER_PORT=443
SERVER_DOMAIN=api.myapp.com
SERVER_CORS_ALLOWED_ORIGINS=https://myapp.com
SERVER_TLS_MIN_VERSION=1.3
SERVER_CERTIFICATE_CERT_FILE=/etc/ssl/certs/production.crt
SERVER_CERTIFICATE_KEY_FILE=/etc/ssl/private/production.key
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=31536000
# Shorter timeouts for high concurrency
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=60s
```

## Security-First Configuration

Maximum security configuration:

```bash
# .env.security-first
SERVER_MODE=https
SERVER_PORT=443
SERVER_DOMAIN=secure-api.myapp.com
# Strict CORS - only specific origins
SERVER_CORS_ALLOWED_ORIGINS=https://app.myapp.com
# TLS 1.3 only
SERVER_TLS_MIN_VERSION=1.3
SERVER_CERTIFICATE_CERT_FILE=/etc/ssl/certs/secure.crt
SERVER_CERTIFICATE_KEY_FILE=/etc/ssl/private/secure.key
SERVER_SECURITY_HEADERS_ENABLED=true
# Long HSTS for security
SERVER_HSTS_MAX_AGE=63072000
# Conservative timeouts
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=15s  
SERVER_IDLE_TIMEOUT=60s
```

## Docker Configuration

Example Docker environment configuration:

```bash
# docker.env
SERVER_MODE=https
SERVER_PORT=8443
SERVER_DOMAIN=app
SERVER_CORS_ALLOWED_ORIGINS=https://localhost:3000
SERVER_TLS_MIN_VERSION=1.2
SERVER_CERTIFICATE_CERT_FILE=/app/certs/server.crt
SERVER_CERTIFICATE_KEY_FILE=/app/certs/server.key
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=31536000
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=120s
```

```dockerfile
# Dockerfile snippet
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/certs ./certs/
CMD ["./server"]
```

## Kubernetes Configuration

Example Kubernetes deployment with ConfigMap:

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: server-config
data:
  SERVER_MODE: "https"
  SERVER_PORT: "8443"
  SERVER_DOMAIN: "api.myapp.com"
  SERVER_CORS_ALLOWED_ORIGINS: "https://myapp.com,https://admin.myapp.com"
  SERVER_TLS_MIN_VERSION: "1.3"
  SERVER_SECURITY_HEADERS_ENABLED: "true"
  SERVER_HSTS_MAX_AGE: "31536000"
  SERVER_READ_TIMEOUT: "30s"
  SERVER_WRITE_TIMEOUT: "30s"
  SERVER_IDLE_TIMEOUT: "120s"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: server-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: server
  template:
    metadata:
      labels:
        app: server
    spec:
      containers:
      - name: server
        image: myapp/server:latest
        ports:
        - containerPort: 8443
        envFrom:
        - configMapRef:
            name: server-config
        volumeMounts:
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
      volumes:
      - name: tls-certs
        secret:
          secretName: tls-certificates
```

## Environment-Specific Loading

Example of loading environment-specific configuration:

```go
// config_loader.go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/joho/godotenv"
)

func loadEnvironmentConfig() error {
    env := os.Getenv("APP_ENV")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf(".env.%s", env)
    
    if _, err := os.Stat(configFile); err == nil {
        return godotenv.Load(configFile)
    }
    
    // Fallback to .env
    return godotenv.Load()
}

func main() {
    if err := loadEnvironmentConfig(); err != nil {
        log.Printf("Warning: Could not load environment config: %v", err)
    }
    
    // Your server code here...
}
```

## Configuration Validation

Example of validating configuration at startup:

```go
// validate_config.go
package main

import (
    "fmt"
    "os"
    "strings"
    "time"
)

func validateConfiguration() error {
    // Validate CORS origins
    origins := os.Getenv("SERVER_CORS_ALLOWED_ORIGINS")
    if origins == "" {
        return fmt.Errorf("SERVER_CORS_ALLOWED_ORIGINS is required")
    }
    
    if strings.Contains(origins, "*") {
        return fmt.Errorf("wildcard CORS origins not allowed in production")
    }
    
    // Validate TLS settings for HTTPS mode
    if os.Getenv("SERVER_MODE") == "https" {
        certFile := os.Getenv("SERVER_CERTIFICATE_CERT_FILE")
        keyFile := os.Getenv("SERVER_CERTIFICATE_KEY_FILE")
        
        if certFile == "" || keyFile == "" {
            return fmt.Errorf("certificate files required for HTTPS mode")
        }
        
        if _, err := os.Stat(certFile); os.IsNotExist(err) {
            return fmt.Errorf("certificate file not found: %s", certFile)
        }
        
        if _, err := os.Stat(keyFile); os.IsNotExist(err) {
            return fmt.Errorf("key file not found: %s", keyFile)
        }
    }
    
    // Validate timeouts
    if timeout := os.Getenv("SERVER_READ_TIMEOUT"); timeout != "" {
        if _, err := time.ParseDuration(timeout); err != nil {
            return fmt.Errorf("invalid SERVER_READ_TIMEOUT: %v", err)
        }
    }
    
    return nil
}
```

## Configuration Best Practices

### 1. Use Environment-Specific Files
- `.env.development` - Local development
- `.env.staging` - Staging environment  
- `.env.production` - Production environment

### 2. Never Commit Secrets
```bash
# .gitignore
.env.production
.env.staging
*.key
*.crt
secrets/
```

### 3. Validate at Startup
- Check required environment variables
- Validate file paths and permissions
- Test certificate loading
- Verify timeout formats

### 4. Use Reasonable Defaults
- Default to secure settings
- Provide sane timeout values
- Enable security features by default

### 5. Document Everything
- Comment configuration files
- Provide example values
- Explain security implications
- Document environment differences

This configuration reference helps you deploy the go-obvious/server framework securely in any environment.