package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hasansino/go42/internal/chat/domain"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

// Service represents the chat service
type Service struct {
	mu      sync.RWMutex
	rooms   map[string]*domain.Room
	clients map[string]*domain.Client
	logger  *slog.Logger
	options serviceOptions
}

// serviceOptions holds configuration options for the service
type serviceOptions struct {
	logger            *slog.Logger
	maxRoomsPerUser   int
	maxMessagesPerMin int
	defaultMaxUsers   int
}

// NewService creates a new chat service instance
func NewService(opts ...Option) *Service {
	options := serviceOptions{
		logger:            slog.Default(),
		maxRoomsPerUser:   10,
		maxMessagesPerMin: 30,
		defaultMaxUsers:   100,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Service{
		rooms:   make(map[string]*domain.Room),
		clients: make(map[string]*domain.Client),
		logger:  options.logger,
		options: options,
	}
}

// CreateRoom creates a new chat room
func (s *Service) CreateRoom(
	ctx context.Context,
	data *domain.CreateRoomData,
	creator domain.UserInfo,
) (*domain.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	roomID := uuid.New().String()

	maxUsers := data.MaxUsers
	if maxUsers <= 0 {
		maxUsers = s.options.defaultMaxUsers
	}

	room := &domain.Room{
		ID:        roomID,
		Name:      data.Name,
		Type:      data.Type,
		MaxUsers:  maxUsers,
		UserCount: 0,
		Users:     make(map[string]domain.UserInfo),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.rooms[roomID] = room

	s.logger.InfoContext(ctx, "room created",
		slog.String("room_id", roomID),
		slog.String("name", data.Name),
		slog.String("type", data.Type),
		slog.String("creator_uuid", creator.UUID))

	return room, nil
}

// GetRoom retrieves a room by ID
func (s *Service) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return nil, domain.ErrRoomNotFound
	}

	return room, nil
}

// ListRooms lists all available rooms
func (s *Service) ListRooms(ctx context.Context, roomType string) ([]*domain.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rooms []*domain.Room
	for _, room := range s.rooms {
		if roomType == "" || room.Type == roomType {
			rooms = append(rooms, room)
		}
	}

	return rooms, nil
}

// JoinRoom adds a user to a room
func (s *Service) JoinRoom(ctx context.Context, roomID string, client *domain.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return domain.ErrRoomNotFound
	}

	if room.UserCount >= room.MaxUsers {
		return domain.ErrRoomFull
	}

	if _, exists := room.Users[client.User.UUID]; exists {
		return domain.ErrUserAlreadyInRoom
	}

	// Add user to room
	userInfo := domain.UserInfo{
		UUID:     client.User.UUID,
		Username: client.User.Username,
		JoinedAt: time.Now(),
	}

	room.Users[client.User.UUID] = userInfo
	room.UserCount++
	room.UpdatedAt = time.Now()

	// Update client
	client.RoomID = roomID
	s.clients[client.ID] = client

	s.logger.InfoContext(ctx, "user joined room",
		slog.String("room_id", roomID),
		slog.String("user_uuid", client.User.UUID))

	// Send join message to room
	joinMessage := &domain.Message{
		ID:        uuid.New().String(),
		Type:      domain.MessageTypeJoin,
		Content:   client.User.Username + " joined the room",
		User:      client.User,
		RoomID:    roomID,
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(roomID, joinMessage)

	return nil
}

// LeaveRoom removes a user from a room
func (s *Service) LeaveRoom(ctx context.Context, clientID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return domain.ErrUserNotFound
	}

	if client.RoomID == "" {
		return nil // User not in any room
	}

	room, exists := s.rooms[client.RoomID]
	if !exists {
		return domain.ErrRoomNotFound
	}

	// Remove user from room
	delete(room.Users, client.User.UUID)
	room.UserCount--
	room.UpdatedAt = time.Now()

	s.logger.InfoContext(ctx, "user left room",
		slog.String("room_id", client.RoomID),
		slog.String("user_uuid", client.User.UUID))

	// Send leave message to room
	leaveMessage := &domain.Message{
		ID:        uuid.New().String(),
		Type:      domain.MessageTypeLeave,
		Content:   client.User.Username + " left the room",
		User:      client.User,
		RoomID:    client.RoomID,
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(client.RoomID, leaveMessage)

	// Clean up client
	delete(s.clients, clientID)

	// Clean up empty rooms (except for private rooms which might be persistent)
	if room.UserCount == 0 && room.Type == domain.RoomTypePublic {
		delete(s.rooms, client.RoomID)
		s.logger.InfoContext(ctx, "empty room cleaned up", slog.String("room_id", client.RoomID))
	}

	return nil
}

// SendMessage sends a message to a room
func (s *Service) SendMessage(ctx context.Context, clientID string, data *domain.SendMessageData) error {
	s.mu.RLock()
	client, exists := s.clients[clientID]
	s.mu.RUnlock()

	if !exists {
		return domain.ErrUserNotFound
	}

	if client.RoomID != data.RoomID {
		return domain.ErrUnauthorized
	}

	message := &domain.Message{
		ID:        uuid.New().String(),
		Type:      data.Type,
		Content:   data.Content,
		User:      client.User,
		RoomID:    data.RoomID,
		Timestamp: time.Now(),
	}

	s.logger.InfoContext(ctx, "message sent",
		slog.String("message_id", message.ID),
		slog.String("room_id", data.RoomID),
		slog.String("user_uuid", client.User.UUID))

	s.broadcastToRoom(data.RoomID, message)

	return nil
}

// RegisterClient registers a new websocket client
func (s *Service) RegisterClient(ctx context.Context, client *domain.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client.ID] = client

	s.logger.InfoContext(ctx, "client registered",
		slog.String("client_id", client.ID),
		slog.String("user_uuid", client.User.UUID))
}

// UnregisterClient unregisters a websocket client
func (s *Service) UnregisterClient(ctx context.Context, clientID string) {
	// Leave room if client was in one
	_ = s.LeaveRoom(ctx, clientID)

	s.logger.InfoContext(ctx, "client unregistered", slog.String("client_id", clientID))
}

// broadcastToRoom sends a message to all clients in a room
func (s *Service) broadcastToRoom(roomID string, message *domain.Message) {
	_, exists := s.rooms[roomID]
	if !exists {
		return
	}

	// Find all clients in this room and send message
	for _, client := range s.clients {
		if client.RoomID == roomID {
			select {
			case client.Send <- s.encodeMessage(message):
			default:
				// Client's send channel is full, skip this client
			}
		}
	}
}

// encodeMessage converts a message to JSON bytes for sending
func (s *Service) encodeMessage(message *domain.Message) []byte {
	data, _ := json.Marshal(message)
	return data
}
