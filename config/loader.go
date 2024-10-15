package config

import (
	"sync"
)

type Configurable interface {
	Load() error
}

var (
	mu   = sync.Mutex{}
	cfgs = make([]Configurable, 0)
)

func Register(cfgs ...Configurable) {
	mu.Lock()
	defer mu.Unlock()
	cfgs = append(cfgs, cfgs...)
}

func Load() error {
	for _, cfg := range cfgs {
		if err := cfg.Load(); err != nil {
			return err
		}
	}
	return nil
}
