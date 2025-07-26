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

	"github.com/hasansino/go42/internal/tools"
)

type Config struct {
	sync.RWMutex
	Core     Core
	Limits   Limits
	Logger   Logger
	Tracing  Tracing
	Sentry   Sentry
	Vault    Vault
	Etcd     Etcd
	Database Database
	Cache    Cache
	Events   Events
	Pprof    Pprof
	Server   Server
	Outbox   Outbox
	Auth     Auth
}

// ╭──────────────────────────────╮
// │             CORE             │
// ╰──────────────────────────────╯

type Core struct {
	ServiceName              string        `env:"SERVICE_NAME"               default:"go42"`
	Environment              string        `env:"ENVIRONMENT"                default:"default"`
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
	LogLevel  string `env:"LOG_LEVEL"      default:"info"   v:"oneof=debug info warn error"`
	LogOutput string `env:"LOG_OUTPUT"     default:"stdout" v:"oneof=none stdout stderr file"`
	LogFormat string `env:"LOG_FORMAT"     default:"json"   v:"oneof=json text tint"`
	AddSource bool   `env:"LOG_ADD_SOURCE" default:"true"`
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
	Enabled    bool          `env:"VAULT_ENABLED"     default:"false"`
	Host       string        `env:"VAULT_HOST"        default:"http://localhost:8200"`
	Timeout    time.Duration `env:"VAULT_TIMEOUT"     default:"5s"`
	AuthType   string        `env:"VAULT_AUTH_TYPE"   default:"token"`
	Token      string        `env:"VAULT_TOKEN"       default:"qwerty"`
	SecretPath string        `env:"VAULT_SECRET_PATH" default:"/secret/data/github.com/hasansino/go42"`
	MountPath  string        `env:"VAULT_MOUNT_PATH"  default:"secret"`
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
	Engine      string `env:"DATABASE_ENGINE"       default:"sqlite"   v:"oneof=sqlite pgsql mysql"`
	MigratePath string `env:"DATABASE_MIGRATE_PATH" default:"/migrate"`
	LogQueries  bool   `env:"DATABASE_LOG_QUERIES"  default:"false"`
	Sqlite      Sqlite
	Pgsql       Pgsql
	Mysql       Mysql
}

func (db Database) FullMigratePath() string {
	return fmt.Sprintf(
		"%s/%s",
		strings.TrimRight(db.MigratePath, "/"),
		db.Engine,
	)
}

type Sqlite struct {
	Mode       string `env:"DATABASE_SQLITE_MODE"       default:"memory"`
	SqliteFile string `env:"DATABASE_SQLITE_PATH"       default:"file::memory:"`
	CacheMode  string `env:"DATABASE_SQLITE_CACHE_MODE" default:"shared"`
}

type Mysql struct {
	Master          MysqlMaster
	Slave           MysqlSlave
	ConnMaxIdleTime time.Duration `env:"DATABASE_MYSQL_CONN_MAX_IDLE_TIME" default:"10m"`
	ConnMaxLifetime time.Duration `env:"DATABASE_MYSQL_CONN_MAX_LIFETIME"  default:"30m"`
	MaxIdleConns    int           `env:"DATABASE_MYSQL_MAX_IDLE_CONNS"     default:"10"`
	MaxOpenConns    int           `env:"DATABASE_MYSQL_MAX_OPEN_CONNS"     default:"100"`
	QueryTimeout    time.Duration `env:"DATABASE_MYSQL_QUERY_TIMEOUT"      default:"10s"`
}

type MysqlMaster struct {
	Host     string `env:"DATABASE_MYSQL_MASTER_HOST"     default:"localhost"`
	Port     int    `env:"DATABASE_MYSQL_MASTER_PORT"     default:"3306"`
	User     string `env:"DATABASE_MYSQL_MASTER_USER"     default:"user"`
	Password string `env:"DATABASE_MYSQL_MASTER_PASSWORD" default:"qwerty"`
	Charset  string `env:"DATABASE_MYSQL_MASTER_CHARSET"  default:"utf8mb4"`
	Name     string `env:"DATABASE_MYSQL_MASTER_NAME"     default:"go42"`
}

func (db MysqlMaster) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=UTC",
		db.User, db.Password, db.Host, db.Port, db.Name, db.Charset,
	)
}

type MysqlSlave struct {
	Host     string `env:"DATABASE_MYSQL_SLAVE_HOST"     default:""`
	Port     int    `env:"DATABASE_MYSQL_SLAVE_PORT"     default:"3306"`
	User     string `env:"DATABASE_MYSQL_SLAVE_USER"     default:"user"`
	Password string `env:"DATABASE_MYSQL_SLAVE_PASSWORD" default:"qwerty"`
	Charset  string `env:"DATABASE_MYSQL_SLAVE_CHARSET"  default:"utf8mb4"`
	Name     string `env:"DATABASE_MYSQL_SLAVE_NAME"     default:"go42"`
}

