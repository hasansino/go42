package adapter

import (
	"github.com/labstack/echo/v4"

	"github.com/hasansino/go42/internal/auth"
	authMiddleware "github.com/hasansino/go42/internal/auth/middleware"
	chatWebSocketAdapterV1 "github.com/hasansino/go42/internal/chat/adapters/websocket/v1"
)

// HTTPAdapter wraps the websocket adapter to work with the HTTP server registration system
type HTTPAdapter struct {
	websocketAdapter *chatWebSocketAdapterV1.Adapter
	authService      *auth.Service
	websocketPath    string
}

// NewHTTPAdapter creates a new HTTP adapter for websocket chat
func NewHTTPAdapter(
	websocketAdapter *chatWebSocketAdapterV1.Adapter,
	authService *auth.Service,
	websocketPath string,
) *HTTPAdapter {
	return &HTTPAdapter{
		websocketAdapter: websocketAdapter,
		authService:      authService,
		websocketPath:    websocketPath,
	}
}

// Register implements the adapterAccessor interface for HTTP server registration
func (a *HTTPAdapter) Register(group *echo.Group) {
	// Add websocket endpoint with auth middleware
	authMiddlewareFunc := authMiddleware.NewAuthMiddleware(a.authService)
	group.GET(a.websocketPath, a.websocketAdapter.HandleWebSocket, authMiddlewareFunc)
}