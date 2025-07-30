package grpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/hasansino/go42/internal/tools"
)

type Option func(*Server)

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithTracing enables/disables tracing.
func WithTracing(enabled bool) Option {
	return func(s *Server) {
		s.tracingEnabled = enabled
	}
}

// WithMaxRecvMsgSize sets the maximum receive message size.
func WithMaxRecvMsgSize(size int) Option {
	return func(s *Server) {
		s.maxRecvMsgSize = size
	}
}

// WithMaxSendMsgSize sets the maximum send message size.
func WithMaxSendMsgSize(size int) Option {
	return func(s *Server) {
		s.maxSendMsgSize = size
	}
}

// WitHealthCheckCtx sets the health-check context.
// Once context is canceled, health-check will return error.
func WitHealthCheckCtx(ctx context.Context) Option {
	return func(s *Server) {
		go func() {
			<-ctx.Done()
			s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		}()
	}
}

// WithReflection enables/disables reflection.
func WithReflection(enabled bool) Option {
	return func(s *Server) {
		s.withReflection = enabled
	}
}

// WithRateLimiter enables/disables rate limiting.
func WithRateLimiter(rate int, burst int) Option {
	return func(s *Server) {
		s.rateLimiter = tools.NewRateLimiter(rate, burst)
	}
}

// WithUnaryInterceptor adds a unary interceptor with a specific priority.
func WithUnaryInterceptor(priority int, interceptor grpc.UnaryServerInterceptor) Option {
	return func(s *Server) {
		s.extraUnaryInterceptors[priority] = append(s.extraUnaryInterceptors[priority], interceptor)
	}
}

// WithStreamInterceptor adds a stream interceptor with a specific priority.
func WithStreamInterceptor(priority int, interceptor grpc.StreamServerInterceptor) Option {
	return func(s *Server) {
		s.extraStreamInterceptors[priority] = append(s.extraStreamInterceptors[priority], interceptor)
	}
}
