package sqlite

import (
	"log/slog"
)

type Option func(w *Sqlite)

func WithLogger(logger *slog.Logger) Option {
	return func(w *Sqlite) {
		w.logger = logger
	}
}

func WithQueryLogging(do bool) Option {
	return func(w *Sqlite) {
		w.queryLogging = do
	}
}

func WithMode(mode string) Option {
	return func(w *Sqlite) {
		w.connOpts = append(w.connOpts, ConnectionOption{"mode", mode})
	}
}

func WithCacheMod(mode string) Option {
	return func(w *Sqlite) {
		w.connOpts = append(w.connOpts, ConnectionOption{"cache", mode})
	}
}
