package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"e5realtimechat/Auth"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
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

// serveWs handles WebSocket requests from the peer (with authentication).
func serveWs(hub *Hub, authService *Auth.AuthService, w http.ResponseWriter, r *http.Request) {
	// Validate token from query parameter or header
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		log.Println("❌ WebSocket connection rejected: missing token")
		return
	}

	// Validate token
	if Auth.IsTokenBlacklisted(token) {
		http.Error(w, "Unauthorized: token revoked", http.StatusUnauthorized)
		log.Println("❌ WebSocket connection rejected: token revoked")
		return
	}

	claims, err := Auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		log.Printf("❌ WebSocket connection rejected: invalid token - %v", err)
		return
	}

	// Get user info
	user, err := authService.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
		log.Printf("❌ WebSocket connection rejected: user not found - %v", err)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ websocket upgrade error: %v", err)
		return
	}

	log.Printf("✅ WebSocket connected: userID=%d, username=%s", user.ID, user.Username)

	// Create client with user info
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   user.ID,
		username: user.Username,
	}
	hub.register <- client

	// Start write pump in a goroutine, run read pump on this goroutine
	// so that when readPump returns, we can exit the handler cleanly.
	go client.writePump()
	client.readPump()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "chatuser")
	dbPass := getEnv("DB_PASSWORD", "chatpass")
	dbName := getEnv("DB_NAME", "chatdb")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("❌ Database ping failed:", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize auth service
	authService := Auth.NewAuthService(db)
	authHandler := Auth.NewHandler(authService)

	// Create hub
	hub := NewHub()
	go hub.run()

	// Setup routes
	mux := http.NewServeMux()

	// Auth routes (no auth required)
	authHandler.RegisterRoutes(mux)

	// WebSocket route (requires auth)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, authService, w, r)
	})

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Friends API (protected)
	mux.HandleFunc("/api/friends", Auth.AuthMiddleware(friendsHandler))

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Println("🚀 E5 Realtime Chat Server")
	log.Printf("📡 WebSocket endpoint: ws://localhost%s/ws?token=YOUR_TOKEN", addr)
	log.Printf("🔐 Auth API: http://localhost%s/api/auth/", addr)
	log.Printf("💚 Health check: http://localhost%s/healthz", addr)
	log.Printf("👥 Friends API: http://localhost%s/api/friends", addr)
	log.Printf("🎯 Server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
