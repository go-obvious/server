# Secure HTTPS Server Example

This example demonstrates the advanced TLS security features of the go-obvious/server framework.

## Features Demonstrated

- **HTTPS/TLS**: Secure connections with TLS 1.2+
- **Secure Cipher Suites**: Modern cryptographic algorithms
- **HSTS Headers**: HTTP Strict Transport Security
- **Security Headers**: Complete security header suite
- **Certificate Management**: Automatic certificate loading
- **TLS Configuration**: Configurable TLS settings

## Quick Start

```bash
cd examples-local/secure-https
go run main.go
```

The server will:
1. Generate a self-signed certificate (server.crt, server.key) if none exists
2. Start on `https://localhost:8443`
3. Enforce TLS 1.2+ with secure cipher suites
4. Add HSTS and security headers automatically

## Test the Secure API

### Secure Endpoint
```bash
curl -k https://localhost:8443/api/secure
```

Response:
```json
{
  "message": "This is a secure HTTPS endpoint",
  "tls_enabled": true,
  "security_features": [
    "TLS 1.2+ enforcement",
    "Secure cipher suites", 
    "HSTS headers",
    "Security headers",
    "CORS protection"
  ]
}
```

### TLS Information
```bash
curl -k https://localhost:8443/api/tls-info
```

Response shows actual TLS negotiation details:
```json
{
  "tls_version": "TLS 1.3",
  "cipher_suite": "TLS_AES_128_GCM_SHA256",
  "server_name": "localhost",
  "negotiated_protocol": "",
  "peer_certificates": 0
}
```

### Protected POST Endpoint
```bash
curl -k -X POST https://localhost:8443/api/protected \
  -H "Content-Type: application/json" \
  -d '{"sensitive": "data", "user_id": 123}'
```

## Security Headers

When you make requests, notice these security headers:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'
```

## TLS Configuration

The example configures secure TLS settings:

```go
// Environment variables set in main()
SERVER_MODE=https
SERVER_PORT=8443
SERVER_TLS_MIN_VERSION=1.2
SERVER_SECURITY_HEADERS_ENABLED=true
SERVER_HSTS_MAX_AGE=31536000
```

## Certificate Management

### Self-Signed Certificates (Development)
The example automatically generates self-signed certificates for testing:
- `server.crt` - Certificate file
- `server.key` - Private key file

### Production Certificates
For production, replace with real certificates:

```bash
# Use your real certificates
SERVER_CERTIFICATE_CERT_FILE=/path/to/your/cert.pem
SERVER_CERTIFICATE_KEY_FILE=/path/to/your/key.pem
go run main.go
```

## Environment Variables

Customize TLS behavior:

```bash
# Enforce TLS 1.3 only
SERVER_TLS_MIN_VERSION=1.3 go run main.go

# Custom timeouts
SERVER_READ_TIMEOUT=60s \
SERVER_WRITE_TIMEOUT=60s \
SERVER_IDLE_TIMEOUT=300s \
go run main.go

# Custom HSTS settings
SERVER_HSTS_MAX_AGE=86400 go run main.go  # 1 day
```

## Security Features

### Automatic TLS Security
- **Minimum TLS 1.2**: Rejects weak TLS versions
- **Secure Cipher Suites**: Only modern AEAD ciphers
- **Perfect Forward Secrecy**: ECDHE key exchange
- **Modern Curves**: X25519, P-256, P-384

### Security Headers
- **HSTS**: Prevents protocol downgrade attacks
- **CSP**: Mitigates XSS attacks  
- **X-Frame-Options**: Prevents clickjacking
- **X-Content-Type-Options**: Prevents MIME sniffing

### Validation
- Certificate files validated at startup
- TLS configuration validated
- Fail-fast on misconfigurations

## Built-in Endpoints

- `GET /version` - Server version (over HTTPS)
- `GET /healthz` - Health check (over HTTPS)

## Production Deployment

For production use:

1. **Use real certificates** from a trusted CA
2. **Configure proper CORS** origins
3. **Set appropriate timeouts** for your load
4. **Monitor TLS metrics** and cipher suite usage
5. **Keep certificates updated** with proper renewal

This example demonstrates enterprise-grade HTTPS security with minimal configuration required.