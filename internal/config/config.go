package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v11"

	"github.com/hasansino/goapp/internal/tools"
)

type Config struct {
	sync.RWMutex
	Core       Core
	Limits     Limits
	Logger     Logger
	Metrics    Metrics
	Tracing    Tracing
	Sentry     Sentry
	Vault      Vault
	Etcd       Etcd
	Database   Database
	Cache      Cache
	Events     Events
	Pprof      Pprof
	HTTPServer Server
	GRPCServer GRPC
}

// ╭──────────────────────────────╮
// │             CORE             │
// ╰──────────────────────────────╯

type Core struct {
	ServiceName              string        `env:"SERVICE_NAME"               default:"{{SERVICE_NAME}}"`
	Environment              string        `env:"ENVIRONMENT"                default:""`
	ShutdownGracePeriod      time.Duration `env:"SHUTDOWN_GRACE_PERIOD"      default:"10s"`
	ShutdownWaitForProbe     time.Duration `env:"SHUTDOWN_WAIT_FOR_PROBE"    default:"2s"`
	ShutdownComponentTimeout time.Duration `env:"SHUTDOWN_COMPONENT_TIMEOUT" default:"3s"`
}

// ╭──────────────────────────────╮
// │            LIMITS            │
// ╰──────────────────────────────╯

type Limits struct {
	AutoMaxProcsEnabled bool    `env:"AUTOMAXPROCS_ENABLED" default:"false"`
	MinMaxProcs         int     `env:"MIN_MAXPROCS"         default:"1"`
	AutoMemLimitEnabled bool    `env:"AUTOMEMLIMIT_ENABLED" default:"false"`
	MemLimitRatio       float64 `env:"MEMLIMIT_RATIO"       default:"0.9"   v:"gte=0.2,lte=1.0"`
}

// ╭──────────────────────────────╮
// │            LOGGER            │
// ╰──────────────────────────────╯

type Logger struct {
	LogLevel  string `env:"LOG_LEVEL"  default:"info"   v:"oneof=debug info warn error"`
	LogOutput string `env:"LOG_OUTPUT" default:"stdout" v:"oneof=none stdout stderr file"`
	LogFormat string `env:"LOG_FORMAT" default:"json"   v:"oneof=json text"`
}

func (l *Logger) Level() slog.Level {
	logLevel := strings.ToLower(l.LogLevel)

	modifierIndex := -1
	modifierSign := 0

	if idx := strings.IndexByte(logLevel, '+'); idx != -1 {
		modifierIndex = idx
		modifierSign = 1
	} else if idx := strings.IndexByte(logLevel, '-'); idx != -1 {
		modifierIndex = idx
		modifierSign = -1
	}

	var baseLevel slog.Level
	var extraLevel int

	if modifierIndex != -1 {
		if val, err := strconv.Atoi(logLevel[modifierIndex+1:]); err == nil {
			extraLevel = val
		}
		logLevel = logLevel[:modifierIndex]
	}

	// Get base log level
	switch logLevel {
	case "debug":
		baseLevel = slog.LevelDebug
	case "info":
		baseLevel = slog.LevelInfo
	case "warn":
		baseLevel = slog.LevelWarn
	case "error":
		baseLevel = slog.LevelError
	default:
		return slog.LevelInfo
	}

	if modifierSign != 0 {
		return baseLevel + slog.Level(modifierSign*extraLevel)
	}

	return baseLevel
}

// ╭──────────────────────────────╮
// │           METRICS            │
// ╰──────────────────────────────╯

type Metrics struct {
	Timeout time.Duration `env:"METRICS_HANDLER_TIMEOUT" default:"10s"`
}

// ╭──────────────────────────────╮
// │           TRACING            │
// ╰──────────────────────────────╯

type Tracing struct {
	Enable  bool          `env:"TRACING_ENABLED" default:"false"`
	DSN     string        `env:"TRACING_DSN"     default:""`
	Timeout time.Duration `env:"TRACING_TIMEOUT" default:"5s"`
}

// ╭──────────────────────────────╮
// │            SENTRY            │
// ╰──────────────────────────────╯

type Sentry struct {
	Enabled    bool    `env:"SENTRY_ENABLED"     default:"false"`
	DSN        string  `env:"SENTRY_DSN"         default:""`
	Debug      bool    `env:"SENTRY_DEBUG"       default:"false"`
	Stacktrace bool    `env:"SENTRY_STACKTRACE"  default:"false"`
	SampleRate float64 `env:"SENTRY_SAMPLE_RATE" default:"1.0"   v:"gte=0.0,lte=1.0"`
}

// ╭──────────────────────────────╮
// │             VAULT            │
// ╰──────────────────────────────╯

type Vault struct {
	Enabled    bool   `env:"VAULT_ENABLED"     default:"false"`
	Host       string `env:"VAULT_HOST"        default:"http://localhost:8200"`
	AuthType   string `env:"VAULT_AUTH_TYPE"   default:"token"`
	Token      string `env:"VAULT_TOKEN"       default:"qwerty"`
	SecretPath string `env:"VAULT_SECRET_PATH" default:"/secret/data/github.com/hasansino/goapp"`
}

