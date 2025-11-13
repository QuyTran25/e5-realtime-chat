package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements token bucket algorithm with Redis
type RateLimiter struct {
	redis *redis.Client
	ctx   context.Context
}

// RateLimitConfig defines rate limiting parameters
type RateLimitConfig struct {
	RequestsPerMinute int           // Max requests per minute
	BurstSize         int           // Max burst size
	Window            time.Duration // Time window
}

// Preset rate limit configs
var (
	// Strict limit for sensitive endpoints (login, register)
	StrictLimit = RateLimitConfig{
		RequestsPerMinute: 30, // Increased from 10
		BurstSize:         50, // Increased from 20
		Window:            time.Minute,
	}

	// Normal limit for regular API endpoints
	NormalLimit = RateLimitConfig{
		RequestsPerMinute: 300, // Increased from 60
		BurstSize:         500, // Increased from 100
		Window:            time.Minute,
	}

	// Relaxed limit for read-only endpoints
	RelaxedLimit = RateLimitConfig{
		RequestsPerMinute: 600,  // Increased from 120
		BurstSize:         1000, // Increased from 200
		Window:            time.Minute,
	}

	// WebSocket message rate limit (per user)
	WSMessageLimit = RateLimitConfig{
		RequestsPerMinute: 120, // Increased from 60
		BurstSize:         50,  // Increased from 10
		Window:            time.Minute,
	}
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

// CheckLimit checks if request is allowed using token bucket algorithm
// Returns: allowed (bool), remaining tokens (int), reset time (time.Time), error
func (rl *RateLimiter) CheckLimit(key string, config RateLimitConfig) (bool, int, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(config.Window.Seconds()))

	// Lua script for atomic rate limiting (token bucket)
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local current = redis.call('GET', key)
		
		if current == false then
			redis.call('SET', key, limit - 1, 'EX', window)
			return {1, limit - 1}
		end
		
		current = tonumber(current)
		if current > 0 then
			redis.call('DECR', key)
			return {1, current - 1}
		end
		
		return {0, 0}
	`

	result, err := rl.redis.Eval(rl.ctx, script, []string{windowKey}, config.RequestsPerMinute, int(config.Window.Seconds())).Result()
	if err != nil {
		log.Printf("⚠️ Rate limiter error: %v", err)
		return true, config.RequestsPerMinute, now.Add(config.Window), nil // Fail open
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := int(resultSlice[1].(int64))
	resetTime := now.Add(config.Window)

	return allowed, remaining, resetTime, nil
}

// RateLimitMiddleware creates HTTP middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (user ID from context or IP address)
			key := rl.getIdentifier(r)

			allowed, remaining, resetTime, err := rl.CheckLimit(key, config)
			if err != nil {
				log.Printf("❌ Rate limit check failed: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(config.Window.Seconds()), 10))
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				log.Printf("🚫 Rate limit exceeded for: %s (endpoint: %s)", key, r.URL.Path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIdentifier extracts user identifier from request
func (rl *RateLimiter) getIdentifier(r *http.Request) string {
	// Try to get user ID from context (set by auth middleware)
	if userID := r.Context().Value("userID"); userID != nil {
		return fmt.Sprintf("user:%v", userID)
	}

	// Fallback to IP address
	ip := rl.getClientIP(r)
	return fmt.Sprintf("ip:%s", ip)
}

// getClientIP extracts real client IP from request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (behind proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

// CheckUserMessageRate checks rate limit for user messages (for WebSocket)
func (rl *RateLimiter) CheckUserMessageRate(userID int) (bool, error) {
	key := fmt.Sprintf("ws:message:user:%d", userID)
	allowed, _, _, err := rl.CheckLimit(key, WSMessageLimit)
	if err != nil {
		return true, err // Fail open
	}
	return allowed, nil
}

// BlockIP temporarily blocks an IP address (for DDoS protection)
func (rl *RateLimiter) BlockIP(ip string, duration time.Duration) error {
	key := fmt.Sprintf("blocked:ip:%s", ip)
	return rl.redis.Set(rl.ctx, key, "1", duration).Err()
}

// IsIPBlocked checks if an IP is blocked
func (rl *RateLimiter) IsIPBlocked(ip string) (bool, error) {
	key := fmt.Sprintf("blocked:ip:%s", ip)
	exists, err := rl.redis.Exists(rl.ctx, key).Result()
	return exists > 0, err
}

// GetRateLimitStatus returns current rate limit status for a key
func (rl *RateLimiter) GetRateLimitStatus(key string, config RateLimitConfig) (int, time.Time, error) {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(config.Window.Seconds()))

	remaining, err := rl.redis.Get(rl.ctx, windowKey).Int()
	if err == redis.Nil {
		return config.RequestsPerMinute, now.Add(config.Window), nil
	}
	if err != nil {
		return 0, time.Time{}, err
	}

	ttl, err := rl.redis.TTL(rl.ctx, windowKey).Result()
	if err != nil {
		return 0, time.Time{}, err
	}

	resetTime := now.Add(ttl)
	return remaining, resetTime, nil
}

// ResetLimit resets rate limit for a key (admin function)
func (rl *RateLimiter) ResetLimit(key string, config RateLimitConfig) error {
	now := time.Now()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now.Unix()/int64(config.Window.Seconds()))
	return rl.redis.Del(rl.ctx, windowKey).Err()
}
