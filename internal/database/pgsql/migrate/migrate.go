package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

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

func Migrate(ctx context.Context, uri string, schemaPath string) error {
	logger := slog.With(slog.String("component", "migrate"))

	db, err := retry.DoWithData[*sql.DB](func() (*sql.DB, error) {
		db, err := sql.Open("pgx", uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}
		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
		return db, nil
	},
		retry.Context(ctx),
		retry.Attempts(10),
		retry.Delay(2*time.Second),
		retry.MaxDelay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logger.WarnContext(
				ctx,
				"database connection attempt failed, retrying...",
				slog.Any("attempt", n+1),
				slog.String("error", err.Error()),
			)
		}),
	)
	if err != nil {
		return err
	}

	// migrations have independent connections, so we can close the connection after migration
	defer db.Close()

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

	return db.Close()
}
