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
	"github.com/hasansino/vault2cfg"
	"github.com/hashicorp/vault-client-go"
	"github.com/lmittmann/tint"
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

	grpcAPI "github.com/hasansino/go42/internal/api/grpc"
	httpAPI "github.com/hasansino/go42/internal/api/http"
	"github.com/hasansino/go42/internal/cache"
	"github.com/hasansino/go42/internal/cache/aerospike"
	"github.com/hasansino/go42/internal/cache/memcached"
	"github.com/hasansino/go42/internal/cache/otter"
	"github.com/hasansino/go42/internal/cache/redis"
	"github.com/hasansino/go42/internal/config"
	"github.com/hasansino/go42/internal/database"
	"github.com/hasansino/go42/internal/database/mysql"
	mysqlMigrate "github.com/hasansino/go42/internal/database/mysql/migrate"
	"github.com/hasansino/go42/internal/database/pgsql"
	pgsqlMigrate "github.com/hasansino/go42/internal/database/pgsql/migrate"
	"github.com/hasansino/go42/internal/database/sqlite"
	sqliteMigrate "github.com/hasansino/go42/internal/database/sqlite/migrate"
	"github.com/hasansino/go42/internal/events"
	"github.com/hasansino/go42/internal/events/gochan"
	"github.com/hasansino/go42/internal/events/kafka"
	"github.com/hasansino/go42/internal/events/nats"
	"github.com/hasansino/go42/internal/events/rabbitmq"
	"github.com/hasansino/go42/internal/example"
	exampleGrpcProviderV1 "github.com/hasansino/go42/internal/example/provider/grpc/v1"
	exampleHttpProviderV1 "github.com/hasansino/go42/internal/example/provider/http/v1"
	exampleRepositoryPkg "github.com/hasansino/go42/internal/example/repository"
	exampleWorkers "github.com/hasansino/go42/internal/example/workers"
	"github.com/hasansino/go42/internal/metrics"
	"github.com/hasansino/go42/internal/metrics/observers"
	metricsprovider "github.com/hasansino/go42/internal/metrics/providers/http"
	"github.com/hasansino/go42/internal/outbox"
	outboxRepositoryPkg "github.com/hasansino/go42/internal/outbox/repository"
	outboxWorkers "github.com/hasansino/go42/internal/outbox/workers"
	"github.com/hasansino/go42/internal/tools"
)

// These variables are passed as arguments to compiler.
var (
	xBuildTag    string
	xBuildCommit string
)

