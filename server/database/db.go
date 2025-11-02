package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the database connection
type DB struct {
	conn *sql.DB
}

// User represents a user in the system
type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email,omitempty"`
	PasswordHash string     `json:"-"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	IsOnline     bool       `json:"is_online"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Message represents a chat message
type Message struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`
	FromUserID   int       `json:"from_user_id"`
	FromUsername string    `json:"from_username,omitempty"`
	ToUserID     *int      `json:"to_user_id,omitempty"`
	RoomID       *int      `json:"room_id,omitempty"`
	Text         string    `json:"text"`
	Value        *float64  `json:"value,omitempty"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
}

// Room represents a chat room
type Room struct {
	ID          int       `json:"id"`
	RoomName    string    `json:"room_name"`
	RoomType    string    `json:"room_type"`
	Description string    `json:"description,omitempty"`
	CreatedBy   *int      `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewDB creates a new database connection
func NewDB(host, port, user, password, dbname string) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("✅ Database connected successfully: %s:%s/%s", host, port, dbname)

	return &DB{conn: conn}, nil
}

// NewDBFromConnection creates a DB wrapper from existing sql.DB connection
func NewDBFromConnection(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// ============================================
// User Methods
// ============================================

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, is_online, last_seen_at, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := db.conn.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsOnline,
		&user.LastSeenAt,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, is_online, last_seen_at, created_at
		FROM users
		WHERE username = $1
	`

	var user User
	err := db.conn.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsOnline,
		&user.LastSeenAt,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (db *DB) CreateUser(username, email, passwordHash, avatarURL string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, avatar_url, is_online, created_at
	`

	var user User
	err := db.conn.QueryRow(query, username, email, passwordHash, avatarURL).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.AvatarURL,
		&user.IsOnline,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// UpdateUserOnlineStatus updates user's online status
func (db *DB) UpdateUserOnlineStatus(userID int, isOnline bool) error {
	query := `
		UPDATE users
		SET is_online = $1, last_seen_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	_, err := db.conn.Exec(query, isOnline, userID)
	if err != nil {
		return fmt.Errorf("failed to update online status: %w", err)
	}

	return nil
}

// ============================================
// Message Methods
// ============================================

// SaveMessage saves a new message to the database
func (db *DB) SaveMessage(msg *Message) error {
	query := `
		INSERT INTO messages (message_type, from_user_id, to_user_id, room_id, message_text, message_value)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := db.conn.QueryRow(
		query,
		msg.Type,
		msg.FromUserID,
		msg.ToUserID,
		msg.RoomID,
		msg.Text,
		msg.Value,
	).Scan(&msg.ID, &msg.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	log.Printf("💾 Message saved: ID=%d, Type=%s, From=%d", msg.ID, msg.Type, msg.FromUserID)
	return nil
}

// GetMessageHistory retrieves message history for a room
func (db *DB) GetMessageHistory(roomID int, limit int) ([]*Message, error) {
	query := `
		SELECT 
			m.id, m.message_type, m.from_user_id, u.username, 
			m.to_user_id, m.room_id, m.message_text, m.message_value, 
			m.is_read, m.created_at
		FROM messages m
		LEFT JOIN users u ON m.from_user_id = u.id
		WHERE m.room_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`

	rows, err := db.conn.Query(query, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.FromUserID,
			&msg.FromUsername,
			&msg.ToUserID,
			&msg.RoomID,
			&msg.Text,
			&msg.Value,
			&msg.IsRead,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetDirectMessageHistory retrieves direct message history between two users
func (db *DB) GetDirectMessageHistory(userID1, userID2, limit int) ([]*Message, error) {
	query := `
		SELECT 
			m.id, m.message_type, m.from_user_id, u.username,
			m.to_user_id, m.room_id, m.message_text, m.message_value,
			m.is_read, m.created_at
		FROM messages m
		LEFT JOIN users u ON m.from_user_id = u.id
		WHERE (m.from_user_id = $1 AND m.to_user_id = $2)
		   OR (m.from_user_id = $2 AND m.to_user_id = $1)
		ORDER BY m.created_at DESC
		LIMIT $3
	`

	rows, err := db.conn.Query(query, userID1, userID2, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct message history: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.FromUserID,
			&msg.FromUsername,
			&msg.ToUserID,
			&msg.RoomID,
			&msg.Text,
			&msg.Value,
			&msg.IsRead,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ============================================
// Room Methods
// ============================================

// GetRoomByName retrieves a room by name
func (db *DB) GetRoomByName(roomName string) (*Room, error) {
	query := `
		SELECT id, room_name, room_type, description, created_by, created_at
		FROM rooms
		WHERE room_name = $1
	`

	var room Room
	err := db.conn.QueryRow(query, roomName).Scan(
		&room.ID,
		&room.RoomName,
		&room.RoomType,
		&room.Description,
		&room.CreatedBy,
		&room.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("room not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	return &room, nil
}

// GetAllRooms retrieves all public rooms
func (db *DB) GetAllRooms() ([]*Room, error) {
	query := `
		SELECT id, room_name, room_type, description, created_by, created_at
		FROM rooms
		WHERE room_type = 'public'
		ORDER BY created_at ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*Room
	for rows.Next() {
		var room Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
			&room.RoomType,
			&room.Description,
			&room.CreatedBy,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, &room)
	}

	return rooms, nil
}

