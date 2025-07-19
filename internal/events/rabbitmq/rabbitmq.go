package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
)

type AMQP struct {
	logger     *slog.Logger
	publisher  *amqp.Publisher
	subscriber *amqp.Subscriber
	subwg      sync.WaitGroup
}

func New(dsn string, opts ...Option) (*AMQP, error) {
	var (
		engine     = new(AMQP)
		amqpConfig = amqp.NewDurableQueueConfig(dsn)
	)

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
	rmq.subwg.Add(1)
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
		rmq.subwg.Done()
	}()
	return nil
}

func (rmq *AMQP) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		var errs []error
		if err := rmq.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := rmq.subscriber.Close(); err != nil {
			errs = append(errs, fmt.Errorf("subscriber close: %w", err))
		}
		rmq.subwg.Wait()
		done <- errors.Join(errs...)
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
