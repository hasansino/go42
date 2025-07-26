package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
	"github.com/hasansino/go42/internal/metrics"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	SaveFruitEventLog(ctx context.Context, event *models.FruitEventLog) error
}

type subscriber interface {
	Subscribe(
		ctx context.Context, topic string,
		handler func(ctx context.Context, event []byte) error,
	) error
}

type FruitEventSubscriber struct {
	logger     *slog.Logger
	repository repository
	subscriber subscriber
}

func NewFruitEventSubscriber(
	repository repository,
	subscriber subscriber,
	opts ...FruitEventSubscriberOption,
) *FruitEventSubscriber {
	sub := &FruitEventSubscriber{
		repository: repository,
		subscriber: subscriber,
	}
	for _, o := range opts {
		o(sub)
	}
	if sub.logger == nil {
		sub.logger = slog.New(slog.DiscardHandler)
	}
	return sub
}

func (s *FruitEventSubscriber) Subscribe(ctx context.Context, subscriber subscriber) error {
	return subscriber.Subscribe(ctx, domain.TopicNameFruitEvents, s.handleEvent)
}

func (s *FruitEventSubscriber) handleEvent(ctx context.Context, eventData []byte) error {
	event := new(outboxDomain.Event)
	err := json.Unmarshal(eventData, &event)
	if err != nil {
		s.logger.Error("failed to unmarshal event data")
		metrics.Counter("application_errors", map[string]interface{}{
			"type": "example_subscriber_error",
		}).Inc()
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	s.logger.Debug("received event", slog.Any("event", nil))

	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		eventLog := &models.FruitEventLog{
			ID:         event.ID,
			OccurredAt: event.CreatedAt,
			FruitID:    event.AggregateID,
			EventType:  event.AggregateType,
			Data:       event.Payload,
			Metadata:   event.Metadata,
		}
		err := s.repository.SaveFruitEventLog(txCtx, eventLog)
		if err != nil {
			s.logger.Error("failed to save event", slog.Any("error", err))
			metrics.Counter("application_errors", map[string]interface{}{
				"type": "example_subscriber_error",
			}).Inc()
			return fmt.Errorf("failed to create fruit: %w", err)
		}
		s.logger.Debug("event saved", slog.Any("event", eventLog))
		metrics.Counter("application_example_subscriber_processed", nil).Inc()
		return nil
	})
}

type FruitEventSubscriberOption func(*FruitEventSubscriber)

func FruitEventSubscriberWithLogger(logger *slog.Logger) FruitEventSubscriberOption {
	return func(o *FruitEventSubscriber) {
		o.logger = logger
	}
}
