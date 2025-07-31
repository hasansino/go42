# Chat WebSocket API Documentation

## Overview

The chat system provides real-time messaging capabilities through WebSocket connections. Users can create rooms, join existing rooms, and exchange messages in real-time.

## Connection

### Endpoint
```
GET /ws/chat
```

### Authentication
All WebSocket connections require authentication via Bearer token in the `Authorization` header or API key in the `x-api-key` header.

### Example Connection
```javascript
const token = 'your-jwt-token';
const ws = new WebSocket('ws://localhost:8080/ws/chat', [], {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

## Message Format

All messages sent and received through the WebSocket follow this JSON format:

```json
{
  "type": "message_type",
  "data": {
    // Type-specific data
  }
}
```

## Client-to-Server Messages

### 1. Create Room

Create a new chat room.

```json
{
  "type": "create_room",
  "data": {
    "name": "My Room",
    "type": "public",
    "max_users": 50
  }
}
```

**Parameters:**
- `name` (string, required): Room name (1-100 characters)
- `type` (string, required): Room type (`public` or `private`)
- `max_users` (integer, optional): Maximum users allowed (2-1000, default: 100)

**Response:**
```json
{
  "type": "room_created",
  "data": {
    "id": "room-uuid",
    "name": "My Room",
    "type": "public",
    "max_users": 50,
    "user_count": 0,
    "users": {},
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 2. List Rooms

Get a list of available rooms.

```json
{
  "type": "list_rooms",
  "data": {
    "type": "public"
  }
}
```

**Parameters:**
- `type` (string, optional): Filter by room type (`public` or `private`)

**Response:**
```json
{
  "type": "rooms_list",
  "data": [
    {
      "id": "room-uuid-1",
      "name": "Room 1",
      "type": "public",
      "max_users": 100,
      "user_count": 5,
      "users": { ... },
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

### 3. Join Room

Join an existing room.

```json
{
  "type": "join_room",
  "data": {
    "room_id": "room-uuid"
  }
}
```

**Parameters:**
- `room_id` (string, required): UUID of the room to join

### 4. Leave Room

Leave the current room.

```json
{
  "type": "leave_room",
  "data": {}
}
```

### 5. Send Message

Send a message to the current room.

```json
{
  "type": "send_message",
  "data": {
    "type": "text",
    "content": "Hello, everyone!",
    "room_id": "room-uuid"
  }
}
```

**Parameters:**
- `type` (string, required): Message type (`text` or `system`)
- `content` (string, required): Message content (max 1000 characters)
- `room_id` (string, required): UUID of the target room

## Server-to-Client Messages

### 1. Chat Message

Received when someone sends a message to the room.

```json
{
  "id": "message-uuid",
  "type": "text",
  "content": "Hello, everyone!",
  "user_id": 123,
  "user_uuid": "user-uuid",
  "user_email": "user@example.com",
  "room_id": "room-uuid",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2. Join Notification

Received when a user joins the room.

```json
{
  "id": "message-uuid",
  "type": "join",
  "content": "user@example.com joined the room",
  "user_id": 123,
  "user_uuid": "user-uuid",
  "user_email": "user@example.com",
  "room_id": "room-uuid",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 3. Leave Notification

Received when a user leaves the room.

```json
{
  "id": "message-uuid",
  "type": "leave",
  "content": "user@example.com left the room",
  "user_id": 123,
  "user_uuid": "user-uuid",
  "user_email": "user@example.com",
  "room_id": "room-uuid",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 4. Error Message

Received when an error occurs.

```json
{
  "type": "error",
  "data": {
    "message": "Error description"
  }
}
```

## Configuration

The chat system can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAT_MAX_ROOMS_PER_USER` | `10` | Maximum rooms a user can create |
| `CHAT_MAX_MESSAGES_PER_MIN` | `30` | Rate limit for messages per minute |
| `CHAT_DEFAULT_MAX_USERS` | `100` | Default maximum users per room |
| `CHAT_WEBSOCKET_PATH` | `/ws/chat` | WebSocket endpoint path |
| `CHAT_READ_TIMEOUT` | `60s` | WebSocket read timeout |
| `CHAT_WRITE_TIMEOUT` | `10s` | WebSocket write timeout |
| `CHAT_PING_PERIOD` | `54s` | Ping interval for keepalive |
| `CHAT_PONG_WAIT` | `60s` | Pong timeout |

## Error Codes

| Error | Description |
|-------|-------------|
| `room not found` | The specified room does not exist |
| `user not found` | The user/client is not registered |
| `invalid message` | The message format is invalid |
| `unauthorized` | User is not authorized for this action |
| `room is full` | The room has reached maximum capacity |
| `user already in room` | User is already a member of the room |

## Example Implementation

### JavaScript Client

```javascript
class ChatClient {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.currentRoom = null;
  }

  connect() {
    this.ws = new WebSocket('ws://localhost:8080/ws/chat', [], {
      headers: {
        'Authorization': `Bearer ${this.token}`
      }
    });

    this.ws.onopen = () => {
      console.log('Connected to chat');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleMessage(message);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from chat');
    };
  }

  handleMessage(message) {
    switch (message.type) {
      case 'room_created':
        console.log('Room created:', message.data);
        break;
      case 'rooms_list':
        console.log('Available rooms:', message.data);
        break;
      case 'error':
        console.error('Chat error:', message.data.message);
        break;
      default:
        // Handle chat messages
        if (message.id) {
          console.log('New message:', message);
        }
    }
  }

  createRoom(name, type = 'public', maxUsers = 100) {
    this.send('create_room', {
      name,
      type,
      max_users: maxUsers
    });
  }

  listRooms(type = null) {
    this.send('list_rooms', { type });
  }

  joinRoom(roomId) {
    this.send('join_room', { room_id: roomId });
    this.currentRoom = roomId;
  }

  leaveRoom() {
    this.send('leave_room', {});
    this.currentRoom = null;
  }

  sendMessage(content) {
    if (!this.currentRoom) {
      console.error('Not in a room');
      return;
    }

    this.send('send_message', {
      type: 'text',
      content,
      room_id: this.currentRoom
    });
  }

  send(type, data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }
}

// Usage
const client = new ChatClient('your-jwt-token');
client.connect();

// Create a room
client.createRoom('My Chat Room');

// List rooms
client.listRooms();

// Join a room
client.joinRoom('room-uuid');

// Send a message
client.sendMessage('Hello, everyone!');
```

## Rate Limiting

The chat system implements rate limiting to prevent abuse:

- Maximum 30 messages per minute per user (configurable)
- Maximum 10 rooms per user (configurable)
- Connection timeouts to prevent resource exhaustion

## Security Considerations

1. **Authentication Required**: All connections must be authenticated
2. **Rate Limiting**: Prevents message spam and abuse
3. **Input Validation**: All messages are validated for type and content
4. **Room Access Control**: Users can only send messages to rooms they've joined
5. **Connection Limits**: Prevents resource exhaustion

## Monitoring and Metrics

The chat service exposes metrics for monitoring:

- Active connections count
- Messages per second
- Room count
- Error rates
- Connection duration

## Limitations

1. Messages are not persisted to database (in-memory only)
2. No message history retrieval
3. No private direct messaging (only room-based chat)
4. No file upload support
5. No message editing or deletion