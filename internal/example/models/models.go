package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Fruit struct {
	ID        int            `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-"`
	Name      string         `json:"name"`
}

func (m *Fruit) TableName() string {
	return "example_fruits"
}

type FruitEventLog struct {
	ID         uuid.UUID
	OccurredAt time.Time
	CreatedAt  time.Time
	FruitID    int
	EventType  string
	Data       []byte
	Metadata   string
}

func (m *FruitEventLog) TableName() string {
	return "example_fruits_events_log"
}
