package migrate

import (
	"database/sql"
	"log/slog"

	"github.com/pressly/goose/v3"

	"github.com/hasansino/go42/internal/database/sqlite"
)

func Migrate(dbPath string, schemaPath string, opts ...sqlite.ConnectionOption) error {
	db, err := sql.Open("sqlite", sqlite.AddConnectionOptions(dbPath, opts))
	if err != nil {
		return err
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	goose.SetLogger(
		slog.NewLogLogger(
			slog.Default().Handler().WithAttrs(
				[]slog.Attr{slog.String("service", "migrate")},
			),
			slog.LevelInfo,
		),
	)

	db.SetMaxOpenConns(1)

	if err := goose.Up(db, schemaPath); err != nil {
		return err
	}

	if dbPath != "file::memory:" {
		return db.Close()
	}

	// Closing in-memory db will destroy all data.
	return nil
}
