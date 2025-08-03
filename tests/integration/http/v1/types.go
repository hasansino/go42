// nolint
package test

import "time"

// Auth related types
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UpdateSelfRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type User struct {
	UUID        string   `json:"uuid"`
	Email       string   `json:"email"`
	CreatedAt   string   `json:"created_at"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Chat related types
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type ChatMessage struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	User      UserInfo  `json:"user"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
}

type UserInfo struct {
	UUID     string    `json:"uuid"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type Room struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      string              `json:"type"`
	MaxUsers  int                 `json:"max_users"`
	UserCount int                 `json:"user_count"`
	Users     map[string]UserInfo `json:"users"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// WebSocket message data types
type CreateRoomData struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	MaxUsers int    `json:"max_users"`
}

type SendMessageData struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	RoomID  string `json:"room_id"`
}
