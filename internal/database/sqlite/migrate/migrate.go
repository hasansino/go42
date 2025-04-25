package migrate

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
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
			slog.Default().
				With(slog.String("service", "migrate")).Handler(),
			slog.LevelInfo,
		),
	)
	if err := goose.Up(db, schemaPath); err != nil {
		return err
	}
	return nil
}
