package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"

	"github.com/mattn/go-sqlite3"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

type Wrapper struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

func NewWrapper(dbPath string) (*Wrapper, error) {
	gormDB, err := gorm.Open(
		sqlite.Open(dbPath),
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

func IsNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func IsDuplicateKeyError(err error) bool {
	var sqErr sqlite3.Error
	if errors.As(err, &sqErr) {
		return sqErr.ExtendedCode == 2067
	}
	return false
}
