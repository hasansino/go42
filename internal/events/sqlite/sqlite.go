package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	wmsqlite "github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitemodernc"
	"github.com/ThreeDotsLabs/watermill/message"
)

type SQLite struct {
	logger     *slog.Logger
	publisher  message.Publisher
	subscriber message.Subscriber
	subwg      sync.WaitGroup
}

func New(db *sql.DB, opts ...Option) (*SQLite, error) {
	var (
		engine = new(SQLite)
		pubCfg = &wmsqlite.PublisherOptions{
			InitializeSchema: true,
		}
		subCfg = &wmsqlite.SubscriberOptions{
			ConsumerGroupMatcher: wmsqlite.NewStaticConsumerGroupMatcher("default"),
			BatchSize:            100,
			PollInterval:         time.Second,
			LockTimeout:          5 * time.Second,
			InitializeSchema:     true,
		}
	)

	for _, o := range opts {
		o(engine, pubCfg, subCfg)
	}

	if engine.logger == nil {
		engine.logger = slog.New(slog.DiscardHandler)
	}

	wmLogger := watermill.NewSlogLogger(engine.logger)
	pubCfg.Logger = wmLogger
	subCfg.Logger = wmLogger

	publisher, err := wmsqlite.NewPublisher(db, *pubCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating sqlite publisher: %w", err)
	}

	subscriber, err := wmsqlite.NewSubscriber(db, *subCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating sqlite subscriber: %w", err)
	}

	engine.publisher = publisher
	engine.subscriber = subscriber

	return engine, nil
}

func (s *SQLite) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	return s.publisher.Publish(topic, msg)
}

func (s *SQLite) Subscribe(
	ctx context.Context, topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	messages, err := s.subscriber.Subscribe(ctx, topic)
	if err != nil {
		return err
	}
	s.subwg.Add(1)
	go func() {
		for msg := range messages {
			err := handler(ctx, msg.Payload)
			if err != nil {
				msg.Nack()
			} else {
				msg.Ack()
			}
		}
		s.subwg.Done()
	}()
	return nil
}

func (s *SQLite) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		var errs []error
		if err := s.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("publisher close: %w", err))
		}
		if err := s.subscriber.Close(); err != nil {
			errs = append(errs, fmt.Errorf("subscriber close: %w", err))
		}
		s.subwg.Wait()
		done <- errors.Join(errs...)
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
