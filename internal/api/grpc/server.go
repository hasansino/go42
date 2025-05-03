package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hasansino/goapp/internal/metrics"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Register(*grpc.Server)
}

type Server struct {
	logger        *slog.Logger
	grpcServer    *grpc.Server
	serverOptions []grpc.ServerOption
	gracePeriod   time.Duration
}

func New(opts ...Option) *Server {
	s := new(Server)
	for _, o := range opts {
		o(s)
	}
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}

	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	prometheus.MustRegister(srvMetrics)

	grpcPanicRecoveryHandler := func(p any) error {
		metrics.Counter("errors", map[string]interface{}{
			"type": "grpc_panic",
		}).Inc()
		s.logger.Error("grpc panic",
			slog.Any("panic", p),
			slog.Any("stack", debug.Stack()),
		)
		return status.Errorf(codes.Internal, "%s", p)
	}

	unaryInterceptors := grpc.ChainUnaryInterceptor(
		srvMetrics.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(interceptorLogger(s.logger)),
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	)
	streamInterceptors := grpc.ChainStreamInterceptor(
		srvMetrics.StreamServerInterceptor(),
		logging.StreamServerInterceptor(interceptorLogger(s.logger)),
		recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	)

	s.serverOptions = append(s.serverOptions, unaryInterceptors, streamInterceptors)
	s.grpcServer = grpc.NewServer(s.serverOptions...)

	return s
}

func (s *Server) Serve(listen string) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(lis)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.gracePeriod)
	defer cancel()

	done := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case <-done:
		return nil
	}
}

func (s *Server) Register(providers ...providerAccessor) {
	for _, p := range providers {
		p.Register(s.grpcServer)
	}
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
