package main

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	vmetrics "github.com/VictoriaMetrics/metrics"
	"github.com/getsentry/sentry-go"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"github.com/hasansino/etcd2cfg"
	"github.com/hasansino/libvault"
	"github.com/hasansino/vault2cfg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	slogmulti "github.com/samber/slog-multi"
	etcdClient "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/automaxprocs/maxprocs"
	"google.golang.org/grpc"

	grpcAPI "github.com/hasansino/goapp/internal/api/grpc"
	httpAPI "github.com/hasansino/goapp/internal/api/http"
	"github.com/hasansino/goapp/internal/cache"
	"github.com/hasansino/goapp/internal/cache/memcached"
	"github.com/hasansino/goapp/internal/cache/miniredis"
	"github.com/hasansino/goapp/internal/cache/redis"
	"github.com/hasansino/goapp/internal/config"
	"github.com/hasansino/goapp/internal/database/pgsql"
	pgsqlMigrate "github.com/hasansino/goapp/internal/database/pgsql/migrate"
	"github.com/hasansino/goapp/internal/database/sqlite"
	sqliteMigrate "github.com/hasansino/goapp/internal/database/sqlite/migrate"
	"github.com/hasansino/goapp/internal/events"
	"github.com/hasansino/goapp/internal/events/gochan"
	"github.com/hasansino/goapp/internal/events/nats"
	"github.com/hasansino/goapp/internal/example"
	exampleGrpcProvider "github.com/hasansino/goapp/internal/example/provider/grpc"
	exampleHttpProvider "github.com/hasansino/goapp/internal/example/provider/http"
	exampleGormRepository "github.com/hasansino/goapp/internal/example/repository/gorm"
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

// ShutdownFn represents graceful shutdown for specific component.
type ShutdownFn func(context.Context) error

