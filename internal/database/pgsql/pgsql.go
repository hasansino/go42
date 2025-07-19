package pgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgx/v5/pgconn"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
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

func Open(ctx context.Context, masterDSN string, slaveDSN string, opts ...Option) (*Postgres, error) {
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
		slogGorm.SetLogLevel(slogGorm.DefaultLogType, slog.LevelDebug),
	}

	if w.queryLogging {
		slogGormOpts = append(slogGormOpts, slogGorm.WithTraceAll())
	} else {
		slogGormOpts = append(slogGormOpts, slogGorm.WithIgnoreTrace())
	}

	// ---

	masterConn, err := w.connect(
		ctx, masterDSN,
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
		slaveConn, err := w.connect(
			ctx, slaveDSN,
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

func (w *Postgres) connect(ctx context.Context, dsn string, config *gorm.Config) (*gorm.DB, error) {
	db, err := retry.DoWithData[*gorm.DB](func() (*gorm.DB, error) {
		conn, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), config)
		if err != nil {
			return nil, fmt.Errorf("failed to open database connection: %w", err)
		}
		connDB, err := conn.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %w", err)
		}
		if err := connDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
		return conn, nil
	},
		retry.Context(ctx),
		retry.Attempts(10),
		retry.Delay(2*time.Second),
		retry.MaxDelay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			w.logger.WarnContext(
				ctx,
				"database connection attempt failed, retrying...",
				slog.Any("attempt", n+1),
				slog.String("error", err.Error()),
			)
		}),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (w *Postgres) Shutdown(ctx context.Context) error {
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

func (w *Postgres) Master() *gorm.DB {
	return w.master
}

func (w *Postgres) Slave() *gorm.DB {
	return w.slave
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
