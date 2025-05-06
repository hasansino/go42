package grpc

import (
	"context"
	"log/slog"
	"net"
	"runtime/debug"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	logger         *slog.Logger
	grpcServer     *grpc.Server
	serverOptions  []grpc.ServerOption
	tracingEnabled bool
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

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		srvMetrics.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(interceptorLogger(s.logger)),
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		srvMetrics.StreamServerInterceptor(),
		logging.StreamServerInterceptor(interceptorLogger(s.logger)),
		recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	}

	s.serverOptions = append(
		s.serverOptions,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	if s.tracingEnabled {
		s.serverOptions = append(s.serverOptions, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

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

func (s *Server) Shutdown(_ context.Context) error {
	s.grpcServer.GracefulStop()
	return nil
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
