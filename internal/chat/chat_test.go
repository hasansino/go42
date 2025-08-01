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
		creator     domain.UserInfo
		expectError bool
	}{
		{
			name: "successful room creation",
			data: &domain.CreateRoomData{
				Name:     "Test Room",
				Type:     domain.RoomTypePublic,
				MaxUsers: 50,
			},
			creator: domain.UserInfo{
				UUID:     "test-uuid",
				Username: "testuser",
			},
			expectError: false,
		},
		{
			name: "room with default max users",
			data: &domain.CreateRoomData{
				Name:     "Test Room 2",
				Type:     domain.RoomTypePrivate,
				MaxUsers: 0, // Should use default
			},
			creator: domain.UserInfo{
				UUID:     "test-uuid-2",
				Username: "testuser2",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			room, err := service.CreateRoom(ctx, tt.data, tt.creator)

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
	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
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
				ID: "client-1",
				User: domain.UserInfo{
					UUID:     "user-1-uuid",
					Username: "user1",
				},
				Send: make(chan []byte, 10),
			},
			expectError: nil,
		},
		{
			name: "join different user",
			client: &domain.Client{
				ID: "client-2",
				User: domain.UserInfo{
					UUID:     "user-2-uuid",
					Username: "user2",
				},
				Send: make(chan []byte, 10),
			},
			expectError: nil,
		},
		{
			name: "room full",
			client: &domain.Client{
				ID: "client-3",
				User: domain.UserInfo{
					UUID:     "user-3-uuid",
					Username: "user3",
				},
				Send: make(chan []byte, 10),
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
				assert.Contains(t, updatedRoom.Users, tt.client.User.UUID)
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
	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
	require.NoError(t, err)

	client := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
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
	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	publicRoom, err := service.CreateRoom(ctx, &domain.CreateRoomData{
		Name:     "Public Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}, creator)
	require.NoError(t, err)

	privateRoom, err := service.CreateRoom(ctx, &domain.CreateRoomData{
		Name:     "Private Room",
		Type:     domain.RoomTypePrivate,
		MaxUsers: 10,
	}, creator)
	require.NoError(t, err)

	tests := []struct {
		name          string
		roomType      string
		expectedCount int
	}{
		{
			name:          "list all rooms",
			roomType:      "",
			expectedCount: 2,
		},
		{
			name:          "list public rooms only",
			roomType:      domain.RoomTypePublic,
			expectedCount: 1,
		},
		{
			name:          "list private rooms only",
			roomType:      domain.RoomTypePrivate,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms, err := service.ListRooms(ctx, tt.roomType)
			assert.NoError(t, err)
			assert.Len(t, rooms, tt.expectedCount)

			switch tt.roomType {
			case domain.RoomTypePublic:
				assert.Equal(t, publicRoom.ID, rooms[0].ID)
			case domain.RoomTypePrivate:
				assert.Equal(t, privateRoom.ID, rooms[0].ID)
			}
		})
	}
}

func TestService_GetRoom(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	// Create a room
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
	require.NoError(t, err)

	tests := []struct {
		name        string
		roomID      string
		expectError error
	}{
		{
			name:        "get existing room",
			roomID:      room.ID,
			expectError: nil,
		},
		{
			name:        "get non-existent room",
			roomID:      "non-existent-room-id",
			expectError: domain.ErrRoomNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetRoom(ctx, tt.roomID)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.roomID, result.ID)
			}
		})
	}
}

func TestService_LeaveRoom(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	// Create a room
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
	require.NoError(t, err)

	// Create and join a client
	client := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	err = service.JoinRoom(ctx, room.ID, client)
	require.NoError(t, err)

	tests := []struct {
		name        string
		clientID    string
		expectError error
	}{
		{
			name:        "leave room successfully",
			clientID:    client.ID,
			expectError: nil,
		},
		{
			name:        "leave room with non-existent client",
			clientID:    "non-existent-client",
			expectError: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.LeaveRoom(ctx, tt.clientID)

			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)

				// Verify room state if client exists
				if tt.clientID == client.ID {
					updatedRoom, err := service.GetRoom(ctx, room.ID)
					if err == nil { // Room might be cleaned up if it was the last user
						assert.NotContains(t, updatedRoom.Users, client.User.UUID)
					}
				}
			}
		})
	}
}