func init() {
	if len(xBuildTag) == 0 {
		xBuildTag = "dev"
	}
	if len(xBuildCommit) == 0 {
		xBuildCommit = "dev"
	}
}

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
	initVault(ctx, cfg)
	etcdCloser := initEtcd(ctx, cfg)
	initLimits(cfg)
	initSentry(ctx, cfg)
	pprofCloser := initProfiling(cfg)
	metricsHandler := initMetrics(cfg)
	tracingCloser := initTracing(cfg)

	// http server
	httpServer := httpAPI.New(
		httpAPI.WitHealthCheckCtx(ctx),
		httpAPI.WithLogger(slog.Default().With(slog.String("component", "http-server"))),
		httpAPI.WithTracing(cfg.Tracing.Enable),
		httpAPI.WithReadTimeout(cfg.Server.HTTP.ReadTimeout),
		httpAPI.WithWriteTimeout(cfg.Server.HTTP.WriteTimeout),
		httpAPI.WithStaticRoot(cfg.Server.HTTP.StaticRoot),
		httpAPI.WithSwaggerRoot(cfg.Server.HTTP.SwaggerRoot),
	)
	httpServer.Register(metricsprovider.New(metricsHandler))

	// grpc server
	grpcServer := grpcAPI.New(
		grpcAPI.WitHealthCheckCtx(ctx),
		grpcAPI.WithLogger(slog.Default().With(slog.String("component", "grpc-server"))),
		grpcAPI.WithTracing(cfg.Tracing.Enable),
		grpcAPI.WithMaxRecvMsgSize(cfg.Server.GRPC.MaxRecvMsgSize),
		grpcAPI.WithMaxSendMsgSize(cfg.Server.GRPC.MaxSendMsgSize),
		grpcAPI.WithReflection(cfg.Server.GRPC.ReflectionEnabled),
	)

	// database engine
	var (
		dbEngine database.Database
	)
	switch cfg.Database.Engine {
	case "sqlite":
		// run database migrations
		slog.Info("running database migrations...")
		err = sqliteMigrate.Migrate(
			cfg.Database.Sqlite.SqliteFile,
			cfg.Database.FullMigratePath(),
			sqlite.ConnectionOption{Key: "mode", Value: cfg.Database.Sqlite.Mode},
			sqlite.ConnectionOption{Key: "cache", Value: cfg.Database.Sqlite.CacheMode},
		)
		if err != nil {
			log.Fatalf("failed to execute migrations: %v\n", err)
		}

		// connect to database
		slog.Info("Connecting to sqlite...")
		var sqliteConnErr error
		dbEngine, sqliteConnErr = sqlite.New(
			cfg.Database.Sqlite.SqliteFile,
			sqlite.WithMode(cfg.Database.Sqlite.Mode),
			sqlite.WithCacheMod(cfg.Database.Sqlite.CacheMode),
			sqlite.WithLogger(slog.Default().With(slog.String("component", "gorm-sqlte"))),
			sqlite.WithQueryLogging(cfg.Database.LogQueries),
		)
		if sqliteConnErr != nil {
			log.Fatalf("failed to connect to sqlite: %v\n", sqliteConnErr)
		}

		slog.Info("connected to sqlite")
	case "pgsql":
		// run database migrations
		slog.Info("running database migrations...")
		err = pgsqlMigrate.Migrate(
			cfg.Database.Pgsql.Master.DSN(),
			cfg.Database.FullMigratePath(),
		)
		if err != nil {
			log.Fatalf("failed to execute migrations: %v\n", err)
		}

		// connect to database
		slog.Info("connecting to PostgreSQL...")
		var pgsqlConnErr error
		dbEngine, pgsqlConnErr = pgsql.New(
			cfg.Database.Pgsql.Master.DSN(),
			cfg.Database.Pgsql.Slave.DSN(),
			pgsql.WithLogger(slog.Default().With(slog.String("component", "gorm-pgsql"))),
			pgsql.WithQueryLogging(cfg.Database.LogQueries),
			pgsql.WithConnMaxIdleTime(cfg.Database.Pgsql.ConnMaxIdleTime),
			pgsql.WithConnMaxLifetime(cfg.Database.Pgsql.ConnMaxLifetime),
			pgsql.WithMaxOpenConns(cfg.Database.Pgsql.MaxOpenConns),
			pgsql.WithMaxIdleConns(cfg.Database.Pgsql.MaxIdleConns),
		)
		if pgsqlConnErr != nil {
			log.Fatalf("failed to connect to pgsql: %v\n", pgsqlConnErr)
		}

		slog.Info("connected to pgsql")
	case "mysql":
		// run database migrations
		slog.Info("running database migrations...")
		err = mysqlMigrate.Migrate(
			cfg.Database.Mysql.Master.DSN(),
			cfg.Database.FullMigratePath(),
		)
		if err != nil {
			log.Fatalf("failed to execute migrations: %v\n", err)
		}

		// connect to database
		slog.Info("connecting to MySQL...")
		var mysqlConnErr error
		dbEngine, mysqlConnErr = mysql.New(
			cfg.Database.Mysql.Master.DSN(),
			cfg.Database.Mysql.Slave.DSN(),
			mysql.WithLogger(slog.Default().With(slog.String("component", "gorm-mysql"))),
			mysql.WithQueryLogging(cfg.Database.LogQueries),
			mysql.WithConnMaxIdleTime(cfg.Database.Mysql.ConnMaxIdleTime),
			mysql.WithConnMaxLifetime(cfg.Database.Mysql.ConnMaxLifetime),
			mysql.WithMaxOpenConns(cfg.Database.Mysql.MaxOpenConns),
			mysql.WithMaxIdleConns(cfg.Database.Mysql.MaxIdleConns),
		)
		if mysqlConnErr != nil {
			log.Fatalf("failed to connect to mysql: %v\n", mysqlConnErr)
		}

		slog.Info("connected to mysql")
	default:
		log.Fatalf("empty or not supported database engine: %v\n", cfg.Database.Engine)
	}

	// database metrics
	{
		masterDB, err := dbEngine.Master().DB()
		if err != nil {
			log.Fatalf("failed to retrieve master db: %v\n", err)
		}
		dbObserverMaster, err := observers.NewDatabaseObserver(
			masterDB,
			observers.WithName("gorm-master"),
		)
		if err != nil {
			log.Fatalf("failed to initialize database metrics: %v\n", err)
		}
		go dbObserverMaster.Observe(ctx)

		slaveDB, err := dbEngine.Slave().DB()
		if err != nil {
			log.Fatalf("failed to retrieve slave db: %v\n", err)
		}
		dbObserverSlave, err := observers.NewDatabaseObserver(
			slaveDB,
			observers.WithName("gorm-slave"),
		)
		if err != nil {
			log.Fatalf("failed to initialize database metrics: %v\n", err)
		}
		go dbObserverSlave.Observe(ctx)
	}

	// cache engine
	var (
		cacheEngine cache.Cache
	)
	switch cfg.Cache.Engine {
	case "none":
		cacheEngine = cache.NewNoop()
		slog.Info("no cache engine initialized")
	case "otter":
		cacheEngine, err = otter.New()
		if err != nil {
			log.Fatalf("failed to initialize otter cache: %v\n", err)
		}
		slog.Info("otter cache engine initialized")
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
		slog.Info("memcached cache initialized")
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
		slog.Info("redis cache initialized")
	case "aerospike":
		var err error
		cacheEngine, err = aerospike.New(
			cfg.Cache.Aerospike.Hosts,
			cfg.Cache.Aerospike.Namespace,
		)
		if err != nil {
			log.Fatalf("failed to initialize aerospike cache: %v\n", err)
		}
		slog.Info("aerospike cache initialized")
	}

	// event engine
	var (
		eventsEngine events.Eventer
	)

	switch cfg.Events.Engine {
	case "none":
		eventsEngine = events.NewNoop()
		slog.Info("no event engine initialized")
	case "gochan":
		eventsEngine = gochan.New(
			gochan.WithLogger(slog.Default().With(slog.String("component", "events-gochan"))),
		)
		slog.Info("gochan event engine initialized")
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
		slog.Info("nats event engine initialized")
	case "rabbitmq":
		eventsEngine, err = rabbitmq.New(
			cfg.Events.RabbitMQ.DSN,
			rabbitmq.WithLogger(slog.Default().With(slog.String("component", "events-rabbitmq"))),
		)
		if err != nil {
			log.Fatalf("failed to initialize rabbitmq event engine: %v\n", err)
		}
		slog.Info("rabbitmq event engine initialized")
	case "kafka":
		eventsEngine, err = kafka.New(
			cfg.Events.Kafka.Brokers,
			cfg.Events.Kafka.ConsumerGroup,
			kafka.WithLogger(slog.Default().With(slog.String("component", "events-kafka"))),
		)
		if err != nil {
			log.Fatalf("failed to initialize kafka event engine: %v\n", err)
		}
		slog.Info("kafka event engine initialized")
	}

	// ---

	// service layer

	{
		// outbox domain
		outboxLogger := slog.Default().With(slog.String("component", "outbox-service"))
		outboxRepository := outboxRepositoryPkg.New(database.NewBaseRepository(dbEngine))
		outboxService := outbox.NewService(
			outboxRepository,
			outbox.WithLogger(outboxLogger),
		)

		outboxPublisher := outboxWorkers.NewOutboxMessagePublisher(
			outboxRepository,
			eventsEngine,
			outboxWorkers.OutboxMessagePublisherWithLogger(
				slog.Default().With(slog.String("component", "outbox-publisher")),
			),
		)
		go outboxPublisher.Run(ctx, 5*time.Second, 100)

		// example domain
		exampleLogger := slog.Default().With(slog.String("component", "example-service"))
		exampleRepository := exampleRepositoryPkg.New(database.NewBaseRepository(dbEngine))
		exampleService := example.NewService(
			exampleRepository,
			outboxService,
			example.WithLogger(exampleLogger),
		)

		fruitEventSubscriber := exampleWorkers.NewFruitEventSubscriber(
			exampleRepository,
			eventsEngine,
			exampleWorkers.FruitEventSubscriberWithLogger(
				slog.Default().With(slog.String("component", "example-subscriber")),
			),
		)
		err := fruitEventSubscriber.Subscribe(ctx, eventsEngine)
		if err != nil {
			log.Fatalf("failed to subscribe to events: %v\n", err)
		}

		// http server
		exampleHttp := exampleHttpProviderV1.New(
			exampleService, cacheEngine,
			exampleHttpProviderV1.WithCache(cacheEngine, 1*time.Second),
		)
		httpServer.RegisterV1(exampleHttp)

		// grpc server
		exampleGrpc := exampleGrpcProviderV1.New(exampleService)
		grpcServer.Register(exampleGrpc)
	}

	// ---

	go func() {
		slog.Info("starting http server...", slog.String("port", cfg.Server.HTTP.Listen))
		if err := httpServer.Start(cfg.Server.HTTP.Listen); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start http server: %v\n", err)
		}
	}()

	go func() {
		slog.Info("starting grpc server...", slog.String("port", cfg.Server.GRPC.Listen))
		if err := grpcServer.Serve(cfg.Server.GRPC.Listen); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("failed to start grpc server: %v\n", err)
		}
	}()

	// entities passed into shutdown are processed in the same order
	shutdown(
		cfg,
		cancel,
		etcdCloser, pprofCloser,
		httpServer, grpcServer, eventsEngine,
		cacheEngine, dbEngine, tracingCloser,
	)
}

