package listener_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server/internal/listener"
	"github.com/go-obvious/server/internal/listener/aws"
)

func TestGetListener(t *testing.T) {
	tests := []struct {
		mode     string
		expected listener.ListenAndServeFunc
	}{
		{mode: listener.AwsGatewayLambda, expected: aws.GatewayLambdaListenAndServe},
		{mode: listener.Http, expected: http.ListenAndServe}, // Added HTTP type check
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			result := listener.GetListener(tt.mode)
			assert.NotNil(t, result)
			assert.Equal(t, funcType(tt.expected), funcType(result))
		})
	}
}

func funcType(f interface{}) string {
	return fmt.Sprintf("%T", f)
}
