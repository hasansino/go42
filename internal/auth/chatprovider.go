package auth

import (
	"context"
	"time"

	chatDomain "github.com/hasansino/go42/internal/chat/domain"
)

// ChatAuthProvider implements chat.AuthProvider interface
// This decouples chat from direct auth service dependencies
type ChatAuthProvider struct {
	service *Service
}

// NewChatAuthProvider creates a new ChatAuthProvider
func NewChatAuthProvider(service *Service) *ChatAuthProvider {
	return &ChatAuthProvider{
		service: service,
	}
}

// ValidateToken validates a JWT token and returns user information for chat
func (p *ChatAuthProvider) ValidateToken(ctx context.Context, token string) (chatDomain.UserInfo, error) {
	// Validate the JWT token using the auth service
	claims, err := p.service.ValidateJWTToken(ctx, token)
	if err != nil {
		return chatDomain.UserInfo{}, err
	}

	// Get user details using the subject (user UUID) from claims
	user, err := p.service.GetUserByUUID(ctx, claims.Subject)
	if err != nil {
		return chatDomain.UserInfo{}, err
	}

	// Convert UUID to string for chat domain
	userUUIDStr := user.UUID.String()

	// Convert to chat domain UserInfo (hiding sensitive data like email, ID)
	return chatDomain.UserInfo{
		UUID:     userUUIDStr,
		Username: user.Email, // Using email as username for now
		JoinedAt: time.Now(),
	}, nil
}