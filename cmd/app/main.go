package main

import (
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	vmetrics "github.com/VictoriaMetrics/metrics"
	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	slogmulti "github.com/samber/slog-multi"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/hasansino/goapp/internal/api"
	"github.com/hasansino/goapp/internal/config"
	"github.com/hasansino/goapp/internal/database/pgsql"
	"github.com/hasansino/goapp/internal/database/pgsql/migrate"
	"github.com/hasansino/goapp/internal/metrics"
	metricsprovider "github.com/hasansino/goapp/internal/metrics/providers/http"
)

// These variables are passed as arguments to compiler.
var (
	xBuildDate   string
	xBuildTag    string
	xBuildCommit string
)

func init() {
	if len(xBuildDate) == 0 {
		xBuildDate = "dev"
	}
	if len(xBuildTag) == 0 {
		xBuildTag = "dev"
	}
	if len(xBuildCommit) == 0 {
		xBuildCommit = "dev"
	}
}

func main() {
	// configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initialize config: %v\n", err)
	}

	// core systems
	initLogging(cfg)
	initLimits(cfg)
	initSentry(cfg)
	pprofCloser := initProfiling(cfg)
	metricsHandler := initMetrics(cfg)

	// http server
	httpServer := api.New(
		api.WithReadTimeout(cfg.Server.ReadTimeout),
		api.WithWriteTimeout(cfg.Server.WriteTimeout),
	)
	httpServer.Register(metricsprovider.New(metricsHandler))

	// run database migrations
	slog.Info("Running database migrations...")
	err = migrate.Migrate(
		cfg.Database.PgsqlDSN(),
		cfg.Database.FullMigratePath(),
	)
	if err != nil {
		log.Fatalf("Failed to execute migrations: %v\n", err)
	}

	// connect to database
	slog.Info("Connecting to PostgreSQL...")
	pgsqlConn, pgsqlConnErr := pgsql.NewWrapper(
		cfg.Database.PgsqlDSN(),
		pgsql.WithConnMaxIdleTime(cfg.Database.ConnMaxIdleTime),
		pgsql.WithConnMaxLifetime(cfg.Database.ConnMaxLifetime),
		pgsql.WithMaxOpenConns(cfg.Database.MaxOpenConns),
		pgsql.WithMaxIdleConns(cfg.Database.MaxIdleConns),
	)
	if pgsqlConnErr != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v\n", pgsqlConnErr)
	}
	slog.Info("Connected to PostgreSQL")

	// register database metrics
	prometheus.DefaultRegisterer.MustRegister(
		collectors.NewDBStatsCollector(pgsqlConn.DB(), "gorm"),
	)

	// ---

	// Business logic

	// ---
	go func() {
		slog.Info("starting http server...", slog.String("port", cfg.Server.Listen))
		if err := httpServer.Start(cfg.Server.Listen); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start HTTP server", slog.Any("error", err))
		}
	}()

	// listen for signals
	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGINT, syscall.SIGTERM)
	shutdown(<-sys,
		pprofCloser, httpServer,
	)
}

func initLogging(cfg *config.Config) {
	var slogOutput io.Writer
	switch cfg.Logger.LogOutput {
	case "stdout":
		slogOutput = os.Stdout
	case "stderr":
		slogOutput = os.Stderr
	case "file":
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("failed to open log file: %s", err)
		}
		slogOutput = file
	}

	var slogHandler slog.Handler
	switch cfg.Logger.LogFormat {
	case "text":
		loggerOpts := &slog.HandlerOptions{
			Level: cfg.Logger.Level(),
		}
		slogHandler = slog.NewTextHandler(slogOutput, loggerOpts)
	case "json":
		loggerOpts := &slog.HandlerOptions{
			Level: cfg.Logger.Level(),
		}
		slogHandler = slog.NewJSONHandler(slogOutput, loggerOpts)
	default:
		log.Fatalf("unsupported logging format: %s", cfg.Logger.LogFormat)
	}

	hostname, _ := os.Hostname()

	logger := slog.New(slogHandler)
	enrichedLogger := logger.With(
		slog.String("service", cfg.ServiceName),
		slog.String("build_date", xBuildDate),
		slog.String("build_tag", xBuildTag),
		slog.String("build_commit", xBuildCommit),
		slog.String("hostname", hostname),
	)

	// for both 'log' and 'slog'
	slog.SetDefault(enrichedLogger)

	// any log calls before this point will be non-structured
	slog.Info("logging initialized", slog.String("log_level", cfg.Logger.Level().String()))
}

