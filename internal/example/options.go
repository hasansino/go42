package example

import (
	"log/slog"
)

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

func WithCache(cache Cache) Option {
	return func(s *Service) {
		s.cache = cache
	}
}

func WithEventer(engine Eventer) Option {
	return func(s *Service) {
		s.events = engine
	}
}
