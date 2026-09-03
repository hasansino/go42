package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hasansino/go42/internal/metrics"
	"github.com/hasansino/go42/internal/outbox/domain"
	"github.com/hasansino/go42/internal/outbox/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	GetUnprocessedMessages(ctx context.Context, limit int) ([]models.Message, error)
	SaveProcessedMessages(ctx context.Context, messages []models.Message) error
	SaveFailedMessages(ctx context.Context, messages []models.Message) error
}

type publisher interface {
	Publish(topic string, event []byte) error
}

type OutboxMessagePublisher struct {
	logger     *slog.Logger
	repository repository
	publisher  publisher
}

func NewOutboxMessagePublisher(
	repository repository,
	publisher publisher,
	opts ...OutboxMessagePublisherOption,
) *OutboxMessagePublisher {
	pub := &OutboxMessagePublisher{
		repository: repository,
		publisher:  publisher,
	}
	for _, opt := range opts {
		opt(pub)
	}
	if pub.logger == nil {
		pub.logger = slog.New(slog.DiscardHandler)
	}
	return pub
}

func (p *OutboxMessagePublisher) Run(
	ctx context.Context, interval time.Duration, batchSize int,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := p.run(ctx, batchSize)
			if ctx.Err() != nil {
				return
			}
			result := "success"
			if err != nil {
				result = "error"
			}
			metrics.Counter("application_outbox_worker_runs_total", map[string]interface{}{
				"result": result,
			}).Inc()
		}
	}
}

func (p *OutboxMessagePublisher) run(ctx context.Context, batchSize int) error {
	err := p.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		p.logger.Debug("running outbox publisher job")

		messages, err := p.repository.GetUnprocessedMessages(txCtx, batchSize)
		if err != nil {
			return fmt.Errorf("failed to get unprocessed messages: %w", err)
		}

		var (
			processed []models.Message
			failed    []models.Message
		)

		for _, message := range messages {
			event := domain.Event{
				ID:            message.ID,
				CreatedAt:     message.CreatedAt,
				AggregateID:   message.AggregateID,
				AggregateType: message.AggregateType,
				Payload:       message.Payload,
				Metadata:      message.Metadata,
			}
			jsonBytes, err := json.Marshal(event)
			if err != nil {
				return fmt.Errorf("failed to marshal event: %w", err)
			}
			err = p.publisher.Publish(message.Topic, jsonBytes)
			if err != nil {
				message.RetryCount++
				message.LastError = err.Error()
				result := "retry"
				if message.RetryCount >= message.MaxRetries {
					message.Status = models.MessageStatusFailed
					result = "permanently_failed"
				}
				observeDelivery(message.CreatedAt, result)
				failed = append(failed, message)
				p.logger.Error("failed to publish message", slog.Any("error", err))
				metrics.Counter("application_errors", map[string]interface{}{
					"type": "outbox_publisher_error",
				}).Inc()
				continue
			}
			observeDelivery(message.CreatedAt, "processed")
			processed = append(processed, message)
			p.logger.Debug("published message", slog.Any("message", message))
		}

		if len(processed) > 0 {
			err := p.repository.SaveProcessedMessages(txCtx, processed)
			if err != nil {
				return fmt.Errorf("failed to save processed messages: %w", err)
			}
		}

		if len(failed) > 0 {
			err := p.repository.SaveFailedMessages(txCtx, failed)
			if err != nil {
				return fmt.Errorf("failed to save failed messages: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		p.logger.ErrorContext(ctx,
			"failed to run outbox publisher job", slog.Any("error", err))
		metrics.Counter("application_errors", map[string]interface{}{
			"type": "outbox_publisher_error",
		}).Inc()
	}
	return err
}

func observeDelivery(createdAt time.Time, result string) {
	metrics.Counter("application_outbox_messages_total", map[string]interface{}{
		"result": result,
	}).Inc()

	delay := time.Since(createdAt).Seconds()
	if delay < 0 {
		delay = 0
	}
	metrics.Histogram("application_outbox_delivery_delay_seconds", map[string]interface{}{
		"result": result,
	}).Update(delay)
}

type OutboxMessagePublisherOption func(*OutboxMessagePublisher)

func OutboxMessagePublisherWithLogger(logger *slog.Logger) OutboxMessagePublisherOption {
	return func(o *OutboxMessagePublisher) {
		o.logger = logger
	}
}
