package listener

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

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

func GetListener(mode string, certs *config.Certificate) ListenAndServeFunc {
	switch mode {
	case AwsGatewayLambda:
		return gateway.ListenAndServeV1
	case AwsGatewayV2Lambda:
		return gateway.ListenAndServeV2
	case Https:
		return func(addr string, router http.Handler) error {
			if certs != nil {
				server := &http.Server{
					Addr:    addr,
					Handler: router,
				}
				if certs.CAFile != "" {
					certPool := x509.NewCertPool()
					caCert, err := os.ReadFile(certs.CAFile)
					if err != nil {
						return err
					}
					if !certPool.AppendCertsFromPEM(caCert) {
						return fmt.Errorf("failed to append CA certificates")
					}
					server.TLSConfig = &tls.Config{
						ClientCAs:  certPool,
						ClientAuth: tls.RequireAndVerifyClientCert,
					}
				}
				return server.ListenAndServeTLS(certs.CertFile, certs.KeyFile)
			}
			return http.ListenAndServe(addr, router)
		}
	default:
		return http.ListenAndServe
	}
}
