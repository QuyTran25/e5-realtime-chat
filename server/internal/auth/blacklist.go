package auth

import (
	"e5realtimechat/internal/cache"
	"log"
	"sync"
	"time"
)

// TokenBlacklist manages blacklisted tokens
type TokenBlacklist struct {
	cacheService *cache.CacheService
	// Fallback in-memory map if Redis is unavailable
	inMemory map[string]time.Time
	mu       sync.RWMutex
	useRedis bool
}

var (
	blacklist     *TokenBlacklist
	blacklistOnce sync.Once
)

// InitTokenBlacklist initializes the global token blacklist
func InitTokenBlacklist(cacheService *cache.CacheService) {
	blacklistOnce.Do(func() {
		useRedis := cacheService != nil
		blacklist = &TokenBlacklist{
			cacheService: cacheService,
			inMemory:     make(map[string]time.Time),
			useRedis:     useRedis,
		}

		if useRedis {
			log.Println("✅ Token blacklist using Redis cache")
		} else {
			log.Println("⚠️  Token blacklist using in-memory fallback")
		}

		// Start cleanup goroutine for in-memory fallback
		if !useRedis {
			go blacklist.cleanupExpired()
		}
	})
}

// GetTokenBlacklist returns the global blacklist instance
func GetTokenBlacklist() *TokenBlacklist {
	if blacklist == nil {
		// Initialize with nil cache (in-memory fallback)
		InitTokenBlacklist(nil)
	}
	return blacklist
}

// BlacklistToken adds a token to the blacklist
func (tb *TokenBlacklist) BlacklistToken(token string, expiration time.Duration) error {
	if tb.useRedis && tb.cacheService != nil {
		// Use Redis
		err := tb.cacheService.BlacklistToken(token, expiration)
		if err != nil {
			log.Printf("⚠️  Redis blacklist failed, using in-memory: %v", err)
			// Fallback to in-memory
			tb.mu.Lock()
			tb.inMemory[token] = time.Now().Add(expiration)
			tb.mu.Unlock()
			return nil
		}
		return nil
	}

	// Use in-memory
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.inMemory[token] = time.Now().Add(expiration)
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (tb *TokenBlacklist) IsTokenBlacklisted(token string) bool {
	if tb.useRedis && tb.cacheService != nil {
		// Check Redis first
		blacklisted, err := tb.cacheService.IsTokenBlacklisted(token)
		if err == nil {
			return blacklisted
		}
		// If Redis fails, check in-memory fallback
		log.Printf("⚠️  Redis check failed, using in-memory: %v", err)
	}

	// Check in-memory
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	expiry, exists := tb.inMemory[token]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(expiry) {
		// Expired, remove it (do this in a separate goroutine to avoid blocking)
		go func() {
			tb.mu.Lock()
			delete(tb.inMemory, token)
			tb.mu.Unlock()
		}()
		return false
	}

	return true
}

// cleanupExpired periodically removes expired tokens from in-memory storage
func (tb *TokenBlacklist) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tb.mu.Lock()
		now := time.Now()
		for token, expiry := range tb.inMemory {
			if now.After(expiry) {
				delete(tb.inMemory, token)
			}
		}
		tb.mu.Unlock()
	}
}

// Legacy functions for backwards compatibility
var (
	// Old in-memory blacklist (deprecated, use TokenBlacklist instead)
	tokenBlacklist = make(map[string]time.Time)
	blacklistMu    sync.RWMutex
)

// BlacklistTokenLegacy adds a token to the legacy in-memory blacklist (deprecated)
func BlacklistTokenLegacy(token string, expiration time.Duration) {
	blacklistMu.Lock()
	defer blacklistMu.Unlock()
	tokenBlacklist[token] = time.Now().Add(expiration)
}

// IsTokenBlacklistedLegacy checks the legacy blacklist (deprecated)
func IsTokenBlacklistedLegacy(token string) bool {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	expiry, exists := tokenBlacklist[token]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// Wrapper functions that use the global blacklist instance
func BlacklistToken(token string, expiration time.Duration) error {
	return GetTokenBlacklist().BlacklistToken(token, expiration)
}

func IsTokenBlacklisted(token string) bool {
	return GetTokenBlacklist().IsTokenBlacklisted(token)
}
