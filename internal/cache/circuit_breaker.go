package cache

import (
	"time"

	"news/internal/metrics"
	"news/internal/middleware"
)

var (
	// Define circuit breakers for cache operations
	cacheGetCircuitBreaker    *middleware.CircuitBreaker
	cacheSetCircuitBreaker    *middleware.CircuitBreaker
	cacheDeleteCircuitBreaker *middleware.CircuitBreaker
)

// Initialize circuit breakers
func init() {
	// Setup circuit breakers with configurations for cache operations
	cacheGetCircuitBreaker = middleware.NewCircuitBreaker(
		"cache_get",
		middleware.WithFailureThreshold(3),
		middleware.WithSuccessThreshold(2),
		middleware.WithTimeout(2*time.Second),
	)

	cacheSetCircuitBreaker = middleware.NewCircuitBreaker(
		"cache_set",
		middleware.WithFailureThreshold(3),
		middleware.WithSuccessThreshold(2),
		middleware.WithTimeout(2*time.Second),
	)

	cacheDeleteCircuitBreaker = middleware.NewCircuitBreaker(
		"cache_delete",
		middleware.WithFailureThreshold(3),
		middleware.WithSuccessThreshold(2),
		middleware.WithTimeout(2*time.Second),
	)
}

// SafeGetCachedNews retrieves a cached news article with circuit breaker protection
func SafeGetCachedNews(key string) (string, error) {
	var value string
	err := cacheGetCircuitBreaker.Execute(func() error {
		var err error
		value, err = GetRedisClient().GetCachedNews(key)

		if err == nil {
			metrics.TrackCacheHit(key)
		} else {
			metrics.TrackCacheMiss(key)
		}

		return err
	})
	return value, err
}

// SafeCacheNews caches a news article with circuit breaker protection
func SafeCacheNews(key string, value interface{}, expiration time.Duration) error {
	return cacheSetCircuitBreaker.Execute(func() error {
		// Track the operation duration
		defer metrics.TrackDatabaseOperation("cache_set")()
		return GetRedisClient().CacheNews(key, value, expiration)
	})
}

// SafeRemoveFromCache removes an item from the cache with circuit breaker protection
func SafeRemoveFromCache(key string) error {
	return cacheDeleteCircuitBreaker.Execute(func() error {
		// Track the operation duration
		defer metrics.TrackDatabaseOperation("cache_delete")()
		return GetRedisClient().RemoveFromCache(key)
	})
}

// SafeBlacklistToken adds a token to the blacklist with circuit breaker protection
func SafeBlacklistToken(token string, expiration time.Duration) error {
	return cacheSetCircuitBreaker.Execute(func() error {
		return GetRedisClient().BlacklistToken(token, expiration)
	})
}

// SafeIsTokenBlacklisted checks if a token is blacklisted with circuit breaker protection
func SafeIsTokenBlacklisted(token string) (bool, error) {
	var isBlacklisted bool
	err := cacheGetCircuitBreaker.Execute(func() error {
		isBlacklisted = GetRedisClient().IsTokenBlacklisted(token)
		return nil
	})
	return isBlacklisted, err
}
