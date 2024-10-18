package listener_test

import (
	"fmt"
	"net/http"
	"testing"

	v1 "github.com/go-obvious/gateway"
	v2 "github.com/go-obvious/gateway/v2"
	"github.com/stretchr/testify/assert"

	"github.com/go-obvious/server/internal/listener"
)

func TestGetListener(t *testing.T) {
	tests := []struct {
		mode     string
		expected listener.ListenAndServeFunc
	}{
		{mode: listener.AwsGatewayV2Lambda, expected: v2.ListenAndServe},
		{mode: listener.AwsGatewayLambda, expected: v1.ListenAndServe},
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
