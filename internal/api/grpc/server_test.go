package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"testing"
	"time"

	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
)

func TestShutdownForcesGRPCServerAfterDeadline(t *testing.T) {
	server := New(WithLogger(slog.New(slog.DiscardHandler)))
	listener := bufconn.Listen(1024 * 1024)
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.grpcServer.Serve(listener)
	}()

	clientConn, err := grpcpkg.NewClient(
		"passthrough:///bufconn",
		grpcpkg.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpcpkg.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("create gRPC client: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })

	watchCtx, cancelWatch := context.WithCancel(context.Background())
	defer cancelWatch()
	stream, err := healthpb.NewHealthClient(clientConn).Watch(
		watchCtx,
		&healthpb.HealthCheckRequest{},
	)
	if err != nil {
		t.Fatalf("start health watch: %v", err)
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatalf("receive initial health status: %v", err)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelShutdown()
	err = server.Shutdown(shutdownCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want %v", err, context.DeadlineExceeded)
	}

	select {
	case err := <-serveResult:
		if err != nil && !errors.Is(err, grpcpkg.ErrServerStopped) {
			t.Errorf("Serve() error = %v, want nil or %v", err, grpcpkg.ErrServerStopped)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for forced gRPC shutdown")
	}
}

func TestShutdownGracefullyStopsIdleGRPCServer(t *testing.T) {
	server := New(WithLogger(slog.New(slog.DiscardHandler)))
	listener := bufconn.Listen(1024 * 1024)
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.grpcServer.Serve(listener)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	select {
	case err := <-serveResult:
		if err != nil && !errors.Is(err, grpcpkg.ErrServerStopped) {
			t.Errorf("Serve() error = %v, want nil or %v", err, grpcpkg.ErrServerStopped)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for graceful gRPC shutdown")
	}
}
