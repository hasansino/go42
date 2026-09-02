package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	wkafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Kafka struct {
	logger     *slog.Logger
	publisher  *wkafka.Publisher
	subscriber *wkafka.Subscriber
}

func New(brokers []string, group string, opts ...Option) (*Kafka, error) {
	var (
		engine = new(Kafka)
		pubCfg = wkafka.DefaultSaramaSyncPublisherConfig()
		subCfg = wkafka.DefaultSaramaSubscriberConfig()
	)

	for _, opt := range opts {
		opt(engine, pubCfg, subCfg)
	}

	pubCfg.Net.MaxOpenRequests = 1
	pubCfg.Producer.RequiredAcks = sarama.WaitForAll
	if pubCfg.Producer.Retry.Max < 1 {
		pubCfg.Producer.Retry.Max = 1
	}
	pubCfg.Producer.Idempotent = true

	if engine.logger == nil {
		engine.logger = slog.New(slog.DiscardHandler)
	}

	publisher, err := wkafka.NewPublisher(
		wkafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             wkafka.DefaultMarshaler{},
			OverwriteSaramaConfig: pubCfg,
		},
		watermill.NewSlogLogger(engine.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka publisher: %v", err)
	}

	subscriber, err := wkafka.NewSubscriber(
		wkafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           wkafka.DefaultMarshaler{},
			OverwriteSaramaConfig: subCfg,
			ConsumerGroup:         group,
		},
		watermill.NewSlogLogger(engine.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating kafka subscriber: %v", err)
	}

	engine.publisher = publisher
	engine.subscriber = subscriber

	return engine, nil
}

func (k *Kafka) Publisher() message.Publisher {
	return k.publisher
}

func (k *Kafka) Subscriber() message.Subscriber {
	return k.subscriber
}

func (k *Kafka) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		var errs []error
		if err := k.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := k.subscriber.Close(); err != nil {
			errs = append(errs, fmt.Errorf("subscriber close: %w", err))
		}
		done <- errors.Join(errs...)
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
