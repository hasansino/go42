package adapter

import (
	"github.com/labstack/echo/v4"

	"github.com/hasansino/go42/internal/auth"
	authMiddleware "github.com/hasansino/go42/internal/auth/middleware"
	chatWebSocketAdapterV1 "github.com/hasansino/go42/internal/chat/adapters/websocket/v1"
)

// Adapter wraps the websocket adapter to work with the HTTP server registration system
type Adapter struct {
	websocketAdapter *chatWebSocketAdapterV1.Adapter
	authService      *auth.Service
	websocketPath    string
}

// NewHTTPAdapter creates a new HTTP adapter for websocket chat
func NewHTTPAdapter(
	websocketAdapter *chatWebSocketAdapterV1.Adapter,
	authService *auth.Service,
	websocketPath string,
) *Adapter {
	return &Adapter{
		websocketAdapter: websocketAdapter,
		authService:      authService,
		websocketPath:    websocketPath,
	}
}

func (a *Adapter) Register(group *echo.Group) {
	authMiddlewareFunc := authMiddleware.NewAuthMiddleware(a.authService)
	group.GET(a.websocketPath, a.websocketAdapter.HandleWebSocket, authMiddlewareFunc)
}