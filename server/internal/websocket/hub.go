package websocket

import (
	"e5realtimechat/internal/cache"
	"encoding/json"
	"log"

	ws "github.com/gorilla/websocket"
)

// Message structure for routing
type WSMessage struct {
	Type       string `json:"type"`         // "message", "join", "leave", "user_status", "heartbeat"
	From       string `json:"from"`         // username of sender
	FromUserID int    `json:"from_user_id"` // user ID of sender
	ToUserID   int    `json:"to_user_id"`   // user ID of recipient (0 = broadcast)
	Text       string `json:"text"`
	User       string `json:"user"`
	UserID     int    `json:"user_id,omitempty"`   // for status updates
	IsOnline   bool   `json:"is_online,omitempty"` // for status updates
	Username   string `json:"username,omitempty"`  // for status updates
}

// // Hub quản lý tất cả client đang kết nối và phân phối tin nhắn giữa họ
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	directMsg    chan *DirectMessage // channel for direct messages
	register     chan *Client
	unregister   chan *Client
	cacheService *cache.CacheService // Redis cache for online status
	rateLimiter  interface {         // Rate limiter for message throttling
		CheckUserMessageRate(userID int) (bool, error)
	}
	redisClient *cache.RedisClient // Redis client for Pub/Sub cross-instance messaging
}

// DirectMessage contains message and target user ID
type DirectMessage struct {
	message  []byte
	toUserID int
}

// NewHub khởi tạo 1 hub mới
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte),
		directMsg:    make(chan *DirectMessage),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		cacheService: nil,
		rateLimiter:  nil,
		redisClient:  nil,
	}
}

// SetCacheService sets the cache service for the hub
func (h *Hub) SetCacheService(cacheService *cache.CacheService) {
	h.cacheService = cacheService
}

// SetRateLimiter sets the rate limiter for the hub
func (h *Hub) SetRateLimiter(rateLimiter interface {
	CheckUserMessageRate(userID int) (bool, error)
}) {
	h.rateLimiter = rateLimiter
}

// SetRedisClient sets the Redis client for Pub/Sub messaging
func (h *Hub) SetRedisClient(redisClient *cache.RedisClient) {
	h.redisClient = redisClient
	// Start Redis subscriber in background if Redis is available
	if h.redisClient != nil {
		go h.subscribeToRedis()
		log.Println("✅ Redis Pub/Sub enabled for cross-instance messaging")
	}
}

// // Run chạy liên tục, xử lý các sự kiện từ các channel
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			//thêm client mới vào danh sách
			h.clients[client] = true

			// Mark user as online in cache
			if h.cacheService != nil && client.userID > 0 {
				if err := h.cacheService.SetUserOnline(client.userID); err != nil {
					log.Printf("⚠️ Failed to set user %d online: %v", client.userID, err)
				} else {
					log.Printf("✅ User %d (%s) is now ONLINE", client.userID, client.username)

					// Broadcast user online status to all clients
					h.broadcastUserStatus(client.userID, client.username, true)
				}
			}

		case client := <-h.unregister:
			//xóa client khi ngắt kết nối
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Mark user as offline in cache
				if h.cacheService != nil && client.userID > 0 {
					if err := h.cacheService.SetUserOffline(client.userID); err != nil {
						log.Printf("⚠️ Failed to set user %d offline: %v", client.userID, err)
					} else {
						log.Printf("👋 User %d (%s) is now OFFLINE", client.userID, client.username)

						// Broadcast user offline status to all clients
						h.broadcastUserStatus(client.userID, client.username, false)
					}
				}
			}

		case message := <-h.broadcast:
			//gửi tin nhắn tới tất cả clients
			log.Printf("📢 Hub broadcasting message to %d clients: %s", len(h.clients), string(message))
			sentCount := 0
			for client := range h.clients {
				select {
				case client.send <- message:
					sentCount++
					log.Printf("✅ Sent broadcast to client %d (%s)", client.userID, client.username)
				default:
					// Nếu không gửi được → đóng kết nối client
					log.Printf("❌ Failed to send to client %d (%s), closing connection", client.userID, client.username)
					close(client.send)
					delete(h.clients, client)
				}
			}
			log.Printf("📢 Broadcast complete: sent to %d/%d clients", sentCount, len(h.clients))

		case directMsg := <-h.directMsg:
			// Gửi tin nhắn riêng tư cho user cụ thể
			log.Printf("🎯 Hub received directMsg for user %d: %s", directMsg.toUserID, string(directMsg.message))
			log.Printf("🔍 Searching for recipient in %d connected clients...", len(h.clients))
			found := false
			for client := range h.clients {
				if client.userID == directMsg.toUserID {
					found = true
					log.Printf("✅ Found recipient: client %d (%s)", client.userID, client.username)
					select {
					case client.send <- directMsg.message:
						log.Printf("✅ Message sent to client %d (%s) successfully", client.userID, client.username)
					default:
						log.Printf("❌ Failed to send to client %d (%s), channel blocked. Closing connection.", client.userID, client.username)
						close(client.send)
						delete(h.clients, client)
					}
					break
				}
			}
			if !found {
				log.Printf("⚠️ Recipient user %d not found in connected clients", directMsg.toUserID)
			}
		}
	}
}

