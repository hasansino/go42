package api

import (
	"log/slog"
	"time"
)

type Option func(s *Server)

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

func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.l = logger
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
