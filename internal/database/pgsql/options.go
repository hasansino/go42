package pgsql

import (
	"log/slog"
	"time"
)

type Option func(w *Wrapper)

func WithLogger(logger *slog.Logger) Option {
	return func(w *Wrapper) {
		w.logger = logger
	}
}

func WithQueryLogging(do bool) Option {
	return func(w *Wrapper) {
		w.queryLogging = do
	}
}

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
