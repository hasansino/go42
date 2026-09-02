package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type AMQP struct {
	logger     *slog.Logger
	publisher  *amqp.Publisher
	subscriber *amqp.Subscriber
}

func New(dsn string, consumerGroup string, opts ...Option) (*AMQP, error) {
	if len(consumerGroup) == 0 {
		return nil, errors.New("consumer group is required")
	}

	var (
		engine     = new(AMQP)
		amqpConfig = amqp.NewDurablePubSubConfig(
			dsn,
			amqp.GenerateQueueNameTopicNameWithSuffix(consumerGroup),
		)
	)

	amqpConfig.Publish.ConfirmDelivery = true

	for _, opt := range opts {
		opt(engine, &amqpConfig)
	}

	if engine.logger == nil {
		engine.logger = slog.New(slog.DiscardHandler)
	}

	publisher, err := amqp.NewPublisher(
		amqpConfig,
		watermill.NewSlogLogger(engine.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating amqp publisher: %w", err)
	}

	subscriber, err := amqp.NewSubscriber(
		amqpConfig,
		watermill.NewSlogLogger(engine.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating amqp subscriber: %w", err)
	}

	engine.publisher = publisher
	engine.subscriber = subscriber

	return engine, nil
}

func (rmq *AMQP) Publisher() message.Publisher {
	return rmq.publisher
}

func (rmq *AMQP) Subscriber() message.Subscriber {
	return rmq.subscriber
}

func (rmq *AMQP) InitializeTopic(topic string) error {
	return rmq.subscriber.SubscribeInitialize(topic)
}

func (rmq *AMQP) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		var errs []error
		if err := rmq.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := rmq.subscriber.Close(); err != nil {
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
