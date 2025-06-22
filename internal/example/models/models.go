package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Fruit struct {
	ID        int            `json:"id"         gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at" gorm:""`
	UpdatedAt time.Time      `json:"updated_at" gorm:""`
	DeletedAt gorm.DeletedAt `json:"-"          gorm:""`
	Name      string         `json:"name"       gorm:""`
}

func (m *Fruit) TableName() string {
	return "example_fruits"
}

type FruitEventLog struct {
	ID         uuid.UUID `gorm:"primaryKey"`
	OccurredAt time.Time `gorm:""`
	CreatedAt  time.Time `gorm:""`
	FruitID    int       `gorm:""`
	EventType  string    `gorm:""`
	Data       []byte    `gorm:""`
	Metadata   string    `gorm:""`
}

func (m *FruitEventLog) TableName() string {
	return "example_fruits_events_log"
}
