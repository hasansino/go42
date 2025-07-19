package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
)

func Migrate(ctx context.Context, uri string, schemaPath string) error {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}
	defer db.Close()

	provider, err := goose.NewProvider(
		goose.DialectMySQL,
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

	return db.Close()
}
