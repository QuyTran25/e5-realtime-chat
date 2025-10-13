package main

//// Hub quản lý tất cả client đang kết nối và phân phối tin nhắn giữa họ
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

//Newhub khởi tạo 1 hub mới
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
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
		}
	}
}
