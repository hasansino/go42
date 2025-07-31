# Chat Package

A WebSocket-based real-time chat system for the go42 application.

## Features

- **Real-time messaging** via WebSocket connections
- **Room-based chat** with public and private rooms  
- **User authentication** required for all connections
- **Rate limiting** to prevent abuse
- **In-memory storage** for fast performance
- **Configurable limits** for rooms and messages
- **JSON message format** for easy integration

## Architecture

The chat package follows the same clean architecture pattern as other packages in go42:

```
internal/chat/
├── chat.go              # Service layer - business logic
├── chat_test.go         # Unit tests for service
├── options.go           # Service configuration options
├── domain/
│   └── domain.go        # Domain models and constants
├── models/
│   └── models.go        # Data models (for future database persistence)
├── adapters/
│   ├── http/v1/
│   │   └── adapter.go   # HTTP adapter for WebSocket registration
│   └── websocket/v1/
│       ├── adapter.go   # WebSocket handler implementation
│       └── options.go   # WebSocket adapter options
```

## Components

### Service Layer (`chat.go`)
- **Room Management**: Create, list, and manage chat rooms
- **User Management**: Handle user joins/leaves, track active users
- **Message Handling**: Process and broadcast messages
- **Client Management**: Track WebSocket connections

### Domain Layer (`domain/domain.go`)
- **Message Types**: Text, join, leave, system messages
- **Room Types**: Public and private rooms
- **Error Definitions**: Chat-specific errors
- **Data Structures**: Message, Room, Client, UserInfo

### WebSocket Adapter (`adapters/websocket/v1/`)
- **Connection Handling**: Upgrade HTTP to WebSocket
- **Message Processing**: Parse and route WebSocket messages
- **Authentication**: Integrate with auth middleware
- **Connection Management**: Handle ping/pong, timeouts

### HTTP Adapter (`adapters/http/v1/`)
- **Registration**: Register WebSocket endpoint with HTTP server
- **Middleware Integration**: Add authentication middleware
- **Route Configuration**: Configure WebSocket path

## Configuration

The chat system is configured through environment variables in the config package:

```go
type Chat struct {
    Enabled             bool          `env:"CHAT_ENABLED"                 default:"false"`
    MaxRoomsPerUser     int           `env:"CHAT_MAX_ROOMS_PER_USER"      default:"10"`
    MaxMessagesPerMin   int           `env:"CHAT_MAX_MESSAGES_PER_MIN"    default:"30"`
    DefaultMaxUsers     int           `env:"CHAT_DEFAULT_MAX_USERS"       default:"100"`
    WebSocketPath       string        `env:"CHAT_WEBSOCKET_PATH"          default:"/ws/chat"`
    ReadTimeout         time.Duration `env:"CHAT_READ_TIMEOUT"            default:"60s"`
    WriteTimeout        time.Duration `env:"CHAT_WRITE_TIMEOUT"           default:"10s"`
    PingPeriod          time.Duration `env:"CHAT_PING_PERIOD"             default:"54s"`
    PongWait            time.Duration `env:"CHAT_PONG_WAIT"               default:"60s"`
}
```

## Usage

### Service Initialization

```go
// In main.go
if cfg.Chat.Enabled {
    chatLogger := slog.Default().With(slog.String("component", "chat-service"))
    chatService = chat.NewService(
        chat.WithLogger(chatLogger),
        chat.WithMaxRoomsPerUser(cfg.Chat.MaxRoomsPerUser),
        chat.WithMaxMessagesPerMin(cfg.Chat.MaxMessagesPerMin),
        chat.WithDefaultMaxUsers(cfg.Chat.DefaultMaxUsers),
    )
}
```

### WebSocket Registration

