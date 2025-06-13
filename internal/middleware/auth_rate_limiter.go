package middleware

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthRateLimiter provides specialized rate limiting for authentication endpoints
// with stricter limits and protection against brute force attacks
type AuthRateLimiter struct {
	// IP-based limiting
	ipLimits      map[string]*IPLimit
	ipMutex       sync.RWMutex
	ipMaxAttempts int           // Maximum attempts per time window
	ipWindowSize  time.Duration // Time window size for IP limiting

	// Username-based limiting (prevents guessing multiple passwords for a specific user)
	usernameLimits      map[string]*UsernameLimit
	usernameMutex       sync.RWMutex
	usernameMaxAttempts int           // Maximum attempts per time window
	usernameWindowSize  time.Duration // Time window size for username limiting
	usernameBlockTime   time.Duration // How long to block after too many attempts

	// Global limiting
	globalFailureCount int // Count of recent failures across all IPs
	globalMutex        sync.RWMutex
	globalWindowSize   time.Duration // Time window for global limiting
	globalThreshold    int           // Threshold for enabling stricter rules
}

// IPLimit tracks login attempts from a specific IP
type IPLimit struct {
	Attempts    int
	LastAttempt time.Time
	Blocked     bool
	BlockUntil  time.Time
}

// UsernameLimit tracks login attempts for a specific username
type UsernameLimit struct {
	Attempts    int
	LastAttempt time.Time
	Blocked     bool
	BlockUntil  time.Time
}

// NewAuthRateLimiter creates a new rate limiter for authentication endpoints
func NewAuthRateLimiter() *AuthRateLimiter {
	limiter := &AuthRateLimiter{
		ipLimits:            make(map[string]*IPLimit),
		ipMaxAttempts:       5,                // 5 attempts
		ipWindowSize:        time.Minute * 10, // per 10 minutes
		usernameLimits:      make(map[string]*UsernameLimit),
		usernameMaxAttempts: 3,                // 3 attempts
		usernameWindowSize:  time.Minute * 10, // per 10 minutes
		usernameBlockTime:   time.Minute * 30, // block for 30 minutes
		globalWindowSize:    time.Minute * 5,  // 5 minute window
		globalThreshold:     100,              // 100 failures globally
	}

	// Start cleanup goroutine to prevent memory leaks
	go limiter.cleanupRoutine()

	return limiter
}

// cleanupRoutine periodically removes old entries from the maps
func (arl *AuthRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(time.Hour) // Clean up once per hour
	defer ticker.Stop()

	for {
		<-ticker.C
		arl.cleanup()
	}
}

// cleanup removes expired entries from the rate limiter maps
func (arl *AuthRateLimiter) cleanup() {
	now := time.Now()

	// Clean up IP limits
	arl.ipMutex.Lock()
	for ip, limit := range arl.ipLimits {
		if now.Sub(limit.LastAttempt) > arl.ipWindowSize*2 && !limit.Blocked {
			delete(arl.ipLimits, ip)
		} else if limit.Blocked && now.After(limit.BlockUntil) {
			delete(arl.ipLimits, ip)
		}
	}
	arl.ipMutex.Unlock()

	// Clean up username limits
	arl.usernameMutex.Lock()
	for username, limit := range arl.usernameLimits {
		if now.Sub(limit.LastAttempt) > arl.usernameWindowSize*2 && !limit.Blocked {
			delete(arl.usernameLimits, username)
		} else if limit.Blocked && now.After(limit.BlockUntil) {
			delete(arl.usernameLimits, username)
		}
	}
	arl.usernameMutex.Unlock()

	// Reset global counter if window has passed
	arl.globalMutex.Lock()
	arl.globalFailureCount = 0
	arl.globalMutex.Unlock()
}

// LimitByIP checks if an IP has exceeded the request limit
func (arl *AuthRateLimiter) LimitByIP(ip string) bool {
	now := time.Now()

	arl.ipMutex.Lock()
	defer arl.ipMutex.Unlock()

	limit, exists := arl.ipLimits[ip]
	if !exists {
		arl.ipLimits[ip] = &IPLimit{
			Attempts:    1,
			LastAttempt: now,
		}
		return false
	}

	// Check if blocked
	if limit.Blocked && now.Before(limit.BlockUntil) {
		return true
	} else if limit.Blocked {
		// Unblock if block period has expired
		limit.Blocked = false
		limit.Attempts = 1
		limit.LastAttempt = now
		return false
	}

	// Reset attempts if window has passed
	if now.Sub(limit.LastAttempt) > arl.ipWindowSize {
		limit.Attempts = 1
		limit.LastAttempt = now
		return false
	}

	// Increment attempts
	limit.Attempts++
	limit.LastAttempt = now

	// Block if too many attempts
	if limit.Attempts > arl.ipMaxAttempts {
		limit.Blocked = true
		limit.BlockUntil = now.Add(arl.usernameBlockTime)
		return true
	}

	return false
}

