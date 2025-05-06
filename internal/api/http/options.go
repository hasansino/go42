package http

import (
	"log/slog"
	"time"
)

type Option func(s *Server)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.l = logger
	}
}

func WithTracing(enabled bool) Option {
	return func(s *Server) {
		s.tracingEnabled = enabled
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.e.Server.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.e.Server.WriteTimeout = d
	}
}

func WithStaticRoot(root string) Option {
	return func(s *Server) {
		s.staticRoot = root
	}
}

func WithSwaggerRoot(root string) Option {
	return func(s *Server) {
		s.swaggerRoot = root
	}
}