func main() {
	// main context of the application
	ctx, cancel := context.WithCancel(context.Background())

	// configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initialize config: %v\n", err)
	}

	// core systems
	initLogging(cfg)
	initVault(cfg)
	etcdCloser := initEtcd(ctx, cfg)
	initLimits(cfg)
	initSentry(cfg)
	pprofCloser := initProfiling(cfg)
	metricsHandler := initMetrics(cfg)
	tracingCloser := initTracing(cfg)

	// http server
	httpServer := httpAPI.New(
		httpAPI.WitHealthCheckCtx(ctx),
		httpAPI.WithLogger(slog.Default().With(slog.String("component", "http-server"))),
		httpAPI.WithTracing(cfg.Tracing.Enable),
		httpAPI.WithReadTimeout(cfg.HTTPServer.ReadTimeout),
		httpAPI.WithWriteTimeout(cfg.HTTPServer.WriteTimeout),
		httpAPI.WithStaticRoot(cfg.HTTPServer.StaticRoot),
		httpAPI.WithSwaggerRoot(cfg.HTTPServer.SwaggerRoot),
	)
	httpServer.Register(metricsprovider.New(metricsHandler))

	// grpc server
	grpcServer := grpcAPI.New(
		grpcAPI.WitHealthCheckCtx(ctx),
		grpcAPI.WithLogger(slog.Default().With(slog.String("component", "grpc-server"))),
		grpcAPI.WithTracing(cfg.Tracing.Enable),
		grpcAPI.WithMaxRecvMsgSize(cfg.GRPCServer.MaxRecvMsgSize),
		grpcAPI.WithMaxSendMsgSize(cfg.GRPCServer.MaxSendMsgSize),
	)

	// cache engine
	var (
		cacheEngine cache.Cache
	)
	switch cfg.Cache.Engine {
	case "redis":
		var err error
		cacheEngine, err = redis.New(
			cfg.Cache.Redis.Host, cfg.Cache.Redis.DB,
			redis.WithClientName(cfg.Core.ServiceName),
			redis.WithUserName(cfg.Cache.Redis.Username),
			redis.WithPassword(cfg.Cache.Redis.Password),
			redis.WithMaxRetries(cfg.Cache.Redis.MaxRetries),
			redis.WithMinRetryBackoff(cfg.Cache.Redis.MinRetryBackoff),
			redis.WithMaxRetryBackoff(cfg.Cache.Redis.MaxRetryBackoff),
			redis.WithDialTimeout(cfg.Cache.Redis.DialTimeout),
			redis.WithReadTimeout(cfg.Cache.Redis.ReadTimeout),
			redis.WithWriteTimeout(cfg.Cache.Redis.WriteTimeout),
			redis.WithContextTimeoutEnabled(cfg.Cache.Redis.ContextTimeoutEnabled),
			redis.WithPoolSize(cfg.Cache.Redis.PoolSize),
			redis.WithPoolTimeout(cfg.Cache.Redis.PoolTimeout),
			redis.WithMinIdleConns(cfg.Cache.Redis.MinIdleConns),
			redis.WithMaxIdleConns(cfg.Cache.Redis.MaxIdleConns),
			redis.WithMaxActiveConns(cfg.Cache.Redis.MaxActiveConns),
			redis.WithConnMaxIdleTime(cfg.Cache.Redis.ConnMaxIdleTime),
			redis.WithConnMaxLifetime(cfg.Cache.Redis.ConnMaxLifetime),
		)
		if err != nil {
			log.Fatalf("failed to initialize redis cache: %v\n", err)
		}
		log.Printf("redis cache initialized\n")
	case "miniredis":
		cacheEngine = miniredis.New()
		log.Printf("miniredis cache initialized\n")
	case "memcached":
		var err error
		cacheEngine, err = memcached.New(
			cfg.Cache.Memcached.Hosts,
			memcached.WithTimeout(cfg.Cache.Memcached.Timeout),
			memcached.WithMaxIdleConns(cfg.Cache.Memcached.MaxIdleConns),
		)
		if err != nil {
			log.Fatalf("failed to initialize memcached cache: %v\n", err)
		}
		log.Printf("memcached cache initialized\n")
	default:
		cacheEngine = cache.NewNoop()
		log.Printf("no cache engine initialized\n")
	}

	// database engine
	var (
		dbCloser ShutdownFn
		// declare required repositories
		exampleRepository example.Repository
	)

	switch cfg.Database.Engine {
	case "sqlite":
		// run database migrations
		slog.Info("Running database migrations...")
		err = sqliteMigrate.Migrate(
			cfg.Database.Sqlite.SqliteFile,
			cfg.Database.FullMigratePath(),
			sqlite.ConnectionOption{Key: "mode", Value: cfg.Database.Sqlite.Mode},
			sqlite.ConnectionOption{Key: "cache", Value: cfg.Database.Sqlite.CacheMode},
		)
		if err != nil {
			log.Fatalf("Failed to execute migrations: %v\n", err)
		}

		// connect to database
		slog.Info("Connecting to sqlite...")
		sqliteConn, sqliteConnErr := sqlite.New(
			cfg.Database.Sqlite.SqliteFile,
			sqlite.WithLogger(slog.Default().With(slog.String("component", "gorm-sqlite"))),
			sqlite.WithMode(cfg.Database.Sqlite.Mode),
			sqlite.WithCacheMod(cfg.Database.Sqlite.CacheMode),
		)
		if sqliteConnErr != nil {
			log.Fatalf("Failed to connect to sqlite: %v\n", sqliteConnErr)
		}
		slog.Info("Connected to sqlite")

		dbCloser = sqliteConn.Shutdown

		// register database metrics
		prometheus.DefaultRegisterer.MustRegister(
			collectors.NewDBStatsCollector(sqliteConn.SqlDB(), "gorm"),
		)

		// initialize repositories
		exampleRepository = exampleGormRepository.New(sqliteConn.GormDB(), sqliteConn)
	case "pgsql":
		// run database migrations
		slog.Info("Running database migrations...")
		err = pgsqlMigrate.Migrate(
			cfg.Database.Pgsql.DSN(),
			cfg.Database.FullMigratePath(),
		)
		if err != nil {
			log.Fatalf("Failed to execute migrations: %v\n", err)
		}

		// connect to database
		slog.Info("Connecting to PostgreSQL...")
		pgsqlConn, pgsqlConnErr := pgsql.New(
			cfg.Database.Pgsql.DSN(),
			pgsql.WithConnMaxIdleTime(cfg.Database.Pgsql.ConnMaxIdleTime),
			pgsql.WithConnMaxLifetime(cfg.Database.Pgsql.ConnMaxLifetime),
			pgsql.WithMaxOpenConns(cfg.Database.Pgsql.MaxOpenConns),
			pgsql.WithMaxIdleConns(cfg.Database.Pgsql.MaxIdleConns),
			pgsql.WithQueryTimeout(cfg.Database.Pgsql.QueryTimeout),
			pgsql.WithLogger(slog.Default().With(slog.String("component", "gorm-pgsql"))),
		)
		if pgsqlConnErr != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v\n", pgsqlConnErr)
		}
		slog.Info("Connected to PostgreSQL")

		dbCloser = pgsqlConn.Shutdown

		// register database metrics
		prometheus.DefaultRegisterer.MustRegister(
			collectors.NewDBStatsCollector(pgsqlConn.SqlDB(), "gorm"),
		)

		// initialize repositories
		exampleRepository = exampleGormRepository.New(pgsqlConn.GormDB(), pgsqlConn)
	}

	// event engine
	var (
		eventsEngine events.Eventer
	)

	switch cfg.Events.Engine {
	case "gochan":
		eventsEngine = gochan.New(
			gochan.WithLogger(slog.Default().With(slog.String("component", "events-gochan"))),
		)
		log.Printf("gochan event engine initialized\n")
	case "nats":
		eventsEngine, err = nats.New(
			cfg.Events.NATS.DSN,
			nats.WithLogger(slog.Default().With(slog.String("component", "events-nats"))),
			nats.WithClientName(cfg.Events.NATS.ClientName),
			nats.WithClientToken(cfg.Events.NATS.Token),
			nats.WithConnectTimeout(cfg.Events.NATS.ConnTimeout),
			nats.WithConnectionRetry(cfg.Events.NATS.ConnRetry),
			nats.WithMaxReconnects(cfg.Events.NATS.MaxRetry),
			nats.WithReconnectDelay(cfg.Events.NATS.RetryDelay),
			nats.WithSubGroupPrefix(cfg.Events.NATS.Subscriber.GroupPrefix),
			nats.WithSubWorkerCount(cfg.Events.NATS.Subscriber.WorkerCount),
			nats.WithSubTimeout(cfg.Events.NATS.Subscriber.Timeout),
			nats.WithSubAckTimeout(cfg.Events.NATS.Subscriber.Timeout),
			nats.WithSubCloseTimeout(cfg.Events.NATS.Subscriber.Timeout),
		)
		if err != nil {
			log.Fatalf("failed to initialize nats event engine: %v\n", err)
		}
		log.Printf("nats event engine initialized\n")
	default:
		eventsEngine = events.NewNoop()
		log.Printf("no event engine initialized\n")
	}

	// ---

	// service layer

	{
		// Example domain
		exampleLogger := slog.Default().With(slog.String("component", "example"))
		exampleService := example.NewService(
			exampleRepository,
			example.WithLogger(exampleLogger),
			example.WithCache(cacheEngine),
			example.WithEventer(eventsEngine),
		)
		// http server
		exampleHttp := exampleHttpProvider.New(exampleService)
		httpServer.RegisterV1(exampleHttp)
		// grpc server
		exampleGrpc := exampleGrpcProvider.New(exampleService)
		grpcServer.Register(exampleGrpc)
		// event consumer
		err := exampleService.Subscribe(ctx)
		if err != nil {
			log.Fatalf("example-service failed to subscribe to events: %v\n", err)
		}
	}

	// ---

	go func() {
		slog.Info("starting http server...", slog.String("port", cfg.HTTPServer.Listen))
		if err := httpServer.Start(cfg.HTTPServer.Listen); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", slog.Any("error", err))
		}
	}()

	go func() {
		slog.Info("starting grpc server...", slog.String("port", cfg.GRPCServer.Listen))
		if err := grpcServer.Serve(cfg.GRPCServer.Listen); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("failed to start grpc server", slog.Any("error", err))
		}
	}()

	// entities passed into shutdown are processed in the same order
	shutdown(
		cfg,
		cancel,
		etcdCloser,
		pprofCloser, httpServer.Shutdown, grpcServer.Shutdown,
		cacheEngine.Shutdown, dbCloser, tracingCloser,
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
		slog.String("service", cfg.Core.ServiceName),
		slog.String("environment", cfg.Core.Environment),
		slog.String("hostname", hostname),
		slog.String("build_date", xBuildDate),
		slog.String("build_tag", xBuildTag),
		slog.String("build_commit", xBuildCommit),
	)

	// for both 'log' and 'slog'
	slog.SetDefault(enrichedLogger)

	// any log calls before this point will be non-structured
	slog.Info("logging initialized", slog.String("log_level", cfg.Logger.Level().String()))
}

