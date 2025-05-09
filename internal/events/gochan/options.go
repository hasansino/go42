package gochan

import (
	"log/slog"
)

type Option func(*GoChan)

func WithLogger(logger *slog.Logger) Option {
	return func(gc *GoChan) {
		gc.logger = logger
	}
}
