package cache

import (
	"fmt"
	"time"
)

// Cache keys prefixes
const (
	// Session/Token cache
	PrefixTokenBlacklist = "token:blacklist:"
	PrefixUserSession    = "session:user:"

	// User data cache
	PrefixUserProfile = "user:profile:"
	PrefixUserFriends = "user:friends:"
	PrefixUserOnline  = "user:online:"

	// Online users set
	KeyOnlineUsers = "online:users"

	// Message cache
	PrefixConversation   = "conversation:"
	PrefixMessageHistory = "messages:history:"
)

// Cache TTL durations
const (
	TTLToken        = 24 * time.Hour   // Token blacklist
	TTLSession      = 24 * time.Hour   // User session
	TTLUserProfile  = 1 * time.Hour    // User profile
	TTLFriendsList  = 5 * time.Minute  // Friends list
	TTLOnlineStatus = 30 * time.Second // Online status
	TTLMessages     = 10 * time.Minute // Message history
)

// CacheService provides high-level caching operations
type CacheService struct {
	redis *RedisClient
}

// NewCacheService creates a new cache service
func NewCacheService(redis *RedisClient) *CacheService {
	return &CacheService{redis: redis}
}

// ==================== TOKEN BLACKLIST ====================

// BlacklistToken adds a token to the blacklist
func (c *CacheService) BlacklistToken(token string, expiration time.Duration) error {
	key := PrefixTokenBlacklist + token
	return c.redis.Set(key, "1", expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (c *CacheService) IsTokenBlacklisted(token string) (bool, error) {
	key := PrefixTokenBlacklist + token
	return c.redis.Exists(key)
}

// ==================== USER SESSION ====================

// UserSession represents cached user session data
type UserSession struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	LoginAt  time.Time `json:"login_at"`
	LastSeen time.Time `json:"last_seen"`
}

// SetUserSession caches user session
func (c *CacheService) SetUserSession(userID int, session *UserSession) error {
	key := fmt.Sprintf("%s%d", PrefixUserSession, userID)
	return c.redis.SetJSON(key, session, TTLSession)
}

// GetUserSession retrieves cached user session
func (c *CacheService) GetUserSession(userID int) (*UserSession, error) {
	key := fmt.Sprintf("%s%d", PrefixUserSession, userID)
	var session UserSession
	err := c.redis.GetJSON(key, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteUserSession removes user session from cache
func (c *CacheService) DeleteUserSession(userID int) error {
	key := fmt.Sprintf("%s%d", PrefixUserSession, userID)
	return c.redis.Delete(key)
}

// ==================== USER PROFILE ====================

// UserProfile represents cached user profile
type UserProfile struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// SetUserProfile caches user profile
func (c *CacheService) SetUserProfile(userID int, profile *UserProfile) error {
	key := fmt.Sprintf("%s%d", PrefixUserProfile, userID)
	return c.redis.SetJSON(key, profile, TTLUserProfile)
}

// GetUserProfile retrieves cached user profile
func (c *CacheService) GetUserProfile(userID int) (*UserProfile, error) {
	key := fmt.Sprintf("%s%d", PrefixUserProfile, userID)
	var profile UserProfile
	err := c.redis.GetJSON(key, &profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// InvalidateUserProfile removes user profile from cache
func (c *CacheService) InvalidateUserProfile(userID int) error {
	key := fmt.Sprintf("%s%d", PrefixUserProfile, userID)
	return c.redis.Delete(key)
}

// ==================== FRIENDS LIST ====================

// Friend represents a cached friend
type Friend struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	IsOnline  bool   `json:"is_online"`
}

// SetFriendsList caches user's friends list
func (c *CacheService) SetFriendsList(userID int, friends []Friend) error {
	key := fmt.Sprintf("%s%d", PrefixUserFriends, userID)
	return c.redis.SetJSON(key, friends, TTLFriendsList)
}

// GetFriendsList retrieves cached friends list
func (c *CacheService) GetFriendsList(userID int) ([]Friend, error) {
	key := fmt.Sprintf("%s%d", PrefixUserFriends, userID)
	var friends []Friend
	err := c.redis.GetJSON(key, &friends)
	if err != nil {
		return nil, err
	}
	return friends, nil
}

// InvalidateFriendsList removes friends list from cache
func (c *CacheService) InvalidateFriendsList(userID int) error {
	key := fmt.Sprintf("%s%d", PrefixUserFriends, userID)
	return c.redis.Delete(key)
}

// ==================== ONLINE STATUS ====================

// SetUserOnline marks a user as online
func (c *CacheService) SetUserOnline(userID int) error {
	// Add to online users set
	if err := c.redis.SAdd(KeyOnlineUsers, userID); err != nil {
		return err
	}

	// Set online flag with expiration (will auto-expire if user doesn't refresh)
	key := fmt.Sprintf("%s%d", PrefixUserOnline, userID)
	return c.redis.Set(key, "1", TTLOnlineStatus)
}

// SetUserOffline marks a user as offline
func (c *CacheService) SetUserOffline(userID int) error {
	// Remove from online users set
	if err := c.redis.SRem(KeyOnlineUsers, userID); err != nil {
		return err
	}

	// Delete online flag
	key := fmt.Sprintf("%s%d", PrefixUserOnline, userID)
	return c.redis.Delete(key)
}

// IsUserOnline checks if a user is online
func (c *CacheService) IsUserOnline(userID int) (bool, error) {
	return c.redis.SIsMember(KeyOnlineUsers, userID)
}

// GetOnlineUsers returns list of all online user IDs
func (c *CacheService) GetOnlineUsers() ([]string, error) {
	return c.redis.SMembers(KeyOnlineUsers)
}

// RefreshUserOnline refreshes user's online status (called on heartbeat)
func (c *CacheService) RefreshUserOnline(userID int) error {
	key := fmt.Sprintf("%s%d", PrefixUserOnline, userID)
	return c.redis.Expire(key, TTLOnlineStatus)
}

// ==================== MESSAGE HISTORY ====================

// Message represents a cached message
type Message struct {
	ID        int       `json:"id"`
	FromID    int       `json:"from_id"`
	ToID      int       `json:"to_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// GetConversationKey generates a consistent key for conversation between two users
func GetConversationKey(userID1, userID2 int) string {
	// Ensure consistent key regardless of order
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	return fmt.Sprintf("%s%d:%d", PrefixConversation, userID1, userID2)
}

// CacheConversationMessages caches recent messages for a conversation
func (c *CacheService) CacheConversationMessages(userID1, userID2 int, messages []Message) error {
	key := GetConversationKey(userID1, userID2)
	return c.redis.SetJSON(key, messages, TTLMessages)
}

// GetConversationMessages retrieves cached conversation messages
func (c *CacheService) GetConversationMessages(userID1, userID2 int) ([]Message, error) {
	key := GetConversationKey(userID1, userID2)
	var messages []Message
	err := c.redis.GetJSON(key, &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// InvalidateConversation removes conversation from cache
func (c *CacheService) InvalidateConversation(userID1, userID2 int) error {
	key := GetConversationKey(userID1, userID2)
	return c.redis.Delete(key)
}

// ==================== UTILITY ====================

// ClearUserCache clears all cache related to a user
func (c *CacheService) ClearUserCache(userID int) error {
	keys := []string{
		fmt.Sprintf("%s%d", PrefixUserSession, userID),
		fmt.Sprintf("%s%d", PrefixUserProfile, userID),
		fmt.Sprintf("%s%d", PrefixUserFriends, userID),
		fmt.Sprintf("%s%d", PrefixUserOnline, userID),
	}

	for _, key := range keys {
		if err := c.redis.Delete(key); err != nil {
			return err
		}
	}

	// Remove from online users
	return c.redis.SRem(KeyOnlineUsers, userID)
}

// Ping tests Redis connection
func (c *CacheService) Ping() error {
	return c.redis.Ping()
}
