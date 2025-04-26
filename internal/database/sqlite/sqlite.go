package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite"
)

type Wrapper struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

func NewWrapper(dbPath string, opts ...Option) (*Wrapper, error) {
	gormDB, err := gorm.Open(
		sqlite.Open(AddConnectionOptions(dbPath, opts...)),
		&gorm.Config{
			PrepareStmt: true,
			Logger: slogGorm.New(
				// slogGorm.WithIgnoreTrace(),
				slogGorm.WithHandler(slog.Default().Handler().WithAttrs(
					[]slog.Attr{slog.String("system", "gorm")},
				)),
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

func AddConnectionOptions(dbPath string, opts ...Option) string {
	if len(opts) == 0 {
		return dbPath
	}
	dbPath += "?"
	for i, opt := range opts {
		key, value := opt()
		dbPath += key + "=" + value
		if i < len(opts)-1 {
			dbPath += "&"
		}
	}
	return dbPath
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicateKeyError(err error) bool {
	return false
}
