package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	logger *slog.Logger
	gormDB *gorm.DB
	sqlDB  *sql.DB

	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxOpenConns    int
	maxIdleConns    int

	queryLogging bool
}

func New(dsn string, opts ...Option) (*Postgres, error) {
	w := new(Postgres)

	for _, opt := range opts {
		opt(w)
	}

	if w.logger == nil {
		w.logger = slog.New(slog.DiscardHandler)
	}

	slogGormOpts := []slogGorm.Option{
		slogGorm.WithHandler(w.logger.Handler()),
		// log level translations: when gormDB sends X level -> slog handles it as Y level
		slogGorm.SetLogLevel(slogGorm.ErrorLogType, slog.LevelError),
		slogGorm.SetLogLevel(slogGorm.SlowQueryLogType, slog.LevelWarn),
		slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelInfo),
	}

	if w.queryLogging {
		slogGormOpts = append(slogGormOpts, slogGorm.WithTraceAll())
	} else {
		slogGormOpts = append(slogGormOpts, slogGorm.WithIgnoreTrace())
	}

	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{DSN: dsn}),
		&gorm.Config{
			PrepareStmt: true,
			Logger:      slogGorm.New(slogGormOpts...),
		})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(w.maxOpenConns)
	sqlDB.SetMaxIdleConns(w.maxIdleConns)
	sqlDB.SetConnMaxLifetime(w.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(w.connMaxIdleTime)

	w.gormDB = gormDB
	w.sqlDB = sqlDB

	return w, nil
}

func (w *Postgres) Shutdown(ctx context.Context) error {
	doneChan := make(chan error)
	go func() {
		doneChan <- w.sqlDB.Close()
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-doneChan:
		return err
	}
}

func (w *Postgres) DB() *sql.DB {
	return w.sqlDB
}

func (w *Postgres) GormDB() *gorm.DB {
	return w.gormDB
}

func (w *Postgres) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (w *Postgres) IsDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
