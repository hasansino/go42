package domain

import (
	"time"

	"github.com/google/uuid"
)

const MaxRetries = 3

type Message struct {
	AggregateID   int    `v:"required,gte=1"`
	AggregateType string `v:"required,min=3,max=100"`
	Payload       []byte `v:"required,min=2"`
	Metadata      string `v:"omitzero,max=1000"`
}

type Event struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	AggregateID   int       `json:"aggregate_id"`
	AggregateType string    `json:"aggregate_type"`
	Payload       []byte    `json:"payload"`
	Metadata      string    `json:"metadata"`
}
