package config_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/config"
)

// TestConfigurable implements the Configurable interface for testing
type TestConfigurable struct {
	name        string
	loaded      bool
	shouldError bool
	loadError   error
	mu          sync.Mutex
}

func NewTestConfigurable(name string) *TestConfigurable {
	return &TestConfigurable{
		name: name,
	}
}

func (t *TestConfigurable) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.shouldError {
		return t.loadError
	}

	t.loaded = true
	return nil
}

func (t *TestConfigurable) IsLoaded() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.loaded
}

func (t *TestConfigurable) SetShouldError(shouldError bool, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.shouldError = shouldError
	t.loadError = err
}

func TestConfig_Register_SingleConfig(t *testing.T) {
	// Reset config registry before test
	config.Reset()

	// Test registering a single configuration
	testConfig := NewTestConfigurable("test1")

	config.Register(testConfig)

	err := config.Load()
	require.NoError(t, err)
	assert.True(t, testConfig.IsLoaded())
}

func TestConfig_Register_MultipleConfigs(t *testing.T) {
	// Reset config registry before test
	config.Reset()

	// Test registering multiple configurations
	testConfig1 := NewTestConfigurable("test1")
	testConfig2 := NewTestConfigurable("test2")
	testConfig3 := NewTestConfigurable("test3")

	// Register multiple configs in one call
	config.Register(testConfig1, testConfig2, testConfig3)

	err := config.Load()
	require.NoError(t, err)

	// All configs should be loaded
	assert.True(t, testConfig1.IsLoaded())
	assert.True(t, testConfig2.IsLoaded())
	assert.True(t, testConfig3.IsLoaded())
}

func TestConfig_Register_SeparateCalls(t *testing.T) {
	config.Reset()

	// Test registering configurations in separate calls
	testConfig1 := NewTestConfigurable("test1")
	testConfig2 := NewTestConfigurable("test2")

	config.Register(testConfig1)
	config.Register(testConfig2)

	err := config.Load()
	require.NoError(t, err)

	assert.True(t, testConfig1.IsLoaded())
	assert.True(t, testConfig2.IsLoaded())
}

func TestConfig_Load_ConfigurationError(t *testing.T) {
	config.Reset()

	// Test error handling when a configuration fails to load
	testConfig1 := NewTestConfigurable("test1")
	testConfig2 := NewTestConfigurable("test2")

	// Make the second config fail
	testConfig2.SetShouldError(true, errors.New("config2 load error"))

	config.Register(testConfig1, testConfig2)

	err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config2 load error")

	// First config should have been loaded before the error
	assert.True(t, testConfig1.IsLoaded())
	// Second config should not be marked as loaded due to error
	assert.False(t, testConfig2.IsLoaded())
}

func TestConfig_Load_FirstConfigurationError(t *testing.T) {
	config.Reset()

	// Test error handling when the first configuration fails
	testConfig1 := NewTestConfigurable("test1")
	testConfig2 := NewTestConfigurable("test2")

	// Make the first config fail
	testConfig1.SetShouldError(true, errors.New("first config error"))

	config.Register(testConfig1, testConfig2)

	err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first config error")

	// Neither config should be loaded due to early failure
	assert.False(t, testConfig1.IsLoaded())
	assert.False(t, testConfig2.IsLoaded())
}

func TestConfig_Load_NoConfigurationsRegistered(t *testing.T) {
	config.Reset()

	// Test loading when no configurations are registered
	// This should not error, just do nothing
	err := config.Load()
	require.NoError(t, err)
}

func TestConfig_Load_EmptyConfigurationList(t *testing.T) {
	config.Reset()

	// Test registering empty configuration list
	config.Register() // No arguments

	err := config.Load()
	require.NoError(t, err)
}

func TestConfig_Register_NilConfiguration(t *testing.T) {
	config.Reset()

	// Test registering nil configuration
	// This should not panic but might not do anything useful
	config.Register(nil)

	err := config.Load()
	// This might error or might not, depending on implementation
	// The key is that it shouldn't panic
	_ = err
}

func TestConfig_ConcurrentRegistration(t *testing.T) {
	config.Reset()

	// Test concurrent registration to ensure thread safety
	const numGoroutines = 10
	const configsPerGoroutine = 5

	var wg sync.WaitGroup

	// Start multiple goroutines registering configurations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineIndex int) {
			defer wg.Done()

			// Register multiple configs from this goroutine
			goroutineConfigs := make([]config.Configurable, configsPerGoroutine)
			for j := 0; j < configsPerGoroutine; j++ {
				cfg := NewTestConfigurable(fmt.Sprintf("goroutine_%d_config_%d", goroutineIndex, j))
				goroutineConfigs[j] = cfg
			}

			config.Register(goroutineConfigs...)
		}(i)
	}

	wg.Wait()

	// Load all configurations
	err := config.Load()
	require.NoError(t, err)

	// The important thing is that config.Load() succeeds without panic or error
}

func TestConfig_LoadMultipleTimes(t *testing.T) {
	config.Reset()

	// Test calling Load() multiple times
	testConfig := NewTestConfigurable("test")
	config.Register(testConfig)

	// First load
	err1 := config.Load()
	require.NoError(t, err1)
	assert.True(t, testConfig.IsLoaded())

	// Second load should also work
	err2 := config.Load()
	require.NoError(t, err2)
	// Config should still be loaded
	assert.True(t, testConfig.IsLoaded())
}

func TestConfig_Integration_WithServerConfig(t *testing.T) {
	config.Reset()

	// Integration test with actual server config
	serverConfig := &config.Server{}
	testConfig := NewTestConfigurable("integration_test")

	// Register both configs
	config.Register(serverConfig, testConfig)

	err := config.Load()
	require.NoError(t, err)

	// Both should be loaded successfully
	assert.True(t, testConfig.IsLoaded())
	// Server config should have default values
	assert.Equal(t, "http", serverConfig.Mode)
	assert.Equal(t, uint(8080), serverConfig.Port)
}
