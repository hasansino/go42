package kafka

import (
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

type Option func(*Kafka, *sarama.Config, *sarama.Config)

func WithLogger(logger *slog.Logger) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		k.logger = logger
	}
}

func WithClientID(clientID string) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if clientID != "" {
			pubCfg.ClientID = clientID
			subCfg.ClientID = clientID
		}
	}
}

func WithDialTimeout(timeout time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Net.DialTimeout = timeout
		subCfg.Net.DialTimeout = timeout
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Net.ReadTimeout = timeout
		subCfg.Net.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Net.WriteTimeout = timeout
		subCfg.Net.WriteTimeout = timeout
	}
}

func WithKeepAlive(keepAlive time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if keepAlive > 0 {
			pubCfg.Net.KeepAlive = keepAlive
			subCfg.Net.KeepAlive = keepAlive
		}
	}
}

func WithProducerRetryMax(retries int) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.Retry.Max = retries
	}
}

func WithProducerRetryBackoff(backoff time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.Retry.Backoff = backoff
	}
}

func WithProducerMaxMessageBytes(bytes int) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.MaxMessageBytes = bytes
	}
}

func WithProducerCompressionLevel(level int) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.CompressionLevel = level
	}
}

func WithProducerCompression(compression string) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		switch compression {
		case "none":
			pubCfg.Producer.Compression = sarama.CompressionNone
		case "gzip":
			pubCfg.Producer.Compression = sarama.CompressionGZIP
		case "snappy":
			pubCfg.Producer.Compression = sarama.CompressionSnappy
		case "lz4":
			pubCfg.Producer.Compression = sarama.CompressionLZ4
		case "zstd":
			pubCfg.Producer.Compression = sarama.CompressionZSTD
		}
	}
}

func WithProducerFlushMessages(messages int) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if messages > 0 {
			pubCfg.Producer.Flush.Messages = messages
		}
	}
}

func WithProducerFlushFrequency(frequency time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if frequency > 0 {
			pubCfg.Producer.Flush.Frequency = frequency
		}
	}
}

func WithProducerRequiredAcks(acks int16) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.RequiredAcks = sarama.RequiredAcks(acks)
	}
}

func WithProducerIdempotent(idempotent bool) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Producer.Idempotent = idempotent
	}
}

func WithConsumerRetryBackoff(backoff time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Retry.Backoff = backoff
	}
}

func WithConsumerFetchMin(bytes int32) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Fetch.Min = bytes
	}
}

func WithConsumerFetchDefault(bytes int32) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Fetch.Default = bytes
	}
}

func WithConsumerFetchMax(bytes int32) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if bytes > 0 {
			subCfg.Consumer.Fetch.Max = bytes
		}
	}
}

func WithConsumerMaxWaitTime(duration time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.MaxWaitTime = duration
	}
}

func WithConsumerMaxProcessingTime(duration time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if duration > 0 {
			subCfg.Consumer.MaxProcessingTime = duration
		}
	}
}

func WithConsumerReturnErrors(returnErrors bool) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Return.Errors = returnErrors
	}
}

func WithConsumerOffsetInitial(offset int64) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Offsets.Initial = offset
	}
}

func WithConsumerGroupSessionTimeout(timeout time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Group.Session.Timeout = timeout
	}
}

func WithConsumerGroupHeartbeatInterval(interval time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		subCfg.Consumer.Group.Heartbeat.Interval = interval
	}
}

func WithConsumerGroupRebalanceStrategy(strategy string) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		switch strategy {
		case "range":
			subCfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
		case "roundrobin":
			subCfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
		case "sticky":
			subCfg.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategySticky()
		}
	}
}

func WithMetadataRefreshFrequency(frequency time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Metadata.RefreshFrequency = frequency
		subCfg.Metadata.RefreshFrequency = frequency
	}
}

func WithMetadataRetryMax(retries int) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Metadata.Retry.Max = retries
		subCfg.Metadata.Retry.Max = retries
	}
}

func WithMetadataRetryBackoff(backoff time.Duration) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		pubCfg.Metadata.Retry.Backoff = backoff
		subCfg.Metadata.Retry.Backoff = backoff
	}
}

func WithKafkaVersion(version string) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		if version != "" {
			if v, err := sarama.ParseKafkaVersion(version); err == nil {
				pubCfg.Version = v
				subCfg.Version = v
			}
		}
	}
}
