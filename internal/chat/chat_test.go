package chat

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hasansino/go42/internal/chat/domain"
)

func TestService_CreateRoom(t *testing.T) {
	service := NewService()

	tests := []struct {
		name        string
		data        *domain.CreateRoomData
		creatorID   int
		creatorUUID string
		creatorEmail string
		expectError bool
	}{
		{
			name: "successful room creation",
			data: &domain.CreateRoomData{
				Name:     "Test Room",
				Type:     domain.RoomTypePublic,
				MaxUsers: 50,
			},
			creatorID:   1,
			creatorUUID: "test-uuid",
			creatorEmail: "test@example.com",
			expectError: false,
		},
		{
			name: "room with default max users",
			data: &domain.CreateRoomData{
				Name:     "Test Room 2",
				Type:     domain.RoomTypePrivate,
				MaxUsers: 0, // Should use default
			},
			creatorID:   1,
			creatorUUID: "test-uuid",
			creatorEmail: "test@example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			room, err := service.CreateRoom(ctx, tt.data, tt.creatorID, tt.creatorUUID, tt.creatorEmail)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, room)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, room)
				assert.Equal(t, tt.data.Name, room.Name)
				assert.Equal(t, tt.data.Type, room.Type)
				assert.NotEmpty(t, room.ID)
				assert.Equal(t, 0, room.UserCount)
				assert.NotNil(t, room.Users)

				if tt.data.MaxUsers > 0 {
					assert.Equal(t, tt.data.MaxUsers, room.MaxUsers)
				} else {
					assert.Equal(t, 100, room.MaxUsers) // Default value
				}
			}
		})
	}
}

func TestService_JoinRoom(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Create a room first
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 2,
	}
	room, err := service.CreateRoom(ctx, roomData, 1, "creator-uuid", "creator@example.com")
	require.NoError(t, err)
	require.NotNil(t, room)

	tests := []struct {
		name        string
		client      *domain.Client
		expectError error
	}{
		{
			name: "successful join",
			client: &domain.Client{
				ID:        "client-1",
				UserID:    1,
				UserUUID:  "user-1-uuid",
				UserEmail: "user1@example.com",
				Send:      make(chan []byte, 10),
			},
			expectError: nil,
		},
		{
			name: "join different user",
			client: &domain.Client{
				ID:        "client-2",
				UserID:    2,
				UserUUID:  "user-2-uuid",
				UserEmail: "user2@example.com",
				Send:      make(chan []byte, 10),
			},
			expectError: nil,
		},
		{
			name: "room full",
			client: &domain.Client{
				ID:        "client-3",
				UserID:    3,
				UserUUID:  "user-3-uuid",
				UserEmail: "user3@example.com",
				Send:      make(chan []byte, 10),
			},
			expectError: domain.ErrRoomFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.JoinRoom(ctx, room.ID, tt.client)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, room.ID, tt.client.RoomID)

				// Check room state
				updatedRoom, err := service.GetRoom(ctx, room.ID)
				assert.NoError(t, err)
				assert.Contains(t, updatedRoom.Users, tt.client.UserID)
			}
		})
	}
}

func TestService_SendMessage(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Create a room and join a client
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 10,
	}
	room, err := service.CreateRoom(ctx, roomData, 1, "creator-uuid", "creator@example.com")
	require.NoError(t, err)

	client := &domain.Client{
		ID:        "client-1",
		UserID:    1,
		UserUUID:  "user-1-uuid",
		UserEmail: "user1@example.com",
		Send:      make(chan []byte, 10),
	}

	err = service.JoinRoom(ctx, room.ID, client)
	require.NoError(t, err)

	tests := []struct {
		name        string
		clientID    string
		data        *domain.SendMessageData
		expectError error
	}{
		{
			name:     "successful message send",
			clientID: client.ID,
			data: &domain.SendMessageData{
				Type:    domain.MessageTypeText,
				Content: "Hello, world!",
				RoomID:  room.ID,
			},
			expectError: nil,
		},
		{
			name:     "client not found",
			clientID: "non-existent-client",
			data: &domain.SendMessageData{
				Type:    domain.MessageTypeText,
				Content: "Hello, world!",
				RoomID:  room.ID,
			},
			expectError: domain.ErrUserNotFound,
		},
		{
			name:     "unauthorized room access",
			clientID: client.ID,
			data: &domain.SendMessageData{
				Type:    domain.MessageTypeText,
				Content: "Hello, world!",
				RoomID:  "different-room-id",
			},
			expectError: domain.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendMessage(ctx, tt.clientID, tt.data)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ListRooms(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	// Create multiple rooms
	publicRoom, err := service.CreateRoom(ctx, &domain.CreateRoomData{
		Name:     "Public Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}, 1, "creator-uuid", "creator@example.com")
	require.NoError(t, err)

	privateRoom, err := service.CreateRoom(ctx, &domain.CreateRoomData{
		Name:     "Private Room",
		Type:     domain.RoomTypePrivate,
		MaxUsers: 10,
	}, 1, "creator-uuid", "creator@example.com")
	require.NoError(t, err)

	tests := []struct {
		name         string
		roomType     string
		expectedCount int
	}{
		{
			name:         "list all rooms",
			roomType:     "",
			expectedCount: 2,
		},
		{
			name:         "list public rooms only",
			roomType:     domain.RoomTypePublic,
			expectedCount: 1,
		},
		{
			name:         "list private rooms only",
			roomType:     domain.RoomTypePrivate,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms, err := service.ListRooms(ctx, tt.roomType)
			assert.NoError(t, err)
			assert.Len(t, rooms, tt.expectedCount)

			if tt.roomType == domain.RoomTypePublic {
				assert.Equal(t, publicRoom.ID, rooms[0].ID)
			} else if tt.roomType == domain.RoomTypePrivate {
				assert.Equal(t, privateRoom.ID, rooms[0].ID)
			}
		})
	}
}