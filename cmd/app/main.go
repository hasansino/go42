package main

import (
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/getsentry/sentry-go"
	"go.uber.org/automaxprocs/maxprocs"

	sentryslog "github.com/getsentry/sentry-go/slog"

	"github.com/hasansino/goapp/internal/config"
	"github.com/hasansino/goapp/internal/utils"
)

// These variables are passed as arguments to compiler.
var (
	xBuildDate   string
	xBuildCommit string
)

func init() {
	if len(xBuildDate) == 0 {
		xBuildDate = "dev"
	}
	if len(xBuildCommit) == 0 {
		xBuildCommit = "dev"
	}
}

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v\n", err)
	}

	initLogging(cfg)
	initLimits(cfg)
	initSentry(cfg)
	pprofCloser := initProfiling(cfg.Server.ListenPprof)

	slog.Error("Starting application...", slog.String("listen", cfg.Server.Listen))

	// listen for signals
	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGINT, syscall.SIGTERM)
	shutdown(<-sys, pprofCloser)
}

func initLogging(cfg *config.Config) {
	var slogOutput io.Writer
	switch cfg.Logger.LogOutput {
	case "stdout":
		slogOutput = os.Stdout
	case "stderr":
		slogOutput = os.Stderr
	case "file":
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %s", err)
		}
		slogOutput = file
	}

	var slogHandler slog.Handler
	switch cfg.Logger.LogFormat {
	case "json":
		loggerOpts := &slog.HandlerOptions{
			Level: cfg.Logger.Level(),
		}
		slogHandler = slog.NewJSONHandler(slogOutput, loggerOpts)
	default:
		log.Fatalf("Unsupported logging format: %s", cfg.Logger.LogFormat)
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to retrieve hostname: %s", err)
	}

	logger := slog.New(slogHandler)
	enrichedLogger := logger.With(
		slog.String("service", cfg.ServiceName),
		slog.String("build_commit", xBuildCommit),
		slog.String("build_date", xBuildDate),
		slog.String("hostname", hostname),
	)

	slog.Info("Logging initialized")

	// for both 'log' and 'slog'
	slog.SetDefault(enrichedLogger)
}

func initLimits(cfg *config.Config) {
	var err error
	if cfg.Limits.AutoMaxProcsEnabled {
		_, err = maxprocs.Set(maxprocs.Logger(log.Printf))
		if err != nil {
			slog.Error("Failed to set maxprocs", slog.Any("error", err.Error()))
		}
	} else {
		slog.Warn("Package `automaxprocs` is disabled")
	}
	if cfg.Limits.AutoMemLimitEnabled {
		_, err = memlimit.SetGoMemLimitWithOpts(
			memlimit.WithLogger(slog.Default()),
			memlimit.WithRatio(cfg.Limits.MemLimitRatio),
			memlimit.WithProvider(
				memlimit.ApplyFallback(
					memlimit.FromCgroup,
					memlimit.FromSystem,
				),
			),
		)
		if err != nil {
			slog.Error("Failed to set memory limits", slog.Any("error", err.Error()))
		}
	} else {
		slog.Warn("Package `automemlimit` is disabled")
	}
}

func initSentry(cfg *config.Config) {
	if cfg.Sentry.DSN == "" {
		slog.Warn("Sentry is disabled")
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to retrieve hostname: %s", err)
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:        cfg.Sentry.DSN,
		ServerName: hostname,
		Release:    xBuildCommit,
		Tags: map[string]string{
			"service":      cfg.ServiceName,
			"build_commit": xBuildCommit,
			"build_date":   xBuildDate,
			"hostname":     hostname,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	sentryHandler := sentryslog.Option{
		Level: slog.LevelError,
	}.NewSentryHandler()

	multiHandler := utils.NewSlogMultiHandler(slog.Default().Handler(), sentryHandler)
	multiLogger := slog.New(multiHandler)
	slog.SetDefault(multiLogger)

	slog.Info("Sentry initialized")
}

func initProfiling(listen string) io.Closer {
	if len(listen) == 0 {
		slog.Warn("Pprof is disabled")
		return nil
	}
	slog.Info("Starting pprof http server...", slog.String("port", listen))
	server := &http.Server{Addr: listen}
	go func() {
		if err := server.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start pprof http server: %s", err.Error())
		}
	}()
	return server
}

// shutdown implements all graceful shutdown logic.
func shutdown(_ os.Signal, closers ...io.Closer) {
	log.Println("Shutting down...")
	for _, c := range closers {
		if err := c.Close(); err != nil {
			log.Printf(
				"Error closing %s: %v",
				reflect.TypeOf(c).String(), err,
			)
		}
	}
	sentry.Flush(2 * time.Second)
	os.Exit(0)
}
