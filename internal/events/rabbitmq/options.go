package rabbitmq

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
)

type Option func(*AMQP, *amqp.Config)

func WithLogger(logger *slog.Logger) Option {
	return func(amqp *AMQP, _ *amqp.Config) {
		amqp.logger = logger
	}
}
