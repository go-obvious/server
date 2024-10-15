package config

import (
	"sync"

	"github.com/sirupsen/logrus"
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

func Load() {
	for _, cfg := range cfgs {
		if err := cfg.Load(); err != nil {
			logrus.WithError(err).Fatal("error while loading configuration")
		}
	}
}
