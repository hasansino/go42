package migrate

import (
	"database/sql"
	"log/slog"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
)

func Migrate(uri string, schemaPath string) error {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("mysql"); err != nil {
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

	return db.Close()
}
