package migrate

import (
	"database/sql"
	"log/slog"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
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
