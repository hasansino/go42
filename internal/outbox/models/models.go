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
	ID            uuid.UUID    `json:"id"             gorm:"primaryKey"`
	AggregateID   int          `json:"aggregate_id"   gorm:""`
	AggregateType string       `json:"aggregate_type" gorm:""`
	Topic         string       `json:"topic"          gorm:""`
	Payload       []byte       `json:"payload"        gorm:""`
	CreatedAt     time.Time    `json:"created_at"     gorm:""`
	ProcessedAt   sql.NullTime `json:"processed_at"   gorm:""`
	Status        string       `json:"status"         gorm:""`
	RetryCount    int          `json:"retry_count"    gorm:""`
	MaxRetries    int          `json:"max_retries"    gorm:""`
	LastError     string       `json:"last_error"     gorm:""`
	Metadata      string       `json:"metadata"       gorm:""`
}

func (m *Message) TableName() string {
	return "transactional_outbox"
}
