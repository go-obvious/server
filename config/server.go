package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Server struct {
	Mode   string `envconfig:"SERVER_MODE" default:"http"`
	Domain string `envconfig:"SERVER_DOMAIN" default:"example.com"`
	Port   uint   `envconfig:"SERVER_PORT" default:"8080"`
	*Certificate
}

type Certificate struct {
	Cert string `envconfig:"SERVER_CERTIFICATE_CERT"`
	Key  string `envconfig:"SERVER_CERTIFICATE_KEY"`
}

func (c *Server) Load() error {
	return envconfig.Process("server", c)
}
