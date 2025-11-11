package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"e5realtimechat/internal/auth"
	"e5realtimechat/internal/database"
)

// GetMessageHistoryHandler returns chat history between current user and another user
func GetMessageHistoryHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get current user from context
		userID, ok := r.Context().Value(auth.UserIDKey).(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get other user ID from query
		otherUserIDStr := r.URL.Query().Get("user_id")
		if otherUserIDStr == "" {
			http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
			return
		}

		otherUserID, err := strconv.Atoi(otherUserIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		// Get limit (default 50)
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		// Get messages from database
		messages, err := db.GetPrivateMessages(userID, otherUserID, limit)
		if err != nil {
			log.Printf("❌ Error getting message history: %v", err)
			http.Error(w, "Failed to get message history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"messages": messages,
		})
	}
}

// GetConversationsHandler returns list of conversations for current user
func GetConversationsHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get current user from context
		userID, ok := r.Context().Value(auth.UserIDKey).(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get conversations from database
		conversations, err := db.GetUserConversations(userID)
		if err != nil {
			log.Printf("❌ Error getting conversations: %v", err)
			http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":       true,
			"conversations": conversations,
		})
	}
}

var dbInstance *database.DB

// SetDBInstance sets the database instance for SaveMessageToDB
func SetDBInstance(db *database.DB) {
	dbInstance = db
}

// SaveMessageToDB saves a message to database (called from WebSocket handler)
func SaveMessageToDB(fromUserID, toUserID int, messageText string) error {
	msg := &database.Message{
		Type:       "message",
		FromUserID: fromUserID,
		ToUserID:   &toUserID,
		Text:       messageText,
	}

	return dbInstance.SaveMessage(msg)
}
