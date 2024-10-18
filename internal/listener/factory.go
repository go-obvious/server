package listener

import (
	"net/http"

	v1 "github.com/go-obvious/gateway"
	v2 "github.com/go-obvious/gateway/v2"
)

const (
	AwsGatewayLambda   = "aws-gateway"
	AwsGatewayV2Lambda = "aws-gateway-v2"
	Https              = "https"
	Http               = "http"
)

type ListenAndServeFunc func(addr string, router http.Handler) error

func GetListener(mode string) ListenAndServeFunc {
	switch mode {
	case AwsGatewayLambda:
		return v1.ListenAndServe
	case AwsGatewayV2Lambda:
		return v2.ListenAndServe
	// case Https:
	// 	return func(addr string, router http.Handler) error {
	// 		return http.ListenAndServeTLS(addr, "path/to/certfile", "path/to/keyfile", router)
	// 	}
	default:
		return http.ListenAndServe
	}
}
