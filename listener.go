package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-obvious/gateway"
	"github.com/go-obvious/server/config"
)

const (
	AwsGatewayLambda   = "aws-gateway-v1"
	AwsGatewayV2Lambda = "aws-gateway-v2"
	Https              = "https"
	Http               = "http"
)

type ListenAndServeFunc func(addr string, router http.Handler) error

func AWSGatewayLambdaListener() ListenAndServeFunc {
	return gateway.ListenAndServeV1
}

func AWSGatewayV2LambdaListener() ListenAndServeFunc {
	return gateway.ListenAndServeV2
}

func HTTPListener() ListenAndServeFunc {
	return http.ListenAndServe
}

func TLSListener(readTimeout, writeTimeout, idleTimeout time.Duration, tlsProvider func() *tls.Config) ListenAndServeFunc {
	return func(addr string, router http.Handler) error {
		server := &http.Server{
			Addr:         addr,
			Handler:      router,
			ErrorLog:     log.New(logAdapter{os.Stderr}, "go-obvious.server TLS Error: ", log.LstdFlags),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
			TLSConfig:    tlsProvider(),
		}
		return server.ListenAndServeTLS("", "")
	}
}

// SecureTLSListener creates a TLS listener with secure defaults and configurable timeouts
func SecureTLSListener(cfg *config.Server) ListenAndServeFunc {
	return TLSListener(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout, func() *tls.Config {
		// Start with secure defaults
		tlsConfig := cfg.GetSecureTLSConfig()

		// If certificate files are provided, load them
		if cfg.Certificate != nil && cfg.CertFile != "" && cfg.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
			if err != nil {
				log.Fatalf("Failed to load TLS certificates: %v", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		return tlsConfig
	})
}
