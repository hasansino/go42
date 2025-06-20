package models

import (
	"time"
)

type Event struct {
	ID        int       `json:"id"   db:"id"`
	CreatedAt time.Time `json:"-"    db:"created_at"`
	Data      string    `json:"data" db:"data"`
}

func (m *Event) TableName() string {
	return "example_fruits_events"
}
