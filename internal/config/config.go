package config

import (
	"encoding/json"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"{{SERVICE_NAME}}"`
}

// New parses environments and creates new instance of config.
func New() (*Config, error) {
	cfg := new(Config)

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) String() string {
	if out, err := json.MarshalIndent(&c, "", "  "); err == nil {
		return string(out)
	}
	return ""
}
