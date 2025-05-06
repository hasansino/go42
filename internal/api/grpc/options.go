package grpc

import (
	"log/slog"

	"google.golang.org/grpc"
)

type Option func(*Server)

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithTracing(enabled bool) Option {
	return func(s *Server) {
		s.tracingEnabled = enabled
	}
}

// WithMaxRecvMsgSize sets the maximum receive message size.
func WithMaxRecvMsgSize(size int) Option {
	return func(s *Server) {
		s.serverOptions = append(s.serverOptions, grpc.MaxRecvMsgSize(size))
	}
}

// WithMaxSendMsgSize sets the maximum send message size.
func WithMaxSendMsgSize(size int) Option {
	return func(s *Server) {
		s.serverOptions = append(s.serverOptions, grpc.MaxSendMsgSize(size))
	}
}
