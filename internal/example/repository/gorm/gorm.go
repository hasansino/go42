package gorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/hasansino/go42/internal/example/domain"
	"github.com/hasansino/go42/internal/example/models"
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

func (r *Repository) getTxKey() txKey {
	return txKey{}
}

func (r *Repository) getTx(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(r.getTxKey()).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

func (r *Repository) Begin(ctx context.Context, isolationLvl sql.IsolationLevel) (context.Context, error) {
	if _, ok := ctx.Value(r.getTxKey()).(*gorm.DB); ok {
		return ctx, errors.New("transaction already exists in context")
	}
	tx := r.db.WithContext(ctx).Begin(&sql.TxOptions{Isolation: isolationLvl})
	if tx.Error != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	return context.WithValue(ctx, r.getTxKey(), tx), nil
}

func (r *Repository) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(r.getTxKey()).(*gorm.DB)
	if !ok {
		return errors.New("no transaction found in context")
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(r.getTxKey()).(*gorm.DB)
	if !ok {
		return errors.New("no transaction found in context")
	}
	if err := tx.Rollback().Error; err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListFruits(ctx context.Context, limit, offset int) ([]*models.Fruit, error) {
	var fruits []*models.Fruit
	result := r.getTx(ctx).Limit(limit).Offset(offset).Order("id ASC").Find(&fruits)
	if result.Error != nil {
		return nil, fmt.Errorf("error listing fruits: %w", result.Error)
	}
	return fruits, nil
}

func (r *Repository) GetFruitByID(ctx context.Context, id int) (*models.Fruit, error) {
	var fruit models.Fruit
	result := r.getTx(ctx).First(&fruit, id)
	if result.Error != nil {
		if r.sqlCore.IsNotFoundError(result.Error) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("error fetching fruit by ID: %w", result.Error)
	}
	return &fruit, nil
}

func (r *Repository) CreateFruit(ctx context.Context, fruit *models.Fruit) error {
	err := r.getTx(ctx).Create(fruit).Error
	if err != nil {
		if r.sqlCore.IsDuplicateKeyError(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("error creating fruit: %w", err)
	}
	return nil
}

func (r *Repository) DeleteFruit(ctx context.Context, fruit *models.Fruit) error {
	result := r.getTx(ctx).Delete(&fruit)
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	if result.Error != nil {
		return fmt.Errorf("error deleting fruit: %w", result.Error)
	}
	return nil
}

func (r *Repository) UpdateFruit(ctx context.Context, fruit *models.Fruit) error {
	tx := r.getTx(ctx).Save(fruit)
	if tx.Error != nil {
		return fmt.Errorf("error updating fruit: %w", tx.Error)
	}
	return nil
}

func (r *Repository) SaveEvent(ctx context.Context, event *models.Event) error {
	err := r.getTx(ctx).Create(event).Error
	if err != nil {
		return fmt.Errorf("error saving event: %w", err)
	}
	return nil
}
