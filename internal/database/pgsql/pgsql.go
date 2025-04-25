package pgsql

import (
	"database/sql"
	"log/slog"
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	gorm    *gorm.DB
	sqlDB   *sql.DB
	timeout time.Duration
}

func NewWrapper(dsn string, opts ...Option) (*Wrapper, error) {
	gormDB, err := gorm.Open(
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

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	wrapper := &Wrapper{}
	for _, opt := range opts {
		opt(wrapper, gormDB, sqlDB)
	}

	wrapper.gorm = gormDB
	wrapper.sqlDB = sqlDB

	return wrapper, nil
}

func (w *Wrapper) Close() error {
	return w.sqlDB.Close()
}

func (w *Wrapper) DB() *sql.DB {
	return w.sqlDB
}
