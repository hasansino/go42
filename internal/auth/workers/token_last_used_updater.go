package workers

import (
	"context"
	"log/slog"
	"time"
)

type TokenLastUsedUpdater struct {
	logger      *slog.Logger
	repository  repository
	authService authService
}

func NewTokenLastUsedUpdater(
	repository repository,
	authService authService,
	opts ...TokenLastUsedUpdaterOption,
) *TokenLastUsedUpdater {
	updater := &TokenLastUsedUpdater{
		repository:  repository,
		authService: authService,
	}
	for _, o := range opts {
		o(updater)
	}
	if updater.logger == nil {
		updater.logger = slog.New(slog.DiscardHandler)
	}
	return updater
}

func (u *TokenLastUsedUpdater) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.run(ctx)
		}
	}
}

func (u *TokenLastUsedUpdater) run(ctx context.Context) {
	tokens := u.authService.RecentlyUsedTokens()
	for id, when := range tokens {
		err := u.repository.WithTransaction(ctx, func(txCtx context.Context) error {
			return u.repository.UpdateTokenLastUsed(txCtx, id, when)
		})
		if err != nil {
			u.logger.ErrorContext(
				ctx, "failed to update token last used time",
				slog.Any("token", id),
				slog.Any("when", when),
				slog.Any("error", err),
			)
		}
	}
}

type TokenLastUsedUpdaterOption func(*TokenLastUsedUpdater)

func TokenLastUsedUpdaterWithLogger(logger *slog.Logger) TokenLastUsedUpdaterOption {
	return func(o *TokenLastUsedUpdater) {
		o.logger = logger
	}
}
