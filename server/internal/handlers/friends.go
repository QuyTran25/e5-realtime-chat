package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"e5realtimechat/internal/auth"
	"e5realtimechat/internal/cache"
)

// Cấu trúc dữ liệu bạn bè
type Friend struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"` // Alias for username
	Email     string `json:"email,omitempty"`
	Avatar    string `json:"avatar"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Online    bool   `json:"online"`
	IsOnline  bool   `json:"is_online"`
	Status    string `json:"status,omitempty"`
}

// FriendsService handles friend-related operations
type FriendsService struct {
	db           *sql.DB
	cacheService *cache.CacheService
}

// NewFriendsService creates a new friends service
func NewFriendsService(db *sql.DB) *FriendsService {
	return &FriendsService{
		db:           db,
		cacheService: nil, // Will be set later
	}
}

// SetCacheService sets the cache service for the friends service
func (s *FriendsService) SetCacheService(cacheService *cache.CacheService) {
	s.cacheService = cacheService
}

// GetUserFriends retrieves all accepted friends for a user (with caching)
func (s *FriendsService) GetUserFriends(userID int) ([]Friend, error) {
	// Try to get from cache first
	if s.cacheService != nil {
		cachedFriends, err := s.cacheService.GetFriendsList(userID)
		if err == nil && len(cachedFriends) > 0 {
			log.Printf("✅ Cache HIT: Friends list for user %d", userID)
			// Convert cache.Friend to handlers.Friend
			friends := make([]Friend, len(cachedFriends))
			for i, cf := range cachedFriends {
				friends[i] = Friend{
					ID:        cf.ID,
					Username:  cf.Username,
					Name:      cf.Username,
					Email:     cf.Email,
					Avatar:    cf.AvatarURL,
					AvatarURL: cf.AvatarURL,
					Online:    cf.IsOnline,
					IsOnline:  cf.IsOnline,
				}
			}
			return friends, nil
		}
		log.Printf("⚠️ Cache MISS: Friends list for user %d", userID)
	}

	// Cache miss or no cache - fetch from database
	query := `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar_url, ''), u.is_online
		FROM users u
		INNER JOIN friendships f ON (u.id = f.friend_id OR u.id = f.user_id)
		WHERE (f.user_id = $1 OR f.friend_id = $1) 
		  AND f.status = 'accepted'
		  AND u.id != $1
		ORDER BY u.is_online DESC, u.username ASC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var friend Friend
		err := rows.Scan(&friend.ID, &friend.Username, &friend.Email, &friend.AvatarURL, &friend.IsOnline)
		if err != nil {
			return nil, err
		}
		// Set aliases for frontend compatibility
		friend.Name = friend.Username
		friend.Avatar = friend.AvatarURL
		friend.Online = friend.IsOnline
		friends = append(friends, friend)
	}

	// Store in cache for next time
	if s.cacheService != nil && len(friends) > 0 {
		cacheFriends := make([]cache.Friend, len(friends))
		for i, f := range friends {
			cacheFriends[i] = cache.Friend{
				ID:        f.ID,
				Username:  f.Username,
				Email:     f.Email,
				AvatarURL: f.AvatarURL,
				IsOnline:  f.IsOnline,
			}
		}
		if err := s.cacheService.SetFriendsList(userID, cacheFriends); err != nil {
			log.Printf("⚠️ Failed to cache friends list: %v", err)
		} else {
			log.Printf("✅ Cached friends list for user %d", userID)
		}
	}

	return friends, nil
}

// SearchUsers searches for users by username or email
func (s *FriendsService) SearchUsers(query string, currentUserID int, limit int) ([]Friend, error) {
	searchQuery := `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar_url, ''), u.is_online,
		       CASE 
		           WHEN f.id IS NOT NULL AND f.status = 'accepted' THEN 'friend'
		           WHEN f.id IS NOT NULL AND f.status = 'pending' THEN 'pending'
		           ELSE 'none'
		       END as friendship_status
		FROM users u
		LEFT JOIN friendships f ON (
		    (f.user_id = $1 AND f.friend_id = u.id) OR 
		    (f.friend_id = $1 AND f.user_id = u.id)
		)
		WHERE u.id != $1 
		  AND (LOWER(u.username) LIKE LOWER($2) OR LOWER(u.email) LIKE LOWER($2))
		ORDER BY u.username ASC
		LIMIT $3
	`

	searchPattern := "%" + query + "%"
	rows, err := s.db.Query(searchQuery, currentUserID, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []Friend
	for rows.Next() {
		var user Friend
		var status string
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL, &user.IsOnline, &status)
		if err != nil {
			return nil, err
		}
		user.Name = user.Username
		user.Avatar = user.AvatarURL
		user.Online = user.IsOnline
		user.Status = status
		users = append(users, user)
	}

	return users, nil
}

