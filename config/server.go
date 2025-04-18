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
	CertFile string `envconfig:"SERVER_CERTIFICATE_CERT_FILE"`
	KeyFile  string `envconfig:"SERVER_CERTIFICATE_KEY_FILE"`
	CAFile   string `envconfig:"SERVER_CERTIFICATE_CA_FILE"`
}

func (c *Server) Load() error {
	return envconfig.Process("server", c)
}
