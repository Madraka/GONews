package cache

import (
	"fmt"

	"news/internal/config"
	"news/internal/metrics"
	"news/internal/middleware"
)

var (
	// Optimized circuit breakers for enhanced cache operations
	optimizedCacheGetCircuitBreaker    *middleware.CircuitBreaker
	optimizedCacheSetCircuitBreaker    *middleware.CircuitBreaker
	optimizedCacheDeleteCircuitBreaker *middleware.CircuitBreaker
)

// Initialize optimized circuit breakers with configurations from environment
func init() {
	// Get cache configuration from environment
	cacheConfig := config.GetCacheConfig()

	// Enhanced circuit breaker for cache GET operations
	optimizedCacheGetCircuitBreaker = middleware.NewCircuitBreaker(
		"optimized_cache_get",
		middleware.WithFailureThreshold(cacheConfig.CircuitBreakerFailureThreshold),
		middleware.WithSuccessThreshold(cacheConfig.CircuitBreakerSuccessThreshold),
		middleware.WithTimeout(cacheConfig.CircuitBreakerTimeout),
	)

	// Enhanced circuit breaker for cache SET operations
	optimizedCacheSetCircuitBreaker = middleware.NewCircuitBreaker(
		"optimized_cache_set",
		middleware.WithFailureThreshold(cacheConfig.CircuitBreakerFailureThreshold),
		middleware.WithSuccessThreshold(cacheConfig.CircuitBreakerSuccessThreshold),
		middleware.WithTimeout(cacheConfig.CircuitBreakerTimeout),
	)

	// Enhanced circuit breaker for cache DELETE operations
	optimizedCacheDeleteCircuitBreaker = middleware.NewCircuitBreaker(
		"optimized_cache_delete",
		middleware.WithFailureThreshold(cacheConfig.CircuitBreakerFailureThreshold),
		middleware.WithSuccessThreshold(cacheConfig.CircuitBreakerSuccessThreshold),
		middleware.WithTimeout(cacheConfig.CircuitBreakerTimeout),
	)
}

// SafeOptimizedGet retrieves cached data with enhanced circuit breaker protection
func SafeOptimizedGet(key string) (interface{}, error) {
	var value interface{}
	var found bool

	err := optimizedCacheGetCircuitBreaker.Execute(func() error {
		cache := GetOptimizedUnifiedCache()
		val, f := cache.SmartGet(key)
		value = val
		found = f

		if found {
			metrics.TrackCacheHit(key)
			return nil
		}

		metrics.TrackCacheMiss(key)
		return fmt.Errorf("cache miss for key: %s", key)
	})

	if err != nil {
		// Circuit breaker is open or operation failed
		if optimizedCacheGetCircuitBreaker.GetState() == middleware.StateOpen {
			metrics.IncrementCounter("cache_circuit_breaker_open")
			// Return graceful degradation - could implement fallback here
			return nil, fmt.Errorf("cache circuit breaker open: %w", err)
		}
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("key not found")
	}

	return value, nil
}

// SafeOptimizedSet stores cached data with enhanced circuit breaker protection
func SafeOptimizedSet(key string, value interface{}, options ...CacheSetOption) error {
	return optimizedCacheSetCircuitBreaker.Execute(func() error {
		cache := GetOptimizedUnifiedCache()
		err := cache.SmartSet(key, value, options...)

		if err != nil {
			metrics.IncrementCounter("cache_set_errors")
			return fmt.Errorf("failed to set cache key %s: %w", key, err)
		}

		metrics.TrackCacheSet(key)
		return nil
	})
}

// SafeOptimizedDelete removes cached data with enhanced circuit breaker protection
func SafeOptimizedDelete(key string) error {
	return optimizedCacheDeleteCircuitBreaker.Execute(func() error {
		cache := GetOptimizedUnifiedCache()
		err := cache.SmartDelete(key)

		if err != nil {
			metrics.IncrementCounter("cache_delete_errors")
			return fmt.Errorf("failed to delete cache key %s: %w", key, err)
		}

		metrics.TrackCacheDelete(key)
		return nil
	})
}

// SafeOptimizedBulkDelete removes multiple cached keys with circuit breaker protection
func SafeOptimizedBulkDelete(keys []string) error {
	return optimizedCacheDeleteCircuitBreaker.Execute(func() error {
		cache := GetOptimizedUnifiedCache()
		err := cache.SmartBulkDelete(keys)

		if err != nil {
			metrics.IncrementCounter("cache_bulk_delete_errors")
			return fmt.Errorf("failed to bulk delete cache keys: %w", err)
		}

		for _, key := range keys {
			metrics.TrackCacheDelete(key)
		}
		return nil
	})
}

// GetOptimizedCircuitBreakerStatus returns status of all optimized circuit breakers
func GetOptimizedCircuitBreakerStatus() map[string]interface{} {
	return map[string]interface{}{
		"cache_get": map[string]interface{}{
			"state":                 optimizedCacheGetCircuitBreaker.GetState(),
			"consecutive_failures":  optimizedCacheGetCircuitBreaker.GetConsecutiveFailures(),
			"consecutive_successes": optimizedCacheGetCircuitBreaker.GetConsecutiveSuccesses(),
			"last_failure_time":     optimizedCacheGetCircuitBreaker.GetLastFailureTime(),
		},
		"cache_set": map[string]interface{}{
			"state":                 optimizedCacheSetCircuitBreaker.GetState(),
			"consecutive_failures":  optimizedCacheSetCircuitBreaker.GetConsecutiveFailures(),
			"consecutive_successes": optimizedCacheSetCircuitBreaker.GetConsecutiveSuccesses(),
			"last_failure_time":     optimizedCacheSetCircuitBreaker.GetLastFailureTime(),
		},
		"cache_delete": map[string]interface{}{
			"state":                 optimizedCacheDeleteCircuitBreaker.GetState(),
			"consecutive_failures":  optimizedCacheDeleteCircuitBreaker.GetConsecutiveFailures(),
			"consecutive_successes": optimizedCacheDeleteCircuitBreaker.GetConsecutiveSuccesses(),
			"last_failure_time":     optimizedCacheDeleteCircuitBreaker.GetLastFailureTime(),
		},
	}
}

// ResetOptimizedCircuitBreakers manually resets all circuit breakers (for admin use)
func ResetOptimizedCircuitBreakers() {
	optimizedCacheGetCircuitBreaker.Reset()
	optimizedCacheSetCircuitBreaker.Reset()
	optimizedCacheDeleteCircuitBreaker.Reset()

	fmt.Println("✅ All optimized cache circuit breakers have been reset")
}
