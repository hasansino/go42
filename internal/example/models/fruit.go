package models

import (
	"time"

	"gorm.io/gorm"
)

type Fruit struct {
	ID        uint32         `json:"id"   gorm:"primarykey"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"    gorm:"index"`
	Name      string         `json:"name"`
}

func (m *Fruit) TableName() string {
	return "example_fruits"
}
