package websocket

import (
	"e5realtimechat/internal/cache"
	"log"

	ws "github.com/gorilla/websocket"
)

// Message structure for routing
type WSMessage struct {
	Type       string `json:"type"`         // "message", "join", "leave"
	From       string `json:"from"`         // username of sender
	FromUserID int    `json:"from_user_id"` // user ID of sender
	ToUserID   int    `json:"to_user_id"`   // user ID of recipient (0 = broadcast)
	Text       string `json:"text"`
	User       string `json:"user"`
}

// // Hub quản lý tất cả client đang kết nối và phân phối tin nhắn giữa họ
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	directMsg    chan *DirectMessage // channel for direct messages
	register     chan *Client
	unregister   chan *Client
	cacheService *cache.CacheService // Redis cache for online status
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
	}
}

// SetCacheService sets the cache service for the hub
func (h *Hub) SetCacheService(cacheService *cache.CacheService) {
	h.cacheService = cacheService
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
					}
				}
			}

		case message := <-h.broadcast:
			//gửi tin nhắn tới tất cả clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Nếu không gửi được → đóng kết nối client
					close(client.send)
					delete(h.clients, client)
				}
			}

		case directMsg := <-h.directMsg:
			// Gửi tin nhắn riêng tư cho user cụ thể
			for client := range h.clients {
				if client.userID == directMsg.toUserID {
					select {
					case client.send <- directMsg.message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

// SendDirectMessage sends a message to a specific user
func (h *Hub) SendDirectMessage(message []byte, toUserID int) {
	h.directMsg <- &DirectMessage{
		message:  message,
		toUserID: toUserID,
	}
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
