package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/avast/retry-go/v4"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	lockID         = 1288990
	lockTimeoutSec = 30
	lockAttemptCnt = 15
)

func Migrate(
	ctx context.Context,
	uri string,
	schemaPath string,
	opts ...Option,
) (returnErr error) {
	config := defaultOptions()

	for _, opt := range opts {
		opt(&config)
	}

	logger := config.logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	retryCtx, cancelRetry := context.WithTimeout(ctx, config.connectRetryTimeout)

	db, err := retry.DoWithData[*sql.DB](func() (*sql.DB, error) {
		db, err := sql.Open("pgx", uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}
		if err := db.PingContext(retryCtx); err != nil {
			pingErr := fmt.Errorf("failed to ping database: %w", err)
			if closeErr := db.Close(); closeErr != nil {
				return nil, errors.Join(
					pingErr,
					fmt.Errorf("failed to close migration database: %w", closeErr),
				)
			}
			return nil, pingErr
		}
		return db, nil
	},
		retry.Context(retryCtx),
		retry.Attempts(0),
		retry.Delay(config.connectRetryInitialBackoff),
		retry.MaxDelay(config.connectRetryMaxBackoff),
		retry.DelayType(retry.FullJitterBackoffDelay),
		retry.WrapContextErrorWithLastError(true),
		retry.OnRetry(func(n uint, err error) {
			if retryCtx.Err() == nil {
				logger.WarnContext(
					ctx,
					"database connection attempt failed, retrying...",
					slog.Any("attempt", n+1),
					slog.Any("error", err),
				)
			}
		}),
	)
	cancelRetry()
	if err != nil {
		return err
	}

	// migrations have independent connections, so we can close the connection after migration
	defer func() {
		if err := db.Close(); err != nil {
			returnErr = errors.Join(
				returnErr,
				fmt.Errorf("failed to close migration database: %w", err),
			)
		}
	}()

	// locker is used to ensure that only one migration process runs at a time
	// this is required to prevent concurrent migrations that could lead to database inconsistencies
	locker, err := lock.NewPostgresSessionLocker(
		lock.WithLockID(lockID),
		lock.WithLockTimeout(lockTimeoutSec, lockAttemptCnt),
		lock.WithUnlockTimeout(lockTimeoutSec, lockAttemptCnt),
	)
	if err != nil {
		return fmt.Errorf("failed to create session locker: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS(schemaPath),
		goose.WithLogger(slog.NewLogLogger(logger.Handler(), slog.LevelInfo)),
		goose.WithVerbose(true),
		goose.WithSessionLocker(locker),
	)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