func (db MysqlSlave) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=UTC",
		db.User, db.Password, db.Host, db.Port, db.Name, db.Charset,
	)
}

type Pgsql struct {
	Master          PgsqlMaster
	Slave           PgsqlSlave
	ConnMaxIdleTime time.Duration `env:"DATABASE_PGSQL_CONN_MAX_IDLE_TIME" default:"10m"`
	ConnMaxLifetime time.Duration `env:"DATABASE_PGSQL_CONN_MAX_LIFETIME"  default:"30m"`
	MaxIdleConns    int           `env:"DATABASE_PGSQL_MAX_IDLE_CONNS"     default:"10"`
	MaxOpenConns    int           `env:"DATABASE_PGSQL_MAX_OPEN_CONNS"     default:"100"`
	QueryTimeout    time.Duration `env:"DATABASE_PGSQL_QUERY_TIMEOUT"      default:"10s"`
}

type PgsqlMaster struct {
	Host     string `env:"DATABASE_PGSQL_MASTER_HOST"     default:"localhost"`
	Port     int    `env:"DATABASE_PGSQL_MASTER_PORT"     default:"5432"`
	User     string `env:"DATABASE_PGSQL_MASTER_USER"     default:"user"`
	Password string `env:"DATABASE_PGSQL_MASTER_PASSWORD" default:"qwerty"`
	Name     string `env:"DATABASE_PGSQL_MASTER_NAME"     default:"go42"`
}

