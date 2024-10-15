package aws

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/drivetopurchase/apigo"
)

func GatewayLambdaListenAndServe(host string, router http.Handler) error {
	gw := apigo.Gateway{
		Proxy: &LambdaProxy{
			apigo.DefaultProxy{Host: host},
		},
		Handler: router,
	}
	gw.ListenAndServe()
	return nil
}

type LambdaProxy struct {
	apigo.DefaultProxy
}

func (p *LambdaProxy) Transform(ctx context.Context, evt events.APIGatewayProxyRequest) (*http.Request, error) {
	r, err := p.DefaultProxy.Transform(ctx, evt)
	if err != nil {
		return nil, err
	}
	for key, value := range evt.Headers {
		r.Header.Set(key, value)
	}
	return r.WithContext(ctx), err
}
