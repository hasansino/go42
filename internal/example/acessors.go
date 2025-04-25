package example

import (
	"context"

	"github.com/hasansino/goapp/internal/example/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/accessors.go

type Repository interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	List(ctx context.Context, limit, offset int) ([]*models.Fruit, error)
	GetByID(ctx context.Context, id int) (*models.Fruit, error)
	Create(ctx context.Context, fruit *models.Fruit) error
}
