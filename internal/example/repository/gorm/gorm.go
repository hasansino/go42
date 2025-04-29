package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/hasansino/goapp/internal/example/domain"
	"github.com/hasansino/goapp/internal/example/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type sqlCoreAccessor interface {
	IsNotFoundError(err error) bool
	IsDuplicateKeyError(err error) bool
}

type Repository struct {
	db      *gorm.DB
	sqlCore sqlCoreAccessor
}

func New(db *gorm.DB, sqlCore sqlCoreAccessor) *Repository {
	return &Repository{db: db, sqlCore: sqlCore}
}

type txKey struct{}

func (r *Repository) getKey() txKey {
	return txKey{}
}

func (r *Repository) Begin(ctx context.Context) (context.Context, error) {
	if _, ok := ctx.Value(r.getKey()).(*gorm.DB); ok {
		return ctx, errors.New("transaction already exists in context")
	}
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	return context.WithValue(ctx, r.getKey(), tx), nil
}

func (r *Repository) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(r.getKey()).(*gorm.DB)
	if !ok {
		return errors.New("no transaction found in context")
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(r.getKey()).(*gorm.DB)
	if !ok {
		return errors.New("no transaction found in context")
	}
	if err := tx.Rollback().Error; err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]*models.Fruit, error) {
	if limit <= 0 {
		limit = 10
	}
	var fruits []*models.Fruit
	result := r.db.Limit(limit).Offset(offset).Order("id ASC").Find(&fruits)
	if result.Error != nil {
		return nil, fmt.Errorf("error listing fruits: %w", result.Error)
	}
	return fruits, nil
}

func (r *Repository) GetByID(ctx context.Context, id int) (*models.Fruit, error) {
	var fruit models.Fruit
	result := r.db.First(&fruit, id)
	if result.Error != nil {
		if r.sqlCore.IsNotFoundError(result.Error) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching fruit by ID: %w", result.Error)
	}
	return &fruit, nil
}

func (r *Repository) Create(ctx context.Context, fruit *models.Fruit) error {
	err := r.db.Create(fruit).Error
	if err != nil {
		if r.sqlCore.IsDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("error creating fruit: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	var fruit models.Fruit
	if result := r.db.First(&fruit, id); result.Error != nil {
		if r.sqlCore.IsNotFoundError(result.Error) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("error fetching fruit by ID: %w", result.Error)
	}
	result := r.db.Delete(&fruit, id)
	if result.Error != nil {
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		return fmt.Errorf("error deleting fruit: %w", result.Error)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, fruit *models.Fruit) error {
	var existingFruit models.Fruit
	if result := r.db.First(&existingFruit, fruit.ID); result.Error != nil {
		if r.sqlCore.IsDuplicateKeyError(result.Error) {
			return domain.ErrAlreadyExists
		}
	}
	if err := r.db.Save(fruit).Error; err != nil {
		return fmt.Errorf("error updating fruit: %w", err)
	}
	return nil
}