func initLogging(cfg *config.Config) {
	var slogOutput io.Writer
	switch cfg.Logger.LogOutput {
	case "none":
		slogOutput = io.Discard
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
	case "tint":
		loggerOpts := &tint.Options{
			Level:      cfg.Logger.Level(),
			TimeFormat: time.Kitchen,
		}
		slogHandler = tint.NewHandler(slogOutput, loggerOpts)
	default:
		log.Fatalf("unsupported logging format: %s", cfg.Logger.LogFormat)
	}

	hostname, _ := os.Hostname()

	logger := slog.New(tools.SlogContextWrapper(slogHandler))
	enrichedLogger := logger.With(
		slog.String("service", cfg.Core.ServiceName),
		slog.String("environment", cfg.Core.Environment),
		slog.String("hostname", hostname),
		slog.String("build_tag", xBuildTag),
		slog.String("build_commit", xBuildCommit),
	)

	// for both 'log' and 'slog'
	slog.SetDefault(enrichedLogger)

	// any log calls before this point will be non-structured
	slog.Info("logging initialized", slog.String("log_level", cfg.Logger.Level().String()))
}

func initVault(ctx context.Context, cfg *config.Config) {
	if !cfg.Vault.Enabled {
		return
	}

	client, err := vault.New(
		vault.WithAddress(cfg.Vault.Host),
		vault.WithRequestTimeout(cfg.Vault.Timeout),
	)
	if err != nil {
		log.Fatalf("failed to initialise vault client: %v", err)
	}

	switch cfg.Vault.AuthType {
	case "token":
		err = client.SetToken(cfg.Vault.Token)
		if err != nil {
			log.Fatalf("failed to authenticate in vault: %v", err)
		}
	default:
		log.Fatalf("unknown vault auth type: %s", cfg.Vault.AuthType)
	}

	slog.Info("connected and authenticated @ vault")

	reqCtx, cancel := context.WithTimeout(ctx, cfg.Vault.Timeout)
	defer cancel()

	data, err := client.Secrets.KvV2Read(
		reqCtx, cfg.Vault.SecretPath, vault.WithMountPath(cfg.Vault.MountPath),
	)
	if err != nil {
		log.Fatalf("failed to read vault secrets: %v", err)
	}

	if err := vault2cfg.Bind(cfg, data.Data.Data); err != nil {
		log.Fatalf("failed to bind vault secrets: %v", err)
	}
}