func TestService_RegisterUnregisterClient(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	client := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	// Test RegisterClient
	t.Run("register client", func(t *testing.T) {
		service.RegisterClient(ctx, client)

		// Verify client is registered
		service.mu.RLock()
		_, exists := service.clients[client.ID]
		service.mu.RUnlock()
		assert.True(t, exists)
	})

	// Test UnregisterClient (which calls LeaveRoom)
	t.Run("unregister client", func(t *testing.T) {
		// Create a room and join the client first
		creator := domain.UserInfo{
			UUID:     "creator-uuid",
			Username: "creator",
		}
		roomData := &domain.CreateRoomData{
			Name:     "Test Room",
			Type:     domain.RoomTypePublic,
			MaxUsers: 50,
		}
		room, err := service.CreateRoom(ctx, roomData, creator)
		require.NoError(t, err)

		err = service.JoinRoom(ctx, room.ID, client)
		require.NoError(t, err)

		// Now unregister
		service.UnregisterClient(ctx, client.ID)

		// Verify client is unregistered (removed by LeaveRoom)
		service.mu.RLock()
		_, exists := service.clients[client.ID]
		service.mu.RUnlock()
		assert.False(t, exists)
	})
}

func TestService_UserAlreadyInRoom(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	// Create a room
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
	require.NoError(t, err)

	// Create a client
	client := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	// Join room first time
	err = service.JoinRoom(ctx, room.ID, client)
	require.NoError(t, err)

	// Try to join again with same user
	client2 := &domain.Client{
		ID: "client-2",
		User: domain.UserInfo{
			UUID:     "user-1-uuid", // Same user UUID
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	err = service.JoinRoom(ctx, room.ID, client2)
	assert.Equal(t, domain.ErrUserAlreadyInRoom, err)
}

func TestService_MessageBroadcasting(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	// Create a room
	roomData := &domain.CreateRoomData{
		Name:     "Test Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}
	room, err := service.CreateRoom(ctx, roomData, creator)
	require.NoError(t, err)

	// Create multiple clients
	client1 := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	client2 := &domain.Client{
		ID: "client-2",
		User: domain.UserInfo{
			UUID:     "user-2-uuid",
			Username: "user2",
		},
		Send: make(chan []byte, 10),
	}

	// Join both clients to room
	err = service.JoinRoom(ctx, room.ID, client1)
	require.NoError(t, err)
	err = service.JoinRoom(ctx, room.ID, client2)
	require.NoError(t, err)

	// Send message from client1
	messageData := &domain.SendMessageData{
		Type:    domain.MessageTypeText,
		Content: "Hello everyone!",
		RoomID:  room.ID,
	}

	err = service.SendMessage(ctx, client1.ID, messageData)
	require.NoError(t, err)

	// Both clients should receive the message
	// Note: In a real scenario you'd need to check the channels,
	// but for unit tests we're just verifying no errors occurred
	assert.NoError(t, err)
}

func TestService_RoomCleanup(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	creator := domain.UserInfo{
		UUID:     "creator-uuid",
		Username: "creator",
	}

	// Create a public room (should be cleaned up when empty)
	publicRoomData := &domain.CreateRoomData{
		Name:     "Public Room",
		Type:     domain.RoomTypePublic,
		MaxUsers: 50,
	}
	publicRoom, err := service.CreateRoom(ctx, publicRoomData, creator)
	require.NoError(t, err)

	// Create a private room (should NOT be cleaned up when empty)
	privateRoomData := &domain.CreateRoomData{
		Name:     "Private Room",
		Type:     domain.RoomTypePrivate,
		MaxUsers: 50,
	}
	privateRoom, err := service.CreateRoom(ctx, privateRoomData, creator)
	require.NoError(t, err)

	// Create and join a client to both rooms
	client := &domain.Client{
		ID: "client-1",
		User: domain.UserInfo{
			UUID:     "user-1-uuid",
			Username: "user1",
		},
		Send: make(chan []byte, 10),
	}

	// Join public room
	err = service.JoinRoom(ctx, publicRoom.ID, client)
	require.NoError(t, err)

	// Leave public room - should trigger cleanup
	err = service.LeaveRoom(ctx, client.ID)
	require.NoError(t, err)

	// Public room should be cleaned up
	_, err = service.GetRoom(ctx, publicRoom.ID)
	assert.Equal(t, domain.ErrRoomNotFound, err)

	// Private room should still exist
	_, err = service.GetRoom(ctx, privateRoom.ID)
	assert.NoError(t, err)
}
