package mysql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	libMysql "github.com/go-sql-driver/mysql"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	logger *slog.Logger
	gormDB *gorm.DB
	sqlDB  *sql.DB

	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxOpenConns    int
	maxIdleConns    int

	queryLogging bool
}

func New(dsn string, opts ...Option) (*Mysql, error) {
	w := new(Mysql)

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
		mysql.Open(dsn),
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

func (w *Mysql) Shutdown(ctx context.Context) error {
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

func (w *Mysql) DB() *sql.DB {
	return w.sqlDB
}

func (w *Mysql) GormDB() *gorm.DB {
	return w.gormDB
}

func (w *Mysql) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (w *Mysql) IsDuplicateKeyError(err error) bool {
	var mysqlErr *libMysql.MySQLError
	if errors.As(err, &mysqlErr) {
		// MySQL error codes for duplicate key violations
		return mysqlErr.Number == 1062 || mysqlErr.Number == 1586
	}
	return false
}