func initEtcd(ctx context.Context, cfg *config.Config) ShutMeDown {
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

	return &ShutMeDownWrap{closer: client}
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

func initSentry(ctx context.Context, cfg *config.Config) ShutMeDown {
	if !cfg.Sentry.Enabled {
		slog.Warn("sentry is disabled")
		return nil
	}

	hostname, _ := os.Hostname()

	client, err := sentry.NewClient(sentry.ClientOptions{
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
			"build_tag":    xBuildTag,
			"build_commit": xBuildCommit,
		},
	})
	sentry.CurrentHub().BindClient(client)

	if err != nil {
		log.Fatalf("failed to initialize sentry: %s", err)
	}

	sentryHandler := sentryslog.Option{
		EventLevel: []slog.Level{slog.LevelError},
	}.NewSentryHandler(ctx)

	multiLogger := slog.New(
		slogmulti.Fanout(
			slog.Default().Handler(),
			sentryHandler,
		),
	)

	slog.SetDefault(multiLogger)
	slog.Info("sentry initialized")

	return &ShutMeDownWrap{
		fn: func(ctx context.Context) error {
			left, ok := ctx.Deadline()
			if ok {
				client.Flush(time.Until(left))
			} else {
				client.Flush(time.Second)
			}
			client.Close()
			return nil
		},
	}
}

