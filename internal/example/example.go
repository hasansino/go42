package example

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	ListFruits(ctx context.Context, limit, offset int) ([]*models.Fruit, error)
	GetFruitByID(ctx context.Context, id int) (*models.Fruit, error)
	CreateFruit(ctx context.Context, fruit *models.Fruit) error
	UpdateFruit(ctx context.Context, fruit *models.Fruit) error
	DeleteFruit(ctx context.Context, fruit *models.Fruit) error
}

type outboxService interface {
	NewOutboxMessage(ctx context.Context, topic string, msg *outboxDomain.Message) error
}

// Service layer of example domain
type Service struct {
	logger        *slog.Logger
	repository    repository
	outboxService outboxService
}

// NewService creates service with given repository
func NewService(
	repository repository,
	outboxService outboxService,
	opts ...Option,
) *Service {
	svc := &Service{
		repository:    repository,
		outboxService: outboxService,
	}
	for _, opt := range opts {
		opt(svc)
	}
	if svc.logger == nil {
		svc.logger = slog.New(slog.DiscardHandler)
	}
	return svc
}

func (s *Service) Fruits(ctx context.Context, limit int, offset int) ([]*models.Fruit, error) {
	return s.repository.ListFruits(ctx, limit, offset)
}

func (s *Service) FruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	return s.repository.GetFruitByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, name string) (*models.Fruit, error) {
	fruit := new(models.Fruit)
	err := s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		fruit.Name = name
		err := s.repository.CreateFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to create fruit: %w", err)
		}
		err = s.sendEvent(txCtx, domain.EventTypeCreated, fruit)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fruit, nil
}

func (s *Service) Update(ctx context.Context, id int, name string) (*models.Fruit, error) {
	var fruit *models.Fruit
	err := s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		fruit, err = s.repository.GetFruitByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get fruit by id: %w", err)
		}
		fruit.Name = name
		err = s.repository.UpdateFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to update fruit: %w", err)
		}
		err = s.sendEvent(txCtx, domain.EventTypeUpdated, fruit)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fruit, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		fruit, err := s.repository.GetFruitByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get fruit by id: %w", err)
		}
		err = s.repository.DeleteFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to delete fruit: %w", err)
		}
		err = s.sendEvent(txCtx, domain.EventTypeDeleted, fruit)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
}

func (s *Service) sendEvent(ctx context.Context, eventType string, fruit *models.Fruit) error {
	payloadJson, err := json.Marshal(fruit)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	outboxMessage := outboxDomain.Message{
		AggregateID:   fruit.ID,
		AggregateType: eventType,
		Payload:       payloadJson,
	}
	err = s.outboxService.NewOutboxMessage(ctx, domain.TopicNameFruitEvents, &outboxMessage)
	if err != nil {
		return fmt.Errorf("failed to send outbox message: %w", err)
	}
	return nil
}
