package config

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	TagNameEnvVarName   = "env"
	TagNameDefaultValue = "default"
)

type Config struct {
	ServiceName string `env:"SERVICE_NAME" default:"{{SERVICE_NAME}}"`
	Environment string `env:"ENVIRONMENT"  default:"dev"`
	Limits      Limits
	Logger      Logger
	Sentry      Sentry
	Metrics     Metrics
	Pprof       Pprof
	Server      Server
}

type Limits struct {
	AutoMaxProcsEnabled bool    `env:"AUTOMAXPROCS_ENABLED" default:"false"`
	MinMaxProcs         int     `env:"MIN_MAXPROCS"         default:"1"`
	AutoMemLimitEnabled bool    `env:"AUTOMEMLIMIT_ENABLED" default:"false"`
	MemLimitRatio       float64 `env:"MEMLIMIT_RATIO"       default:"0.9"`
}

type Logger struct {
	LogLevel  string `env:"LOG_LEVEL"  default:"info"`
	LogOutput string `env:"LOG_OUTPUT" default:"stdout"`
	LogFormat string `env:"LOG_FORMAT" default:"json"`
}

func (sl *Logger) Level() slog.Level {
	switch strings.ToLower(sl.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type Sentry struct {
	DSN        string  `env:"SENTRY_DSN"         default:""`
	Debug      bool    `env:"SENTRY_DEBUG"       default:"false"`
	Stacktrace bool    `env:"SENTRY_STACKTRACE"  default:"false"`
	SampleRate float64 `env:"SENTRY_SAMPLE_RATE" default:"1.0"`
}

type Metrics struct {
	Timeout time.Duration `env:"METRICS_TIMEOUT" default:"10s"`
}

type Pprof struct {
	Enabled      bool          `env:"PPROF_ENABLED"       default:"false"`
	Listen       string        `env:"PPROF_LISTEN"        default:":6060"`
	Prefix       string        `env:"PPROF_PREFIX"        default:"/debug/pprof"`
	ReadTimeout  time.Duration `env:"PPROF_READ_TIMEOUT"  default:"5s"`
	WriteTimeout time.Duration `env:"PPROF_WRITE_TIMEOUT" default:"60s"`
}

type Server struct {
	Listen       string        `env:"SERVER_LISTEN"        default:":8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT"  default:"5s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" default:"5s"`
	StaticRoot   string        `env:"SERVER_STATIC_ROOT"   default:"/usr/share/www"`
	SwaggerRoot  string        `env:"SERVER_SWAGGER_ROOT"  default:"/usr/share/www/api"`
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
