package healthz_test

import (
	"errors"
	"testing"

	"github.com/go-obvious/server/healthz"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {

	var firstCheckCalled bool = false
	healthz.Register("check1", func() error {
		firstCheckCalled = true
		return nil
	})
	healthz.Register("check2", func() error { return errors.New("check2 failed") })

	err := healthz.NewHealthz().Run()
	if err == nil {
		t.Errorf("Expected an error but got none")
	}
	assert.True(t, firstCheckCalled, "Expected the first check to be called")
}
