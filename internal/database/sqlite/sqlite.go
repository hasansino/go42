package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/glebarez/sqlite"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	sqlitelib "modernc.org/sqlite/lib"
)

type Sqlite struct {
	logger *slog.Logger

	gormDB *gorm.DB
	sqlDB  *sql.DB

	connOpts []ConnectionOption

	queryLogging bool
}

type ConnectionOption struct {
	Key   string
	Value string
}

func Open(dbPath string, opts ...Option) (*Sqlite, error) {
	w := new(Sqlite)

	for _, opt := range opts {
		opt(w)
	}

	if w.logger == nil {
		w.logger = slog.New(slog.DiscardHandler)
	}

	slogGormOpts := []slogGorm.Option{
		slogGorm.WithHandler(w.logger.Handler()),
		// log level translations: when gormDB sends X level -> slog handles it as Y level
		slogGorm.SetLogLevel(slogGorm.ErrorLogType, slog.LevelWarn),
		slogGorm.SetLogLevel(slogGorm.SlowQueryLogType, slog.LevelWarn),
		slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelDebug),
	}

	if w.queryLogging {
		slogGormOpts = append(slogGormOpts, slogGorm.WithTraceAll())
	} else {
		slogGormOpts = append(slogGormOpts, slogGorm.WithIgnoreTrace())
	}

	gormDB, err := gorm.Open(
		sqlite.Open(AddConnectionOptions(dbPath, w.connOpts)),
		&gorm.Config{
			PrepareStmt: true,
			Logger:      slogGorm.New(slogGormOpts...),
		})
	if err != nil {
		return nil, err
	}

	if err := gormDB.Use(tracing.NewPlugin(
		tracing.WithDBSystem("sqlite"),
		tracing.WithoutServerAddress(),
		tracing.WithoutMetrics(),
	)); err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)

	return &Sqlite{gormDB: gormDB, sqlDB: sqlDB}, nil
}

func (w *Sqlite) Shutdown(ctx context.Context) error {
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

func (w *Sqlite) Master() *gorm.DB {
	return w.gormDB
}

func (w *Sqlite) Slave() *gorm.DB {
	return w.gormDB
}

func AddConnectionOptions(dbPath string, connOpts []ConnectionOption) string {
	if len(connOpts) == 0 {
		return dbPath
	}
	dbPath += "?"
	for i := range connOpts {
		dbPath += connOpts[i].Key + "=" + connOpts[i].Value
		if i < len(connOpts)-1 {
			dbPath += "&"
		}
	}
	return dbPath
}

func (w *Sqlite) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (w *Sqlite) IsDuplicateKeyError(err error) bool {
	sqliteErr, ok := err.(interface{ Code() int })
	if ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}
	return false
}