func initVault(cfg *config.Config) {
	if !cfg.Vault.Enabled {
		return
	}

	vaultClient, err := libvault.New(cfg.Vault.Host)
	if err != nil {
		log.Fatalf("failed to initialise vault: %v", err)
	}

	switch cfg.Vault.AuthType {
	case "token":
		err = vaultClient.TokenAuth(cfg.Vault.Token)
		if err != nil {
			log.Fatalf("failed to authenticate in vault: %v", err)
		}
	default:
		log.Fatalf("unknown vault auth type: %s", cfg.Vault.AuthType)
	}

	slog.Info("connected to vault")

	data, err := vaultClient.Retrieve(cfg.Vault.SecretPath)
	if err != nil {
		log.Fatalf("failed to retrieve vault data: %v", err)
	}

	vault2cfg.Bind(cfg, data)
}

func initEtcd(ctx context.Context, cfg *config.Config) ShutdownFn {
	if !cfg.Etcd.Enabled {
		slog.Warn("etcd is disabled")
		return nil
	}

	// Connect to etcd
	client, err := etcdClient.New(etcdClient.Config{
		Endpoints:   cfg.Etcd.Hosts,
		DialTimeout: cfg.Etcd.Timeout,
		// forces to use log.Default()
		Logger:    nil,
		LogConfig: nil,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}

	syncCtx, cancel := context.WithTimeout(ctx, cfg.Etcd.Timeout)
	defer cancel()

	err = client.Sync(syncCtx)
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}

	etcdLogger := slog.Default().With(slog.String("component", "etcd"))

	switch cfg.Etcd.Method {
	case "bind":
		err = etcd2cfg.Bind(
			cfg, client,
			etcd2cfg.WithLogger(etcdLogger),
			etcd2cfg.WithClientTimeout(cfg.Etcd.Timeout),
		)
		if err != nil {
			log.Fatalf("Failed to bind config: %v", err)
		}
	case "sync":
		err = etcd2cfg.Sync(
			ctx, cfg, client,
			etcd2cfg.WithLogger(etcdLogger),
			etcd2cfg.WithClientTimeout(cfg.Etcd.Timeout),
			etcd2cfg.WithRunInterval(cfg.Etcd.SyncInterval),
		)
		if err != nil {
			log.Fatalf("Failed to bind config: %v", err)
		}
	}

	slog.Info("connected to etcd")

	return func(context.Context) error {
		return client.Close()
	}
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
			memlimit.WithLogger(slog.Default().With(slog.String("component", "memlimit"))),
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
	if !cfg.Sentry.Enabled {
		slog.Warn("sentry is disabled")
		return
	}

	hostname, _ := os.Hostname()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		ServerName:       hostname,
		Environment:      cfg.Core.Environment,
		Release:          xBuildTag,
		SampleRate:       cfg.Sentry.SampleRate,
		Debug:            cfg.Sentry.Debug,
		AttachStacktrace: cfg.Sentry.Stacktrace,
		Tags: map[string]string{
			"service":      cfg.Core.ServiceName,
			"hostname":     hostname,
			"build_date":   xBuildDate,
			"build_tag":    xBuildTag,
			"build_commit": xBuildCommit,
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

func initProfiling(cfg *config.Config) ShutdownFn {
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
				slog.String("component", "pprof"),
			).Handler(), slog.LevelError),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start pprof http server", slog.Any("error", err))
		}
	}()

	return server.Shutdown
}

