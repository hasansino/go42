package example

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/hasansino/goapp/internal/example/domain"
	"github.com/hasansino/goapp/internal/example/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type Cache interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type Repository interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	List(ctx context.Context, limit, offset int) ([]*models.Fruit, error)
	GetByID(ctx context.Context, id int) (*models.Fruit, error)
	Create(ctx context.Context, fruit *models.Fruit) error
	Update(ctx context.Context, fruit *models.Fruit) error
	Delete(ctx context.Context, id int) error
}

// Service layer of example domain
type Service struct {
	cache      Cache
	repository Repository
}

// NewService creates service with given repository
func NewService(repository Repository, cache Cache) *Service {
	return &Service{repository: repository, cache: cache}
}

// withTransaction abstracts the transaction management pattern
func (s *Service) withTransaction(
	ctx context.Context,
	fn func(txCtx context.Context) error,
) error {
	txCtx, err := s.repository.Begin(ctx)
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
	return s.repository.List(ctx, limit, offset)
}

func (s *Service) FruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, req *domain.CreateFruitRequest) (*models.Fruit, error) {
	fruit := new(models.Fruit)
	fruit.Name = req.Name
	err := s.repository.Create(ctx, fruit)
	if err != nil {
		return nil, fmt.Errorf("failed to create fruit: %w", err)
	}
	return fruit, nil
}

func (s *Service) Delete(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int, req *domain.UpdateFruitRequest) (*models.Fruit, error) {
	fruit, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get fruit by id: %w", err)
	}
	fruit.Name = req.Name
	err = s.repository.Update(ctx, fruit)
	if err != nil {
		return nil, fmt.Errorf("failed to update fruit: %w", err)
	}
	return fruit, nil
}
