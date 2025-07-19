package rabbitmq

import (
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
)

type Option func(*AMQP, *amqp.Config)

func WithLogger(logger *slog.Logger) Option {
	return func(amqp *AMQP, _ *amqp.Config) {
		amqp.logger = logger
	}
}

func WithReconnectBackoffInitialInterval(interval time.Duration) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if cfg.Connection.Reconnect == nil {
			cfg.Connection.Reconnect = &amqp.ReconnectConfig{}
		}
		cfg.Connection.Reconnect.BackoffInitialInterval = interval
	}
}

func WithReconnectBackoffMultiplier(multiplier float64) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if cfg.Connection.Reconnect == nil {
			cfg.Connection.Reconnect = &amqp.ReconnectConfig{}
		}
		cfg.Connection.Reconnect.BackoffMultiplier = multiplier
	}
}

func WithReconnectBackoffMaxInterval(interval time.Duration) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if cfg.Connection.Reconnect == nil {
			cfg.Connection.Reconnect = &amqp.ReconnectConfig{}
		}
		cfg.Connection.Reconnect.BackoffMaxInterval = interval
	}
}

func WithExchangeName(name string) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if name != "" {
			cfg.Exchange.GenerateName = func(topic string) string {
				return name
			}
		}
	}
}

func WithExchangeType(exchangeType string) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Exchange.Type = exchangeType
	}
}

func WithExchangeDurable(durable bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Exchange.Durable = durable
	}
}

func WithExchangeAutoDelete(autoDelete bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Exchange.AutoDeleted = autoDelete
	}
}

func WithQueueName(name string) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if name != "" {
			cfg.Queue.GenerateName = func(topic string) string {
				return name
			}
		}
	}
}

func WithQueueDurable(durable bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Queue.Durable = durable
	}
}

func WithQueueAutoDelete(autoDelete bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Queue.AutoDelete = autoDelete
	}
}

func WithQueueExclusive(exclusive bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Queue.Exclusive = exclusive
	}
}

func WithPublishMandatory(mandatory bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Publish.Mandatory = mandatory
	}
}

func WithPublishImmediate(immediate bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Publish.Immediate = immediate
	}
}

func WithPublishTransactional(transactional bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Publish.Transactional = transactional
	}
}

func WithPublishChannelPoolSize(size int) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Publish.ChannelPoolSize = size
	}
}

func WithPublishConfirmDelivery(confirm bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Publish.ConfirmDelivery = confirm
	}
}

func WithConsumeNoRequeueOnNack(noRequeue bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.NoRequeueOnNack = noRequeue
	}
}

func WithConsumeConsumerName(name string) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if name != "" {
			cfg.Consume.Consumer = name
		}
	}
}

func WithConsumeExclusive(exclusive bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.Exclusive = exclusive
	}
}

func WithConsumeNoLocal(noLocal bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.NoLocal = noLocal
	}
}

func WithConsumeQosPrefetchCount(count int) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.Qos.PrefetchCount = count
	}
}

func WithConsumeQosPrefetchSize(size int) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.Qos.PrefetchSize = size
	}
}

func WithConsumeQosGlobal(global bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		cfg.Consume.Qos.Global = global
	}
}

func WithNotPersistentDeliveryMode(notPersistent bool) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if m, ok := cfg.Marshaler.(amqp.DefaultMarshaler); ok {
			m.NotPersistentDeliveryMode = notPersistent
			cfg.Marshaler = m
		}
	}
}

func WithMessageUUIDHeaderKey(headerKey string) Option {
	return func(_ *AMQP, cfg *amqp.Config) {
		if headerKey != "" {
			if m, ok := cfg.Marshaler.(amqp.DefaultMarshaler); ok {
				m.MessageUUIDHeaderKey = headerKey
				cfg.Marshaler = m
			}
		}
	}
}