func initMetrics(cfg *config.Config) http.Handler {
	hostname, _ := os.Hostname()
	metrics.RegisterGlobalLabels(map[string]interface{}{
		"service":     cfg.Core.ServiceName,
		"environment": cfg.Core.Environment,
		"hostname":    hostname,
	})
	metrics.Gauge("application_build", map[string]interface{}{
		"build_date":   xBuildDate,
		"build_tag":    xBuildTag,
		"build_commit": xBuildCommit,
	}).Set(1)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write metrics from `github.com/prometheus/client_golang` collectors
		// they are initialised in init() of prom package
		promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				Registry: prometheus.DefaultRegisterer,
				ErrorLog: log.Default(), // @todo: use named logger
				Timeout:  cfg.Metrics.Timeout,
			}).ServeHTTP(w, r)
		// append metrics from `github.com/VictoriaMetrics/metrics`
		vmetrics.WritePrometheus(w, false)
	})
}

func initTracing(cfg *config.Config) ShutdownFn {
	if !cfg.Tracing.Enable {
		slog.Warn("tracing is disabled")
		return nil
	}

	exporter, err := zipkin.New(
		cfg.Tracing.DSN,
		zipkin.WithClient(&http.Client{
			Timeout: cfg.Tracing.Timeout,
		}),
	)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	// resource is collection of default labels
	hostname, _ := os.Hostname()
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Core.ServiceName),
			semconv.ServiceVersionKey.String(xBuildTag),
			semconv.DeploymentEnvironmentKey.String(cfg.Core.Environment),
			semconv.HostNameKey.String(hostname),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	// provider creates new spans
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// allow to propagate trace context through http requests
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("tracing initialized")

	return tp.Shutdown
}

