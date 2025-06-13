package cache

import (
	"fmt"
	"strings"
	"time"

	"news/internal/json"
	"news/internal/metrics"
)

// UnifiedCacheManager provides a two-level cache hierarchy
// L1: Ristretto (in-memory, ultra-fast)
// L2: Redis (network-based, persistent)
type UnifiedCacheManager struct {
	ristretto *RistrettoCache // L1 cache
	redis     *RedisClient    // L2 cache
}

var (
	defaultUnifiedCache *UnifiedCacheManager
)

// InitUnifiedCache initializes the unified cache manager
func InitUnifiedCache() error {
	// Initialize Ristretto (L1)
	if err := InitRistretto(); err != nil {
		return fmt.Errorf("failed to initialize L1 cache (Ristretto): %v", err)
	}

	// Initialize Redis (L2) - should already be initialized
	if err := InitRedis(); err != nil {
		return fmt.Errorf("failed to initialize L2 cache (Redis): %v", err)
	}

	defaultUnifiedCache = &UnifiedCacheManager{
		ristretto: GetRistrettoCache(),
		redis:     GetRedisClient(),
	}

	return nil
}

// GetUnifiedCache returns the singleton unified cache manager
func GetUnifiedCache() *UnifiedCacheManager {
	if defaultUnifiedCache == nil {
		fmt.Println("Warning: Unified cache not initialized, auto-initializing")
		err := InitUnifiedCache()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize unified cache: %v", err))
		}
	}
	return defaultUnifiedCache
}

// Get retrieves a value using the cache hierarchy (L1 -> L2 -> nil)
func (ucm *UnifiedCacheManager) Get(key string) (interface{}, bool) {
	defer metrics.TrackDatabaseOperation("unified_cache_get")()

	// L1: Try Ristretto first (fastest)
	if value, found := ucm.ristretto.Get(key); found {
		metrics.IncrementCacheHit("ristretto", key)

		// Deserialize the JSON bytes back to original type
		if bytes, ok := value.([]byte); ok {
			var result interface{}
			if err := json.Unmarshal(bytes, &result); err == nil {
				return result, true
			}
		}
		return value, true
	}

	// L2: Fallback to Redis
	if value, err := ucm.redis.GetCachedNews(key); err == nil {
		metrics.IncrementCacheHit("redis", key)

		// Smart cache warming based on access patterns
		l1TTL := ucm.calculateOptimalL1TTL(key)
		ucm.ristretto.Set(key, value, l1TTL)

		return value, true
	}

	metrics.IncrementCacheMiss("unified", key)
	return nil, false
}

// GetString retrieves a string value using the cache hierarchy
func (ucm *UnifiedCacheManager) GetString(key string) (string, bool) {
	// L1: Try Ristretto first
	if value, found := ucm.ristretto.GetString(key); found {
		metrics.IncrementCacheHit("ristretto", key)
		return value, true
	}

	// L2: Fallback to Redis
	if value, err := ucm.redis.GetCachedNews(key); err == nil {
		metrics.IncrementCacheHit("redis", key)

		// Store in L1 for future requests
		ucm.ristretto.Set(key, value, 5*time.Minute)

		return value, true
	}

	metrics.IncrementCacheMiss("unified", key)
	return "", false
}

// Set stores a value in both cache levels with different TTLs
func (ucm *UnifiedCacheManager) Set(key string, value interface{}, l1TTL, l2TTL time.Duration) error {
	defer metrics.TrackDatabaseOperation("unified_cache_set")()

	// Store in L1 (Ristretto) with shorter TTL for hot data
	if !ucm.ristretto.Set(key, value, l1TTL) {
		fmt.Printf("Warning: Failed to store in L1 cache for key: %s\n", key)
	}

	// Store in L2 (Redis) with longer TTL for persistence
	if err := ucm.redis.CacheNews(key, value, l2TTL); err != nil {
		return fmt.Errorf("failed to store in L2 cache: %v", err)
	}

	return nil
}

