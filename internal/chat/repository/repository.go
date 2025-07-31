package repository

import (
	"context"
	"fmt"

	"github.com/hasansino/go42/internal/chat/domain"
	"github.com/hasansino/go42/internal/chat/models"
	"github.com/hasansino/go42/internal/database"
)

type Repository struct {
	*database.BaseRepository
}

func New(baseRepository *database.BaseRepository) *Repository {
	return &Repository{
		BaseRepository: baseRepository,
	}
}

func (r *Repository) CreateRoom(ctx context.Context, room *models.ChatRoom) error {
	err := r.GetTx(ctx).Create(room).Error
	if err != nil {
		if r.IsDuplicateKeyError(err) {
			return domain.ErrRoomAlreadyExists
		}
		return fmt.Errorf("error creating room: %w", err)
	}
	return nil
}

func (r *Repository) GetRoomByID(ctx context.Context, id int) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.GetReadDB(ctx).Where("id = ?", id).First(&room).Error
	if r.IsNotFoundError(err) {
		return nil, domain.ErrRoomNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting room: %w", err)
	}
	return &room, nil
}

func (r *Repository) GetRoomByUUID(ctx context.Context, uuid string) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.GetReadDB(ctx).Where("uuid = ?", uuid).First(&room).Error
	if r.IsNotFoundError(err) {
		return nil, domain.ErrRoomNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting room: %w", err)
	}
	return &room, nil
}

func (r *Repository) ListRooms(ctx context.Context, roomType string, limit, offset int) ([]*models.ChatRoom, error) {
	var rooms []*models.ChatRoom
	
	query := r.GetReadDB(ctx).Limit(limit).Offset(offset).Order("created_at DESC")
	if roomType != "" {
		query = query.Where("type = ?", roomType)
	}
	
	err := query.Find(&rooms).Error
	if err != nil {
		return nil, fmt.Errorf("error listing rooms: %w", err)
	}
	
	return rooms, nil
}

func (r *Repository) CreateMessage(ctx context.Context, message *models.ChatMessage) error {
	err := r.GetTx(ctx).Create(message).Error
	if err != nil {
		return fmt.Errorf("error creating message: %w", err)
	}
	return nil
}

func (r *Repository) GetMessagesByRoomID(ctx context.Context, roomID int, limit, offset int) ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	
	err := r.GetReadDB(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	
	return messages, nil
}

func (r *Repository) AddRoomMember(ctx context.Context, member *models.ChatRoomMember) error {
	err := r.GetTx(ctx).Create(member).Error
	if err != nil {
		return fmt.Errorf("error adding room member: %w", err)
	}
	return nil
}

func (r *Repository) RemoveRoomMember(ctx context.Context, roomID, userID int) error {
	result := r.GetTx(ctx).
		Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).
		Update("left_at", "NOW()")
	
	if result.Error != nil {
		return fmt.Errorf("error removing room member: %w", result.Error)
	}
	
	return nil
}

func (r *Repository) IsUserInRoom(ctx context.Context, roomID, userID int) (bool, error) {
	var count int64
	err := r.GetReadDB(ctx).
		Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ? AND left_at IS NULL", roomID, userID).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("error checking room membership: %w", err)
	}
	
	return count > 0, nil
}

func (r *Repository) GetRoomMemberCount(ctx context.Context, roomID int) (int, error) {
	var count int64
	err := r.GetReadDB(ctx).
		Model(&models.ChatRoomMember{}).
		Where("room_id = ? AND left_at IS NULL", roomID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("error getting room member count: %w", err)
	}
	
	return int(count), nil
}