// SendFriendRequest sends a friend request
func (s *FriendsService) SendFriendRequest(fromUserID, toUserID int) error {
	// Check if friendship already exists
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM friendships 
			WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)
		)
	`, fromUserID, toUserID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return auth.ErrUserExists // Reuse error, or create new one
	}

	// Insert friend request
	_, err = s.db.Exec(`
		INSERT INTO friendships (user_id, friend_id, status, created_at)
		VALUES ($1, $2, 'pending', NOW())
	`, fromUserID, toUserID)

	// Invalidate cache for both users
	if s.cacheService != nil {
		s.cacheService.InvalidateFriendsList(fromUserID)
		s.cacheService.InvalidateFriendsList(toUserID)
	}

	return err
}

// AcceptFriendRequest accepts a friend request
func (s *FriendsService) AcceptFriendRequest(userID, friendID int) error {
	result, err := s.db.Exec(`
		UPDATE friendships 
		SET status = 'accepted', updated_at = NOW()
		WHERE friend_id = $1 AND user_id = $2 AND status = 'pending'
	`, userID, friendID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return auth.ErrUserNotFound // No pending request found
	}

	// Invalidate cache for both users
	if s.cacheService != nil {
		s.cacheService.InvalidateFriendsList(userID)
		s.cacheService.InvalidateFriendsList(friendID)
	}

	return nil
}

// RejectFriendRequest rejects a friend request
func (s *FriendsService) RejectFriendRequest(userID, friendID int) error {
	result, err := s.db.Exec(`
		DELETE FROM friendships 
		WHERE friend_id = $1 AND user_id = $2 AND status = 'pending'
	`, userID, friendID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return auth.ErrUserNotFound
	}

	// Invalidate cache for both users
	if s.cacheService != nil {
		s.cacheService.InvalidateFriendsList(userID)
		s.cacheService.InvalidateFriendsList(friendID)
	}

	return nil
}

// GetFriendRequests gets pending friend requests for a user
func (s *FriendsService) GetFriendRequests(userID int) ([]Friend, error) {
	query := `
		SELECT u.id, u.username, u.email, COALESCE(u.avatar_url, ''), u.is_online
		FROM users u
		INNER JOIN friendships f ON u.id = f.user_id
		WHERE f.friend_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Friend
	for rows.Next() {
		var user Friend
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL, &user.IsOnline)
		if err != nil {
			return nil, err
		}
		user.Name = user.Username
		user.Avatar = user.AvatarURL
		user.Online = user.IsOnline
		requests = append(requests, user)
	}

	return requests, nil
}

// HTTP Handlers

var friendsService *FriendsService

// friendsHandler returns list of user's friends
func friendsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user from context
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friends, err := friendsService.GetUserFriends(claims.UserID)
	if err != nil {
		log.Printf("❌ Error getting friends: %v", err)
		http.Error(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(friends)
}

// searchUsersHandler searches for users
func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode([]Friend{})
		return
	}

	users, err := friendsService.SearchUsers(query, claims.UserID, 20)
	if err != nil {
		log.Printf("❌ Error searching users: %v", err)
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(users)
}

// sendFriendRequestHandler sends a friend request
func sendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FriendID int `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.FriendID == claims.UserID {
		http.Error(w, "Cannot send friend request to yourself", http.StatusBadRequest)
		return
	}

	err := friendsService.SendFriendRequest(claims.UserID, req.FriendID)
	if err != nil {
		log.Printf("❌ Error sending friend request: %v", err)
		http.Error(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Friend request sent",
	})
}

// getFriendRequestsHandler gets pending friend requests
func getFriendRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := friendsService.GetFriendRequests(claims.UserID)
	if err != nil {
		log.Printf("❌ Error getting friend requests: %v", err)
		http.Error(w, "Failed to get friend requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(requests)
}

// acceptFriendRequestHandler accepts a friend request
func acceptFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FriendID int `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := friendsService.AcceptFriendRequest(claims.UserID, req.FriendID)
	if err != nil {
		log.Printf("❌ Error accepting friend request: %v", err)
		http.Error(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Friend request accepted",
	})
}

// rejectFriendRequestHandler rejects a friend request
func rejectFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FriendID int `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := friendsService.RejectFriendRequest(claims.UserID, req.FriendID)
	if err != nil {
		log.Printf("❌ Error rejecting friend request: %v", err)
		http.Error(w, "Failed to reject friend request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Friend request rejected",
	})
}

// Exported handlers for main.go

// FriendsHandler returns handler for getting friends list
func FriendsHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return friendsHandler
}

// SearchUsersHandler returns handler for searching users
func SearchUsersHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return searchUsersHandler
}

// SendFriendRequestHandler returns handler for sending friend requests
func SendFriendRequestHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return sendFriendRequestHandler
}

// GetFriendRequestsHandler returns handler for getting friend requests
func GetFriendRequestsHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return getFriendRequestsHandler
}

// AcceptFriendRequestHandler returns handler for accepting friend requests
func AcceptFriendRequestHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return acceptFriendRequestHandler
}

// RejectFriendRequestHandler returns handler for rejecting friend requests
func RejectFriendRequestHandler(service *FriendsService) http.HandlerFunc {
	friendsService = service
	return rejectFriendRequestHandler
}
