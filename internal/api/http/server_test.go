package http

import (
	"context"
	"errors"
	"log/slog"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
)

func TestReadyReturnsDependencyStatus(t *testing.T) {
	tests := []struct {
		name       string
		check      func(context.Context) error
		wantStatus int
	}{
		{
			name: "healthy",
			check: func(context.Context) error {
				return nil
			},
			wantStatus: nethttp.StatusOK,
		},
		{
			name: "unhealthy",
			check: func(context.Context) error {
				return errors.New("dependency unavailable")
			},
			wantStatus: nethttp.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newTestServer(t, WithReadinessCheck(test.check))
			server.readyStatus.Store(ReadyStatusServing)

			if got := getReadyStatus(server); got != test.wantStatus {
				t.Errorf("GET /ready status = %d, want %d", got, test.wantStatus)
			}
		})
	}
}

func TestReadyReturnsServiceUnavailableWhenCheckTimesOut(t *testing.T) {
	checkCanceled := make(chan struct{})
	server := newTestServer(
		t,
		WithReadinessCheckTimeout(25*time.Millisecond),
		WithReadinessCheck(func(ctx context.Context) error {
			<-ctx.Done()
			close(checkCanceled)
			return ctx.Err()
		}),
	)
	server.readyStatus.Store(ReadyStatusServing)

	if got := getReadyStatus(server); got != nethttp.StatusServiceUnavailable {
		t.Errorf("GET /ready status = %d, want %d", got, nethttp.StatusServiceUnavailable)
	}
	waitForSignal(t, checkCanceled, "readiness check cancellation")
}

func TestReadyReturnsServiceUnavailableAfterShutdown(t *testing.T) {
	server := newTestServer(t)
	_, serveResult := startTestServer(t, server)

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-serveResult; err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if got := getReadyStatus(server); got != nethttp.StatusServiceUnavailable {
		t.Errorf("GET /ready status = %d, want %d", got, nethttp.StatusServiceUnavailable)
	}
}

func TestShutdownWaitsForActiveHTTPRequest(t *testing.T) {
	server := newTestServer(t)
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	server.root.GET("/block", func(c *echo.Context) error {
		close(requestStarted)
		<-releaseRequest
		return c.NoContent(nethttp.StatusNoContent)
	})

	address, serveResult := startTestServer(t, server)
	requestResult := make(chan error, 1)
	go func() {
		response, err := nethttp.Get("http://" + address + "/block") //nolint:gosec,noctx
		if response != nil {
			_ = response.Body.Close()
		}
		requestResult <- err
	}()
	waitForSignal(t, requestStarted, "HTTP request to start")

	shutdownResult := make(chan error, 1)
	go func() {
		shutdownResult <- server.Shutdown(context.Background())
	}()
	select {
	case err := <-shutdownResult:
		t.Fatalf("Shutdown() returned before request completed: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(releaseRequest)
	if err := <-shutdownResult; err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-requestResult; err != nil {
		t.Fatalf("HTTP request error = %v", err)
	}
	if err := <-serveResult; err != nil {
		t.Fatalf("Start() error = %v", err)
	}
}

func TestShutdownReturnsAfterEchoGracefulTimeout(t *testing.T) {
	server := newTestServer(t, WithGracefulTimeout(50*time.Millisecond))
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	server.root.GET("/block", func(c *echo.Context) error {
		close(requestStarted)
		<-releaseRequest
		return c.NoContent(nethttp.StatusNoContent)
	})

	address, serveResult := startTestServer(t, server)
	requestResult := make(chan error, 1)
	go func() {
		response, err := nethttp.Get("http://" + address + "/block") //nolint:gosec,noctx
		if response != nil {
			_ = response.Body.Close()
		}
		requestResult <- err
	}()
	waitForSignal(t, requestStarted, "HTTP request to start")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	shutdownStarted := time.Now()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if elapsed := time.Since(shutdownStarted); elapsed < 25*time.Millisecond {
		t.Errorf("Shutdown() returned after %s, before Echo's graceful timeout", elapsed)
	}
	if err := <-serveResult; err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	close(releaseRequest)
	if err := <-requestResult; err != nil {
		t.Fatalf("HTTP request error = %v", err)
	}
}

func TestCORSAllowsAPIKeyHeader(t *testing.T) {
	server := newTestServer(t, WithCORSAllowOrigins([]string{"https://example.com"}))
	server.root.GET("/protected", func(c *echo.Context) error {
		return c.NoContent(nethttp.StatusNoContent)
	})

	address, serveResult := startTestServer(t, server)
	req, err := nethttp.NewRequest(
		nethttp.MethodOptions,
		"http://"+address+"/protected",
		strings.NewReader(""),
	)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", nethttp.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "x-api-key")

	resp, err := nethttp.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("preflight request error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != nethttp.StatusNoContent {
		t.Errorf("preflight status = %d, want %d", resp.StatusCode, nethttp.StatusNoContent)
	}
	allowedHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	if !strings.Contains(strings.ToLower(allowedHeaders), "x-api-key") {
		t.Errorf("Access-Control-Allow-Headers = %q, want x-api-key", allowedHeaders)
	}

	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-serveResult; err != nil {
		t.Fatalf("Start() error = %v", err)
	}
}

func TestClientIPUsesDirectPeerByDefault(t *testing.T) {
	server := newTestServer(t)
	request := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	request.RemoteAddr = "198.51.100.10:1234"
	request.Header.Set(echo.HeaderXForwardedFor, "203.0.113.10")

	if got := server.e.IPExtractor(request); got != "198.51.100.10" {
		t.Errorf("IPExtractor() = %q, want direct peer IP", got)
	}
}

func TestClientIPUsesForwardedAddressFromTrustedProxy(t *testing.T) {
	server := newTestServer(t, WithTrustedProxyCIDRs([]string{"192.0.2.0/24"}))
	request := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set(echo.HeaderXForwardedFor, "203.0.113.10")

	if got := server.e.IPExtractor(request); got != "203.0.113.10" {
		t.Errorf("IPExtractor() = %q, want forwarded client IP", got)
	}
}

func TestClientIPIgnoresForwardedAddressFromUntrustedPeer(t *testing.T) {
	server := newTestServer(t, WithTrustedProxyCIDRs([]string{"192.0.2.0/24"}))
	request := httptest.NewRequest(nethttp.MethodGet, "/", nil)
	request.RemoteAddr = "198.51.100.10:1234"
	request.Header.Set(echo.HeaderXForwardedFor, "203.0.113.10")

	if got := server.e.IPExtractor(request); got != "198.51.100.10" {
		t.Errorf("IPExtractor() = %q, want direct peer IP", got)
	}
}

func newTestServer(t *testing.T, extraOptions ...Option) *Server {
	t.Helper()
	options := []Option{
		WithLogger(slog.New(slog.DiscardHandler)),
		WithStaticRoot(t.TempDir()),
		WithSwaggerRoot(t.TempDir()),
		WithCORSAllowOrigins([]string{"*"}),
		WithGracefulTimeout(time.Second),
	}
	options = append(options, extraOptions...)
	return New(options...)
}

func startTestServer(t *testing.T, server *Server) (string, <-chan error) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	serveResult := make(chan error, 1)
	go func() {
		serveResult <- server.start(echo.StartConfig{Listener: listener})
	}()
	return listener.Addr().String(), serveResult
}

func waitForSignal(t *testing.T, signal <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func getReadyStatus(server *Server) int {
	request := httptest.NewRequest(nethttp.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	server.e.ServeHTTP(response, request)
	return response.Code
}
