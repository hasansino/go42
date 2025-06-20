package models

import (
	"time"

	"gorm.io/gorm"
)

type Fruit struct {
	ID        int            `json:"id"   db:"id"`
	CreatedAt time.Time      `json:"-"    db:"created_at"`
	UpdatedAt time.Time      `json:"-"    db:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"    db:"deleted_at"`
	Name      string         `json:"name" db:"name"`
}

func (m *Fruit) TableName() string {
	return "example_fruits"
}
