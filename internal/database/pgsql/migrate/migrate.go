package migrate

import (
	"database/sql"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Migrate(uri string, schemaPath string) error {
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return err
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(
		slog.NewLogLogger(
			slog.Default().
				With("service", "migrate").Handler(),
			slog.LevelInfo,
		),
	)
	if err := goose.Up(db, schemaPath); err != nil {
		return err
	}
	return nil
}
