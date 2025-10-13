// Định nghĩa struct Client
// TEMPORARY STUB - Will be properly implemented by Người 1 + 2
package main

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// các hằng số cấu hình cho việc đọc/ghi message
// cần đặt timeout và tần suất ping/pong
const (
	writeWait = 10 * time.Second      // thời gian tối đa để ghi message xuống client
	pongWait = 60 * time.Second       // thời gian chờ nhận pong từ client
	pingPeriod = (pongWait * 9) / 10  // gửi ping đều đặn để giữ kết nối
	maxMessageSize = 512              // giới hạn kích thước message
)


var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Định nghĩa struct Client
// Một Client đại diện cho một kết nối websocket tới một user cụ thể
// Nó sẽ đọc tin nhắn từ kết nối và gửi tin nhắn từ Hub xuống kết nối
type Client struct {
	hub  *Hub              // tham chiếu tới Hub (quản lý chung)
	conn *websocket.Conn   // kết nối websocket thật sự
	send chan []byte       // kênh để nhận tin nhắn từ Hub và gửi xuống client
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ lỗi đọc message: %v", err)
			}
			break
		}

		// làm sạch message
		message = bytes.TrimSpace(bytes.Replace(message, []byte("\n"), []byte(" "), -1))

		// gửi message tới Hub để broadcast
		c.hub.broadcast <- message
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
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
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