```go
// WebSocket adapter
chatWebSocketAdapter := chatWebSocketAdapterV1.New(
    chatService,
    chatWebSocketAdapterV1.WithLogger(logger),
    chatWebSocketAdapterV1.WithReadTimeout(cfg.Chat.ReadTimeout),
    chatWebSocketAdapterV1.WithWriteTimeout(cfg.Chat.WriteTimeout),
    chatWebSocketAdapterV1.WithPingPeriod(cfg.Chat.PingPeriod),
    chatWebSocketAdapterV1.WithPongWait(cfg.Chat.PongWait),
)

// HTTP adapter for registration
chatHTTPAdapter := chatHTTPAdapterV1.NewHTTPAdapter(
    chatWebSocketAdapter,
    authService,
    cfg.Chat.WebSocketPath,
)

// Register with HTTP server
httpServer.Register(chatHTTPAdapter)
```

### Basic Operations

```go
// Create a room
room, err := service.CreateRoom(ctx, &domain.CreateRoomData{
    Name:     "General Chat",
    Type:     domain.RoomTypePublic,
    MaxUsers: 100,
}, userID, userUUID, userEmail)

// Join a room
client := &domain.Client{
    ID:        uuid.New().String(),
    UserID:    userID,
    UserUUID:  userUUID,
    UserEmail: userEmail,
    Send:      make(chan []byte, 256),
}
err = service.JoinRoom(ctx, roomID, client)

// Send a message
err = service.SendMessage(ctx, clientID, &domain.SendMessageData{
    Type:    domain.MessageTypeText,
    Content: "Hello, world!",
    RoomID:  roomID,
})
```

## Message Flow

1. **Client Connection**: Client connects to WebSocket endpoint with authentication
2. **Client Registration**: Service registers the client and assigns a unique ID
3. **Room Operations**: Client can create rooms, list rooms, join/leave rooms
4. **Message Exchange**: Clients in the same room can exchange real-time messages
5. **Broadcasting**: Messages are broadcast to all clients in the room
6. **Cleanup**: Client disconnection triggers automatic cleanup and leave notifications

## Error Handling

The package defines specific errors for different scenarios:

```go
var (
    ErrRoomNotFound      = errors.New("room not found")
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidMessage    = errors.New("invalid message")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrRoomFull          = errors.New("room is full")
    ErrUserAlreadyInRoom = errors.New("user already in room")
)
```

## Security Features

- **Authentication Required**: All WebSocket connections require valid JWT or API token
- **Rate Limiting**: Configurable limits on messages per minute
- **Input Validation**: All messages are validated for type and content
- **Room Access Control**: Users can only message rooms they've joined
- **Connection Timeouts**: Prevent resource exhaustion

## Performance Considerations

- **In-Memory Storage**: Fast access but not persistent across restarts
- **Concurrent Safety**: All operations are thread-safe using mutexes
- **Channel-Based Broadcasting**: Efficient message distribution
- **Connection Limits**: Configurable to prevent resource exhaustion
- **Graceful Cleanup**: Automatic cleanup of disconnected clients

## Testing

The package includes comprehensive unit tests:

```bash
# Run chat tests
go test -v ./internal/chat/

# Run all tests
make test-unit
```

Tests cover:
- Room creation and management
- User join/leave operations
- Message sending and broadcasting
- Error conditions and edge cases

## Future Enhancements

Potential improvements for the chat system:

1. **Database Persistence**: Store messages and rooms in database
2. **Message History**: Retrieve chat history for rooms
3. **Private Messaging**: Direct user-to-user messaging
4. **File Uploads**: Support for file sharing
5. **Message Reactions**: Emoji reactions to messages
6. **User Presence**: Online/offline status tracking
7. **Typing Indicators**: Show when users are typing
8. **Message Search**: Search through chat history
9. **Push Notifications**: Notify users of mentions/DMs
10. **Admin Features**: Moderation tools and user management

## Dependencies

The chat package depends on:

- **gorilla/websocket**: WebSocket implementation
- **echo/v4**: HTTP framework for routing
- **google/uuid**: UUID generation
- **slog**: Structured logging
- **auth package**: User authentication
- **config package**: Configuration management

## API Documentation

For detailed WebSocket API documentation, see [doc/chat/websocket-api.md](../../doc/chat/websocket-api.md).