func shutdown(cfg *config.Config, mainCancel context.CancelFunc, closers ...ShutdownFn) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// wait for signal
	sig := <-sigChan
	log.Printf("Received %s, shutting down...\n", sig.String())

	// allows second signal to bypass graceful shutdown and terminate application immediately
	signal.Stop(sigChan)

	// total timeout for graceful shutdown
	ctx, done := context.WithTimeout(context.Background(), cfg.Core.ShutdownGracePeriod)
	defer done()

	doneChan := make(chan struct{})
	go func(ctx context.Context) {
		// calling cancel() on main context disables health-checks for http and grpc servers
		mainCancel()
		time.Sleep(cfg.Core.ShutdownWaitForProbe)
		for _, c := range closers {
			if c == nil {
				continue
			}
			// we assume that ShutdownFn is blocking and final operation -
			// that is, when ShutdownFn returns, all resources are released
			if err := c(ctx); err != nil {
				slog.Error("shutdown error", slog.Any("error", err))
			}
		}
		left, _ := ctx.Deadline()
		sentry.Flush(time.Until(left))
		close(doneChan)
	}(ctx)

	select {
	case <-doneChan:
		log.Println("graceful shutdown completed")
	case <-ctx.Done():
		log.Println("graceful shutdown timed out")
	}

	os.Exit(0)
}
