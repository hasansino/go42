package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/hasansino/go42/internal/tools"
)

type Option func(s *Server)

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.l = logger
	}
}

// WithTracing enables/disables tracing.
func WithTracing(enabled bool) Option {
	return func(s *Server) {
		s.tracingEnabled = enabled
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.e.Server.ReadTimeout = d
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.e.Server.WriteTimeout = d
	}
}

// WithStaticRoot sets the root directory for static files.
func WithStaticRoot(root string) Option {
	return func(s *Server) {
		s.staticRoot = root
	}
}

// WithSwaggerRoot sets the root for swagger files.
// @note given directory should contain folders: v1, v2, etc.
func WithSwaggerRoot(root string) Option {
	return func(s *Server) {
		s.swaggerRoot = root
	}
}

// WitHealthCheckCtx sets the health-check context.
// Once context is canceled, health-check will return error.
func WitHealthCheckCtx(ctx context.Context) Option {
	return func(s *Server) {
		go func() {
			<-ctx.Done()
			s.readyStatus.Store(false)
		}()
	}
}

// WithRateLimiter enables/disables rate limiting.
func WithRateLimiter(rate int, burst int) Option {
	return func(s *Server) {
		s.rateLimiter = tools.NewRateLimiter(rate, burst)
	}
}
