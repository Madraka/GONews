package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"news/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
)

// IsRateLimitDisabled checks if rate limiting is disabled via environment variables
func IsRateLimitDisabled() bool {
	disabled := os.Getenv("DISABLE_RATE_LIMITS")
	enabled := os.Getenv("RATE_LIMIT_ENABLED")

	// If DISABLE_RATE_LIMITS is set to true, disable rate limiting
	if disabled == "true" || disabled == "1" {
		return true
	}

	// If RATE_LIMIT_ENABLED is set to false, disable rate limiting
	if enabled == "false" || enabled == "0" {
		return true
	}

	// For development environment, check if it's explicitly disabled
	if os.Getenv("ENVIRONMENT") == "development" && enabled != "true" {
		return true
	}

	return false
}

// RateLimiterStore defines an interface for rate limiter storage backends
type RateLimiterStore interface {
	Allow(key string) (bool, error)
	IsDistributed() bool
}

// MemoryRateLimiter implements in-memory rate limiting
type MemoryRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewMemoryRateLimiter(r rate.Limit, b int) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *MemoryRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

func (i *MemoryRateLimiter) Allow(key string) (bool, error) {
	return i.getLimiter(key).Allow(), nil
}

func (i *MemoryRateLimiter) IsDistributed() bool {
	return false
}

// RedisRateLimiter implements Redis-based distributed rate limiting
type RedisRateLimiter struct {
	client     *redis.Client
	window     time.Duration
	maxRequest int
}

func NewRedisRateLimiter(client *redis.Client, window time.Duration, maxRequest int) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:     client,
		window:     window,
		maxRequest: maxRequest,
	}
}

// Allow implements the sliding window rate limiting algorithm using Redis
func (r *RedisRateLimiter) Allow(key string) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixNano() / int64(time.Millisecond)
	windowStart := now - int64(r.window/time.Millisecond)

	// Use Redis transaction to ensure atomic operations
	pipe := r.client.TxPipeline()

	// Remove requests older than the current window
	pipe.ZRemRangeByScore(ctx, "rate:"+key, "0", strconv.FormatInt(windowStart, 10))

	// Count requests in the current window
	pipe.ZCard(ctx, "rate:"+key)

	// Add the current request timestamp
	pipe.ZAdd(ctx, "rate:"+key, &redis.Z{Score: float64(now), Member: now})

	// Set the key expiration
	pipe.Expire(ctx, "rate:"+key, r.window*2)

	// Execute the pipeline
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Get the number of requests in the current window
	count := cmds[1].(*redis.IntCmd).Val()

	// Check if the rate limit has been exceeded
	return count < int64(r.maxRequest), nil // Corrected condition
}

func (r *RedisRateLimiter) IsDistributed() bool {
	return true
}

// RateLimit middleware with configurable storage
func RateLimit(requestsPerSecond float64, burst int, pathSpecific bool) gin.HandlerFunc {
	var limiterStore RateLimiterStore

	// Check if Redis is available for distributed rate limiting
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		// Parse Redis URL if it contains protocol
		redisAddr := redisURL
		if strings.HasPrefix(redisURL, "redis://") {
			// Extract host:port from redis://host:port/db
			redisAddr = strings.TrimPrefix(redisURL, "redis://")
			if strings.Contains(redisAddr, "/") {
				// Remove database suffix (e.g., "/0")
				redisAddr = strings.Split(redisAddr, "/")[0]
			}
		}

		// Use Redis for rate limiting in distributed environments
		client := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: os.Getenv("REDIS_PASSWORD"),
		})

		// Test Redis connection
		if err := client.Ping(context.Background()).Err(); err == nil {
			window := time.Second
			// Convert rate from requests per second to total in the window
			maxRequests := int(requestsPerSecond)
			limiterStore = NewRedisRateLimiter(client, window, maxRequests)
			log.Println("Using Redis-based distributed rate limiting")
		} else {
			log.Printf("Redis connection failed: %v. Falling back to in-memory rate limiting", err)
			limiterStore = NewMemoryRateLimiter(rate.Limit(requestsPerSecond), burst)
		}
	} else {
		// Use in-memory rate limiting
		log.Println("Using in-memory rate limiting")
		limiterStore = NewMemoryRateLimiter(rate.Limit(requestsPerSecond), burst)
	}

	return func(c *gin.Context) {
		// Skip rate limiting in test mode or if disabled
		if IsTestMode() || IsRateLimitDisabled() {
			c.Next()
			return
		}

		// Build rate limit key - either just the IP or IP+path
		ip := c.ClientIP()
		var key string

		if pathSpecific {
			// Include path in the key for more granular control
			key = fmt.Sprintf("%s:%s", ip, c.Request.URL.Path)
		} else {
			// Just use the IP for global limiting
			key = ip
		}

		// Apply rate limiting
		allowed, err := limiterStore.Allow(key)
		if err != nil {
			log.Printf("Rate limiting error: %v", err)
			// Allow the request in case of errors
			c.Next()
			return
		}

		if !allowed {
			retryAfter := 1 // Default 1 second

			// Set headers according to standards
			c.Header("RateLimit-Limit", strconv.Itoa(burst))
			c.Header("RateLimit-Remaining", "0")
			c.Header("RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(retryAfter)*time.Second).Unix(), 10))
			c.Header("Retry-After", strconv.Itoa(retryAfter))

			c.JSON(429, gin.H{
				"error":       "Too many requests",
				"retry_after": retryAfter,
				"message":     "Rate limit exceeded. Please try again later.",
			})

			// Log the rate limiting event
			log.Printf("Rate limit exceeded for %s on %s", ip, c.Request.URL.Path)

			// Track rate limit exceeded in metrics
			metrics.TrackRateLimitExceeded(c.Request.URL.Path, ip)

			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyRateLimit provides rate limiting based on API keys with different tiers
func APIKeyRateLimit() gin.HandlerFunc {
	// Implementation of API key-based rate limiting can be added here
	// This would allow different rate limits for different API key tiers

	return func(c *gin.Context) {
		// For now, just pass through
		c.Next()
	}
}
