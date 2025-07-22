package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"github.com/hasansino/go42/internal/database"
	"github.com/hasansino/go42/internal/outbox/models"
)

type Repository struct {
	*database.BaseRepository
}

func New(baseRepository *database.BaseRepository) *Repository {
	return &Repository{baseRepository}
}

func (r *Repository) NewOutboxMessage(ctx context.Context, msg *models.Message) error {
	err := r.GetTx(ctx).Create(msg).Error
	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}
	return nil
}

func (r *Repository) GetUnprocessedMessages(ctx context.Context, limit int) ([]models.Message, error) {
	var messages []models.Message
	result := r.
		GetTx(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsSkipLocked}).
		Where("status = ?", models.MessageStatusPending).
		Limit(limit).Find(&messages)
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching messages: %w", result.Error)
	}
	return messages, nil
}

func (r *Repository) SaveProcessedMessages(ctx context.Context, messages []models.Message) error {
	var ids []uuid.UUID
	for _, message := range messages {
		ids = append(ids, message.ID)
	}
	result := r.
		GetTx(ctx).
		Model(&models.Message{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":       models.MessageStatusProcessed,
			"processed_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("error updating processed messages: %w", result.Error)
	}
	return nil
}

func (r *Repository) SaveFailedMessages(ctx context.Context, messages []models.Message) error {
	for _, message := range messages {
		result := r.GetTx(ctx).Save(&message)
		if result.Error != nil {
			return fmt.Errorf("error saving message with ID %d: %w", message.ID, result.Error)
		}
	}
	return nil
}
