package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	chatDomain "github.com/hasansino/go42/internal/chat/domain"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type serviceAccessor interface {
	CreateRoom(
		ctx context.Context,
		data *chatDomain.CreateRoomData,
		creator chatDomain.UserInfo,
	) (*chatDomain.Room, error)
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
	authService   chatDomain.AuthService
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
	authService chatDomain.AuthService,
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
	group.GET(a.websocketPath, a.HandleWebSocket)
}

// HandleWebSocket handles websocket connections
func (a *Adapter) HandleWebSocket(c echo.Context) error {
	// Extract JWT token from query parameter or Authorization header
	var token string

	// Try query parameter first (for WebSocket connections)
	token = c.QueryParam("token")

	// If not in query, try Authorization header
	if token == "" {
		authHeader := c.Request().Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		a.logger.ErrorContext(c.Request().Context(), "authentication required - no token provided")
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	// Validate token and get user UUID using AuthService
	userUUID, err := a.authService.ValidateJWTToken(c.Request().Context(), token)
	if err != nil {
		a.logger.ErrorContext(c.Request().Context(), "token validation failed", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	// Get basic user information
	userUUIDStr, email, err := a.authService.GetBasicUserInfo(c.Request().Context(), userUUID)
	if err != nil {
		a.logger.ErrorContext(c.Request().Context(), "failed to get user info", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	}

	// Create chat user info
	userInfo := chatDomain.UserInfo{
		UUID:     userUUIDStr,
		Username: email, // Using email as username for now
		JoinedAt: time.Now(),
	}

	a.logger.DebugContext(c.Request().Context(), "WebSocket authentication successful",
		slog.String("user_uuid", userInfo.UUID))

	// Log request headers for debugging
	a.logger.DebugContext(c.Request().Context(), "WebSocket upgrade request headers",
		slog.String("connection", c.Request().Header.Get("Connection")),
		slog.String("upgrade", c.Request().Header.Get("Upgrade")),
		slog.String("sec_websocket_version", c.Request().Header.Get("Sec-Websocket-Version")),
		slog.String("sec_websocket_key", c.Request().Header.Get("Sec-Websocket-Key")),
		slog.String("user_agent", c.Request().Header.Get("User-Agent")))

	// Upgrade connection to websocket
	// Echo and middleware may wrap the response writer, so we need to get the original one
	// that implements http.Hijacker interface
	resp := c.Response()
	writer := resp.Writer

	// Log debug info about the response writer type
	a.logger.DebugContext(c.Request().Context(), "Response writer info",
		slog.String("writer_type", fmt.Sprintf("%T", writer)),
		slog.Bool("implements_hijacker", func() bool {
			_, ok := writer.(http.Hijacker)
			return ok
		}()))

	// Try to unwrap any middleware wrappers to get the underlying http.ResponseWriter
	// that implements http.Hijacker. Use reflection to access the embedded field.

	// For the responseRecorder from our middleware, try to access the embedded ResponseWriter
	// using reflection since the struct is not exported
	if underlyingWriter := getUnderlyingHijacker(writer); underlyingWriter != nil {
		a.logger.DebugContext(c.Request().Context(), "Successfully found hijacker-capable writer",
			slog.String("unwrapped_writer_type", fmt.Sprintf("%T", underlyingWriter)))
		writer = underlyingWriter
	}
	conn, err := a.upgrader.Upgrade(writer, c.Request(), nil)
	if err != nil {
		a.logger.ErrorContext(c.Request().Context(), "failed to upgrade websocket",
			slog.Any("error", err),
			slog.String("error_type", fmt.Sprintf("%T", err)),
			slog.String("error_message", err.Error()))
		return err
	}

	client := &chatDomain.Client{
		ID:       uuid.New().String(),
		User:     userInfo, // Use the userInfo from AuthProvider
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
		_ = conn.Close()
	}()

	_ = conn.SetReadDeadline(time.Now().Add(a.options.pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(a.options.pongWait))
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
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(a.options.writeTimeout))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(a.options.writeTimeout))
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

// getUnderlyingHijacker tries to extract an http.ResponseWriter that implements http.Hijacker
// from wrapped response writers (like middleware recorders)
func getUnderlyingHijacker(w http.ResponseWriter) http.ResponseWriter {
	// Check if the current writer implements Hijacker
	if _, ok := w.(http.Hijacker); ok {
		return w
	}

	// Try to unwrap common wrapper patterns
	// Pattern 1: responseRecorder with embedded ResponseWriter field
	if wrapped := unwrapResponseRecorder(w); wrapped != nil {
		if _, ok := wrapped.(http.Hijacker); ok {
			return wrapped
		}
	}

	return nil
}

// unwrapResponseRecorder tries to extract the embedded ResponseWriter from a responseRecorder
func unwrapResponseRecorder(w http.ResponseWriter) http.ResponseWriter {
	// Check if this is our responseRecorder type and it has the GetUnderlyingWriter method
	if recorder, ok := w.(interface{ GetUnderlyingWriter() http.ResponseWriter }); ok {
		return recorder.GetUnderlyingWriter()
	}

	return nil
}