// SendDirectMessage sends a message to a specific user
func (h *Hub) SendDirectMessage(message []byte, toUserID int) {
	log.Printf("🎯 Hub.SendDirectMessage called: toUserID=%d, message=%s", toUserID, string(message))
	h.directMsg <- &DirectMessage{
		message:  message,
		toUserID: toUserID,
	}
	log.Printf("✅ Message queued in directMsg channel for user %d", toUserID)
}

// SendBroadcast sends a message to all connected clients
func (h *Hub) SendBroadcast(message []byte) {
	h.broadcast <- message
}

// Register registers a new client
func (h *Hub) Register(conn interface{}, userID int, username string) *Client {
	client := &Client{
		hub:      h,
		conn:     conn.(*ws.Conn),
		send:     make(chan []byte, 256),
		userID:   userID,
		username: username,
	}
	h.register <- client
	return client
}

// StartClient starts the read and write pumps for a client
func (c *Client) StartClient() {
	go c.writePump()
	c.readPump()
}

// ==================== REDIS PUB/SUB FOR HORIZONTAL SCALING ====================

const redisBroadcastChannel = "chat:broadcast"
const redisDirectMsgChannel = "chat:direct"

// BroadcastViaRedis publishes a broadcast message to Redis
// This allows messages to be received by clients connected to other server instances
func (h *Hub) BroadcastViaRedis(message []byte) error {
	log.Printf("📡 BroadcastViaRedis called with message: %s", string(message))
	if h.redisClient == nil {
		log.Printf("⚠️ Redis client is nil, using local broadcast only")
		// Fallback to local broadcast only - use select with timeout
		select {
		case h.broadcast <- message:
			log.Printf("✅ Message sent to local broadcast channel")
		default:
			log.Printf("⚠️ Broadcast channel full, message dropped")
		}
		return nil
	}

	// Publish to Redis for cross-instance delivery
	log.Printf("📤 Publishing to Redis channel: %s", redisBroadcastChannel)
	if err := h.redisClient.Publish(redisBroadcastChannel, string(message)); err != nil {
		log.Printf("⚠️ Failed to publish message to Redis: %v", err)
		// Fallback to local broadcast
		select {
		case h.broadcast <- message:
			log.Printf("✅ Fallback: Message sent to local broadcast channel")
		default:
			log.Printf("⚠️ Fallback: Broadcast channel full, message dropped")
		}
		return err
	}
	log.Printf("✅ Message published to Redis successfully")

	// Also send to local clients immediately (don't wait for Redis echo)
	select {
	case h.broadcast <- message:
		log.Printf("✅ Message also sent to local broadcast channel")
	default:
		log.Printf("⚠️ Local broadcast channel full, skipping local delivery (will rely on Redis)")
	}
	return nil
}

// subscribeToRedis listens for messages published by other server instances
func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.Subscribe(redisBroadcastChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	log.Printf("📡 Listening on Redis channel: %s", redisBroadcastChannel)

	for msg := range ch {
		// Received message from another instance, broadcast to local clients
		message := []byte(msg.Payload)

		// Send to local clients only (don't re-publish to avoid loops)
		for client := range h.clients {
			select {
			case client.send <- message:
			default:
				// Client channel full, skip this message for this client
				log.Printf("⚠️ Client %d channel full, skipping message", client.userID)
			}
		}
	}
}

// broadcastUserStatus broadcasts user online/offline status to all connected clients
func (h *Hub) broadcastUserStatus(userID int, username string, isOnline bool) {
	statusMsg := WSMessage{
		Type:     "user_status",
		UserID:   userID,
		Username: username,
		IsOnline: isOnline,
	}

	msgBytes, err := json.Marshal(statusMsg)
	if err != nil {
		log.Printf("⚠️ Failed to marshal status message: %v", err)
		return
	}

	// Broadcast via Redis for cross-instance delivery
	if err := h.BroadcastViaRedis(msgBytes); err != nil {
		log.Printf("⚠️ Failed to broadcast status via Redis: %v", err)
	}

	log.Printf("📢 Broadcasted status: user=%d (%s) online=%v", userID, username, isOnline)
}
