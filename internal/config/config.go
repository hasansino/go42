package config

import (
	"encoding/json"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	TagNameEnvVarName   = "env"
	TagNameDefaultValue = "default"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" default:"{{SERVICE_NAME}}"`
	Logger      Logger
	Server      Server
}

type Server struct {
	Listen       string        `env:"SERVER_LISTEN_HTTP" default:":8080"`
	ListenPprof  string        `env:"SERVER_LISTEN_PPROF" default:":6060"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TO" default:"5s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TO" default:"5s"`
}

type Logger struct {
	Level  string `env:"LOG_LEVEL" default:"info"`
	Output string `env:"LOG_OUTPUT" default:"stdout"`
	Format string `env:"LOG_FORMAT" default:"json"`
}

// New parses environments and creates new instance of config.
func New() (*Config, error) {
	cfg := new(Config)

	err := env.ParseWithOptions(cfg, env.Options{
		TagName:             TagNameEnvVarName,
		DefaultValueTagName: TagNameDefaultValue,
	})
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
