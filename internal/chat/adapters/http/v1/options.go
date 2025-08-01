package adapter

import (
	"log/slog"
	"time"
)

// Option represents a configuration option for the Adapter.
type Option func(*adapterOptions)

// WithLogger sets the logger for the adapter.
func WithLogger(logger *slog.Logger) Option {
	return func(o *adapterOptions) {
		o.logger = logger
	}
}

// WithReadTimeout sets the read timeout for websocket connections.
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *adapterOptions) {
		o.readTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout for websocket connections.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *adapterOptions) {
		o.writeTimeout = timeout
	}
}

// WithPingPeriod sets the ping period for websocket connections.
func WithPingPeriod(period time.Duration) Option {
	return func(o *adapterOptions) {
		o.pingPeriod = period
	}
}

// WithPongWait sets the pong wait timeout for websocket connections.
func WithPongWait(wait time.Duration) Option {
	return func(o *adapterOptions) {
		o.pongWait = wait
	}
}