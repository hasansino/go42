package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const (
	MessageStatusPending   string = "pending"
	MessageStatusProcessed string = "processed"
	MessageStatusFailed    string = "failed"
)

type Message struct {
	ID            uuid.UUID    `json:"id"`
	AggregateID   int          `json:"aggregate_id"`
	AggregateType string       `json:"aggregate_type"`
	Topic         string       `json:"topic"`
	Payload       []byte       `json:"payload"`
	CreatedAt     time.Time    `json:"created_at"`
	ProcessedAt   sql.NullTime `json:"processed_at"`
	Status        string       `json:"status"`
	RetryCount    int          `json:"retry_count"`
	MaxRetries    int          `json:"max_retries"`
	LastError     string       `json:"last_error"`
	Metadata      string       `json:"metadata"`
}

func (m *Message) TableName() string {
	return "transactional_outbox"
}
