package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	lockID         = 1288990
	lockTimeoutSec = 30
	lockAttemptCnt = 10
)

func Migrate(ctx context.Context, uri string, schemaPath string) error {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return err
	}

	logger := slog.NewLogLogger(
		slog.Default().Handler().WithAttrs(
			[]slog.Attr{slog.String("component", "migrate")},
		),
		slog.LevelInfo,
	)

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
		goose.WithLogger(logger),
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