// ╭──────────────────────────────╮
// │             ETCD             │
// ╰──────────────────────────────╯

type Etcd struct {
	Enabled      bool          `env:"ETCD_ENABLED"       default:"false"`
	Hosts        []string      `env:"ETCD_HOST"          default:"localhost:2379"`
	Timeout      time.Duration `env:"ETCD_TIMEOUT"       default:"5s"`
	Method       string        `env:"ETCD_METHOD"        default:"bind"`
	SyncInterval time.Duration `env:"ETCD_SYNC_INTERVAL" default:"5m"`
}

// ╭──────────────────────────────╮
// │           DATABASE           │
// ╰──────────────────────────────╯

type Database struct {
	Engine      string `env:"DATABASE_ENGINE"       default:"sqlite"   v:"oneof=sqlite pgsql"`
	MigratePath string `env:"DATABASE_MIGRATE_PATH" default:"/migrate"`
	LogQueries  bool   `env:"DATABASE_LOG_QUERIES"  default:"false"`
	Pgsql       Pgsql
	Sqlite      Sqlite
}

func (db Database) FullMigratePath() string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(db.MigratePath, "/"),
		db.Engine,
	)
}

type Pgsql struct {
	Host            string        `env:"DATABASE_PGSQL_HOST"               default:"localhost"`
	Port            int           `env:"DATABASE_PGSQL_PORT"               default:"5432"`
	User            string        `env:"DATABASE_PGSQL_USER"               default:"user"`
	Password        string        `env:"DATABASE_PGSQL_PASSWORD"           default:"qwerty"`
	Name            string        `env:"DATABASE_PGSQL_NAME"               default:"goapp"`
	ConnMaxIdleTime time.Duration `env:"DATABASE_PGSQL_CONN_MAX_IDLE_TIME" default:"10m"`
	ConnMaxLifetime time.Duration `env:"DATABASE_PGSQL_CONN_MAX_LIFETIME"  default:"30m"`
	MaxIdleConns    int           `env:"DATABASE_PGSQL_MAX_IDLE_CONNS"     default:"10"`
	MaxOpenConns    int           `env:"DATABASE_PGSQL_MAX_OPEN_CONNS"     default:"100"`
	QueryTimeout    time.Duration `env:"DATABASE_PGSQL_QUERY_TIMEOUT"      default:"10s"`
}

func (db Pgsql) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

type Sqlite struct {
	Mode       string `env:"DATABASE_SQLITE_MODE"       default:"memory"`
	SqliteFile string `env:"DATABASE_SQLITE_PATH"       default:"file::memory:"`
	CacheMode  string `env:"DATABASE_SQLITE_CACHE_MODE" default:"shared"`
}

// ╭──────────────────────────────╮
// │            CACHE             │
// ╰──────────────────────────────╯

type Cache struct {
	Engine    string `env:"CACHE_ENGINE" default:"none" v:"oneof=none redis miniredis memcached"`
	Redis     Redis
	Memcached Memcached
}

type Redis struct {
	Host                  string        `env:"CACHE_REDIS_HOST"                    default:"localhost:6379"`
	DB                    int           `env:"CACHE_REDIS_DB"                      default:"0"`
	Username              string        `env:"CACHE_REDIS_USERNAME"                default:""`
	Password              string        `env:"CACHE_REDIS_PASSWORD"                default:""`
	MaxRetries            int           `env:"CACHE_REDIS_MAX_RETRIES"             default:"3"`
	MinRetryBackoff       time.Duration `env:"CACHE_REDIS_MIN_RETRY_BACKOFF"       default:"8ms"`
	MaxRetryBackoff       time.Duration `env:"CACHE_REDIS_MAX_RETRY_BACKOFF"       default:"512ms"`
	DialTimeout           time.Duration `env:"CACHE_REDIS_DIAL_TIMEOUT"            default:"5s"`
	ReadTimeout           time.Duration `env:"CACHE_REDIS_READ_TIMEOUT"            default:"3s"`
	WriteTimeout          time.Duration `env:"CACHE_REDIS_WRITE_TIMEOUT"           default:"3s"`
	ContextTimeoutEnabled bool          `env:"CACHE_REDIS_CONTEXT_TIMEOUT_ENABLED" default:"true"`
	PoolSize              int           `env:"CACHE_REDIS_POOL_SIZE"               default:"10"`
	PoolTimeout           time.Duration `env:"CACHE_REDIS_POOL_TIMEOUT"            default:"4s"`
	MinIdleConns          int           `env:"CACHE_REDIS_MIN_IDLE_CONNS"          default:"0"`
	MaxIdleConns          int           `env:"CACHE_REDIS_MAX_IDLE_CONNS"          default:"0"`
	MaxActiveConns        int           `env:"CACHE_REDIS_MAX_ACTIVE_CONNS"        default:"0"`
	ConnMaxIdleTime       time.Duration `env:"CACHE_REDIS_CONN_MAX_IDLE_TIME"      default:"0s"`
	ConnMaxLifetime       time.Duration `env:"CACHE_REDIS_CONN_MAX_LIFETIME"       default:"0s"`
}

