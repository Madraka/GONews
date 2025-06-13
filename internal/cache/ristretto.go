package cache

import (
	"fmt"
	"time"

	"news/internal/json"
	"news/internal/metrics"

	"github.com/dgraph-io/ristretto"
)

// RistrettoCache provides high-performance in-memory caching
type RistrettoCache struct {
	cache *ristretto.Cache
}

var (
	defaultRistrettoCache *RistrettoCache
)

// InitRistretto initializes the Ristretto cache
func InitRistretto() error {
	config := &ristretto.Config{
		NumCounters: 1e7,     // 10M keys tracking (optimal for high-traffic)
		MaxCost:     2 << 30, // 2GB max memory (increased for better caching)
		BufferItems: 128,     // Increased async buffer for better throughput
		Metrics:     true,    // Enable detailed metrics
		KeyToHash:   nil,     // Use default hash function (fastest)
		Cost: func(value interface{}) int64 {
			// Dynamic cost calculation based on serialized size
			if bytes, ok := value.([]byte); ok {
				return int64(len(bytes))
			}
			// Estimate cost for other types
			return 1
		},
	}

	cache, err := ristretto.NewCache(config)
	if err != nil {
		return fmt.Errorf("failed to initialize Ristretto cache: %v", err)
	}

	defaultRistrettoCache = &RistrettoCache{
		cache: cache,
	}

	return nil
}

// GetRistrettoCache returns the singleton Ristretto cache instance
func GetRistrettoCache() *RistrettoCache {
	if defaultRistrettoCache == nil {
		// Auto-initialize if needed but log a warning
		fmt.Println("Warning: Ristretto cache not initialized, auto-initializing")
		err := InitRistretto()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize Ristretto: %v", err))
		}
	}
	return defaultRistrettoCache
}

// Set stores a value in the cache with TTL
func (rc *RistrettoCache) Set(key string, value interface{}, ttl time.Duration) bool {
	defer metrics.TrackDatabaseOperation("ristretto_set")()

	// Serialize the value to JSON for consistent storage
	serialized, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("Error serializing value for key %s: %v\n", key, err)
		return false
	}

	cost := int64(len(serialized)) // Use serialized size as cost
	return rc.cache.SetWithTTL(key, serialized, cost, ttl)
}

// Get retrieves a value from the cache
func (rc *RistrettoCache) Get(key string) (interface{}, bool) {
	defer metrics.TrackDatabaseOperation("ristretto_get")()

	value, found := rc.cache.Get(key)
	if !found {
		metrics.TrackCacheMiss(key)
		return nil, false
	}

	metrics.TrackCacheHit(key)
	return value, true
}

// GetString retrieves a string value from the cache and deserializes it
func (rc *RistrettoCache) GetString(key string) (string, bool) {
	value, found := rc.Get(key)
	if !found {
		return "", false
	}

	// Value is stored as []byte (JSON), convert back to string
	if bytes, ok := value.([]byte); ok {
		var result string
		if err := json.Unmarshal(bytes, &result); err != nil {
			fmt.Printf("Error deserializing string for key %s: %v\n", key, err)
			return "", false
		}
		return result, true
	}

	return "", false
}

// Delete removes a value from the cache
func (rc *RistrettoCache) Delete(key string) {
	defer metrics.TrackDatabaseOperation("ristretto_delete")()
	rc.cache.Del(key)
}

// Clear removes all values from the cache
func (rc *RistrettoCache) Clear() {
	defer metrics.TrackDatabaseOperation("ristretto_clear")()
	rc.cache.Clear()
}

// GetMetrics returns cache metrics
func (rc *RistrettoCache) GetMetrics() *ristretto.Metrics {
	if rc.cache != nil {
		return rc.cache.Metrics
	}
	return nil
}

// Stats returns detailed cache statistics
func (rc *RistrettoCache) Stats() map[string]interface{} {
	metrics := rc.GetMetrics()
	if metrics == nil {
		return map[string]interface{}{
			"status": "not_initialized",
		}
	}

	return map[string]interface{}{
		"hit_ratio":     metrics.Ratio(),
		"hits":          metrics.Hits(),
		"misses":        metrics.Misses(),
		"keys_added":    metrics.KeysAdded(),
		"keys_updated":  metrics.KeysUpdated(),
		"keys_evicted":  metrics.KeysEvicted(),
		"cost_added":    metrics.CostAdded(),
		"cost_evicted":  metrics.CostEvicted(),
		"sets_dropped":  metrics.SetsDropped(),
		"sets_rejected": metrics.SetsRejected(),
		"gets_kept":     metrics.GetsKept(),
		"gets_dropped":  metrics.GetsDropped(),
	}
}

// Close gracefully closes the Ristretto cache
func (rc *RistrettoCache) Close() {
	if rc.cache != nil {
		rc.cache.Close()
	}
}

// CloseRistretto gracefully closes the global Ristretto cache
func CloseRistretto() {
	if defaultRistrettoCache != nil {
		defaultRistrettoCache.Close()
		defaultRistrettoCache = nil
	}
}
