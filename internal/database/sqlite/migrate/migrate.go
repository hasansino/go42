package migrate

import (
	"database/sql"
	"log/slog"

	"github.com/pressly/goose/v3"
)

func Migrate(dbPath string, schemaPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
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
	if err := goose.Up(db, schemaPath); err != nil {
		return err
	}
	return nil
}
