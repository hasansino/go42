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

	master     *gorm.DB
	masterConn *sql.DB
	slave      *gorm.DB
	slaveConn  *sql.DB

	connMaxIdleTime time.Duration
	connMaxLifetime time.Duration
	maxOpenConns    int
	maxIdleConns    int

	queryLogging bool
}

func New(masterDSN string, slaveDSN string, opts ...Option) (*Mysql, error) {
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

	// ---

	masterConn, err := gorm.Open(
		mysql.Open(masterDSN),
		&gorm.Config{
			PrepareStmt: true,
			Logger:      slogGorm.New(slogGormOpts...),
		})
	if err != nil {
		return nil, err
	}

	masterConnDB, err := masterConn.DB()
	if err != nil {
		return nil, err
	}

	masterConnDB.SetMaxOpenConns(w.maxOpenConns)
	masterConnDB.SetMaxIdleConns(w.maxIdleConns)
	masterConnDB.SetConnMaxLifetime(w.connMaxLifetime)
	masterConnDB.SetConnMaxIdleTime(w.connMaxIdleTime)

	w.master = masterConn
	w.masterConn = masterConnDB

	// ---

	if len(slaveDSN) > 0 {
		slaveConn, err := gorm.Open(
			mysql.Open(masterDSN),
			&gorm.Config{
				PrepareStmt: true,
				Logger:      slogGorm.New(slogGormOpts...),
			})
		if err != nil {
			return nil, err
		}

		slaveConnDB, err := slaveConn.DB()
		if err != nil {
			return nil, err
		}

		slaveConnDB.SetMaxOpenConns(w.maxOpenConns)
		slaveConnDB.SetMaxIdleConns(w.maxIdleConns)
		slaveConnDB.SetConnMaxLifetime(w.connMaxLifetime)
		slaveConnDB.SetConnMaxIdleTime(w.connMaxIdleTime)

		w.slave = slaveConn
		w.slaveConn = slaveConnDB
	} else {
		w.slave = masterConn
		w.slaveConn = masterConnDB
	}

	return w, nil
}

func (w *Mysql) Shutdown(ctx context.Context) error {
	doneChan := make(chan error)
	go func() {
		masterErr := w.masterConn.Close()
		slaveErr := w.slaveConn.Close()
		doneChan <- errors.Join(masterErr, slaveErr)
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-doneChan:
		return err
	}
}

func (w *Mysql) Master() *gorm.DB {
	return w.master
}

func (w *Mysql) Slave() *gorm.DB {
	return w.slave
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
