package domain

import (
	"errors"
	"time"
)

// Message types for chat
const (
	MessageTypeText   = "text"
	MessageTypeJoin   = "join"
	MessageTypeLeave  = "leave"
	MessageTypeSystem = "system"
)

var MessageTypes = []string{
	MessageTypeText,
	MessageTypeJoin,
	MessageTypeLeave,
	MessageTypeSystem,
}

// Room types
const (
	RoomTypePublic  = "public"
	RoomTypePrivate = "private"
)

var RoomTypes = []string{
	RoomTypePublic,
	RoomTypePrivate,
}

// Errors
var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidMessage    = errors.New("invalid message")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrRoomFull          = errors.New("room is full")
	ErrUserAlreadyInRoom = errors.New("user already in room")
)

// UserInfo represents basic user information for chat without sensitive data
type UserInfo struct {
	UUID     string    `json:"uuid"`     // Non-sensitive unique identifier
	Username string    `json:"username"` // Display name for chat
	JoinedAt time.Time `json:"joined_at"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	User      UserInfo  `json:"user"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Room represents a chat room
type Room struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	MaxUsers  int                 `json:"max_users"`
	UserCount int                 `json:"user_count"`
	Users     map[string]UserInfo `json:"users"` // keyed by user UUID
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// Client represents a websocket client
type Client struct {
	ID       string
	User     UserInfo
	RoomID   string
	Send     chan []byte
	JoinedAt time.Time
}

// CreateRoomData represents data for creating a new room
type CreateRoomData struct {
	Name     string `json:"name"      validate:"required,min=1,max=100"`
	Type     string `json:"type"      validate:"required,oneof=public private"`
	MaxUsers int    `json:"max_users" validate:"min=2,max=1000"`
}

// SendMessageData represents data for sending a message
type SendMessageData struct {
	Type    string `json:"type"    validate:"required,oneof=text system"`
	Content string `json:"content" validate:"required,max=1000"`
	RoomID  string `json:"room_id" validate:"required"`
}
