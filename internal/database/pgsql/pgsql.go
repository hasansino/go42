package pgsql

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Wrapper struct {
	logger          *slog.Logger
	gormDB          *gorm.DB
	sqlDB           *sql.DB
	timeout         time.Duration
	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxOpenConns    int
	maxIdleConns    int
}

func New(dsn string, opts ...Option) (*Wrapper, error) {
	w := new(Wrapper)
	for _, opt := range opts {
		opt(w)
	}
	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{DSN: dsn}),
		&gorm.Config{
			PrepareStmt: true,
			Logger: slogGorm.New(
				// slogGorm.WithIgnoreTrace(),
				slogGorm.WithHandler(w.logger.Handler()),
				// log level translations: when gormDB sends X level -> slog handles it as Y level
				slogGorm.SetLogLevel(slogGorm.ErrorLogType, slog.LevelDebug), // exposes query
				slogGorm.SetLogLevel(slogGorm.SlowQueryLogType, slog.LevelWarn),
				slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelInfo),
			),
		})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	w.gormDB = gormDB
	w.sqlDB = sqlDB

	return w, nil
}

func (w *Wrapper) Close() error {
	return w.sqlDB.Close()
}

func (w *Wrapper) GormDB() *gorm.DB {
	return w.gormDB
}

func (w *Wrapper) SqlDB() *sql.DB {
	return w.sqlDB
}

func (w *Wrapper) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (w *Wrapper) IsDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
