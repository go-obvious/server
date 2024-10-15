package listener

import (
	"net/http"

	"github.com/go-obvious/server/internal/listener/aws"
)

const (
	AwsGatewayLambda = "aws-gw-lambda"
	Https            = "https"
	Http             = "http"
)

type ListenAndServeFunc func(addr string, router http.Handler) error

func GetListener(mode string) ListenAndServeFunc {
	switch mode {
	case AwsGatewayLambda:
		return aws.GatewayLambdaListenAndServe
	// case Https:
	// 	return func(addr string, router http.Handler) error {
	// 		return http.ListenAndServeTLS(addr, "path/to/certfile", "path/to/keyfile", router)
	// 	}
	default:
		return http.ListenAndServe
	}
}
