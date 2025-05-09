package rabbitmq

// @todo config options

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

func New(dsn string, opts ...Option) (*AMQP, error) {
	var (
		engine     = new(AMQP)
		amqpConfig = amqp.NewDurableQueueConfig(dsn)
	)

	for _, opt := range opts {
		opt(engine, &amqpConfig)
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

func (rmq *AMQP) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return rmq.publisher.Publish(topic, msg)
}

func (rmq *AMQP) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := rmq.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
	}()
	return nil
}

func (rmq *AMQP) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		if err := rmq.publisher.Close(); err != nil {
			done <- err
		}
		if err := rmq.subscriber.Close(); err != nil {
			done <- err
		}
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