func (db PgsqlMaster) DSN() string {
	if db.Host == "" {
		return ""
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

type PgsqlSlave struct {
	Host     string `env:"DATABASE_PGSQL_SLAVE_HOST"     default:""`
	Port     int    `env:"DATABASE_PGSQL_SLAVE_PORT"     default:"5432"`
	User     string `env:"DATABASE_PGSQL_SLAVE_USER"     default:"user"`
	Password string `env:"DATABASE_PGSQL_SLAVE_PASSWORD" default:"qwerty"`
	Name     string `env:"DATABASE_PGSQL_SLAVE_NAME"     default:"go42"`
}

func (db PgsqlSlave) DSN() string {
	if db.Host == "" {
		return ""
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

// ╭──────────────────────────────╮
// │            CACHE             │
// ╰──────────────────────────────╯

type Cache struct {
	Engine    string `env:"CACHE_ENGINE" default:"none" v:"oneof=none otter memcached redis aerospike"`
	Redis     Redis
	Memcached Memcached
	Aerospike Aerospike
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

type Aerospike struct {
	Hosts     []string `env:"CACHE_AEROSPIKE_HOSTS"     default:"localhost:3000"`
	Namespace string   `env:"CACHE_AEROSPIKE_NAMESPACE" default:"go42"`
}

// ╭──────────────────────────────╮
// │           EVENTS             │
// ╰──────────────────────────────╯

type Events struct {
	Engine   string `env:"EVENTS_ENGINE" default:"gochan" v:"oneof=none gochan nats rabbitmq kafka"`
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
	DSN                      string        `env:"RABBITMQ_DSN"                        default:"amqp://guest:guest@localhost:5672/"`
	ReconnectInitialInterval time.Duration `env:"RABBITMQ_RECONNECT_INITIAL_INTERVAL" default:"500ms"`
	ReconnectMultiplier      float64       `env:"RABBITMQ_RECONNECT_MULTIPLIER"       default:"1.5"`
	ReconnectMaxInterval     time.Duration `env:"RABBITMQ_RECONNECT_MAX_INTERVAL"     default:"30s"`
	ExchangeName             string        `env:"RABBITMQ_EXCHANGE_NAME"              default:""`
	ExchangeType             string        `env:"RABBITMQ_EXCHANGE_TYPE"              default:"topic"`
	ExchangeDurable          bool          `env:"RABBITMQ_EXCHANGE_DURABLE"           default:"true"`
	ExchangeAutoDelete       bool          `env:"RABBITMQ_EXCHANGE_AUTO_DELETE"       default:"false"`
	QueueName                string        `env:"RABBITMQ_QUEUE_NAME"                 default:""`
	QueueDurable             bool          `env:"RABBITMQ_QUEUE_DURABLE"              default:"true"`
	QueueAutoDelete          bool          `env:"RABBITMQ_QUEUE_AUTO_DELETE"          default:"false"`
	QueueExclusive           bool          `env:"RABBITMQ_QUEUE_EXCLUSIVE"            default:"false"`
	PublishMandatory         bool          `env:"RABBITMQ_PUBLISH_MANDATORY"          default:"false"`
	PublishImmediate         bool          `env:"RABBITMQ_PUBLISH_IMMEDIATE"          default:"false"`
	PublishTransactional     bool          `env:"RABBITMQ_PUBLISH_TRANSACTIONAL"      default:"false"`
	PublishChannelPool       int           `env:"RABBITMQ_PUBLISH_CHANNEL_POOL"       default:"5"`
	PublishConfirm           bool          `env:"RABBITMQ_PUBLISH_CONFIRM"            default:"false"`
	ConsumeNoRequeue         bool          `env:"RABBITMQ_CONSUME_NO_REQUEUE"         default:"false"`
	ConsumeConsumerName      string        `env:"RABBITMQ_CONSUME_CONSUMER_NAME"      default:""`
	ConsumeExclusive         bool          `env:"RABBITMQ_CONSUME_EXCLUSIVE"          default:"false"`
	ConsumeNoLocal           bool          `env:"RABBITMQ_CONSUME_NO_LOCAL"           default:"false"`
	ConsumePrefetchCount     int           `env:"RABBITMQ_CONSUME_PREFETCH_COUNT"     default:"1"`
	ConsumePrefetchSize      int           `env:"RABBITMQ_CONSUME_PREFETCH_SIZE"      default:"0"`
	ConsumeQosGlobal         bool          `env:"RABBITMQ_CONSUME_QOS_GLOBAL"         default:"false"`
	NotPersistentMode        bool          `env:"RABBITMQ_NOT_PERSISTENT_MODE"        default:"false"`
	MessageUUIDHeader        string        `env:"RABBITMQ_MESSAGE_UUID_HEADER"        default:""`
}

type Kafka struct {
	Brokers                   []string      `env:"KAFKA_BROKERS"                      default:"localhost:9092"`
	ConsumerGroup             string        `env:"KAFKA_CONSUMER_GROUP"               default:"example"`
	ClientID                  string        `env:"KAFKA_CLIENT_ID"                    default:""`
	Version                   string        `env:"KAFKA_VERSION"                      default:"4.0.0"`
	DialTimeout               time.Duration `env:"KAFKA_DIAL_TIMEOUT"                 default:"30s"`
	ReadTimeout               time.Duration `env:"KAFKA_READ_TIMEOUT"                 default:"30s"`
	WriteTimeout              time.Duration `env:"KAFKA_WRITE_TIMEOUT"                default:"30s"`
	KeepAlive                 time.Duration `env:"KAFKA_KEEP_ALIVE"                   default:"0s"`
	ProducerRetryMax          int           `env:"KAFKA_PRODUCER_RETRY_MAX"           default:"3"`
	ProducerRetryBackoff      time.Duration `env:"KAFKA_PRODUCER_RETRY_BACKOFF"       default:"100ms"`
	ProducerMaxMessageBytes   int           `env:"KAFKA_PRODUCER_MAX_MESSAGE_BYTES"   default:"1000000"`
	ProducerCompression       string        `env:"KAFKA_PRODUCER_COMPRESSION"         default:"none"`
	ProducerCompressionLevel  int           `env:"KAFKA_PRODUCER_COMPRESSION_LEVEL"   default:"-1"`
	ProducerFlushMessages     int           `env:"KAFKA_PRODUCER_FLUSH_MESSAGES"      default:"0"`
	ProducerFlushFrequency    time.Duration `env:"KAFKA_PRODUCER_FLUSH_FREQUENCY"     default:"0s"`
	ProducerRequiredAcks      int16         `env:"KAFKA_PRODUCER_REQUIRED_ACKS"       default:"1"`
	ProducerIdempotent        bool          `env:"KAFKA_PRODUCER_IDEMPOTENT"          default:"false"`
	ConsumerRetryBackoff      time.Duration `env:"KAFKA_CONSUMER_RETRY_BACKOFF"       default:"2s"`
	ConsumerFetchMin          int32         `env:"KAFKA_CONSUMER_FETCH_MIN"           default:"1"`
	ConsumerFetchDefault      int32         `env:"KAFKA_CONSUMER_FETCH_DEFAULT"       default:"1048576"`
	ConsumerFetchMax          int32         `env:"KAFKA_CONSUMER_FETCH_MAX"           default:"0"`
	ConsumerMaxWaitTime       time.Duration `env:"KAFKA_CONSUMER_MAX_WAIT_TIME"       default:"250ms"`
	ConsumerMaxProcessingTime time.Duration `env:"KAFKA_CONSUMER_MAX_PROCESSING_TIME" default:"100ms"`
	ConsumerReturnErrors      bool          `env:"KAFKA_CONSUMER_RETURN_ERRORS"       default:"false"`
	ConsumerOffsetInitial     int64         `env:"KAFKA_CONSUMER_OFFSET_INITIAL"      default:"-1"`
	ConsumerSessionTimeout    time.Duration `env:"KAFKA_CONSUMER_SESSION_TIMEOUT"     default:"10s"`
	ConsumerHeartbeatInterval time.Duration `env:"KAFKA_CONSUMER_HEARTBEAT_INTERVAL"  default:"3s"`
	ConsumerRebalanceStrategy string        `env:"KAFKA_CONSUMER_REBALANCE_STRATEGY"  default:"range"`
	MetadataRefreshFrequency  time.Duration `env:"KAFKA_METADATA_REFRESH_FREQUENCY"   default:"10m"`
	MetadataRetryMax          int           `env:"KAFKA_METADATA_RETRY_MAX"           default:"3"`
	MetadataRetryBackoff      time.Duration `env:"KAFKA_METADATA_RETRY_BACKOFF"       default:"250ms"`
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
	HTTP HTTP
	GRPC GRPC
}

type HTTP struct {
	Listen       string        `env:"SERVER_HTTP_LISTEN"        default:":8080"`
	ReadTimeout  time.Duration `env:"SERVER_HTTP_READ_TIMEOUT"  default:"5s"`
	WriteTimeout time.Duration `env:"SERVER_HTTP_WRITE_TIMEOUT" default:"5s"`
	StaticRoot   string        `env:"SERVER_HTTP_STATIC_ROOT"   default:"/usr/share/www"`
	SwaggerRoot  string        `env:"SERVER_HTTP_SWAGGER_ROOT"  default:"/usr/share/www/api"`
	SwaggerDark  bool          `env:"SERVER_HTTP_SWAGGER_DARK"  default:"true"`
	BodyLimitKB  int           `env:"SERVER_HTTP_BODY_LIMIT_KB" default:"1024"`
	RateLimiter  HTTPRateLimiter
}

type HTTPRateLimiter struct {
	Enabled bool `env:"SERVER_HTTP_RATE_LIMITER_ENABLED" default:"false"`
	Rate    int  `env:"SERVER_HTTP_RATE_LIMITER_RATE"    default:"100"`
	Burst   int  `env:"SERVER_HTTP_RATE_LIMITER_BURST"   default:"10"`
}

type GRPC struct {
	Listen            string `env:"SERVER_GRPC_LISTEN"             default:":50051"`
	MaxRecvMsgSize    int    `env:"SERVER_GRPC_MAX_RECV_MSG_SIZE"  default:"1024"`
	MaxSendMsgSize    int    `env:"SERVER_GRPC_MAX_SEND_MSG_SIZE"  default:"1024"`
	ReflectionEnabled bool   `env:"SERVER_GRPC_REFLECTION_ENABLED" default:"false"`
	RateLimiter       GRPCRateLimiter
}

type GRPCRateLimiter struct {
	Enabled bool `env:"SERVER_GRPC_RATE_LIMITER_ENABLED" default:"false"`
	Rate    int  `env:"SERVER_GRPC_RATE_LIMITER_RATE"    default:"100"`
	Burst   int  `env:"SERVER_GRPC_RATE_LIMITER_BURST"   default:"10"`
}

// ╭──────────────────────────────╮
// │            OUTBOX            │
// ╰──────────────────────────────╯

type Outbox struct {
	WorkerRunInterval time.Duration `env:"OUTBOX_WORKER_INTERVAL"   default:"5s"`
	WorkerBatchSize   int           `env:"OUTBOX_WORKER_BATCH_SIZE" default:"1000"`
}

// ╭──────────────────────────────╮
// │             AUTH             │
// ╰──────────────────────────────╯

type Auth struct {
	JWTSecret            string        `env:"AUTH_JWT_SECRET"             default:"0128899"`
	JWTAccessTokenTTL    time.Duration `env:"AUTH_JWT_ACCESS_TOKEN_TTL"   default:"15m"`
	JWTRefreshTokenTTL   time.Duration `env:"AUTH_JWT_REFRESH_TOKEN_TTL"  default:"168h"`
	JWTIssuer            string        `env:"AUTH_JWT_ISSUER"             default:"go42"`
	JWTAudience          []string      `env:"AUTH_JWT_AUDIENCE"           default:"go42"`
	APICacheTTL          time.Duration `env:"AUTH_API_CACHE_TTL"          default:"60m"`
	TokenUpdaterInterval time.Duration `env:"AUTH_TOKEN_UPDATER_INTERVAL" default:"5m"`
}

// ---

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

	return cfg, tools.ValidateStructCompact(cfg)
}

func (c *Config) String() string {
	if out, err := json.MarshalIndent(&c, "", "  "); err == nil {
		return string(out)
	}
	return ""
}
