package pgsql

import (
	"log/slog"
	"time"
)

type Option func(w *Wrapper)

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(w *Wrapper) {
		w.connMaxIdleTime = d
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(w *Wrapper) {
		w.connMaxLifetime = d
	}
}

func WithMaxIdleConns(n int) Option {
	return func(w *Wrapper) {
		w.maxIdleConns = n
	}
}

func WithMaxOpenConns(n int) Option {
	return func(w *Wrapper) {
		w.maxOpenConns = n
	}
}

func WithQueryTimeout(d time.Duration) Option {
	return func(w *Wrapper) {
		w.timeout = d
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(w *Wrapper) {
		w.logger = logger
	}
}
