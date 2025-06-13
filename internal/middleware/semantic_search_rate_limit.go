package middleware

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// SemanticSearchLimiter provides rate limiting specifically for semantic search
// with OpenAI API cost control and fallback to local search
type SemanticSearchLimiter struct {
	// Per-user limits (authenticated users)
	userLimits      map[uint]*UserSearchLimit
	userMutex       sync.RWMutex
	userMaxRequests int           // Max AI requests per user per day
	userWindowSize  time.Duration // Time window (24 hours)

	// Per-IP limits (unauthenticated users)
	ipLimits      map[string]*IPSearchLimit
	ipMutex       sync.RWMutex
	ipMaxRequests int           // Max AI requests per IP per day
	ipWindowSize  time.Duration // Time window (24 hours)

	// Global limits for cost control
	globalRequests    int
	globalMaxRequests int // Global daily limit for AI requests
	globalMutex       sync.RWMutex
	globalWindowSize  time.Duration // Time window (24 hours)
	globalResetTime   time.Time

	redisClient *redis.Client
}

// UserSearchLimit tracks search requests for authenticated users
type UserSearchLimit struct {
	UserID        uint
	AIRequests    int // AI-powered searches (uses OpenAI)
	LocalRequests int // Local searches (ElasticSearch only)
	LastRequest   time.Time
	WindowStart   time.Time
	Blocked       bool
	BlockUntil    time.Time
}

// IPSearchLimit tracks search requests for unauthenticated users (by IP)
type IPSearchLimit struct {
	IP            string
	AIRequests    int // AI-powered searches
	LocalRequests int // Local searches
	LastRequest   time.Time
	WindowStart   time.Time
	Blocked       bool
	BlockUntil    time.Time
}

// SearchLimitConfig holds configuration for search rate limiting
type SearchLimitConfig struct {
	// Authenticated user limits
	UserAIRequestsPerDay    int // Default: 50 AI requests per day per user
	UserLocalRequestsPerDay int // Default: 500 local requests per day per user

	// Unauthenticated IP limits
	IPAIRequestsPerDay    int // Default: 5 AI requests per day per IP
	IPLocalRequestsPerDay int // Default: 50 local requests per day per IP

	// Global limits
	GlobalAIRequestsPerDay int // Default: 10000 AI requests per day globally

	// Block durations
	UserBlockDuration time.Duration // Default: 1 hour
	IPBlockDuration   time.Duration // Default: 6 hours
}

// DefaultSearchLimitConfig returns default configuration
func DefaultSearchLimitConfig() *SearchLimitConfig {
	return &SearchLimitConfig{
		UserAIRequestsPerDay:    50,            // Generous for authenticated users
		UserLocalRequestsPerDay: 500,           // Very generous for local search
		IPAIRequestsPerDay:      5,             // Limited for unauthenticated
		IPLocalRequestsPerDay:   50,            // Moderate for unauthenticated
		GlobalAIRequestsPerDay:  10000,         // Cost control
		UserBlockDuration:       time.Hour,     // 1 hour block
		IPBlockDuration:         6 * time.Hour, // 6 hours block for IPs
	}
}

