package models

import (
	"time"
)

type Event struct {
	ID        int       `json:"id"   gorm:"primarykey"`
	CreatedAt time.Time `json:"-"`
	Data      string    `json:"data"`
}

func (m *Event) TableName() string {
	return "example_events"
}
