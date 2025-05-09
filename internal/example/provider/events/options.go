package events

import (
	"log/slog"
)

type Option func(*Provider)

func WithLogger(logger *slog.Logger) Option {
	return func(p *Provider) {
		p.logger = logger
	}
}