func initProfiling(cfg *config.Config) ShutMeDown {
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

	return server
}

func initMetrics(cfg *config.Config) http.Handler {
	hostname, _ := os.Hostname()
	metrics.RegisterGlobalLabels(map[string]interface{}{
		"service":     cfg.Core.ServiceName,
		"environment": cfg.Core.Environment,
		"hostname":    hostname,
	})
	metrics.Gauge("application_build", map[string]interface{}{
		"build_tag":    xBuildTag,
		"build_commit": xBuildCommit,
	}).Set(1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vmetrics.WritePrometheus(w, true)
		vmetrics.WriteFDMetrics(w)
	})
}

func initTracing(cfg *config.Config) ShutMeDown {
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

	return tp
}

func shutdown(cfg *config.Config, mainCancel context.CancelFunc, closers ...ShutMeDown) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// wait for signal
	sig := <-sigChan
	slog.Info("Received signal, shutting down...", slog.String("signal", sig.String()))

	// allows second signal to bypass graceful shutdown and terminate application immediately
	signal.Stop(sigChan)
	close(sigChan)

	// total timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Core.ShutdownGracePeriod)
	defer cancel()

	done := make(chan struct{})
	go func(ctx context.Context) {
		// Calling cancel() on main context disables health-checks for http and grpc servers.
		mainCancel()
		time.Sleep(cfg.Core.ShutdownWaitForProbe)
		for _, c := range closers {
			if c == nil {
				continue
			}
			cmpCtx, cmpCancel := context.WithTimeout(ctx, cfg.Core.ShutdownComponentTimeout)
			// We assume that Shutdown() is blocking and final operation -
			// that is, when Shutdown() returns, all resources are released
			// or operation timed out, and we should not wait for it anymore.
			if err := c.Shutdown(cmpCtx); err != nil {
				slog.Error("shutdown error", slog.Any("error", err))
			}
			cmpCancel()
		}
		close(done)
	}(ctx)

	select {
	case <-done:
		slog.Info("shutdown completed")
	case <-ctx.Done():
		slog.Info("shutdown timed out")
	}

	// When an application receives a signal (SIGINT/SIGTERM) and catches it using a signal handler,
	// the typical and expected behavior is exiting with a non-zero status.
	// Go runtime(?) will enforce exit code 1 even if os.Exit() is called with a different code.
}

// ShutMeDown implements graceful shutdown for specific component.
type ShutMeDown interface {
	Shutdown(context.Context) error
}

// ShutMeDownWrap wraps io.Closer or plan function and implements ShutMeDown interface.
// Only one of the fields is used: closer or fn.
// If both are nil, Shutdown() returns nil.
type ShutMeDownWrap struct {
	closer io.Closer
	fn     func(ctx context.Context) error
}

// Shutdown implements graceful shutdown for specific component.
// It should be blocking and final.
func (s *ShutMeDownWrap) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		if s.closer != nil {
			done <- s.closer.Close()
		} else if s.fn != nil {
			done <- s.fn(ctx)
		} else {
			done <- nil
		}
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
