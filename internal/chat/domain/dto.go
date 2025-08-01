package domain

import "time"

// UserInfo represents basic user information for chat without sensitive data
type UserInfo struct {
	UUID      string    `json:"uuid"`      // Non-sensitive unique identifier
	Username  string    `json:"username"`  // Display name for chat
	JoinedAt  time.Time `json:"joined_at"`
}

// ClientInfo represents basic client information
type ClientInfo struct {
	ClientID string   `json:"client_id"`
	User     UserInfo `json:"user"`
}