// nolint
package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hasansino/go42/tests/integration"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chat WebSocket Integration Tests", func() {
	var client *http.Client
	var testEmail1, testEmail2 string
	var testPassword string
	var accessToken1, accessToken2 string

	BeforeEach(func() {
		client = &http.Client{Timeout: 10 * time.Second}
		testEmail1 = fmt.Sprintf("test-user1-%s@example.com", integration.GenerateRandomString("user"))
		testEmail2 = fmt.Sprintf("test-user2-%s@example.com", integration.GenerateRandomString("user"))
		testPassword = "TestPassword123!"
	})

	Describe("WebSocket Chat Authentication", func() {
		BeforeEach(func() {
			// Create test users
			createUser := func(email string) string {
				signupReq := SignupRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err := json.Marshal(signupReq)
				Expect(err).ToNot(HaveOccurred())

				resp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/signup",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				// Login to get access token
				loginReq := LoginRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err = json.Marshal(loginReq)
				Expect(err).ToNot(HaveOccurred())

				loginResp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/login",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer loginResp.Body.Close()
				Expect(loginResp.StatusCode).To(Equal(http.StatusOK))

				var tokens Tokens
				err = json.NewDecoder(loginResp.Body).Decode(&tokens)
				Expect(err).ToNot(HaveOccurred())

				return tokens.AccessToken
			}

			accessToken1 = createUser(testEmail1)
			accessToken2 = createUser(testEmail2)
		})

		It("should connect with valid JWT token", func() {
			wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
			u, err := url.Parse(wsURL)
			Expect(err).ToNot(HaveOccurred())
			
			q := u.Query()
			q.Set("token", accessToken1)
			u.RawQuery = q.Encode()

			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			Expect(err).ToNot(HaveOccurred())
			defer conn.Close()

			// Connection should be successful
			Expect(conn).ToNot(BeNil())
		})

		It("should reject connection without JWT token", func() {
			wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
			
			conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).To(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			if conn != nil {
				conn.Close()
			}
		})

		It("should reject connection with invalid JWT token", func() {
			wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
			u, err := url.Parse(wsURL)
			Expect(err).ToNot(HaveOccurred())
			
			q := u.Query()
			q.Set("token", "invalid-token")
			u.RawQuery = q.Encode()

			conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
			Expect(err).To(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			if conn != nil {
				conn.Close()
			}
		})
	})

	Describe("Chat Room Management", func() {
		var conn1, conn2 *websocket.Conn
		var roomID string

		BeforeEach(func() {
			// Create test users
			createUser := func(email string) string {
				signupReq := SignupRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err := json.Marshal(signupReq)
				Expect(err).ToNot(HaveOccurred())

				resp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/signup",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				loginReq := LoginRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err = json.Marshal(loginReq)
				Expect(err).ToNot(HaveOccurred())

				loginResp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/login",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer loginResp.Body.Close()

				var tokens Tokens
				err = json.NewDecoder(loginResp.Body).Decode(&tokens)
				Expect(err).ToNot(HaveOccurred())

				return tokens.AccessToken
			}

			accessToken1 = createUser(testEmail1)
			accessToken2 = createUser(testEmail2)

			// Connect both clients
			connectClient := func(token string) *websocket.Conn {
				wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
				u, err := url.Parse(wsURL)
				Expect(err).ToNot(HaveOccurred())
				
				q := u.Query()
				q.Set("token", token)
				u.RawQuery = q.Encode()

				conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				Expect(err).ToNot(HaveOccurred())
				return conn
			}

			conn1 = connectClient(accessToken1)
			conn2 = connectClient(accessToken2)
		})

		AfterEach(func() {
			if conn1 != nil {
				conn1.Close()
			}
			if conn2 != nil {
				conn2.Close()
			}
		})

		It("should list rooms initially empty", func() {
			// Send list rooms message
			listMsg := WebSocketMessage{
				Type: "list_rooms",
				Data: map[string]interface{}{},
			}
			err := conn1.WriteJSON(listMsg)
			Expect(err).ToNot(HaveOccurred())

			// Read response
			var response WebSocketMessage
			err = conn1.ReadJSON(&response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Type).To(Equal("rooms_list"))
			
			// Should be empty array or nil
			rooms, ok := response.Data.([]interface{})
			if ok {
				Expect(rooms).To(HaveLen(0))
			} else {
				Expect(response.Data).To(BeNil())
			}
		})

		It("should create a new room", func() {
			createRoomMsg := WebSocketMessage{
				Type: "create_room",
				Data: CreateRoomData{
					Name:     "Test Room",
					Type:     "public",
					MaxUsers: 10,
				},
			}
			err := conn1.WriteJSON(createRoomMsg)
			Expect(err).ToNot(HaveOccurred())

			// Read room created response
			var response WebSocketMessage
			err = conn1.ReadJSON(&response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Type).To(Equal("room_created"))
			
			// Parse room data
			roomData, err := json.Marshal(response.Data)
			Expect(err).ToNot(HaveOccurred())
			
			var room Room
			err = json.Unmarshal(roomData, &room)
			Expect(err).ToNot(HaveOccurred())
			
			Expect(room.Name).To(Equal("Test Room"))
			Expect(room.Type).To(Equal("public"))
			Expect(room.MaxUsers).To(Equal(10))
			Expect(room.ID).ToNot(BeEmpty())
			
			roomID = room.ID
		})

		It("should list created rooms", func() {
			// Create a room first
			createRoomMsg := WebSocketMessage{
				Type: "create_room",
				Data: CreateRoomData{
					Name:     "Test Room for Listing",
					Type:     "public",
					MaxUsers: 5,
				},
			}
			err := conn1.WriteJSON(createRoomMsg)
			Expect(err).ToNot(HaveOccurred())

			// Read room created response
			var createResponse WebSocketMessage
			err = conn1.ReadJSON(&createResponse)
			Expect(err).ToNot(HaveOccurred())

			// List rooms
			listMsg := WebSocketMessage{
				Type: "list_rooms",
				Data: map[string]interface{}{},
			}
			err = conn1.WriteJSON(listMsg)
			Expect(err).ToNot(HaveOccurred())

			// Read rooms list response
			var listResponse WebSocketMessage
			err = conn1.ReadJSON(&listResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Type).To(Equal("rooms_list"))
			
			// Parse rooms data
			roomsData, err := json.Marshal(listResponse.Data)
			Expect(err).ToNot(HaveOccurred())
			
			var rooms []Room
			err = json.Unmarshal(roomsData, &rooms)
			Expect(err).ToNot(HaveOccurred())
			
			Expect(rooms).To(HaveLen(1))
			Expect(rooms[0].Name).To(Equal("Test Room for Listing"))
		})

		It("should join and leave rooms", func() {
			// Create a room first
			createRoomMsg := WebSocketMessage{
				Type: "create_room",
				Data: CreateRoomData{
					Name:     "Join Test Room",
					Type:     "public",
					MaxUsers: 5,
				},
			}
			err := conn1.WriteJSON(createRoomMsg)
			Expect(err).ToNot(HaveOccurred())

			// Read room created response
			var createResponse WebSocketMessage
			err = conn1.ReadJSON(&createResponse)
			Expect(err).ToNot(HaveOccurred())
			
			// Extract room ID
			roomData, err := json.Marshal(createResponse.Data)
			Expect(err).ToNot(HaveOccurred())
			
			var room Room
			err = json.Unmarshal(roomData, &room)
			Expect(err).ToNot(HaveOccurred())
			roomID = room.ID

			// Join room with second client
			joinMsg := WebSocketMessage{
				Type: "join_room",
				Data: map[string]interface{}{
					"room_id": roomID,
				},
			}
			err = conn2.WriteJSON(joinMsg)
			Expect(err).ToNot(HaveOccurred())

			// Should receive join notification on conn1 (room creator)
			var joinNotification ChatMessage
			err = conn1.ReadJSON(&joinNotification)
			Expect(err).ToNot(HaveOccurred())
			Expect(joinNotification.Type).To(Equal("join"))
			Expect(joinNotification.RoomID).To(Equal(roomID))

			// Leave room
			leaveMsg := WebSocketMessage{
				Type: "leave_room",
				Data: map[string]interface{}{},
			}
			err = conn2.WriteJSON(leaveMsg)
			Expect(err).ToNot(HaveOccurred())

			// Should receive leave notification
			var leaveNotification ChatMessage
			err = conn1.ReadJSON(&leaveNotification)
			Expect(err).ToNot(HaveOccurred())
			Expect(leaveNotification.Type).To(Equal("leave"))
			Expect(leaveNotification.RoomID).To(Equal(roomID))
		})
	})

	Describe("Chat Messaging", func() {
		var conn1, conn2 *websocket.Conn
		var roomID string

		BeforeEach(func() {
			// Create test users and connect
			createUser := func(email string) string {
				signupReq := SignupRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err := json.Marshal(signupReq)
				Expect(err).ToNot(HaveOccurred())

				resp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/signup",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				loginReq := LoginRequest{
					Email:    email,
					Password: testPassword,
				}
				bodyBytes, err = json.Marshal(loginReq)
				Expect(err).ToNot(HaveOccurred())

				loginResp, err := client.Post(
					integration.HTTPServerAddress()+"/api/v1/auth/login",
					"application/json",
					strings.NewReader(string(bodyBytes)),
				)
				Expect(err).ToNot(HaveOccurred())
				defer loginResp.Body.Close()

				var tokens Tokens
				err = json.NewDecoder(loginResp.Body).Decode(&tokens)
				Expect(err).ToNot(HaveOccurred())

				return tokens.AccessToken
			}

			accessToken1 = createUser(testEmail1)
			accessToken2 = createUser(testEmail2)

			// Connect both clients
			connectClient := func(token string) *websocket.Conn {
				wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
				u, err := url.Parse(wsURL)
				Expect(err).ToNot(HaveOccurred())
				
				q := u.Query()
				q.Set("token", token)
				u.RawQuery = q.Encode()

				conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				Expect(err).ToNot(HaveOccurred())
				return conn
			}

			conn1 = connectClient(accessToken1)
			conn2 = connectClient(accessToken2)

			// Create a room and have both users join
			createRoomMsg := WebSocketMessage{
				Type: "create_room",
				Data: CreateRoomData{
					Name:     "Message Test Room",
					Type:     "public",
					MaxUsers: 10,
				},
			}
			err := conn1.WriteJSON(createRoomMsg)
			Expect(err).ToNot(HaveOccurred())

			var createResponse WebSocketMessage
			err = conn1.ReadJSON(&createResponse)
			Expect(err).ToNot(HaveOccurred())
			
			roomData, err := json.Marshal(createResponse.Data)
			Expect(err).ToNot(HaveOccurred())
			
			var room Room
			err = json.Unmarshal(roomData, &room)
			Expect(err).ToNot(HaveOccurred())
			roomID = room.ID

			// Join room with second client
			joinMsg := WebSocketMessage{
				Type: "join_room",
				Data: map[string]interface{}{
					"room_id": roomID,
				},
			}
			err = conn2.WriteJSON(joinMsg)
			Expect(err).ToNot(HaveOccurred())

			// Consume join notification
			var joinNotification ChatMessage
			err = conn1.ReadJSON(&joinNotification)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if conn1 != nil {
				conn1.Close()
			}
			if conn2 != nil {
				conn2.Close()
			}
		})

		It("should send and receive text messages", func() {
			// Send message from conn1
			sendMsg := WebSocketMessage{
				Type: "send_message",
				Data: SendMessageData{
					Type:    "text",
					Content: "Hello from user 1!",
					RoomID:  roomID,
				},
			}
			err := conn1.WriteJSON(sendMsg)
			Expect(err).ToNot(HaveOccurred())

			// Receive message on conn2
			var receivedMsg ChatMessage
			err = conn2.ReadJSON(&receivedMsg)
			Expect(err).ToNot(HaveOccurred())
			
			Expect(receivedMsg.Type).To(Equal("text"))
			Expect(receivedMsg.Content).To(Equal("Hello from user 1!"))
			Expect(receivedMsg.RoomID).To(Equal(roomID))
			Expect(receivedMsg.ID).ToNot(BeEmpty())
			Expect(receivedMsg.User.UUID).ToNot(BeEmpty())
			Expect(receivedMsg.Timestamp).ToNot(BeZero())

			// Receive same message on conn1 (sender should also receive)
			var senderMsg ChatMessage
			err = conn1.ReadJSON(&senderMsg)
			Expect(err).ToNot(HaveOccurred())
			Expect(senderMsg.Content).To(Equal("Hello from user 1!"))
		})

		It("should handle multiple messages in sequence", func() {
			messages := []string{
				"First message",
				"Second message",
				"Third message",
			}

			for i, content := range messages {
				sendMsg := WebSocketMessage{
					Type: "send_message",
					Data: SendMessageData{
						Type:    "text",
						Content: content,
						RoomID:  roomID,
					},
				}
				
				if i%2 == 0 {
					err := conn1.WriteJSON(sendMsg)
					Expect(err).ToNot(HaveOccurred())
				} else {
					err := conn2.WriteJSON(sendMsg)
					Expect(err).ToNot(HaveOccurred())
				}

				// Both clients should receive each message
				var msg1, msg2 ChatMessage
				err := conn1.ReadJSON(&msg1)
				Expect(err).ToNot(HaveOccurred())
				err = conn2.ReadJSON(&msg2)
				Expect(err).ToNot(HaveOccurred())

				Expect(msg1.Content).To(Equal(content))
				Expect(msg2.Content).To(Equal(content))
			}
		})
	})

	Describe("Error Handling", func() {
		var conn *websocket.Conn

		BeforeEach(func() {
			// Create test user and connect
			signupReq := SignupRequest{
				Email:    testEmail1,
				Password: testPassword,
			}
			bodyBytes, err := json.Marshal(signupReq)
			Expect(err).ToNot(HaveOccurred())

			resp, err := client.Post(
				integration.HTTPServerAddress()+"/api/v1/auth/signup",
				"application/json",
				strings.NewReader(string(bodyBytes)),
			)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			loginReq := LoginRequest{
				Email:    testEmail1,
				Password: testPassword,
			}
			bodyBytes, err = json.Marshal(loginReq)
			Expect(err).ToNot(HaveOccurred())

			loginResp, err := client.Post(
				integration.HTTPServerAddress()+"/api/v1/auth/login",
				"application/json",
				strings.NewReader(string(bodyBytes)),
			)
			Expect(err).ToNot(HaveOccurred())
			defer loginResp.Body.Close()

			var tokens Tokens
			err = json.NewDecoder(loginResp.Body).Decode(&tokens)
			Expect(err).ToNot(HaveOccurred())

			wsURL := strings.Replace(integration.HTTPServerAddress(), "http", "ws", 1) + "/ws/chat"
			u, err := url.Parse(wsURL)
			Expect(err).ToNot(HaveOccurred())
			
			q := u.Query()
			q.Set("token", tokens.AccessToken)
			u.RawQuery = q.Encode()

			conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if conn != nil {
				conn.Close()
			}
		})

		It("should return error for invalid message format", func() {
			// Send invalid JSON
			err := conn.WriteMessage(websocket.TextMessage, []byte("{invalid json"))
			Expect(err).ToNot(HaveOccurred())

			// Connection should close due to invalid JSON
			_, _, err = conn.ReadMessage()
			Expect(err).To(HaveOccurred())
		})

		It("should return error for unknown message type", func() {
			unknownMsg := WebSocketMessage{
				Type: "unknown_type",
				Data: map[string]interface{}{},
			}
			err := conn.WriteJSON(unknownMsg)
			Expect(err).ToNot(HaveOccurred())

			// Should receive error response
			var response WebSocketMessage
			err = conn.ReadJSON(&response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Type).To(Equal("error"))
		})

		It("should return error when joining non-existent room", func() {
			joinMsg := WebSocketMessage{
				Type: "join_room",
				Data: map[string]interface{}{
					"room_id": "non-existent-room-id",
				},
			}
			err := conn.WriteJSON(joinMsg)
			Expect(err).ToNot(HaveOccurred())

			var response WebSocketMessage
			err = conn.ReadJSON(&response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Type).To(Equal("error"))
		})

		It("should return error when sending message without joining room", func() {
			sendMsg := WebSocketMessage{
				Type: "send_message",
				Data: SendMessageData{
					Type:    "text",
					Content: "This should fail",
					RoomID:  "any-room-id",
				},
			}
			err := conn.WriteJSON(sendMsg)
			Expect(err).ToNot(HaveOccurred())

			var response WebSocketMessage
			err = conn.ReadJSON(&response)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Type).To(Equal("error"))
		})
	})
})