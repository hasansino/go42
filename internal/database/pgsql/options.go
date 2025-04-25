package pgsql

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Option func(w *Wrapper, gorm *gorm.DB, db *sql.DB)

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(w *Wrapper, gorm *gorm.DB, db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(w *Wrapper, gorm *gorm.DB, db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

func WithMaxIdleConns(n int) Option {
	return func(w *Wrapper, gorm *gorm.DB, db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithMaxOpenConns(n int) Option {
	return func(w *Wrapper, gorm *gorm.DB, db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}

func WithQueryTimeout(d time.Duration) Option {
	return func(w *Wrapper, gorm *gorm.DB, db *sql.DB) {
		w.timeout = d
	}
}