type Memcached struct {
	Hosts        []string      `env:"CACHE_MEMCACHED_HOSTS"          default:"localhost:11211"`
	Timeout      time.Duration `env:"CACHE_MEMCACHED_TIMEOUT"        default:"1s"`
	MaxIdleConns int           `env:"CACHE_MEMCACHED_MAX_IDLE_CONNS" default:"100"`
}

// ╭──────────────────────────────╮
// │           EVENTS             │
// ╰──────────────────────────────╯

type Events struct {
	Engine   string `env:"EVENTS_ENGINE" default:"none" v:"oneof=none gochan nats rabbitmq kafka"`
	NATS     NATS
	RabbitMQ RabbitMQ
	Kafka    Kafka
}

type NATS struct {
	DSN         string        `env:"NATS_DSN"          default:"nats://localhost:4222"`
	ClientName  string        `env:"NATS_CLIENT_NAME"  default:""`
	Token       string        `env:"NATS_TOKEN"        default:""`
	ConnTimeout time.Duration `env:"NATS_CONN_TIMEOUT" default:"5s"`
	ConnRetry   bool          `env:"NATS_CONN_RETRY"   default:"false"`
	MaxRetry    int           `env:"NATS_MAX_RETRY"    default:"10"`
	RetryDelay  time.Duration `env:"NATS_RETRY_DELAY"  default:"1s"`
	Subscriber  NATSSubscriber
}

type NATSSubscriber struct {
	GroupPrefix  string        `env:"NATS_SUB_QUEUE_GROUP_PREFIX" default:""`
	WorkerCount  int           `env:"NATS_SUB_WORKER_COUNT"       default:"1"`
	Timeout      time.Duration `env:"NATS_SUB_TIMEOUT"            default:"30s"`
	AckTimeout   time.Duration `env:"NATS_SUB_ACK_TIMEOUT"        default:"30s"`
	CloseTimeout time.Duration `env:"NATS_SUB_CLOSE_TIMEOUT"      default:"30s"`
}

type RabbitMQ struct {
	DSN string `env:"RABBITMQ_DSN" default:"amqp://guest:guest@localhost:5672/ "`
}

type Kafka struct {
	Brokers       []string `env:"KAFKA_BROKERS"        default:"localhost:9092"`
	ConsumerGroup string   `env:"KAFKA_CONSUMER_GROUP" default:"example"`
}

// ╭──────────────────────────────╮
// │            PPROF             │
// ╰──────────────────────────────╯

type Pprof struct {
	Enabled      bool          `env:"PPROF_ENABLED"       default:"false"`
	Listen       string        `env:"PPROF_LISTEN"        default:":6060"`
	Prefix       string        `env:"PPROF_PREFIX"        default:"/debug/pprof"`
	ReadTimeout  time.Duration `env:"PPROF_READ_TIMEOUT"  default:"5s"`
	WriteTimeout time.Duration `env:"PPROF_WRITE_TIMEOUT" default:"60s"`
}

// ╭──────────────────────────────╮
// │            SERVER            │
// ╰──────────────────────────────╯

type Server struct {
	Listen       string        `env:"SERVER_LISTEN"        default:":8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT"  default:"5s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" default:"5s"`
	StaticRoot   string        `env:"SERVER_STATIC_ROOT"   default:"/usr/share/www"`
	SwaggerRoot  string        `env:"SERVER_SWAGGER_ROOT"  default:"/usr/share/www/api"`
}

// ╭──────────────────────────────╮
// │             GRPCServer             │
// ╰──────────────────────────────╯

type GRPC struct {
	Listen         string `env:"GRPC_LISTEN"            default:":50051"`
	MaxRecvMsgSize int    `env:"GRPC_MAX_RECV_MSG_SIZE" default:"1024"`
	MaxSendMsgSize int    `env:"GRPC_MAX_SEND_MSG_SIZE" default:"1024"`
}

const (
	TagNameEnvVarName   = "env"
	TagNameDefaultValue = "default"
)

func New() (*Config, error) {
	cfg := new(Config)
	err := env.ParseWithOptions(cfg, env.Options{
		TagName:             TagNameEnvVarName,
		DefaultValueTagName: TagNameDefaultValue,
	})
	if err != nil {
		return nil, err
	}

	vErrs := tools.ValidateStruct(cfg)
	if len(vErrs) > 0 {
		var line string
		for _, vErr := range vErrs {
			line += vErr.Compact() + ","
		}
		line = strings.TrimSuffix(line, ",")
		return nil, fmt.Errorf("validation errors: %s", line)
	}

	return cfg, nil
}

func (c *Config) String() string {
	if out, err := json.MarshalIndent(&c, "", "  "); err == nil {
		return string(out)
	}
	return ""
}
