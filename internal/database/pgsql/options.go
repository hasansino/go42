package pgsql

import (
	"log/slog"
	"time"
)

type Option func(w *Postgres)

func WithLogger(logger *slog.Logger) Option {
	return func(w *Postgres) {
		w.logger = logger
	}
}

func WithQueryLogging(do bool) Option {
	return func(w *Postgres) {
		w.queryLogging = do
	}
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(w *Postgres) {
		w.connMaxIdleTime = d
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(w *Postgres) {
		w.connMaxLifetime = d
	}
}

func WithMaxIdleConns(n int) Option {
	return func(w *Postgres) {
		w.maxIdleConns = n
	}
}

func WithMaxOpenConns(n int) Option {
	return func(w *Postgres) {
		w.maxOpenConns = n
	}
}
