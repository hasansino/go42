package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	"github.com/hasansino/go42/internal/metrics"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
)

type AuthEventSubscriber struct {
	logger     *slog.Logger
	repository repository
}

func NewAuthEventSubscriber(
	repository repository,
	opts ...AuthEventSubscriberOption,
) *AuthEventSubscriber {
	sub := &AuthEventSubscriber{
		repository: repository,
	}
	for _, o := range opts {
		o(sub)
	}
	if sub.logger == nil {
		sub.logger = slog.New(slog.DiscardHandler)
	}
	return sub
}

func (s *AuthEventSubscriber) Subscribe(ctx context.Context, subscriber subscriber) error {
	return subscriber.Subscribe(ctx, domain.TopicNameAuthEvents, s.handleEvent)
}

func (s *AuthEventSubscriber) handleEvent(ctx context.Context, eventData []byte) error {
	event := new(outboxDomain.Event)
	err := json.Unmarshal(eventData, &event)
	if err != nil {
		s.logger.Error("failed to unmarshal event data")
		metrics.Counter("application_errors", map[string]interface{}{
			"type": "auth_event_subscriber_error",
		}).Inc()
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	s.logger.Debug("received event", slog.Any("event", nil))

	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		eventLog := &models.UserHistoryRecord{
			ID:         event.ID,
			OccurredAt: event.CreatedAt,
			UserID:     event.AggregateID,
			EventType:  event.AggregateType,
			Data:       event.Payload,
			Metadata:   event.Metadata,
		}
		err := s.repository.SaveUserHistoryRecord(txCtx, eventLog)
		if err != nil {
			s.logger.Error("failed to save event", slog.Any("error", err))
			metrics.Counter("application_errors", map[string]interface{}{
				"type": "auth_event_subscriber_error",
			}).Inc()
			return fmt.Errorf("failed to save log: %w", err)
		}
		s.logger.Debug("event saved", slog.Any("event", eventLog))
		metrics.Counter("application_auth_event_subscriber_processed", nil).Inc()
		return nil
	})
}

type AuthEventSubscriberOption func(*AuthEventSubscriber)

func AuthEventSubscriberWithLogger(logger *slog.Logger) AuthEventSubscriberOption {
	return func(o *AuthEventSubscriber) {
		o.logger = logger
	}
}
