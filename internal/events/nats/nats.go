package nats

// @bug with WithConnectionRetry(true) option passed, driver will initialise successfully
// without error and will try to reconnect (according to reconnection options). If all
// attempts will fail, driver will call ClosedHandler(), but not fail in any way.

// @todo jetstream support

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	wnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	natsgo "github.com/nats-io/nats.go"
)

type NATS struct {
	logger     *slog.Logger
	publisher  *wnats.Publisher
	subscriber *wnats.Subscriber
	subwg      sync.WaitGroup
}

func New(dsn string, opts ...Option) (*NATS, error) {
	var (
		engine = new(NATS)
		pubCfg = &wnats.PublisherConfig{
			URL:       dsn,
			JetStream: wnats.JetStreamConfig{Disabled: true},
			Marshaler: new(wnats.GobMarshaler),
		}
		subCfg = &wnats.SubscriberConfig{
			URL:         dsn,
			JetStream:   wnats.JetStreamConfig{Disabled: true},
			Unmarshaler: new(wnats.GobMarshaler),
		}
	)

	for _, o := range opts {
		o(engine, pubCfg, subCfg)
	}

	if engine.logger == nil {
		engine.logger = slog.New(slog.DiscardHandler)
	}

	pubCfg.NatsOptions = append(pubCfg.NatsOptions, handlers(engine.logger)...)
	subCfg.NatsOptions = append(pubCfg.NatsOptions, handlers(engine.logger)...)

	publisher, err := wnats.NewPublisher(*pubCfg, watermill.NewSlogLogger(engine.logger))
	if err != nil {
		return nil, fmt.Errorf("error creating nats publisher: %w", err)
	}

	subscriber, err := wnats.NewSubscriber(*subCfg, watermill.NewSlogLogger(engine.logger))
	if err != nil {
		return nil, fmt.Errorf("error creating nats subscriber: %w", err)
	}

	engine.publisher = publisher
	engine.subscriber = subscriber

	return engine, nil
}

func (n *NATS) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return n.publisher.Publish(topic, msg)
}

func (n *NATS) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := n.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}
	n.subwg.Add(1)
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
		n.subwg.Done()
	}()
	return nil
}

func (n *NATS) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		var errs []error
		if err := n.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := n.subscriber.Close(); err != nil {
			errs = append(errs, fmt.Errorf("subscriber close: %w", err))
		}
		n.subwg.Wait()
		done <- errors.Join(errs...)
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}

func handlers(l *slog.Logger) []natsgo.Option {
	return []natsgo.Option{
		natsgo.ConnectHandler(func(conn *natsgo.Conn) {
			l.Debug("connection established")
		}),
		natsgo.ErrorHandler(func(conn *natsgo.Conn, sub *natsgo.Subscription, err error) {
			l.Debug("connection error", slog.String("error", err.Error()))
		}),
		natsgo.DisconnectErrHandler(func(conn *natsgo.Conn, err error) {
			l.Debug("disconnection error", slog.String("error", err.Error()))
		}),
		natsgo.LameDuckModeHandler(func(conn *natsgo.Conn) {
			l.Debug("server entering lame duck mode")
		}),
		natsgo.ClosedHandler(func(conn *natsgo.Conn) {
			l.Debug("connection closed")
		}),
		natsgo.ReconnectHandler(func(conn *natsgo.Conn) {
			l.Debug("reconnected")
		}),
		natsgo.ReconnectErrHandler(func(conn *natsgo.Conn, err error) {
			l.Debug("reconnect error", slog.String("error", err.Error()))
		}),
	}
}
