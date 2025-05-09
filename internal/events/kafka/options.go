package kafka

import (
	"log/slog"

	"github.com/IBM/sarama"
)

type Option func(*Kafka, *sarama.Config, *sarama.Config)

func WithLogger(logger *slog.Logger) Option {
	return func(k *Kafka, pubCfg *sarama.Config, subCfg *sarama.Config) {
		k.logger = logger
	}
}
