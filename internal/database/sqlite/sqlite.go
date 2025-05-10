package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"

	"github.com/glebarez/sqlite"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/gorm"
	sqlitelib "modernc.org/sqlite/lib"
)

type Wrapper struct {
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

func New(dbPath string, opts ...Option) (*Wrapper, error) {
	w := new(Wrapper)

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
		sqlite.Open(AddConnectionOptions(dbPath, w.connOpts)),
		&gorm.Config{
			PrepareStmt: true,
			Logger:      slogGorm.New(slogGormOpts...),
		})

	if err != nil {
		log.Fatalf("failed to connect to SQLite database: %v", err)
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)

	return &Wrapper{gormDB: gormDB, sqlDB: sqlDB}, nil
}

func (w *Wrapper) Shutdown(ctx context.Context) error {
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

func (w *Wrapper) GormDB() *gorm.DB {
	return w.gormDB
}

func (w *Wrapper) SqlDB() *sql.DB {
	return w.sqlDB
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

func (w *Wrapper) IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func (w *Wrapper) IsDuplicateKeyError(err error) bool {
	sqliteErr, ok := err.(interface{ Code() int })
	if ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}
	return false
}
