package grpc

import (
	"context"
	"log/slog"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
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
