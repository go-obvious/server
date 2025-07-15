# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### 🚀 New Features
- **Rate Limiting Middleware**: Added comprehensive request rate limiting with support for multiple algorithms (token bucket, sliding window, fixed window) and configurable key extractors
- **API Lifecycle Management**: Implemented graceful shutdown with SIGTERM/SIGINT handling and optional startup/shutdown hooks for APIs
- **Configuration Registry**: Added centralized configuration loading and validation system with fail-fast validation
- **Enhanced Error Context**: Improved error handling with request/correlation IDs and enhanced context propagation

### 🔒 Security Enhancements
- **Enterprise-grade TLS Hardening**: Added configurable TLS minimum version, secure cipher suites, and certificate validation
- **Security Headers Middleware**: Automatic security headers including HSTS, X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, Referrer-Policy, and Content-Security-Policy
- **Enhanced CORS Configuration**: Improved CORS handling with security validations and protocol enforcement

### 🛠️ Infrastructure & Tooling
- **Zerolog Adapter**: Added structured logging integration with zerolog
- **Connection Log Filter**: Exposed connection-level logging controls
- **Configurable Timeouts**: Added support for HTTP read, write, and idle timeout configuration

### 📊 Testing & Quality
- **100% Test Coverage**: Achieved complete test coverage for API and configuration packages
- **Comprehensive Request Package Tests**: Added extensive test coverage (75.5%) for request handling
- **Middleware Integration Tests**: Added comprehensive integration tests for middleware stack

### 📚 Documentation
- **Enhanced Documentation**: Polished documentation for public release readiness
- **Architecture Guidelines**: Updated CLAUDE.md with comprehensive architecture overview and configuration details
- **Security Configuration Guide**: Added detailed security configuration documentation

### 🧹 Maintenance
- **Code Formatting**: Applied consistent code formatting across the codebase
- **Ignore Files**: Limited and optimized ignore file patterns
- **Examples**: Added comprehensive examples for common use cases

## [v0.1.12] - 2025-04-22

### 🚀 New Features
- **API Initialization Split**: Refactored API initialization to separate concerns and improve modularity
- **Enhanced API Structure**: Improved API organization with better separation of health checks, version endpoints, and custom APIs

### 🛠️ Infrastructure & Tooling
- **Middleware Improvements**: Enhanced middleware stack with better panic recovery and request ID handling
- **Error Handling**: Improved error handling throughout the request lifecycle

## [v0.1.11] - 2025-04-22

### 🔧 Dependencies
- **Chi Router Update**: Updated chi router to latest version for improved performance and security

## [v0.1.10] - 2025-04-22

### 🔧 Dependencies
- **Chi Router Update**: Updated chi router to latest version for improved performance and security

## [v0.1.9] - 2025-04-22

### 🚀 New Features
- **Listener Separation**: Separated listener initialization from server initialization for better modularity
- **Improved Architecture**: Enhanced server architecture with cleaner separation of concerns

### 🔧 Dependencies
- **Testify Update**: Bumped github.com/stretchr/testify from 1.9.0 to 1.10.0

## [v0.1.8] - 2025-04-17

### 🚀 New Features
- **TLS Listener Support**: Added comprehensive TLS listener support with configurable TLS settings
- **HTTPS Server**: Full HTTPS server implementation with certificate management

### 🔒 Security Enhancements
- **TLS Configuration**: Configurable TLS minimum version and cipher suites
- **Certificate Validation**: Enhanced certificate validation and error handling

## [v0.1.7] - 2025-02-05

### 🚀 New Features
- **Middleware Support**: Added comprehensive middleware support for the HTTP server
- **Request Processing**: Enhanced request processing pipeline with middleware integration

### 🛠️ Infrastructure & Tooling
- **Middleware Stack**: Implemented configurable middleware stack for extensibility

## [v0.1.6] - 2024-11-09

### 🧹 Maintenance
- **Test Harness**: Fixed test harness issues for improved testing reliability
- **Testing Infrastructure**: Enhanced testing infrastructure and reliability

## [v0.1.5] - 2024-11-09

### 🔧 Dependencies
- **Chi Router Update**: Updated chi router to latest version for improved performance

## [v0.1.4] - 2024-11-09

### 🧹 Maintenance
- **Server Improvements**: Updated server.go with various improvements and optimizations

## [v0.1.3] - 2024-11-08

### 🔧 Dependencies
- **Gateway Update**: Updated gateway dependencies for improved AWS Lambda support

## [v0.1.2] - 2024-10-21

### 🚀 New Features
- **AWS Lambda Gateway**: Added support for both v1 and v2 AWS Lambda gateway integration
- **Multi-version Support**: Enhanced gateway logic to support multiple AWS API Gateway versions

### 🛠️ Infrastructure & Tooling
- **Gateway Logic**: Improved gateway logic with better version handling

## [v0.1.0] - 2024-10-15

### 🚀 New Features
- **Configuration System**: Initial implementation of comprehensive configuration system
- **Environment-based Config**: Added support for environment-based configuration using envconfig
- **Server Configuration**: Comprehensive server configuration with CORS, timeouts, and security settings

### 🔒 Security Enhancements
- **CORS Configuration**: Secure CORS configuration with validation
- **Security Headers**: Basic security headers implementation

## [v0.0.1] - 2024-10-15

### 🚀 Initial Release
- **HTTP Server Framework**: Initial release of Go HTTP service framework
- **Chi Router Integration**: Built on chi router with fluent interface pattern
- **Basic Middleware**: Panic recovery, CORS, and request ID middleware
- **Health and Version Endpoints**: Built-in `/healthz` and `/version` endpoints
- **AWS Lambda Support**: Basic AWS Lambda gateway support

---

**Legend:**
- 🚀 **New Features**: New functionality and capabilities
- 🔒 **Security Enhancements**: Security-related improvements
- 🛠️ **Infrastructure & Tooling**: Development and deployment improvements
- 📊 **Testing & Quality**: Testing and code quality improvements
- 📚 **Documentation**: Documentation updates and improvements
- 🧹 **Maintenance**: Code maintenance and cleanup
- 🔧 **Dependencies**: Dependency updates and management