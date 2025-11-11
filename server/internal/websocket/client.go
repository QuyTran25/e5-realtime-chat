// Định nghĩa struct Client
// TEMPORARY STUB - Will be properly implemented by Người 1 + 2
package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	ws "github.com/gorilla/websocket"
)

// Upgrader upgrades HTTP requests to WebSocket connections
var Upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// For demo/local dev we allow all origins. In production, restrict this.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// các hằng số cấu hình cho việc đọc/ghi message
// cần đặt timeout và tần suất ping/pong
const (
	writeWait      = 10 * time.Second    // thời gian tối đa để ghi message xuống client
	pongWait       = 60 * time.Second    // thời gian chờ nhận pong từ client
	pingPeriod     = (pongWait * 9) / 10 // gửi ping đều đặn để giữ kết nối
	maxMessageSize = 512                 // giới hạn kích thước message
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Định nghĩa struct Client
// Một Client đại diện cho một kết nối websocket tới một user cụ thể
// Nó sẽ đọc tin nhắn từ kết nối và gửi tin nhắn từ Hub xuống kết nối
type Client struct {
	hub      *Hub        // tham chiếu tới Hub (quản lý chung)
	conn     *ws.Conn    // kết nối websocket thật sự
	send     chan []byte // kênh để nhận tin nhắn từ Hub và gửi xuống client
	userID   int         // ID của user đang kết nối
	username string      // Tên của user đang kết nối
}

// SaveMessageFunc is a function type for saving messages to database
type SaveMessageFunc func(fromUserID, toUserID int, messageText string) error

var saveMessageToDB SaveMessageFunc

// SetSaveMessageFunc sets the function for saving messages
func SetSaveMessageFunc(fn SaveMessageFunc) {
	saveMessageToDB = fn
}

// Hàm readPump() – Đọc tin nhắn từ Client
// Hàm này chạy ở 1 goroutine riêng. Nó:
// Liên tục đọc message từ client.
// Khi đọc được message → gửi message đó vào hub.broadcast.
// Nếu client ngắt kết nối hoặc lỗi → unregister client.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c // thông báo Hub biết client rời đi
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("❌ lỗi đọc message: %v", err)
			}
			break
		}

		// làm sạch message
		message = bytes.TrimSpace(bytes.Replace(message, []byte("\n"), []byte(" "), -1))

		// Parse message to check if it's a direct message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err == nil {
			// Check rate limit for this user
			if c.hub.rateLimiter != nil && c.userID > 0 {
				allowed, err := c.hub.rateLimiter.CheckUserMessageRate(c.userID)
				if err != nil {
					log.Printf("⚠️ Rate limit check error for user %d: %v", c.userID, err)
				} else if !allowed {
					log.Printf("🚫 Rate limit exceeded for user %d (%s)", c.userID, c.username)
					// Send rate limit error back to client
					errorMsg := WSMessage{
						Type:       "error",
						Text:       "Rate limit exceeded. Please slow down.",
						FromUserID: 0,
						From:       "System",
					}
					if errBytes, err := json.Marshal(errorMsg); err == nil {
						c.send <- errBytes
					}
					continue // Skip this message
				}
			}

			// Add sender info
			wsMsg.FromUserID = c.userID
			wsMsg.From = c.username

			// Re-encode message with sender info
			enhancedMsg, err := json.Marshal(wsMsg)
			if err == nil {
				message = enhancedMsg
			}

			// Save message to database if it's a chat message
			if wsMsg.Type == "message" && wsMsg.Text != "" {
				if wsMsg.ToUserID > 0 {
					// Private message - save to database
					if saveMessageToDB != nil {
						if err := saveMessageToDB(c.userID, wsMsg.ToUserID, wsMsg.Text); err != nil {
							log.Printf("❌ Error saving message to DB: %v", err)
						}
					}
				}
			}

			// Check if this is a direct message
			if wsMsg.ToUserID > 0 {
				// Send to specific user
				c.hub.SendDirectMessage(message, wsMsg.ToUserID)
				// Also send back to sender for confirmation
				c.send <- message
			} else {
				// Broadcast to all instances via Redis Pub/Sub
				if err := c.hub.BroadcastViaRedis(message); err != nil {
					log.Printf("⚠️ Failed to broadcast via Redis: %v", err)
				}
			}
		} else {
			// If parse fails, broadcast via Redis
			if err := c.hub.BroadcastViaRedis(message); err != nil {
				log.Printf("⚠️ Failed to broadcast via Redis: %v", err)
			}
		}
	}
}

// Hàm writePump() – Gửi tin nhắn tới Client
// Hàm này chạy ở 1 goroutine riêng. Nó:
// Liên tục lắng nghe kênh c.send để gửi tin nhắn tới client.
// Gửi tin ping định kỳ để giữ kết nối.
// Nếu kênh c.send bị đóng hoặc lỗi khi gửi tin nhắn → đóng kết nối.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel bị đóng => đóng kết nối
				c.conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// gửi các tin trong queue còn lại trong channel (nếu có)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Gửi ping để giữ kết nối
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
