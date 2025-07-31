package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatRoom represents a chat room in the database
type ChatRoom struct {
	ID        int
	UUID      uuid.UUID
	Name      string
	Type      string // public, private
	MaxUsers  int
	CreatedBy int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt

	// Relationships
	Messages []ChatMessage `gorm:"foreignKey:RoomID"`
}

func (*ChatRoom) TableName() string { return "chat_rooms" }

// ChatMessage represents a chat message in the database
type ChatMessage struct {
	ID        int
	UUID      uuid.UUID
	Type      string // text, join, leave, system
	Content   string
	UserID    int
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (*ChatMessage) TableName() string { return "chat_messages" }

// ChatRoomMember represents room membership
type ChatRoomMember struct {
	RoomID    int
	UserID    int
	JoinedAt  time.Time
	LeftAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (*ChatRoomMember) TableName() string { return "chat_room_members" }