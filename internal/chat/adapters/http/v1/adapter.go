package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/hasansino/go42/internal/auth"
	authDomain "github.com/hasansino/go42/internal/auth/domain"
	authMiddleware "github.com/hasansino/go42/internal/auth/middleware"
	chatDomain "github.com/hasansino/go42/internal/chat/domain"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	CreateRoom(ctx context.Context, data *chatDomain.CreateRoomData, creator chatDomain.UserInfo) (*chatDomain.Room, error)
	GetRoom(ctx context.Context, roomID string) (*chatDomain.Room, error)
	ListRooms(ctx context.Context, roomType string) ([]*chatDomain.Room, error)
	JoinRoom(ctx context.Context, roomID string, client *chatDomain.Client) error
	LeaveRoom(ctx context.Context, clientID string) error
	SendMessage(ctx context.Context, clientID string, data *chatDomain.SendMessageData) error
	RegisterClient(ctx context.Context, client *chatDomain.Client)
	UnregisterClient(ctx context.Context, clientID string)
}

// Adapter handles HTTP and WebSocket connections for chat
type Adapter struct {
	service       serviceAccessor
	authService   *auth.Service
	websocketPath string
	upgrader      websocket.Upgrader
	logger        *slog.Logger
	options       adapterOptions
}

type adapterOptions struct {
	logger         *slog.Logger
	readTimeout    time.Duration
	writeTimeout   time.Duration
	pingPeriod     time.Duration
	pongWait       time.Duration
	allowedOrigins []string
}

// New creates a new HTTP adapter for chat
func New(
	service serviceAccessor,
	authService *auth.Service,
	websocketPath string,
	opts ...Option,
) *Adapter {
	options := adapterOptions{
		logger:         slog.Default(),
		readTimeout:    60 * time.Second,
		writeTimeout:   10 * time.Second,
		pingPeriod:     54 * time.Second,
		pongWait:       60 * time.Second,
		allowedOrigins: []string{"*"},
	}

	for _, opt := range opts {
		opt(&options)
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if len(options.allowedOrigins) == 1 && options.allowedOrigins[0] == "*" {
				return true
			}
			for _, allowedOrigin := range options.allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		},
	}

	return &Adapter{
		service:       service,
		authService:   authService,
		websocketPath: websocketPath,
		upgrader:      upgrader,
		logger:        options.logger,
		options:       options,
	}
}

// WebSocketMessage represents a websocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (a *Adapter) Register(group *echo.Group) {
	chatAuthMiddleware := authMiddleware.NewJWTOnlyMiddleware(a.authService)
	group.GET(a.websocketPath, a.HandleWebSocket, chatAuthMiddleware)
}

// HandleWebSocket handles websocket connections
func (a *Adapter) HandleWebSocket(c echo.Context) error {
	// Get authenticated user from context  
	authInfo := auth.RetrieveAuthFromContext(c.Request().Context())
	if authInfo == nil {
		a.logger.ErrorContext(c.Request().Context(), "authentication required - no auth info in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	// Only allow JWT-based authentication (authInfo.Type should be credentials)
	if authInfo.Type != authDomain.AuthenticationTypeCredentials {
		a.logger.ErrorContext(c.Request().Context(), "only JWT authentication is allowed for chat", 
			slog.String("auth_type", string(authInfo.Type)))
		return echo.NewHTTPError(http.StatusUnauthorized, "only JWT authentication is allowed for chat")
	}

	a.logger.DebugContext(c.Request().Context(), "WebSocket authentication successful", 
		slog.String("user_uuid", authInfo.UUID))

	// Log request headers for debugging
	a.logger.DebugContext(c.Request().Context(), "WebSocket upgrade request headers",
		slog.String("connection", c.Request().Header.Get("Connection")),
		slog.String("upgrade", c.Request().Header.Get("Upgrade")),
		slog.String("sec_websocket_version", c.Request().Header.Get("Sec-Websocket-Version")),
		slog.String("sec_websocket_key", c.Request().Header.Get("Sec-Websocket-Key")),
		slog.String("user_agent", c.Request().Header.Get("User-Agent")))

	// Upgrade connection to websocket
	conn, err := a.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		a.logger.ErrorContext(c.Request().Context(), "failed to upgrade websocket", 
			slog.Any("error", err),
			slog.String("error_type", fmt.Sprintf("%T", err)))
		return err
	}

	// Create UserInfo for chat (without exposing sensitive data)
	userInfo := chatDomain.UserInfo{
		UUID:     authInfo.UUID,
		Username: authInfo.UUID, // Use UUID as username for now since we don't have access to display name
		JoinedAt: time.Now(),
	}

	client := &chatDomain.Client{
		ID:       uuid.New().String(),
		User:     userInfo,
		Send:     make(chan []byte, 256),
		JoinedAt: time.Now(),
	}

	a.service.RegisterClient(c.Request().Context(), client)

	go a.writePump(c.Request().Context(), conn, client)
	go a.readPump(c.Request().Context(), conn, client)

	return nil
}

func (a *Adapter) readPump(ctx context.Context, conn *websocket.Conn, client *chatDomain.Client) {
	defer func() {
		a.service.UnregisterClient(ctx, client.ID)
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(a.options.pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(a.options.pongWait))
		return nil
	})

	for {
		var wsMsg WebSocketMessage
		err := conn.ReadJSON(&wsMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				a.logger.ErrorContext(ctx, "websocket error", slog.Any("error", err))
			}
			break
		}

		if err := a.handleMessage(ctx, client, &wsMsg); err != nil {
			a.logger.ErrorContext(ctx, "failed to handle message",
				slog.Any("error", err),
				slog.String("client_id", client.ID))
			
			errMsg := WebSocketMessage{
				Type: "error",
				Data: map[string]string{"message": err.Error()},
			}
			select {
			case client.Send <- a.encodeMessage(&errMsg):
			default:
				return
			}
		}
	}
}

