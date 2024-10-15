package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/drivetopurchase/apigo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/internal/listener/aws"
)

func TestTransform(t *testing.T) {
	proxy := &aws.LambdaProxy{
		DefaultProxy: apigo.DefaultProxy{Host: "test-host"},
	}

	ctx := context.Background()
	evt := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	req, err := proxy.Transform(ctx, evt)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "GET", req.Method)
	assert.Equal(t, "/test", req.URL.Path)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, ctx, req.Context())
}
