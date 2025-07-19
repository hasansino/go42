package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/hasansino/go42/internal/database/sqlite"
)

func Migrate(ctx context.Context, dbPath string, schemaPath string, opts ...sqlite.ConnectionOption) error {
	db, err := sql.Open("sqlite", sqlite.AddConnectionOptions(dbPath, opts))
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)

	provider, err := goose.NewProvider(
		goose.DialectSQLite3,
		db,
		os.DirFS(schemaPath),
		goose.WithLogger(
			slog.NewLogLogger(
				slog.Default().Handler().WithAttrs(
					[]slog.Attr{slog.String("component", "migrate")},
				),
				slog.LevelInfo,
			),
		),
		goose.WithVerbose(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create goose provider: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if dbPath != "file::memory:" {
		return db.Close()
	}

	// Closing in-memory db will destroy all data.
	return nil
}