func (a *Adapter) writePump(ctx context.Context, conn *websocket.Conn, client *chatDomain.Client) {
	ticker := time.NewTicker(a.options.pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(a.options.writeTimeout))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(a.options.writeTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming websocket messages
func (a *Adapter) handleMessage(ctx context.Context, client *chatDomain.Client, wsMsg *WebSocketMessage) error {
	switch wsMsg.Type {
	case "join_room":
		return a.handleJoinRoom(ctx, client, wsMsg.Data)
	case "leave_room":
		return a.handleLeaveRoom(ctx, client)
	case "send_message":
		return a.handleSendMessage(ctx, client, wsMsg.Data)
	case "create_room":
		return a.handleCreateRoom(ctx, client, wsMsg.Data)
	case "list_rooms":
		return a.handleListRooms(ctx, client, wsMsg.Data)
	default:
		return chatDomain.ErrInvalidMessage
	}
}

func (a *Adapter) handleJoinRoom(ctx context.Context, client *chatDomain.Client, data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return chatDomain.ErrInvalidMessage
	}

	roomID, ok := dataMap["room_id"].(string)
	if !ok {
		return chatDomain.ErrInvalidMessage
	}

	return a.service.JoinRoom(ctx, roomID, client)
}

func (a *Adapter) handleLeaveRoom(ctx context.Context, client *chatDomain.Client) error {
	return a.service.LeaveRoom(ctx, client.ID)
}

func (a *Adapter) handleSendMessage(ctx context.Context, client *chatDomain.Client, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return chatDomain.ErrInvalidMessage
	}

	var msgData chatDomain.SendMessageData
	if err := json.Unmarshal(dataBytes, &msgData); err != nil {
		return chatDomain.ErrInvalidMessage
	}

	return a.service.SendMessage(ctx, client.ID, &msgData)
}

func (a *Adapter) handleCreateRoom(ctx context.Context, client *chatDomain.Client, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return chatDomain.ErrInvalidMessage
	}

	var roomData chatDomain.CreateRoomData
	if err := json.Unmarshal(dataBytes, &roomData); err != nil {
		return chatDomain.ErrInvalidMessage
	}

	room, err := a.service.CreateRoom(ctx, &roomData, client.User)
	if err != nil {
		return err
	}

	response := WebSocketMessage{
		Type: "room_created",
		Data: room,
	}

	select {
	case client.Send <- a.encodeMessage(&response):
	default:
		return chatDomain.ErrInvalidMessage
	}

	return nil
}

func (a *Adapter) handleListRooms(ctx context.Context, client *chatDomain.Client, data interface{}) error {
	var roomType string
	if data != nil {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if rt, exists := dataMap["type"].(string); exists {
				roomType = rt
			}
		}
	}

	rooms, err := a.service.ListRooms(ctx, roomType)
	if err != nil {
		return err
	}

	response := WebSocketMessage{
		Type: "rooms_list",
		Data: rooms,
	}

	select {
	case client.Send <- a.encodeMessage(&response):
	default:
		return chatDomain.ErrInvalidMessage
	}

	return nil
}

func (a *Adapter) encodeMessage(msg *WebSocketMessage) []byte {
	data, _ := json.Marshal(msg)
	return data
}