package pgsql

import (
	"database/sql"
	"time"
)

type Option func(db *sql.DB)

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(d)
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

func WithMaxIdleConns(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(n)
	}
}

func WithMaxOpenConns(n int) Option {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(n)
	}
}
