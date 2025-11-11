package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	// In-memory token blacklist (use Redis in production)
	tokenBlacklist  = make(map[string]time.Time)
	blacklistMutex  sync.RWMutex
	ErrInvalidCreds = errors.New("invalid credentials")
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// AuthService handles authentication business logic
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(req RegisterRequest) (*User, error) {
	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("all fields are required")
	}

	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if user exists
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)",
		req.Email, req.Username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	user := &User{}

	query := `
		INSERT INTO users (username, email, password_hash, is_online, created_at)
		VALUES ($1, $2, $3, false, NOW())
		RETURNING id, username, email, COALESCE(avatar_url, ''), is_online, COALESCE(last_seen_at, created_at), created_at
	`
	err = s.db.QueryRow(query, req.Username, req.Email, string(hashedPassword)).
		Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL,
			&user.IsOnline, &user.LastSeenAt, &user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// LoginUser authenticates a user and returns user data
func (s *AuthService) LoginUser(req LoginRequest) (*User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCreds
	}

	user := &User{}
	var passwordHash string
	var lastSeenAt sql.NullTime

	query := `
		SELECT id, username, email, password_hash, COALESCE(avatar_url, ''), is_online, last_seen_at, created_at
		FROM users
		WHERE email = $1
	`
	err := s.db.QueryRow(query, req.Email).
		Scan(&user.ID, &user.Username, &user.Email, &passwordHash,
			&user.AvatarURL, &user.IsOnline, &lastSeenAt, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidCreds
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Set last seen at
	if lastSeenAt.Valid {
		user.LastSeenAt = lastSeenAt.Time
	} else {
		user.LastSeenAt = time.Now()
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCreds
	}

	// Update online status
	_, err = s.db.Exec("UPDATE users SET is_online = true, last_seen_at = NOW() WHERE id = $1", user.ID)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to update online status: %v\n", err)
	}

	user.IsOnline = true
	return user, nil
}

// LogoutUser marks user as offline
func (s *AuthService) LogoutUser(userID int) error {
	_, err := s.db.Exec("UPDATE users SET is_online = false, last_seen_at = NOW() WHERE id = $1", userID)
	return err
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(userID int) (*User, error) {
	user := &User{}
	var lastSeenAt sql.NullTime
	query := `
		SELECT id, username, email, COALESCE(avatar_url, ''), is_online, last_seen_at, created_at
		FROM users
		WHERE id = $1
	`
	err := s.db.QueryRow(query, userID).
		Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL,
			&user.IsOnline, &lastSeenAt, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastSeenAt.Valid {
		user.LastSeenAt = lastSeenAt.Time
	} else {
		user.LastSeenAt = time.Now()
	}

	return user, nil
}

// BlacklistToken adds token to blacklist
func BlacklistToken(token string, expiry time.Time) {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	tokenBlacklist[token] = expiry
}

// IsTokenBlacklisted checks if token is blacklisted
func IsTokenBlacklisted(token string) bool {
	blacklistMutex.RLock()
	defer blacklistMutex.RUnlock()

	expiry, exists := tokenBlacklist[token]
	if !exists {
		return false
	}

	// Clean up expired tokens
	if time.Now().After(expiry) {
		delete(tokenBlacklist, token)
		return false
	}

	return true
}
