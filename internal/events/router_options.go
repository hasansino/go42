package events

import "log/slog"

type Option func(*Router)

func WithLogger(logger *slog.Logger) Option {
	return func(r *Router) {
		r.logger = logger
	}
}
