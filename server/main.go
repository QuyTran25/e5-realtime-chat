package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"e5realtimechat/internal/auth"
	"e5realtimechat/internal/cache"
	"e5realtimechat/internal/database"
	"e5realtimechat/internal/handlers"
	"e5realtimechat/internal/queue"
	"e5realtimechat/internal/websocket"

	_ "github.com/lib/pq"
)

// Global database instance
var db *database.DB

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

// serveWs handles WebSocket requests from the peer (with authentication).
func serveWs(hub *websocket.Hub, authService *auth.AuthService, w http.ResponseWriter, r *http.Request) {
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
	if auth.IsTokenBlacklisted(token) {
		http.Error(w, "Unauthorized: token revoked", http.StatusUnauthorized)
		log.Println("❌ WebSocket connection rejected: token revoked")
		return
	}

	claims, err := auth.ValidateToken(token)
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
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ websocket upgrade error: %v", err)
		return
	}

	log.Printf("✅ WebSocket connected: userID=%d, username=%s", user.ID, user.Username)

	// Create client with user info
	client := hub.Register(conn, user.ID, user.Username)

	// Start write pump in a goroutine, run read pump on this goroutine
	// so that when readPump returns, we can exit the handler cleanly.
	client.StartClient()
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

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Database ping failed:", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Initialize Redis cache
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	redisClient, err := cache.NewRedisClient(redisAddr, "", 0)
	if err != nil {
		log.Printf("⚠️  Failed to connect to Redis: %v", err)
		log.Println("⚠️  Running without Redis cache (using in-memory fallback)")
		redisClient = nil
	}

	// Initialize cache service
	var cacheService *cache.CacheService
	if redisClient != nil {
		cacheService = cache.NewCacheService(redisClient)
		defer redisClient.Close()
	}

	// Initialize token blacklist with cache
	auth.InitTokenBlacklist(cacheService)

	// Initialize RabbitMQ
	rabbitmqHost := getEnv("RABBITMQ_HOST", "localhost")
	rabbitmqPort := getEnv("RABBITMQ_PORT", "5672")
	rabbitmqUser := getEnv("RABBITMQ_USER", "chatuser")
	rabbitmqPass := getEnv("RABBITMQ_PASS", "chatpass")

	rabbitMQ, err := queue.NewRabbitMQ(rabbitmqHost, rabbitmqPort, rabbitmqUser, rabbitmqPass)
	if err != nil {
		log.Printf("⚠️  Failed to connect to RabbitMQ: %v", err)
		log.Println("⚠️  Running without message queue")
		rabbitMQ = nil
	} else {
		defer rabbitMQ.Close()
	}

	// Initialize message service
	var messageService *queue.MessageService
	if rabbitMQ != nil {
		messageService = queue.NewMessageService(rabbitMQ)

		// Start consumers
		if err := messageService.StartMessageNotificationConsumer(); err != nil {
			log.Printf("⚠️  Failed to start message notification consumer: %v", err)
		}
		if err := messageService.StartEmailNotificationConsumer(); err != nil {
			log.Printf("⚠️  Failed to start email notification consumer: %v", err)
		}
		if err := messageService.StartMessageProcessingConsumer(); err != nil {
			log.Printf("⚠️  Failed to start message processing consumer: %v", err)
		}
		if err := messageService.StartEventBroadcastConsumer(); err != nil {
			log.Printf("⚠️  Failed to start event broadcast consumer: %v", err)
		}

		log.Println("✅ All RabbitMQ consumers started")
	}

	// Initialize database wrapper
	db = database.NewDBFromConnection(sqlDB)

	// Initialize auth service
	authService := auth.NewAuthService(sqlDB)
	authHandler := auth.NewHandler(authService)

	// Initialize friends service
	friendsService := handlers.NewFriendsService(sqlDB)

	// Set cache service for friends service
	if cacheService != nil {
		friendsService.SetCacheService(cacheService)
	}

	// Set DB instance for handlers
	handlers.SetDBInstance(db)

	// Set save message function for websocket
	websocket.SetSaveMessageFunc(handlers.SaveMessageToDB)

	// Create hub
	hub := websocket.NewHub()

	// Set cache service for online status tracking
	if cacheService != nil {
		hub.SetCacheService(cacheService)
	}

	go hub.Run()

	// Setup routes
	mux := http.NewServeMux()

	// Auth routes (no auth required)
	authHandler.RegisterRoutes(mux)

	// WebSocket route (requires auth)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, authService, w, r)
	})

	// Health check with instance info
	instanceID := getEnv("HOSTNAME", "unknown")
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{"status":"ok","instance":"%s"}`, instanceID)
		_, _ = w.Write([]byte(response))
	})

	// Friends API (protected)
	mux.HandleFunc("/api/friends", auth.AuthMiddleware(handlers.FriendsHandler(friendsService)))
	mux.HandleFunc("/api/friends/search", auth.AuthMiddleware(handlers.SearchUsersHandler(friendsService)))
	mux.HandleFunc("/api/friends/request", auth.AuthMiddleware(handlers.SendFriendRequestHandler(friendsService)))
	mux.HandleFunc("/api/friends/requests", auth.AuthMiddleware(handlers.GetFriendRequestsHandler(friendsService)))
	mux.HandleFunc("/api/friends/accept", auth.AuthMiddleware(handlers.AcceptFriendRequestHandler(friendsService)))
	mux.HandleFunc("/api/friends/reject", auth.AuthMiddleware(handlers.RejectFriendRequestHandler(friendsService)))

	// Messages API (protected)
	mux.HandleFunc("/api/messages/history", auth.AuthMiddleware(handlers.GetMessageHistoryHandler(db)))
	mux.HandleFunc("/api/conversations", auth.AuthMiddleware(handlers.GetConversationsHandler(db)))

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	log.Println("🚀 E5 Realtime Chat Server")
	log.Printf("📡 WebSocket endpoint: ws://localhost%s/ws?token=YOUR_TOKEN", addr)
	log.Printf("🔐 Auth API: http://localhost%s/api/auth/", addr)
	log.Printf("💚 Health check: http://localhost%s/healthz", addr)
	log.Printf("👥 Friends API: http://localhost%s/api/friends", addr)
	log.Printf("🔍 Search API: http://localhost%s/api/friends/search?q=keyword", addr)
	log.Printf("📨 Friend Requests: http://localhost%s/api/friends/requests", addr)
	log.Printf("🎯 Server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
