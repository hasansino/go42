package migrate

import (
	"log/slog"
	"time"
)

const (
	defaultConnectRetryTimeout        = time.Minute
	defaultConnectRetryInitialBackoff = 500 * time.Millisecond
	defaultConnectRetryMaxBackoff     = 5 * time.Second
)

type options struct {
	logger                     *slog.Logger
	connectRetryTimeout        time.Duration
	connectRetryInitialBackoff time.Duration
	connectRetryMaxBackoff     time.Duration
}

type Option func(opts *options)

func WithLogger(logger *slog.Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func WithConnectRetryTimeout(timeout time.Duration) Option {
	return func(opts *options) {
		opts.connectRetryTimeout = timeout
	}
}

func WithConnectRetryBackoff(initial time.Duration, max time.Duration) Option {
	return func(opts *options) {
		opts.connectRetryInitialBackoff = initial
		opts.connectRetryMaxBackoff = max
	}
}

func defaultOptions() options {
	return options{
		connectRetryTimeout:        defaultConnectRetryTimeout,
		connectRetryInitialBackoff: defaultConnectRetryInitialBackoff,
		connectRetryMaxBackoff:     defaultConnectRetryMaxBackoff,
	}
}
