package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	wkafka "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/avast/retry-go/v4"

	"github.com/go42-dev/go42/internal/metrics"
)

const (
	defaultConnectRetryTimeout        = time.Minute
	defaultConnectRetryInitialBackoff = 500 * time.Millisecond
	defaultConnectRetryMaxBackoff     = 5 * time.Second
)

type Kafka struct {
	logger     *slog.Logger
	publisher  *wkafka.Publisher
	subscriber *wkafka.Subscriber

	connectRetryTimeout        time.Duration
	connectRetryInitialBackoff time.Duration
	connectRetryMaxBackoff     time.Duration
}

type connectionResult struct {
	publisher  *wkafka.Publisher
	subscriber *wkafka.Subscriber
	err        error
}

func New(ctx context.Context, brokers []string, group string, opts ...Option) (*Kafka, error) {
	var (
		engine = &Kafka{
			connectRetryTimeout:        defaultConnectRetryTimeout,
			connectRetryInitialBackoff: defaultConnectRetryInitialBackoff,
			connectRetryMaxBackoff:     defaultConnectRetryMaxBackoff,
		}
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

	retryCtx, cancel := context.WithTimeout(ctx, engine.connectRetryTimeout)
	defer cancel()

	err := retry.Do(func() error {
		resultLabel := "failure"
		defer func() {
			metrics.Counter("application_event_backend_connection_attempts_total", map[string]any{
				"backend": "kafka",
				"result":  resultLabel,
			}).Inc()
		}()

		result, err := connect(retryCtx, brokers, group, pubCfg, subCfg, engine.logger)
		if err != nil {
			return err
		}
		engine.publisher = result.publisher
		engine.subscriber = result.subscriber
		resultLabel = "success"
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

// connect enforces a hard startup deadline around the synchronous Watermill/Sarama constructors.
// Those constructors do not accept a context and may otherwise continue blocking after the
// configured connection retry timeout has expired.
func connect(
	ctx context.Context,
	brokers []string,
	group string,
	pubCfg *sarama.Config,
	subCfg *sarama.Config,
	logger *slog.Logger,
) (connectionResult, error) {
	resultChan := make(chan connectionResult, 1)
	go func() {
		resultChan <- openConnections(brokers, group, pubCfg, subCfg, logger)
	}()

	select {
	case result := <-resultChan:
		if err := ctx.Err(); err != nil {
			closeConnections(result)
			return connectionResult{}, err
		}
		return result, result.err
	case <-ctx.Done():
		// The constructors cannot be interrupted. Return to the caller immediately, then close
		// any connections they create when the in-flight attempt eventually completes.
		go func() {
			closeConnections(<-resultChan)
		}()
		return connectionResult{}, ctx.Err()
	}
}

func openConnections(
	brokers []string,
	group string,
	pubCfg *sarama.Config,
	subCfg *sarama.Config,
	logger *slog.Logger,
) connectionResult {
	publisher, err := wkafka.NewPublisher(
		wkafka.PublisherConfig{
			Brokers:               brokers,
			Marshaler:             wkafka.DefaultMarshaler{},
			OverwriteSaramaConfig: pubCfg,
		},
		watermill.NewSlogLogger(logger),
	)
	if err != nil {
		return connectionResult{err: fmt.Errorf("error creating kafka publisher: %w", err)}
	}

	subscriber, err := wkafka.NewSubscriber(
		wkafka.SubscriberConfig{
			Brokers:               brokers,
			Unmarshaler:           wkafka.DefaultMarshaler{},
			OverwriteSaramaConfig: subCfg,
			ConsumerGroup:         group,
		},
		watermill.NewSlogLogger(logger),
	)
	if err != nil {
		return connectionResult{
			err: errors.Join(
				fmt.Errorf("error creating kafka subscriber: %w", err),
				publisher.Close(),
			),
		}
	}

	return connectionResult{publisher: publisher, subscriber: subscriber}
}

func closeConnections(result connectionResult) {
	if result.publisher != nil {
		_ = result.publisher.Close()
	}
	if result.subscriber != nil {
		_ = result.subscriber.Close()
	}
}

func (k *Kafka) Publisher() message.Publisher {
	return k.publisher
}

func (k *Kafka) Subscriber() message.Subscriber {
	return k.subscriber
}

func (k *Kafka) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
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
