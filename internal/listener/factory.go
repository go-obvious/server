package listener

import (
	"net/http"

	"github.com/go-obvious/gateway"
)

const (
	AwsGatewayLambda   = "aws-gateway-v1"
	AwsGatewayV2Lambda = "aws-gateway-v2"
	Https              = "https"
	Http               = "http"
)

type ListenAndServeFunc func(addr string, router http.Handler) error

func GetListener(mode string) ListenAndServeFunc {
	switch mode {
	case AwsGatewayLambda:
		return gateway.ListenAndServeV1
	case AwsGatewayV2Lambda:
		return gateway.ListenAndServeV2
	default:
		return http.ListenAndServe
	}
}
