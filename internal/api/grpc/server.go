package grpc

import (
	"context"
	"log/slog"
	"net"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"github.com/hasansino/go42/internal/metrics"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type providerAccessor interface {
	Register(*grpc.Server)
}

type Server struct {
	logger     *slog.Logger
	grpcServer *grpc.Server

	maxRecvMsgSize int
	maxSendMsgSize int
	tracingEnabled bool
	healthServer   *health.Server
}

func New(opts ...Option) *Server {
	s := new(Server)
	for _, o := range opts {
		o(s)
	}
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}

	grpcPanicRecoveryHandler := func(p any) error {
		metrics.Counter("application_errors", map[string]interface{}{
			"type": "grpc_panic",
		}).Inc()
		s.logger.Error("grpc panic",
			slog.Any("panic", p),
			slog.Any("stack", debug.Stack()),
		)
		return status.Errorf(codes.Internal, "%s", p)
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		logging.UnaryServerInterceptor(interceptorLogger(s.logger)),
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		logging.StreamServerInterceptor(interceptorLogger(s.logger)),
		recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
	}

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}

	if s.maxRecvMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxRecvMsgSize(s.maxRecvMsgSize))
	}
	if s.maxSendMsgSize > 0 {
		serverOptions = append(serverOptions, grpc.MaxSendMsgSize(s.maxSendMsgSize))
	}
	if s.tracingEnabled {
		serverOptions = append(serverOptions, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	s.grpcServer = grpc.NewServer(serverOptions...)

	s.healthServer = health.NewServer()
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s.grpcServer, s.healthServer)

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
