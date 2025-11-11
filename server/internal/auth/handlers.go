package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const tokenDuration = 24 * time.Hour

// Handler wraps AuthService with HTTP handlers
type Handler struct {
	service *AuthService
}

// NewHandler creates a new auth handler
func NewHandler(service *AuthService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all auth routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/register", h.handleRegister)
	mux.HandleFunc("/api/auth/login", h.handleLogin)
	mux.HandleFunc("/api/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/auth/me", h.handleMe)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	log.Printf("📝 Registration attempt: username=%s, email=%s", req.Username, req.Email)

	user, err := h.service.RegisterUser(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrUserExists {
			status = http.StatusConflict
		}
		log.Printf("❌ Registration failed: %v", err)
		writeJSON(w, status, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	token, err := GenerateToken(user, tokenDuration)
	if err != nil {
		log.Printf("❌ Token generation failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "failed to generate token",
		})
		return
	}

	log.Printf("✅ Registration successful: userID=%d, username=%s", user.ID, user.Username)

	writeJSON(w, http.StatusCreated, AuthResponse{
		Success: true,
		Message: "registration successful",
		Token:   token,
		User:    user,
	})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "invalid request body",
		})
		return
	}

	log.Printf("🔐 Login attempt: email=%s", req.Email)

	user, err := h.service.LoginUser(req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInvalidCreds {
			status = http.StatusUnauthorized
		}
		log.Printf("❌ Login failed: %v", err)
		writeJSON(w, status, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	token, err := GenerateToken(user, tokenDuration)
	if err != nil {
		log.Printf("❌ Token generation failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "failed to generate token",
		})
		return
	}

	log.Printf("✅ Login successful: userID=%d, username=%s", user.ID, user.Username)

	writeJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "login successful",
		Token:   token,
		User:    user,
	})
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "missing authorization header",
		})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "invalid authorization header format",
		})
		return
	}

	tokenString := parts[1]
	claims, err := ValidateToken(tokenString)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "invalid token",
		})
		return
	}

	log.Printf("🚪 Logout: userID=%d, username=%s", claims.UserID, claims.Username)

	// Blacklist the token
	BlacklistToken(tokenString, claims.ExpiresAt.Time)

	// Update user online status
	if err := h.service.LogoutUser(claims.UserID); err != nil {
		log.Printf("⚠️ Failed to update online status: %v", err)
	}

	log.Printf("✅ Logout successful: userID=%d", claims.UserID)

	writeJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "logout successful",
	})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, AuthResponse{
			Success: false,
			Message: "method not allowed",
		})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "missing authorization header",
		})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "invalid authorization header format",
		})
		return
	}

	tokenString := parts[1]

	// Check if token is blacklisted
	if IsTokenBlacklisted(tokenString) {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "token has been revoked",
		})
		return
	}

	claims, err := ValidateToken(tokenString)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "invalid token",
		})
		return
	}

	user, err := h.service.GetUserByID(claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "user not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		User:    user,
	})
}
