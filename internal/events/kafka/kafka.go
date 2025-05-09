package kafka

// @todo config options

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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

func (k *Kafka) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return k.publisher.Publish(topic, msg)
}

func (k *Kafka) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := k.subscriber.Subscribe(ctx, topic)
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

func (k *Kafka) Shutdown(ctx context.Context) error {
	doneChan := make(chan error)
	go func() {
		if err := k.publisher.Close(); err != nil {
			doneChan <- err
		}
		if err := k.subscriber.Close(); err != nil {
			doneChan <- err
		}
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-doneChan:
		return err
	}
}
