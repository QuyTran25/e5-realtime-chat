package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// Message represents the JSON payload exchanged over WebSocket.
// Examples:
// {"type":"message","from":"user1","text":"hello"}
// {"type":"join","user":"user2"}
// {"type":"leave","user":"user3"}
// {"type":"data","value":32.5}
type Message struct {
	Type  string  `json:"type"`
	From  string  `json:"from,omitempty"`
	Text  string  `json:"text,omitempty"`
	User  string  `json:"user,omitempty"`
	Value float64 `json:"value,omitempty"`
}

// Upgrader upgrades HTTP requests to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// For demo/local dev we allow all origins. In production, restrict this.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// serveWs handles WebSocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	// Start write pump in a goroutine, run read pump on this goroutine
	// so that when readPump returns, we can exit the handler cleanly.
	go client.writePump()
	client.readPump()
}


func main() {
	hub := NewHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// 👉 Thêm dòng này để xử lý API danh sách bạn bè
	http.HandleFunc("/api/friends", friendsHandler)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Printf("WebSocket server starting on %s (endpoint: /ws)", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}