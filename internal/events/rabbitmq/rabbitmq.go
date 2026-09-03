package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/avast/retry-go/v4"
)

const (
	defaultConnectRetryTimeout        = time.Minute
	defaultConnectRetryInitialBackoff = 500 * time.Millisecond
	defaultConnectRetryMaxBackoff     = 5 * time.Second
)

type AMQP struct {
	logger     *slog.Logger
	publisher  *amqp.Publisher
	subscriber *amqp.Subscriber

	connectRetryTimeout        time.Duration
	connectRetryInitialBackoff time.Duration
	connectRetryMaxBackoff     time.Duration
}

func New(ctx context.Context, dsn string, consumerGroup string, opts ...Option) (*AMQP, error) {
	if len(consumerGroup) == 0 {
		return nil, errors.New("consumer group is required")
	}

	var (
		engine = &AMQP{
			connectRetryTimeout:        defaultConnectRetryTimeout,
			connectRetryInitialBackoff: defaultConnectRetryInitialBackoff,
			connectRetryMaxBackoff:     defaultConnectRetryMaxBackoff,
		}
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

	retryCtx, cancel := context.WithTimeout(ctx, engine.connectRetryTimeout)
	defer cancel()

	err := retry.Do(func() error {
		publisher, err := amqp.NewPublisher(
			amqpConfig,
			watermill.NewSlogLogger(engine.logger),
		)
		if err != nil {
			return fmt.Errorf("error creating amqp publisher: %w", err)
		}

		subscriber, err := amqp.NewSubscriber(
			amqpConfig,
			watermill.NewSlogLogger(engine.logger),
		)
		if err != nil {
			return errors.Join(
				fmt.Errorf("error creating amqp subscriber: %w", err),
				publisher.Close(),
			)
		}

		engine.publisher = publisher
		engine.subscriber = subscriber
		return nil
	},
		retry.Context(retryCtx),
		retry.Attempts(0),
		retry.Delay(engine.connectRetryInitialBackoff),
		retry.MaxDelay(engine.connectRetryMaxBackoff),
		retry.DelayType(retry.FullJitterBackoffDelay),
		retry.WrapContextErrorWithLastError(true),
		retry.OnRetry(func(n uint, err error) {
			if retryCtx.Err() == nil {
				engine.logger.WarnContext(
					ctx,
					"broker connection attempt failed, retrying...",
					slog.Any("attempt", n+1),
					slog.Any("error", err),
				)
			}
		}),
	)
	if err != nil {
		return nil, err
	}

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
