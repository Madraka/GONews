package cache

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// mockRedis is a simple in-memory storage for test data
type mockRedis struct {
	data map[string]string
}

var testRedis = &mockRedis{
	data: make(map[string]string),
}

// MockRedisCache sets up a mock Redis client for testing by disabling Redis functionality
// MockRedisCache sets up a mock Redis client for testing by disabling Redis operations
func MockRedisCache() {
	// Set default client to nil to effectively disable Redis operations
	defaultClient = nil
}

// MockCacheNews mocks the CacheNews function for testing
func MockCacheNews(key string, value interface{}, expiration time.Duration) error {
	// For testing purposes, we'll just store the key
	testRedis.data[key] = "test-value"
	return nil
}

// MockGetCachedNews mocks the GetCachedNews function for testing
func MockGetCachedNews(key string) (string, error) {
	if val, exists := testRedis.data[key]; exists {
		return val, nil
	}
	return "", redis.Nil
}

// MockRetryCache mocks the RetryCache function for testing
func MockRetryCache(key string, value interface{}, expiration time.Duration, maxRetries int) error {
	return MockCacheNews(key, value, expiration)
}

// MockRemoveFromCache mocks the RemoveFromCache function for testing
func MockRemoveFromCache(key string) error {
	delete(testRedis.data, key)
	return nil
}

// MockBlacklistToken mocks the BlacklistToken function for testing
func MockBlacklistToken(token string, expiration time.Duration) error {
	testRedis.data["blacklist:"+token] = "revoked"
	return nil
}

// MockIsTokenBlacklisted mocks the IsTokenBlacklisted function for testing
func MockIsTokenBlacklisted(token string) bool {
	_, exists := testRedis.data["blacklist:"+token]
	return exists
}

// ClearMockCache clears all mock cache data for testing
func ClearMockCache() {
	testRedis.data = make(map[string]string)
}
