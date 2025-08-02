package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, uri string, schemaPath string) error {
	logger := slog.With(slog.String("component", "migrate"))

	slog2mysql := &slog2mysql{logger, slog.LevelWarn}
	if err := mysql.SetLogger(slog2mysql); err != nil {
		return fmt.Errorf("failed to set MySQL slog2mysql: %w", err)
	}

	db, err := retry.DoWithData[*sql.DB](func() (*sql.DB, error) {
		db, err := sql.Open("mysql", uri)
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

	provider, err := goose.NewProvider(
		goose.DialectMySQL,
		db,
		os.DirFS(schemaPath),
		goose.WithLogger(slog.NewLogLogger(logger.Handler(), slog.LevelInfo)),
		goose.WithVerbose(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return db.Close()
}

// slog2mysql is a wrapper to adapt slog logger to the MySQL logger interface.
type slog2mysql struct {
	logger *slog.Logger
	level  slog.Level
}

func (l *slog2mysql) Print(v ...any) {
	l.logger.Log(context.Background(), l.level, fmt.Sprint(v...))
}
