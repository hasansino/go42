package gorm

import (
	"context"
	"fmt"

	"github.com/hasansino/go42/internal/database"
	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
)

type Repository struct {
	*database.BaseRepository
}

func New(baseRepository *database.BaseRepository) *Repository {
	return &Repository{baseRepository}
}

func (r *Repository) ListFruits(ctx context.Context, limit, offset int) ([]*models.Fruit, error) {
	var fruits []*models.Fruit
	result := r.GetReadDB(ctx).Limit(limit).Offset(offset).Order("id ASC").Find(&fruits)
	if result.Error != nil {
		return nil, fmt.Errorf("error listing fruits: %w", result.Error)
	}
	return fruits, nil
}

func (r *Repository) GetFruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	var fruit models.Fruit
	result := r.GetReadDB(ctx).First(&fruit, id)
	if result.Error != nil {
		if r.IsNotFoundError(result.Error) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching fruit by ID: %w", result.Error)
	}
	return &fruit, nil
}

func (r *Repository) CreateFruit(ctx context.Context, fruit *models.Fruit) error {
	err := r.GetTx(ctx).Create(fruit).Error
	if err != nil {
		if r.IsDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("error creating fruit: %w", err)
	}
	return nil
}

func (r *Repository) DeleteFruit(ctx context.Context, fruit *models.Fruit) error {
	result := r.GetTx(ctx).Delete(fruit)
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error deleting fruit: %w", result.Error)
	}
	return nil
}

func (r *Repository) UpdateFruit(ctx context.Context, fruit *models.Fruit) error {
	result := r.GetTx(ctx).Save(fruit)
	if result.Error != nil {
		return fmt.Errorf("error updating fruit: %w", result.Error)
	}
	return nil
}

func (r *Repository) SaveFruitEventLog(ctx context.Context, event *models.FruitEventLog) error {
	err := r.GetTx(ctx).Create(event).Error
	if err != nil {
		return fmt.Errorf("error saving event: %w", err)
	}
	return nil
}
