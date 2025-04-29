package sqlite

import (
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
	logger   *slog.Logger
	gormDB   *gorm.DB
	sqlDB    *sql.DB
	connOpts []ConnectionOption
}

type ConnectionOption struct {
	Key   string
	Value string
}

func NewWrapper(dbPath string, opts ...Option) (*Wrapper, error) {
	w := new(Wrapper)
	for _, opt := range opts {
		opt(w)
	}
	gormDB, err := gorm.Open(
		sqlite.Open(AddConnectionOptions(dbPath, w.connOpts)),
		&gorm.Config{
			PrepareStmt: true,
			Logger: slogGorm.New(
				// slogGorm.WithIgnoreTrace(),
				slogGorm.WithHandler(w.logger.Handler()),
				// log level translations: when gorm sends X level -> slog handles it as Y level
				slogGorm.SetLogLevel(slogGorm.ErrorLogType, slog.LevelDebug), // exposes query
				slogGorm.SetLogLevel(slogGorm.SlowQueryLogType, slog.LevelWarn),
				slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelInfo),
			),
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

func (w *Wrapper) Close() error {
	return w.sqlDB.Close()
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

func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicateKeyError(err error) bool {
	sqliteErr, ok := err.(interface{ Code() int })
	if ok {
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}
	return false
}
