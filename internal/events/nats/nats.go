package nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	wnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	natsgo "github.com/nats-io/nats.go"
)

type NATS struct {
	logger     *slog.Logger
	publisher  *wnats.Publisher
	subscriber *wnats.Subscriber
}

func New(dsn string, opts ...Option) (*NATS, error) {
	var (
		engine          = new(NATS)
		jetStreamConfig = wnats.JetStreamConfig{
			AutoProvision: true,
			SubscribeOptions: []natsgo.SubOpt{
				natsgo.DeliverAll(),
				natsgo.AckExplicit(),
			},
			TrackMsgID: true,
			AckAsync:   false,
		}
		pubCfg = &wnats.PublisherConfig{
			URL:       dsn,
			JetStream: jetStreamConfig,
			Marshaler: new(wnats.GobMarshaler),
		}
		subCfg = &wnats.SubscriberConfig{
			URL:         dsn,
			JetStream:   jetStreamConfig,
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
	subCfg.NatsOptions = append(subCfg.NatsOptions, handlers(engine.logger)...)

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

func (n *NATS) Publisher() message.Publisher {
	return n.publisher
}

func (n *NATS) Subscriber() message.Subscriber {
	return n.subscriber
}

func (n *NATS) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		var errs []error
		if err := n.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := n.subscriber.Close(); err != nil {
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

func handlers(l *slog.Logger) []natsgo.Option {
	return []natsgo.Option{
		natsgo.ConnectHandler(func(conn *natsgo.Conn) {
			l.Info("connection established")
		}),
		natsgo.ErrorHandler(func(conn *natsgo.Conn, sub *natsgo.Subscription, err error) {
			if err != nil {
				l.Warn("connection error", slog.String("error", err.Error()))
			}
		}),
		natsgo.DisconnectErrHandler(func(conn *natsgo.Conn, err error) {
			if err != nil {
				l.Warn("disconnection error", slog.String("error", err.Error()))
			}
		}),
		natsgo.LameDuckModeHandler(func(conn *natsgo.Conn) {
			l.Warn("server entering lame duck mode")
		}),
		natsgo.ClosedHandler(func(conn *natsgo.Conn) {
			l.Error("connection closed")
		}),
		natsgo.ReconnectHandler(func(conn *natsgo.Conn) {
			l.Info("reconnected")
		}),
		natsgo.ReconnectErrHandler(func(conn *natsgo.Conn, err error) {
			l.Debug("reconnect error", slog.String("error", err.Error()))
		}),
	}
}