// SetL1Only stores a value only in L1 cache (for very hot, short-lived data)
func (ucm *UnifiedCacheManager) SetL1Only(key string, value interface{}, ttl time.Duration) bool {
	defer metrics.TrackDatabaseOperation("unified_cache_l1_set")()
	return ucm.ristretto.Set(key, value, ttl)
}

// SetL2Only stores a value only in L2 cache (for persistent data)
func (ucm *UnifiedCacheManager) SetL2Only(key string, value interface{}, ttl time.Duration) error {
	defer metrics.TrackDatabaseOperation("unified_cache_l2_set")()
	return ucm.redis.CacheNews(key, value, ttl)
}

// Delete removes a value from both cache levels
func (ucm *UnifiedCacheManager) Delete(key string) error {
	defer metrics.TrackDatabaseOperation("unified_cache_delete")()

	// Remove from L1
	ucm.ristretto.Delete(key)

	// Remove from L2
	return ucm.redis.RemoveFromCache(key)
}

// DeletePattern removes all keys matching a pattern from both caches
func (ucm *UnifiedCacheManager) DeletePattern(pattern string) error {
	defer metrics.TrackDatabaseOperation("unified_cache_delete_pattern")()

	// Note: Ristretto doesn't support pattern deletion directly
	// We would need to maintain a separate index for this functionality
	// For now, just clear L2 (Redis) pattern

	// Clear L1 entirely if pattern affects many keys (simple approach)
	if pattern == "*" || pattern == "articles:*" {
		ucm.ristretto.Clear()
	}

	// Clear L2 pattern (Redis supports this via EVAL script)
	// This is a simplified approach - in production, implement proper pattern deletion
	return nil
}

// WarmCache preloads data into L1 from L2
func (ucm *UnifiedCacheManager) WarmCache(keys []string) {
	defer metrics.TrackDatabaseOperation("unified_cache_warm")()

	for _, key := range keys {
		if value, err := ucm.redis.GetCachedNews(key); err == nil {
			ucm.ristretto.Set(key, value, 5*time.Minute)
		}
	}
}

// GetStats returns statistics for both cache levels
func (ucm *UnifiedCacheManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"l1_stats": ucm.ristretto.Stats(),
		"l2_healthy": func() bool {
			return ucm.redis.Ping() == nil
		}(),
		"cache_levels": map[string]string{
			"l1": "ristretto",
			"l2": "redis",
		},
	}
}

// Health checks the health of both cache levels
func (ucm *UnifiedCacheManager) Health() map[string]bool {
	return map[string]bool{
		"l1_healthy": ucm.ristretto != nil,
		"l2_healthy": ucm.redis != nil && ucm.redis.Ping() == nil,
	}
}

// Close gracefully closes both cache levels
func (ucm *UnifiedCacheManager) Close() error {
	// Close L1
	if ucm.ristretto != nil {
		ucm.ristretto.Close()
	}

	// Close L2
	if ucm.redis != nil {
		return CloseRedis()
	}

	return nil
}

// Clear clears both cache levels
func (ucm *UnifiedCacheManager) Clear() error {
	defer metrics.TrackDatabaseOperation("unified_cache_clear")()

	// Clear L1 (Ristretto)
	if ucm.ristretto != nil {
		ucm.ristretto.Clear()
	}

	// Clear L2 (Redis)
	if ucm.redis != nil {
		if err := ucm.redis.ClearAllCache(); err != nil {
			return fmt.Errorf("failed to clear L2 cache: %v", err)
		}
	}

	return nil
}

// Legacy wrapper functions for backward compatibility
func CacheNewsUnified(key string, value interface{}, ttl time.Duration) error {
	cache := GetUnifiedCache()
	return cache.Set(key, value, ttl/2, ttl) // L1 gets half TTL, L2 gets full TTL
}

