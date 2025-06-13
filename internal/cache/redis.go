package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"news/internal/metrics"

	"github.com/go-redis/redis/v8"
)

// RedisClient provides Redis client functionality with additional security features
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

var (
	// defaultClient is the global Redis client instance
	defaultClient *RedisClient
	// Support test mode to prevent Redis errors in tests
	inTestMode = false
	// In-memory token blacklist for test mode
	mockBlacklist = make(map[string]bool)
)

const (
	// TokenBlacklistPrefix is used for blacklisted tokens
	TokenBlacklistPrefix = "blacklist:token:"
)

// SetTestMode enables or disables test mode for Redis operations
func SetTestMode(enabled bool) {
	inTestMode = enabled
	if enabled {
		// Initialize an empty mock blacklist when enabling test mode
		mockBlacklist = make(map[string]bool)
	}
}

// IsTestMode returns whether the Redis client is in test mode
func IsTestMode() bool {
	return inTestMode
}

// buildRedisURL processes a Redis URL/address and returns the host:port format
func buildRedisURL(redisURL string) string {
	if redisURL == "" {
		// Try building from separate REDIS_HOST and REDIS_PORT
		redisHost := os.Getenv("REDIS_HOST")
		redisPort := os.Getenv("REDIS_PORT")

		if redisHost != "" && redisPort != "" {
			return redisHost + ":" + redisPort
		}
		return "localhost:6379" // Default fallback
	}

	// If URL has redis:// prefix, parse it
	if strings.HasPrefix(redisURL, "redis://") {
		// Extract host:port from redis://host:port/db
		addr := strings.TrimPrefix(redisURL, "redis://")

		// Remove database suffix (e.g., "/0")
		if strings.Contains(addr, "/") {
			addr = strings.Split(addr, "/")[0]
		}
		return addr
	}

	// If it already looks like host:port, return as-is
	if strings.Contains(redisURL, ":") && !strings.Contains(redisURL, "://") {
		return redisURL
	}

	// For other cases, assume it's just a hostname and add default port
	return redisURL + ":6379"
}

func InitRedis() error {
	// In test mode, just create a dummy client - we won't actually connect to Redis
	if inTestMode {
		defaultClient = &RedisClient{
			client: nil,
			ctx:    context.Background(),
		}
		return nil
	}

	redisAddr := os.Getenv("REDIS_URL")

	// Parse Redis URL to extract host:port
	finalAddr := buildRedisURL(redisAddr)

	client := redis.NewClient(&redis.Options{
		Addr:     finalAddr,
		Password: os.Getenv("REDIS_PASSWORD"), // Will be empty string if not set
		DB:       0,
	})

	// Create our enhanced Redis client
	defaultClient = &RedisClient{
		client: client,
		ctx:    context.Background(),
	}

	// Test the connection
	_, err := defaultClient.client.Ping(defaultClient.ctx).Result()
	return err
}

// GetRedisClient returns the singleton Redis client instance
func GetRedisClient() *RedisClient {
	if defaultClient == nil {
		// Auto-initialize if needed but log a warning
		fmt.Println("Warning: Redis client not initialized, auto-initializing")
		err := InitRedis()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize Redis: %v", err))
		}
	}
	return defaultClient
}

// CacheNews caches a news article
func (rc *RedisClient) CacheNews(key string, value interface{}, expiration time.Duration) error {
	// Track the operation duration
	defer metrics.TrackDatabaseOperation("cache_set")()
	return rc.client.Set(rc.ctx, key, value, expiration).Err()
}

// GetCachedNews retrieves a cached news article
func (rc *RedisClient) GetCachedNews(key string) (string, error) {
	val, err := rc.client.Get(rc.ctx, key).Result()
	if err == redis.Nil {
		// Cache miss
		metrics.TrackCacheMiss(key)
	} else if err == nil {
		// Cache hit
		metrics.TrackCacheHit(key)
	}
	return val, err
}

// RemoveFromCache removes an item from the cache
func (rc *RedisClient) RemoveFromCache(key string) error {
	return rc.client.Del(rc.ctx, key).Err()
}

// BlacklistToken adds a token to the blacklist with expiration
func (rc *RedisClient) BlacklistToken(tokenID string, expiration time.Duration) error {
	key := TokenBlacklistPrefix + tokenID

	if inTestMode {
		// Use in-memory map in test mode
		mockBlacklist[key] = true
		return nil
	}

	return rc.client.Set(rc.ctx, key, "revoked", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (rc *RedisClient) IsTokenBlacklisted(tokenID string) bool {
	key := TokenBlacklistPrefix + tokenID

	if inTestMode {
		// Use in-memory map in test mode
		_, exists := mockBlacklist[key]
		return exists
	}

	_, err := rc.client.Get(rc.ctx, key).Result()
	return err == nil // If no error, the token exists in blacklist
}

// IsConnectionError checks if the error is a Redis connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check different types of Redis connection errors
	return redis.Nil != err && (err == redis.ErrClosed ||
		err.Error() == "redis: client is closed" ||
		err.Error() == "redis: connection pool timeout" ||
		err.Error() == "redis: connection closed")
}

// RetryCache attempts to cache data with retries in case of connection issues
func (rc *RedisClient) RetryCache(key string, value interface{}, expiration time.Duration, maxRetries int) error {
	var err error

	for i := 0; i < maxRetries; i++ {
		err = rc.CacheNews(key, value, expiration)
		if err == nil || !IsConnectionError(err) {
			return err // Success or non-connection error
		}

		// Wait a bit before retrying (exponential backoff)
		time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
	}

	return err
}

// Ping checks the Redis connection
func (rc *RedisClient) Ping() error {
	if inTestMode {
		return nil // Always return success in test mode
	}
	_, err := rc.client.Ping(rc.ctx).Result()
	return err
}

// GetClient returns the underlying redis.Client instance
func (rc *RedisClient) GetClient() *redis.Client {
	if rc == nil {
		return nil
	}
	return rc.client
}

// ClearAllCache clears all keys from Redis (use with caution)
func (rc *RedisClient) ClearAllCache() error {
	if inTestMode {
		// Clear mock data in test mode
		testRedis.data = make(map[string]string)
		mockBlacklist = make(map[string]bool)
		return nil
	}

	if rc.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Use FLUSHDB to clear current database
	return rc.client.FlushDB(rc.ctx).Err()
}

// Legacy function wrappers to maintain backward compatibility
func CacheNews(key string, value interface{}, expiration time.Duration) error {
	return GetRedisClient().CacheNews(key, value, expiration)
}

func GetCachedNews(key string) (string, error) {
	return GetRedisClient().GetCachedNews(key)
}

func RemoveFromCache(key string) error {
	return GetRedisClient().RemoveFromCache(key)
}

func BlacklistToken(token string, expiration time.Duration) error {
	return GetRedisClient().BlacklistToken(token, expiration)
}

func IsTokenBlacklisted(token string) bool {
	return GetRedisClient().IsTokenBlacklisted(token)
}

func RetryCache(key string, value interface{}, expiration time.Duration, maxRetries int) error {
	return GetRedisClient().RetryCache(key, value, expiration, maxRetries)
}

// CloseRedis gracefully closes the Redis connection
func CloseRedis() error {
	if defaultClient == nil {
		return nil // Already closed or never initialized
	}

	if inTestMode {
		// In test mode, just clear the mock blacklist and reset client
		mockBlacklist = make(map[string]bool)
		defaultClient = nil
		return nil
	}

	if defaultClient.client != nil {
		err := defaultClient.client.Close()
		defaultClient = nil
		return err
	}

	return nil
}
