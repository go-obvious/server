package config

import (
	"sync"
)

type Configurable interface {
	Load() error
}

var (
	mu             = sync.Mutex{}
	configurations = make([]Configurable, 0)
)

func Register(cfgs ...Configurable) {
	mu.Lock()
	defer mu.Unlock()
	configurations = append(configurations, cfgs...)
}

func Load() error {
	for _, cfg := range configurations {
		if err := cfg.Load(); err != nil {
			return err
		}
	}
	return nil
}
