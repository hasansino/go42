package database

import (
	"context"

	"gorm.io/gorm"
)

type Database interface {
	Master() *gorm.DB
	Slave() *gorm.DB
	Ping(ctx context.Context) error
	Shutdown(ctx context.Context) error
	IsNotFoundError(err error) bool
	IsDuplicateKeyError(err error) bool
}