// LimitByUsername checks if a username has exceeded the request limit
func (arl *AuthRateLimiter) LimitByUsername(username string) bool {
	now := time.Now()

	arl.usernameMutex.Lock()
	defer arl.usernameMutex.Unlock()

	limit, exists := arl.usernameLimits[username]
	if !exists {
		arl.usernameLimits[username] = &UsernameLimit{
			Attempts:    1,
			LastAttempt: now,
		}
		return false
	}

	// Check if blocked
	if limit.Blocked && now.Before(limit.BlockUntil) {
		return true
	} else if limit.Blocked {
		// Unblock if block period has expired
		limit.Blocked = false
		limit.Attempts = 1
		limit.LastAttempt = now
		return false
	}

	// Reset attempts if window has passed
	if now.Sub(limit.LastAttempt) > arl.usernameWindowSize {
		limit.Attempts = 1
		limit.LastAttempt = now
		return false
	}

	// Increment attempts
	limit.Attempts++
	limit.LastAttempt = now

	// Block if too many attempts
	if limit.Attempts > arl.usernameMaxAttempts {
		limit.Blocked = true
		limit.BlockUntil = now.Add(arl.usernameBlockTime)
		return true
	}

	return false
}

// RecordFailedAttempt records a failed authentication attempt
func (arl *AuthRateLimiter) RecordFailedAttempt(ip, username string) {
	arl.LimitByIP(ip)
	if username != "" {
		arl.LimitByUsername(username)
	}

	// Increment global failure count
	arl.globalMutex.Lock()
	arl.globalFailureCount++
	arl.globalMutex.Unlock()
}

// ShouldBlock determines if a request should be blocked
func (arl *AuthRateLimiter) ShouldBlock(ip, username string) bool {
	// Check for IP-based blocking
	if arl.LimitByIP(ip) {
		return true
	}

	// Check for username-based blocking
	if username != "" && arl.LimitByUsername(username) {
		return true
	}

	// Check if we're under a potential attack (many global failures)
	arl.globalMutex.RLock()
	potentialAttack := arl.globalFailureCount > arl.globalThreshold
	arl.globalMutex.RUnlock()

	if potentialAttack {
		// Apply stricter rules during potential attacks
		// For example, we could block IPs with fewer attempts
		// or implement additional security measures
		// TODO: Implement enhanced security measures
		log.Printf("Potential attack detected: global failure count %d exceeds threshold %d",
			arl.globalFailureCount, arl.globalThreshold)
	}

	return false
}

// AuthRateLimiterMiddleware is a middleware that adds rate limiting for authentication endpoints
func AuthRateLimiterMiddleware(limiter *AuthRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract IP address
		ip := getClientIP(c)

		// For login routes, extract username to apply username-based limiting
		var username string
		if c.Request.Method == "POST" && (c.FullPath() == "/login" || c.FullPath() == "/register") {
			// Try to extract username from request body
			// This might require request body binding which can be complex in middleware
			// For now, we'll just use IP-based limiting
			// TODO: Implement username extraction from request body for enhanced rate limiting
			log.Printf("Auth endpoint accessed from IP: %s", ip)
		}

		// Check if request should be blocked
		if limiter.ShouldBlock(ip, username) {
			c.JSON(429, gin.H{"error": "Too many authentication attempts. Please try again later."})
			c.Abort()
			return
		}

		// Continue processing
		c.Next()

		// Record failed attempts after request processing
		if c.Writer.Status() == 401 && (c.FullPath() == "/login" || c.FullPath() == "/register") {
			limiter.RecordFailedAttempt(ip, username)
		}
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check for X-Forwarded-For header first (for proxied requests)
	clientIP := c.Request.Header.Get("X-Forwarded-For")
	if clientIP != "" {
		// X-Forwarded-For might contain multiple IPs, take the first one
		ips := net.ParseIP(strings.Split(clientIP, ",")[0])
		if ips != nil {
			return ips.String()
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
