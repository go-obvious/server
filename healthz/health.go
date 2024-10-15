package healthz

import (
	"errors"
	"sync"
)

type HealthCheck func() error

type Healthz interface {
	Run() error
}

// Register a health check function
func Register(name string, fn HealthCheck) {
	NewHealthz().(*checker).add(name, fn)
}

var (
	h    *checker
	once sync.Once
)

type checker struct {
	mu     sync.Mutex
	checks map[string]HealthCheck
}

func NewHealthz() Healthz {
	once.Do(func() {
		h = &checker{
			checks: make(map[string]HealthCheck),
		}
	})
	return h
}

func (x *checker) add(name string, fn HealthCheck) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.checks[name] = fn
}

func (x *checker) Checks() map[string]HealthCheck {
	x.mu.Lock()
	defer x.mu.Unlock()
	return x.checks
}

func (x *checker) Run() error {
	x.mu.Lock()
	checks := x.checks
	x.mu.Unlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(checks))

	for _, check := range checks {
		wg.Add(1)
		go func(check HealthCheck) {
			defer wg.Done()
			if err := check(); err != nil {
				errCh <- err
			}
		}(check)
	}

	wg.Wait()
	close(errCh)

	var errHistory []error
	for err := range errCh {
		errHistory = append(errHistory, err)
	}

	return errors.Join(errHistory...)
}