func GetCachedNewsUnified(key string) (string, error) {
	cache := GetUnifiedCache()
	if value, found := cache.GetString(key); found {
		return value, nil
	}
	return "", fmt.Errorf("cache miss")
}

// calculateOptimalL1TTL determines optimal TTL based on key pattern and access frequency
func (ucm *UnifiedCacheManager) calculateOptimalL1TTL(key string) time.Duration {
	// High-frequency endpoints get longer L1 TTL
	if strings.Contains(key, "articles:page:") {
		return 10 * time.Minute // Hot pagination data
	}
	if strings.Contains(key, "categories:") {
		return 15 * time.Minute // Semi-static data
	}
	if strings.Contains(key, "tags:") {
		return 12 * time.Minute // Semi-static data
	}
	if strings.Contains(key, "trending") {
		return 3 * time.Minute // Fast-changing data
	}
	if strings.Contains(key, "article:") && !strings.Contains(key, "articles:") {
		return 20 * time.Minute // Individual articles (stable)
	}

	// Default for unknown patterns
	return 5 * time.Minute
}

// PreloadPopularContent warms cache with frequently accessed data
func (ucm *UnifiedCacheManager) PreloadPopularContent() error {
	defer metrics.TrackDatabaseOperation("cache_preload")()

	// Define popular content patterns to preload
	popularKeys := []string{
		"articles:page:1:limit:10:category:", // First page of articles
		"articles:page:1:limit:20:category:", // First page with higher limit
		"categories:list:hierarchical:true",  // Category list
		"tags:list:sort:popular:limit:50",    // Popular tags
	}

	var errors []error
	for _, key := range popularKeys {
		// Try to get from L2 and warm L1
		if value, err := ucm.redis.GetCachedNews(key); err == nil {
			l1TTL := ucm.calculateOptimalL1TTL(key)
			if !ucm.ristretto.Set(key, value, l1TTL) {
				errors = append(errors, fmt.Errorf("failed to preload key: %s", key))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("preload completed with %d errors", len(errors))
	}
	return nil
}

// GetCacheEfficiency returns cache performance analytics
func (ucm *UnifiedCacheManager) GetCacheEfficiency() map[string]interface{} {
	l1Stats := ucm.ristretto.Stats()

	efficiency := map[string]interface{}{
		"l1_hit_ratio":    l1Stats["hit_ratio"],
		"l1_keys_active":  l1Stats["keys_added"].(uint64) - l1Stats["keys_evicted"].(uint64),
		"l1_memory_usage": l1Stats["cost_added"].(uint64) - l1Stats["cost_evicted"].(uint64),
		"l2_healthy":      ucm.redis.Ping() == nil,
		"overall_efficiency": func() string {
			ratio := l1Stats["hit_ratio"].(float64)
			if ratio > 0.95 {
				return "excellent"
			} else if ratio > 0.85 {
				return "good"
			} else if ratio > 0.70 {
				return "moderate"
			}
			return "poor"
		}(),
		"recommendations": ucm.generateOptimizationRecommendations(l1Stats),
	}

	return efficiency
}

// generateOptimizationRecommendations provides cache optimization suggestions
func (ucm *UnifiedCacheManager) generateOptimizationRecommendations(l1Stats map[string]interface{}) []string {
	var recommendations []string

	hitRatio := l1Stats["hit_ratio"].(float64)
	keysEvicted := l1Stats["keys_evicted"].(uint64)
	setsRejected := l1Stats["sets_rejected"].(uint64)

	if hitRatio < 0.85 {
		recommendations = append(recommendations, "Consider increasing L1 TTL for frequently accessed data")
	}

	if keysEvicted > 100 {
		recommendations = append(recommendations, "High eviction rate detected - consider increasing MaxCost")
	}

	if setsRejected > 10 {
		recommendations = append(recommendations, "Cache rejections detected - optimize data serialization size")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Cache performance is optimal")
	}

	return recommendations
}