// ============================================
// Friend Methods
// ============================================

// GetUserFriends retrieves all accepted friends for a user
func (db *DB) GetUserFriends(userID int) ([]*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.avatar_url, u.is_online, u.last_seen_at
		FROM users u
		INNER JOIN friendships f ON u.id = f.friend_id
		WHERE f.user_id = $1 AND f.status = 'accepted'
		ORDER BY u.username ASC
	`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	defer rows.Close()

	var friends []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.AvatarURL,
			&user.IsOnline,
			&user.LastSeenAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend: %w", err)
		}
		friends = append(friends, &user)
	}

	return friends, nil
}

// GetPrivateMessages gets private messages between two users (alias for GetDirectMessageHistory)
func (db *DB) GetPrivateMessages(userID1, userID2, limit int) ([]*Message, error) {
	return db.GetDirectMessageHistory(userID1, userID2, limit)
}

// Conversation represents a conversation summary
type Conversation struct {
	UserID        int       `json:"user_id"`
	Username      string    `json:"username"`
	AvatarURL     string    `json:"avatar_url"`
	IsOnline      bool      `json:"is_online"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
}

// GetUserConversations retrieves list of conversations for a user
func (db *DB) GetUserConversations(userID int) ([]*Conversation, error) {
	query := `
		WITH recent_messages AS (
			SELECT DISTINCT ON (
				CASE 
					WHEN from_user_id = $1 THEN to_user_id
					ELSE from_user_id
				END
			)
			CASE 
				WHEN from_user_id = $1 THEN to_user_id
				ELSE from_user_id
			END as other_user_id,
			message_text,
			created_at,
			is_read,
			from_user_id
			FROM messages
			WHERE (from_user_id = $1 OR to_user_id = $1)
			  AND to_user_id IS NOT NULL
			ORDER BY 
				CASE 
					WHEN from_user_id = $1 THEN to_user_id
					ELSE from_user_id
				END,
				created_at DESC
		)
		SELECT 
			u.id,
			u.username,
			COALESCE(u.avatar_url, '') as avatar_url,
			u.is_online,
			rm.message_text,
			rm.created_at,
			COALESCE(
				(SELECT COUNT(*) 
				 FROM messages 
				 WHERE from_user_id = u.id 
				   AND to_user_id = $1 
				   AND is_read = false),
				0
			) as unread_count
		FROM recent_messages rm
		JOIN users u ON u.id = rm.other_user_id
		ORDER BY rm.created_at DESC
	`

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(
			&conv.UserID,
			&conv.Username,
			&conv.AvatarURL,
			&conv.IsOnline,
			&conv.LastMessage,
			&conv.LastMessageAt,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, &conv)
	}

	return conversations, nil
}
