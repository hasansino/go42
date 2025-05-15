package example

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type Publisher interface {
	Publish(topic string, event []byte) error
}

type Subscriber interface {
	Subscribe(
		ctx context.Context, topic string,
		handler func(ctx context.Context, event []byte) error,
	) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

type Repository interface {
	Begin(ctx context.Context, isolationLvl sql.IsolationLevel) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	ListFruits(ctx context.Context, limit, offset int) ([]*models.Fruit, error)
	GetFruitByID(ctx context.Context, id int) (*models.Fruit, error)
	CreateFruit(ctx context.Context, fruit *models.Fruit) error
	UpdateFruit(ctx context.Context, fruit *models.Fruit) error
	DeleteFruit(ctx context.Context, fruit *models.Fruit) error
	SaveEvent(ctx context.Context, event *models.Event) error
}

// Service layer of example domain
type Service struct {
	logger     *slog.Logger
	repository Repository
	publisher  Publisher
	cache      Cache
}

// NewService creates service with given repository
func NewService(repository Repository, publisher Publisher, opts ...Option) *Service {
	svc := &Service{
		repository: repository,
		publisher:  publisher,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// withTransaction abstracts the transaction management pattern
func (s *Service) withTransaction(
	ctx context.Context,
	fn func(txCtx context.Context) error,
) error {
	txCtx, err := s.repository.Begin(ctx, sql.LevelDefault)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = s.repository.Rollback(txCtx)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := s.repository.Rollback(txCtx); rbErr != nil {
			return fmt.Errorf("error executing transaction (rollback failed: %v): %w", rbErr, err)
		}
		return err
	}

	if err := s.repository.Commit(txCtx); err != nil {
		_ = s.repository.Rollback(txCtx)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) Fruits(ctx context.Context, limit int, offset int) ([]*models.Fruit, error) {
	return s.repository.ListFruits(ctx, limit, offset)
}

func (s *Service) FruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	return s.repository.GetFruitByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req *domain.CreateFruitRequest) (*models.Fruit, error) {
	fruit := new(models.Fruit)
	err := s.withTransaction(ctx, func(txCtx context.Context) error {
		fruit.Name = req.Name
		err := s.repository.CreateFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to create fruit: %w", err)
		}
		err = s.sendEvent(domain.EventTypeCreated, fruit)
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
	return s.withTransaction(ctx, func(txCtx context.Context) error {
		var err error
		fruit, err := s.repository.GetFruitByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get fruit by id: %w", err)
		}
		err = s.repository.DeleteFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to delete fruit: %w", err)
		}
		err = s.sendEvent(domain.EventTypeDeleted, fruit)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
}

func (s *Service) Update(ctx context.Context, id int, req *domain.UpdateFruitRequest) (*models.Fruit, error) {
	var fruit *models.Fruit
	err := s.withTransaction(ctx, func(txCtx context.Context) error {
		var err error
		fruit, err = s.repository.GetFruitByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get fruit by id: %w", err)
		}
		fruit.Name = req.Name
		err = s.repository.UpdateFruit(txCtx, fruit)
		if err != nil {
			return fmt.Errorf("failed to update fruit: %w", err)
		}
		err = s.sendEvent(domain.EventTypeUpdated, fruit)
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

func (s *Service) sendEvent(eventType int, payload any) error {
	event := &domain.ExampleEvent{Type: eventType, Payload: payload}
	payloadJson, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return s.publisher.Publish("example-topic", payloadJson)
}

func (s *Service) Subscribe(ctx context.Context, subscriber Subscriber) error {
	return subscriber.Subscribe(ctx, "example-topic", s.handleEvent)
}

func (s *Service) handleEvent(ctx context.Context, eventData []byte) error {
	event := new(domain.ExampleEvent)
	err := event.Unmarshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	s.logger.Info("received event", slog.Any("event", event.Type))

	if _, ok := domain.EventTypes[event.Type]; !ok {
		return fmt.Errorf("invalid event type: %d", event.Type)
	}

	dbEvent := new(models.Event)
	return s.withTransaction(ctx, func(txCtx context.Context) error {
		dbEvent.Data = string(eventData)
		err := s.repository.SaveEvent(txCtx, dbEvent)
		if err != nil {
			return fmt.Errorf("failed to create fruit: %w", err)
		}
		return nil
	})
}
