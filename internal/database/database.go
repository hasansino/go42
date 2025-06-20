package database

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type Database interface {
	DB() *sql.DB
	GormDB() *gorm.DB
	Shutdown(ctx context.Context) error
	IsNotFoundError(err error) bool
	IsDuplicateKeyError(err error) bool
}
