package pgsql

import (
	"log/slog"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	*gorm.DB
}

func NewWrapper(dsn string, opts ...Option) (*Wrapper, error) {
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true,
			},
		),
		&gorm.Config{
			PrepareStmt: true,
			Logger:      NewSlogAdapter(slog.Default()),
		})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(sqlDB)
	}

	return &Wrapper{db}, nil
}

func (w *Wrapper) Close() error {
	sqlDB, err := w.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
