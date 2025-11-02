package main

// Message structure for routing
type WSMessage struct {
	Type       string `json:"type"`         // "message", "join", "leave"
	From       string `json:"from"`         // username of sender
	FromUserID int    `json:"from_user_id"` // user ID of sender
	ToUserID   int    `json:"to_user_id"`   // user ID of recipient (0 = broadcast)
	Text       string `json:"text"`
	User       string `json:"user"`
}

//// Hub quản lý tất cả client đang kết nối và phân phối tin nhắn giữa họ
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	directMsg  chan *DirectMessage // channel for direct messages
	register   chan *Client
	unregister chan *Client
}

// DirectMessage contains message and target user ID
type DirectMessage struct {
	message  []byte
	toUserID int
}

//Newhub khởi tạo 1 hub mới
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		directMsg:  make(chan *DirectMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

//// run() chạy liên tục, xử lý các sự kiện từ các channel
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			//thêm client mới vào danh sách
			h.clients[client] = true

		case client := <-h.unregister:
			//xóa client khi ngắt kết nối
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
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