func initLimits(cfg *config.Config) {
	var err error
	if cfg.Limits.AutoMaxProcsEnabled {
		_, err = maxprocs.Set(maxprocs.Logger(log.Printf))
		if err != nil {
			slog.Error("failed to set maxprocs", slog.Any("error", err))
		}
	} else {
		slog.Warn("package `automaxprocs` is disabled")
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
			slog.Error("failed to set memory limits", slog.Any("error", err))
		}
	} else {
		slog.Warn("package `automemlimit` is disabled")
	}
}

func initSentry(cfg *config.Config) {
	if cfg.Sentry.DSN == "" {
		slog.Warn("sentry is disabled")
		return
	}

	hostname, _ := os.Hostname()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		ServerName:       hostname,
		Environment:      cfg.Environment,
		Release:          xBuildTag,
		SampleRate:       cfg.Sentry.SampleRate,
		Debug:            cfg.Sentry.Debug,
		AttachStacktrace: cfg.Sentry.Stacktrace,
		Tags: map[string]string{
			"service":      cfg.ServiceName,
			"build_date":   xBuildDate,
			"build_tag":    xBuildTag,
			"build_commit": xBuildCommit,
			"hostname":     hostname,
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize sentry: %s", err)
	}

	sentryHandler := sentryslog.Option{
		Level: slog.LevelError,
	}.NewSentryHandler()

	multiLogger := slog.New(
		slogmulti.Fanout(
			slog.Default().Handler(),
			sentryHandler,
		),
	)

	slog.SetDefault(multiLogger)
	slog.Info("sentry initialized")
}

func initProfiling(cfg *config.Config) io.Closer {
	if !cfg.Pprof.Enabled {
		slog.Warn("pprof is disabled")
		return nil
	}

	slog.Info("starting pprof http server...", slog.String("port", cfg.Pprof.Listen))

	prefix := strings.TrimRight(cfg.Pprof.Prefix, " /")

	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc(prefix+"/", pprof.Index)
	pprofMux.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc(prefix+"/profile", pprof.Profile)
	pprofMux.HandleFunc(prefix+"/symbol", pprof.Symbol)
	pprofMux.HandleFunc(prefix+"/trace", pprof.Trace)

	server := &http.Server{
		Addr:         cfg.Pprof.Listen,
		ReadTimeout:  cfg.Pprof.ReadTimeout,
		WriteTimeout: cfg.Pprof.WriteTimeout,
		Handler:      pprofMux,
		ErrorLog: slog.NewLogLogger(
			slog.Default().With(
				slog.String("service", "pprof"),
			).Handler(), slog.LevelError),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start pprof http server", slog.Any("error", err))
		}
	}()

	return server
}

func initMetrics(cfg *config.Config) http.Handler {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	prometheus.DefaultRegisterer = reg

	hostname, _ := os.Hostname()

	metrics.RegisterGlobalLabels(map[string]interface{}{
		"hostname": hostname,
		"service":  cfg.ServiceName,
	})

	metrics.Gauge("application_build", map[string]interface{}{
		"build_date":   xBuildDate,
		"build_tag":    xBuildTag,
		"build_commit": xBuildCommit,
	}).Set(1)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write metrics from `github.com/prometheus/client_golang` collectors
		promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				Registry: reg,
				ErrorLog: log.Default(),
				Timeout:  cfg.Metrics.Timeout,
			}).ServeHTTP(w, r)
		// append metrics from `github.com/VictoriaMetrics/metrics`
		vmetrics.WritePrometheus(w, false)
	})
}

// shutdown implements all graceful shutdown logic.
func shutdown(_ os.Signal, closers ...io.Closer) {
	log.Println("Shutting down...")
	for _, c := range closers {
		if c == nil {
			continue
		}
		if err := c.Close(); err != nil {
			log.Printf(
				"shutdown: error closing %s: %v",
				reflect.TypeOf(c).String(), err,
			)
		}
	}
	sentry.Flush(2 * time.Second)
	os.Exit(0)
}