// NewSemanticSearchLimiter creates a new semantic search rate limiter
func NewSemanticSearchLimiter(config *SearchLimitConfig, redisClient *redis.Client) *SemanticSearchLimiter {
	if config == nil {
		config = DefaultSearchLimitConfig()
	}

	limiter := &SemanticSearchLimiter{
		userLimits:        make(map[uint]*UserSearchLimit),
		userMaxRequests:   config.UserAIRequestsPerDay,
		userWindowSize:    24 * time.Hour,
		ipLimits:          make(map[string]*IPSearchLimit),
		ipMaxRequests:     config.IPAIRequestsPerDay,
		ipWindowSize:      24 * time.Hour,
		globalMaxRequests: config.GlobalAIRequestsPerDay,
		globalWindowSize:  24 * time.Hour,
		globalResetTime:   time.Now().Add(24 * time.Hour),
		redisClient:       redisClient,
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// CanUseAI checks if a user/IP can make AI-powered search requests
func (ssl *SemanticSearchLimiter) CanUseAI(userID *uint, ip string) (bool, string) {
	now := time.Now()

	// Check global limit first
	ssl.globalMutex.Lock()
	if now.After(ssl.globalResetTime) {
		ssl.globalRequests = 0
		ssl.globalResetTime = now.Add(ssl.globalWindowSize)
	}

	if ssl.globalRequests >= ssl.globalMaxRequests {
		ssl.globalMutex.Unlock()
		return false, "Global AI search limit exceeded. Using local search."
	}
	ssl.globalMutex.Unlock()

	if userID != nil {
		// Authenticated user
		return ssl.canUserUseAI(*userID, now)
	} else {
		// Unauthenticated IP
		return ssl.canIPUseAI(ip, now)
	}
}

// canUserUseAI checks if an authenticated user can use AI search
func (ssl *SemanticSearchLimiter) canUserUseAI(userID uint, now time.Time) (bool, string) {
	ssl.userMutex.Lock()
	defer ssl.userMutex.Unlock()

	limit, exists := ssl.userLimits[userID]
	if !exists {
		ssl.userLimits[userID] = &UserSearchLimit{
			UserID:      userID,
			AIRequests:  0,
			WindowStart: now,
			LastRequest: now,
		}
		return true, ""
	}

	// Reset window if 24 hours passed
	if now.Sub(limit.WindowStart) >= ssl.userWindowSize {
		limit.AIRequests = 0
		limit.LocalRequests = 0
		limit.WindowStart = now
		limit.Blocked = false
	}

	// Check if blocked
	if limit.Blocked && now.Before(limit.BlockUntil) {
		return false, fmt.Sprintf("User temporarily blocked until %s", limit.BlockUntil.Format("15:04"))
	}

	// Check AI request limit
	if limit.AIRequests >= ssl.userMaxRequests {
		return false, fmt.Sprintf("Daily AI search limit exceeded (%d/%d). Using local search.", limit.AIRequests, ssl.userMaxRequests)
	}

	return true, ""
}

// canIPUseAI checks if an IP can use AI search
func (ssl *SemanticSearchLimiter) canIPUseAI(ip string, now time.Time) (bool, string) {
	ssl.ipMutex.Lock()
	defer ssl.ipMutex.Unlock()

	limit, exists := ssl.ipLimits[ip]
	if !exists {
		ssl.ipLimits[ip] = &IPSearchLimit{
			IP:          ip,
			AIRequests:  0,
			WindowStart: now,
			LastRequest: now,
		}
		return true, ""
	}

	// Reset window if 24 hours passed
	if now.Sub(limit.WindowStart) >= ssl.ipWindowSize {
		limit.AIRequests = 0
		limit.LocalRequests = 0
		limit.WindowStart = now
		limit.Blocked = false
	}

	// Check if blocked
	if limit.Blocked && now.Before(limit.BlockUntil) {
		return false, fmt.Sprintf("IP temporarily blocked until %s", limit.BlockUntil.Format("15:04"))
	}

	// Check AI request limit
	if limit.AIRequests >= ssl.ipMaxRequests {
		return false, fmt.Sprintf("Daily AI search limit exceeded for unauthenticated users (%d/%d). Using local search.", limit.AIRequests, ssl.ipMaxRequests)
	}

	return true, ""
}

// RecordAIRequest records an AI search request
func (ssl *SemanticSearchLimiter) RecordAIRequest(userID *uint, ip string) {
	now := time.Now()

	// Record global request
	ssl.globalMutex.Lock()
	ssl.globalRequests++
	ssl.globalMutex.Unlock()

	if userID != nil {
		ssl.recordUserAIRequest(*userID, now)
	} else {
		ssl.recordIPAIRequest(ip, now)
	}
}

// recordUserAIRequest records AI request for authenticated user
func (ssl *SemanticSearchLimiter) recordUserAIRequest(userID uint, now time.Time) {
	ssl.userMutex.Lock()
	defer ssl.userMutex.Unlock()

	if limit, exists := ssl.userLimits[userID]; exists {
		limit.AIRequests++
		limit.LastRequest = now
	}
}

// recordIPAIRequest records AI request for IP
func (ssl *SemanticSearchLimiter) recordIPAIRequest(ip string, now time.Time) {
	ssl.ipMutex.Lock()
	defer ssl.ipMutex.Unlock()

	if limit, exists := ssl.ipLimits[ip]; exists {
		limit.AIRequests++
		limit.LastRequest = now
	}
}

// RecordLocalRequest records a local search request
func (ssl *SemanticSearchLimiter) RecordLocalRequest(userID *uint, ip string) {
	now := time.Now()

	if userID != nil {
		ssl.recordUserLocalRequest(*userID, now)
	} else {
		ssl.recordIPLocalRequest(ip, now)
	}
}

func (ssl *SemanticSearchLimiter) recordUserLocalRequest(userID uint, now time.Time) {
	ssl.userMutex.Lock()
	defer ssl.userMutex.Unlock()

	if limit, exists := ssl.userLimits[userID]; exists {
		limit.LocalRequests++
		limit.LastRequest = now
	}
}

func (ssl *SemanticSearchLimiter) recordIPLocalRequest(ip string, now time.Time) {
	ssl.ipMutex.Lock()
	defer ssl.ipMutex.Unlock()

	if limit, exists := ssl.ipLimits[ip]; exists {
		limit.LocalRequests++
		limit.LastRequest = now
	}
}

// GetLimitStatus returns current limit status for user/IP
func (ssl *SemanticSearchLimiter) GetLimitStatus(userID *uint, ip string) map[string]interface{} {
	now := time.Now()

	if userID != nil {
		ssl.userMutex.RLock()
		limit, exists := ssl.userLimits[*userID]
		ssl.userMutex.RUnlock()

		if !exists {
			return map[string]interface{}{
				"ai_requests_remaining": ssl.userMaxRequests,
				"ai_requests_used":      0,
				"local_requests_used":   0,
				"window_reset_time":     now.Add(ssl.userWindowSize),
				"user_type":             "authenticated",
			}
		}

		return map[string]interface{}{
			"ai_requests_remaining": ssl.userMaxRequests - limit.AIRequests,
			"ai_requests_used":      limit.AIRequests,
			"local_requests_used":   limit.LocalRequests,
			"window_reset_time":     limit.WindowStart.Add(ssl.userWindowSize),
			"user_type":             "authenticated",
		}
	} else {
		ssl.ipMutex.RLock()
		limit, exists := ssl.ipLimits[ip]
		ssl.ipMutex.RUnlock()

		if !exists {
			return map[string]interface{}{
				"ai_requests_remaining": ssl.ipMaxRequests,
				"ai_requests_used":      0,
				"local_requests_used":   0,
				"window_reset_time":     now.Add(ssl.ipWindowSize),
				"user_type":             "unauthenticated",
			}
		}

		return map[string]interface{}{
			"ai_requests_remaining": ssl.ipMaxRequests - limit.AIRequests,
			"ai_requests_used":      limit.AIRequests,
			"local_requests_used":   limit.LocalRequests,
			"window_reset_time":     limit.WindowStart.Add(ssl.ipWindowSize),
			"user_type":             "unauthenticated",
		}
	}
}

// cleanupRoutine periodically cleans up old entries
func (ssl *SemanticSearchLimiter) cleanupRoutine() {
	ticker := time.NewTicker(6 * time.Hour) // Clean up every 6 hours
	defer ticker.Stop()

	for {
		<-ticker.C
		ssl.cleanup()
	}
}

// cleanup removes old entries
func (ssl *SemanticSearchLimiter) cleanup() {
	now := time.Now()

	// Clean up user limits
	ssl.userMutex.Lock()
	for userID, limit := range ssl.userLimits {
		if now.Sub(limit.LastRequest) > 48*time.Hour { // Keep for 48 hours
			delete(ssl.userLimits, userID)
		}
	}
	ssl.userMutex.Unlock()

	// Clean up IP limits
	ssl.ipMutex.Lock()
	for ip, limit := range ssl.ipLimits {
		if now.Sub(limit.LastRequest) > 48*time.Hour { // Keep for 48 hours
			delete(ssl.ipLimits, ip)
		}
	}
	ssl.ipMutex.Unlock()
}

// SemanticSearchRateLimit middleware for semantic search endpoint
func SemanticSearchRateLimit(limiter *SemanticSearchLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user info
		userIDInterface, userExists := c.Get("userID")
		var userID *uint
		if userExists {
			if uid, ok := userIDInterface.(uint); ok {
				userID = &uid
			}
		}

		// Get IP
		ip := c.ClientIP()

		// Check if AI search is allowed
		canUseAI, reason := limiter.CanUseAI(userID, ip)

		// Set headers for rate limit information
		status := limiter.GetLimitStatus(userID, ip)
		c.Header("X-RateLimit-AI-Remaining", strconv.Itoa(status["ai_requests_remaining"].(int)))
		c.Header("X-RateLimit-AI-Used", strconv.Itoa(status["ai_requests_used"].(int)))
		c.Header("X-RateLimit-Local-Used", strconv.Itoa(status["local_requests_used"].(int)))
		c.Header("X-RateLimit-Reset", status["window_reset_time"].(time.Time).Format(time.RFC3339))
		c.Header("X-RateLimit-User-Type", status["user_type"].(string))

		// Set context for handler to know whether to use AI or local search
		c.Set("use_ai_search", canUseAI)
		c.Set("rate_limit_reason", reason)
		c.Set("search_limiter", limiter)
		c.Set("search_user_id", userID)
		c.Set("search_ip", ip)

		c.Next()
	}
}
