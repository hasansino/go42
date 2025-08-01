package workers

import (
	"context"
	"time"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	UpdateTokenLastUsed(ctx context.Context, tokenID int, when time.Time) error
	SaveUserHistoryRecord(ctx context.Context, record *models.UserHistoryRecord) error
}

type authService interface {
	RotateJWTSecret(newSecret string)
	RecentlyUsedTokensChan() <-chan domain.TokenWasUsed
}

type subscriber interface {
	Subscribe(
		ctx context.Context, topic string,
		handler func(ctx context.Context, event []byte) error,
	) error
}
