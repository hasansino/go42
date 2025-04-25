package pgsql

import (
	"database/sql"
	"log/slog"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	gorm  *gorm.DB
	sqlDB *sql.DB
}

func NewWrapper(dsn string, opts ...Option) (*Wrapper, error) {
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{DSN: dsn},
		),
		&gorm.Config{
			PrepareStmt: true,
			Logger: slogGorm.New(
				slogGorm.WithHandler(slog.Default().Handler()),
				slogGorm.SetLogLevel(slogGorm.ErrorLogType, slog.LevelError),
				slogGorm.SetLogLevel(slogGorm.SlowQueryLogType, slog.LevelInfo),
				slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelDebug),
				slogGorm.WithContextValue("system", "gorm"),
			),
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

	return &Wrapper{db, sqlDB}, nil
}

func (w *Wrapper) Close() error {
	return w.sqlDB.Close()
}

func (w *Wrapper) DB() *sql.DB {
	return w.sqlDB
}
