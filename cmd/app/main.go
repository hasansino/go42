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
	"reflect"
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

func main() {
	// configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to initialize config: %v\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// core systems
	initLogging(cfg)
	initVault(cfg)
	etcdCloser := initEtcd(ctx, cfg)
	initLimits(cfg)
	initSentry(cfg)
	pprofCloser := initProfiling(cfg)
	metricsHandler := initMetrics(cfg)

	// http server
	httpServer := httpAPI.New(
		httpAPI.WithLogger(slog.Default().With(slog.String("system", "api"))),
		httpAPI.WithReadTimeout(cfg.Server.ReadTimeout),
		httpAPI.WithWriteTimeout(cfg.Server.WriteTimeout),
		httpAPI.WithStaticRoot(cfg.Server.StaticRoot),
		httpAPI.WithSwaggerRoot(cfg.Server.SwaggerRoot),
	)
	httpServer.Register(metricsprovider.New(metricsHandler))

	// grpc server
	grpcServer := grpcAPI.New(
		grpcAPI.WithLogger(slog.Default().With(slog.String("system", "grpc"))),
		grpcAPI.WithMaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
		grpcAPI.WithMaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
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
			redis.WithClientName(cfg.ServiceName),
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

	// declare required repositories
	var (
		dbCloser          io.Closer
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
			sqlite.WithLogger(slog.Default().With(slog.String("system", "gorm-sqlite"))),
			sqlite.WithMode(cfg.Database.Sqlite.Mode),
			sqlite.WithCacheMod(cfg.Database.Sqlite.CacheMode),
		)
		if sqliteConnErr != nil {
			log.Fatalf("Failed to connect to sqlite: %v\n", sqliteConnErr)
		}
		slog.Info("Connected to sqlite")

		dbCloser = sqliteConn

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
			pgsql.WithLogger(slog.Default().With(slog.String("system", "gorm-pgsql"))),
		)
		if pgsqlConnErr != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v\n", pgsqlConnErr)
		}
		slog.Info("Connected to PostgreSQL")

		dbCloser = pgsqlConn

		// register database metrics
		prometheus.DefaultRegisterer.MustRegister(
			collectors.NewDBStatsCollector(pgsqlConn.SqlDB(), "gorm"),
		)

		// initialize repositories
		exampleRepository = exampleGormRepository.New(pgsqlConn.GormDB(), pgsqlConn)
	}

	// ---

	// service layer

	{
		// Example domain
		exampleLogger := slog.Default().With(slog.String("system", "example"))
		exampleService := example.NewService(
			exampleRepository,
			example.WithLogger(exampleLogger),
			example.WithCache(cacheEngine),
		)
		// http
		exampleHttp := exampleHttpProvider.New(exampleService)
		httpServer.RegisterV1(exampleHttp)
		// grpc
		exampleGrpc := exampleGrpcProvider.New(exampleService)
		grpcServer.Register(exampleGrpc)
	}

	// ---

	go func() {
		slog.Info("starting http server...", slog.String("port", cfg.Server.Listen))
		if err := httpServer.Start(cfg.Server.Listen); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start http server", slog.Any("error", err))
		}
	}()

	go func() {
		slog.Info("starting grpc server...", slog.String("port", cfg.GRPC.Listen))
		if err := grpcServer.Serve(cfg.GRPC.Listen); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("failed to start grpc server", slog.Any("error", err))
		}
	}()

	// listen for signals
	sys := make(chan os.Signal, 1)
	signal.Notify(sys, syscall.SIGINT, syscall.SIGTERM)
	shutdown(
		<-sys, cancel,
		// same order will be used to call Close()
		etcdCloser,
		pprofCloser, httpServer, grpcServer,
		cacheEngine, dbCloser,
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

func initEtcd(ctx context.Context, cfg *config.Config) io.Closer {
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

	etcdLogger := slog.Default().With(slog.String("system", "etcd"))

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

	return client
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
			memlimit.WithLogger(slog.Default().With(slog.String("system", "memlimit"))),
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
				slog.String("system", "pprof"),
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
		// runtime metrics are collected by collectors.NewGoCollector() instead
		vmetrics.WritePrometheus(w, false)
	})
}

// shutdown implements all graceful shutdown logic.
func shutdown(_ os.Signal, cancel context.CancelFunc, closers ...io.Closer) {
	log.Println("Shutting down...")
	cancel()
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
