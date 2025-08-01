package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageTypes(t *testing.T) {
	expectedTypes := []string{
		MessageTypeText,
		MessageTypeJoin,
		MessageTypeLeave,
		MessageTypeSystem,
	}

	assert.Equal(t, expectedTypes, MessageTypes)
	assert.Equal(t, "text", MessageTypeText)
	assert.Equal(t, "join", MessageTypeJoin)
	assert.Equal(t, "leave", MessageTypeLeave)
	assert.Equal(t, "system", MessageTypeSystem)
}

func TestRoomTypes(t *testing.T) {
	expectedTypes := []string{
		RoomTypePublic,
		RoomTypePrivate,
	}

	assert.Equal(t, expectedTypes, RoomTypes)
	assert.Equal(t, "public", RoomTypePublic)
	assert.Equal(t, "private", RoomTypePrivate)
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "room not found",
			err:  ErrRoomNotFound,
			msg:  "room not found",
		},
		{
			name: "room already exists",
			err:  ErrRoomAlreadyExists,
			msg:  "room already exists",
		},
		{
			name: "user not found",
			err:  ErrUserNotFound,
			msg:  "user not found",
		},
		{
			name: "invalid message",
			err:  ErrInvalidMessage,
			msg:  "invalid message",
		},
		{
			name: "unauthorized",
			err:  ErrUnauthorized,
			msg:  "unauthorized",
		},
		{
			name: "room full",
			err:  ErrRoomFull,
			msg:  "room is full",
		},
		{
			name: "user already in room",
			err:  ErrUserAlreadyInRoom,
			msg:  "user already in room",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestUserInfo(t *testing.T) {
	now := time.Now()
	userInfo := UserInfo{
		UUID:     "test-uuid",
		Username: "testuser",
		JoinedAt: now,
	}

	assert.Equal(t, "test-uuid", userInfo.UUID)
	assert.Equal(t, "testuser", userInfo.Username)
	assert.Equal(t, now, userInfo.JoinedAt)
}

func TestMessage(t *testing.T) {
	now := time.Now()
	user := UserInfo{
		UUID:     "user-uuid",
		Username: "testuser",
		JoinedAt: now,
	}

	message := Message{
		ID:        "msg-123",
		Type:      MessageTypeText,
		Content:   "Hello world",
		User:      user,
		RoomID:    "room-123",
		Timestamp: now,
	}

	assert.Equal(t, "msg-123", message.ID)
	assert.Equal(t, MessageTypeText, message.Type)
	assert.Equal(t, "Hello world", message.Content)
	assert.Equal(t, user, message.User)
	assert.Equal(t, "room-123", message.RoomID)
	assert.Equal(t, now, message.Timestamp)
}

func TestRoom(t *testing.T) {
	now := time.Now()
	users := make(map[string]UserInfo)
	users["user1"] = UserInfo{
		UUID:     "user1",
		Username: "User 1",
		JoinedAt: now,
	}

	room := Room{
		ID:        "room-123",
		Name:      "Test Room",
		Type:      RoomTypePublic,
		MaxUsers:  50,
		UserCount: 1,
		Users:     users,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "room-123", room.ID)
	assert.Equal(t, "Test Room", room.Name)
	assert.Equal(t, RoomTypePublic, room.Type)
	assert.Equal(t, 50, room.MaxUsers)
	assert.Equal(t, 1, room.UserCount)
	assert.Len(t, room.Users, 1)
	assert.Contains(t, room.Users, "user1")
	assert.Equal(t, now, room.CreatedAt)
	assert.Equal(t, now, room.UpdatedAt)
}

func TestClient(t *testing.T) {
	now := time.Now()
	user := UserInfo{
		UUID:     "user-uuid",
		Username: "testuser",
		JoinedAt: now,
	}

	sendChan := make(chan []byte, 10)
	client := Client{
		ID:       "client-123",
		User:     user,
		RoomID:   "room-123",
		Send:     sendChan,
		JoinedAt: now,
	}

	assert.Equal(t, "client-123", client.ID)
	assert.Equal(t, user, client.User)
	assert.Equal(t, "room-123", client.RoomID)
	assert.Equal(t, sendChan, client.Send)
	assert.Equal(t, now, client.JoinedAt)
}

func TestCreateRoomData(t *testing.T) {
	data := CreateRoomData{
		Name:     "Test Room",
		Type:     RoomTypePublic,
		MaxUsers: 50,
	}

	assert.Equal(t, "Test Room", data.Name)
	assert.Equal(t, RoomTypePublic, data.Type)
	assert.Equal(t, 50, data.MaxUsers)
}

func TestSendMessageData(t *testing.T) {
	data := SendMessageData{
		Type:    MessageTypeText,
		Content: "Hello world",
		RoomID:  "room-123",
	}

	assert.Equal(t, MessageTypeText, data.Type)
	assert.Equal(t, "Hello world", data.Content)
	assert.Equal(t, "room-123", data.RoomID)
}

func TestMessageTypesContainment(t *testing.T) {
	// Test that all message type constants are in the MessageTypes slice
	expectedTypes := map[string]bool{
		MessageTypeText:   false,
		MessageTypeJoin:   false,
		MessageTypeLeave:  false,
		MessageTypeSystem: false,
	}

	for _, msgType := range MessageTypes {
		if _, exists := expectedTypes[msgType]; exists {
			expectedTypes[msgType] = true
		} else {
			t.Errorf("Unexpected message type in MessageTypes slice: %s", msgType)
		}
	}

	for msgType, found := range expectedTypes {
		if !found {
			t.Errorf("Message type constant %s not found in MessageTypes slice", msgType)
		}
	}
}

func TestRoomTypesContainment(t *testing.T) {
	// Test that all room type constants are in the RoomTypes slice
	expectedTypes := map[string]bool{
		RoomTypePublic:  false,
		RoomTypePrivate: false,
	}

	for _, roomType := range RoomTypes {
		if _, exists := expectedTypes[roomType]; exists {
			expectedTypes[roomType] = true
		} else {
			t.Errorf("Unexpected room type in RoomTypes slice: %s", roomType)
		}
	}

	for roomType, found := range expectedTypes {
		if !found {
			t.Errorf("Room type constant %s not found in RoomTypes slice", roomType)
		}
	}
}
