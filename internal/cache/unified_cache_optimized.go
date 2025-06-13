package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"news/internal/json"
	"news/internal/metrics"

	"golang.org/x/sync/singleflight"
)

// OptimizedUnifiedCacheManager provides enterprise-grade two-level cache hierarchy
// L1: Ristretto (in-memory, ultra-fast)
// L2: Optimized Redis (network-based, persistent, with singleflight)
type OptimizedUnifiedCacheManager struct {
	ristretto     *RistrettoCache       // L1 cache
	redis         *OptimizedRedisClient // L2 optimized cache
	singleflight  *singleflight.Group   // Prevents duplicate database calls
	config        *CacheConfig
	healthMonitor *CacheHealthMonitor
}

// CacheConfig holds unified cache configuration
type CacheConfig struct {
	// L1 (Ristretto) settings
	L1DefaultTTL   time.Duration
	L1MaxCostRatio float64

	// L2 (Redis) settings
	L2DefaultTTL time.Duration
	L2LongTTL    time.Duration

	// Singleflight settings
	EnableSingleflight bool
	SingleflightTTL    time.Duration

	// Health monitoring
	HealthCheckInterval time.Duration
	MaxFailureRate      float64
}

// CacheHealthMonitor tracks cache layer health and performance
type CacheHealthMonitor struct {
	L1Healthy        bool          `json:"l1_healthy"`
	L2Healthy        bool          `json:"l2_healthy"`
	L1HitRate        float64       `json:"l1_hit_rate"`
	L2HitRate        float64       `json:"l2_hit_rate"`
	OverallHitRate   float64       `json:"overall_hit_rate"`
	AvgLatencyL1     time.Duration `json:"avg_latency_l1"`
	AvgLatencyL2     time.Duration `json:"avg_latency_l2"`
	L1RequestCount   int64         `json:"l1_request_count"`
	L2RequestCount   int64         `json:"l2_request_count"`
	L1HitCount       int64         `json:"l1_hit_count"`
	L2HitCount       int64         `json:"l2_hit_count"`
	SingleflightHits int64         `json:"singleflight_hits"`
	LastHealthCheck  time.Time     `json:"last_health_check"`
	mutex            sync.RWMutex
}

var (
	optimizedUnifiedCache *OptimizedUnifiedCacheManager
	unifiedCacheOnce      sync.Once
)

// GetOptimalCacheConfig returns enterprise-grade cache configuration
func GetOptimalCacheConfig() *CacheConfig {
	return &CacheConfig{
		// L1 Ristretto optimization for hot data
		L1DefaultTTL:   5 * time.Minute, // Keep hot data for 5 minutes
		L1MaxCostRatio: 0.8,             // Use 80% of available memory

		// L2 Redis optimization for persistent data
		L2DefaultTTL: 1 * time.Hour,  // Standard TTL for most data
		L2LongTTL:    24 * time.Hour, // Long TTL for stable data

		// Singleflight optimization to prevent duplicate calls
		EnableSingleflight: true,
		SingleflightTTL:    10 * time.Second, // Prevent duplicate calls for 10s

		// Health monitoring for proactive management
		HealthCheckInterval: 30 * time.Second,
		MaxFailureRate:      0.05, // Alert if >5% failure rate
	}
}

// InitOptimizedUnifiedCache initializes the optimized unified cache manager
func InitOptimizedUnifiedCache() error {
	var initErr error

	unifiedCacheOnce.Do(func() {
		// Initialize L1 (Ristretto)
		if err := InitRistretto(); err != nil {
			initErr = fmt.Errorf("failed to initialize L1 cache (Ristretto): %v", err)
			return
		}

		// Initialize optimized Redis (L2)
		if err := InitOptimizedRedis(); err != nil {
			initErr = fmt.Errorf("failed to initialize optimized L2 cache (Redis): %v", err)
			return
		}

		config := GetOptimalCacheConfig()

		optimizedUnifiedCache = &OptimizedUnifiedCacheManager{
			ristretto:    GetRistrettoCache(),
			redis:        GetOptimizedRedisClient(),
			singleflight: &singleflight.Group{},
			config:       config,
			healthMonitor: &CacheHealthMonitor{
				L1Healthy:       true,
				L2Healthy:       true,
				LastHealthCheck: time.Now(),
			},
		}

		// Start health monitoring
		go optimizedUnifiedCache.healthMonitorLoop()

		fmt.Println("✅ Optimized unified cache initialized successfully")
	})

	return initErr
}

// GetOptimizedUnifiedCache returns the singleton optimized unified cache manager
func GetOptimizedUnifiedCache() *OptimizedUnifiedCacheManager {
	if optimizedUnifiedCache == nil {
		if err := InitOptimizedUnifiedCache(); err != nil {
			panic(fmt.Sprintf("Failed to initialize optimized unified cache: %v", err))
		}
	}
	return optimizedUnifiedCache
}

// SmartGet retrieves a value using optimized cache hierarchy with singleflight protection
func (oucm *OptimizedUnifiedCacheManager) SmartGet(key string) (interface{}, bool) {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_get")()

	start := time.Now()

	// L1: Try Ristretto first (fastest)
	if value, found := oucm.ristretto.Get(key); found {
		latency := time.Since(start)
		oucm.updateL1Metrics(true, latency)
		metrics.IncrementCacheHit("ristretto", key)

		// Deserialize JSON if needed
		if bytes, ok := value.([]byte); ok {
			var result interface{}
			if err := json.Unmarshal(bytes, &result); err == nil {
				return result, true
			}
		}
		return value, true
	}

	// L2: Try optimized Redis with singleflight protection
	if oucm.config.EnableSingleflight {
		value, err, shared := oucm.singleflight.Do(key, func() (interface{}, error) {
			return oucm.redis.SafeGet(key)
		})

		if shared {
			oucm.updateSingleflightMetrics()
		}

		if err == nil {
			latency := time.Since(start)
			oucm.updateL2Metrics(true, latency)
			metrics.IncrementCacheHit("redis", key)

			// Smart cache warming: promote to L1 based on access patterns
			l1TTL := oucm.calculateOptimalL1TTL(key)
			oucm.ristretto.Set(key, value, l1TTL)

			return value, true
		}
	} else {
		// Direct Redis access without singleflight
		if value, err := oucm.redis.SafeGet(key); err == nil {
			latency := time.Since(start)
			oucm.updateL2Metrics(true, latency)
			metrics.IncrementCacheHit("redis", key)

			l1TTL := oucm.calculateOptimalL1TTL(key)
			oucm.ristretto.Set(key, value, l1TTL)

			return value, true
		}
	}

	// Cache miss on both levels
	latency := time.Since(start)
	oucm.updateL1Metrics(false, latency)
	oucm.updateL2Metrics(false, latency)
	metrics.IncrementCacheMiss("unified", key)

	return nil, false
}

// SmartSet stores a value in both cache levels with intelligent TTL optimization
func (oucm *OptimizedUnifiedCacheManager) SmartSet(key string, value interface{}, options ...CacheSetOption) error {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_set")()

	// Parse options
	opts := &CacheSetOptions{
		L1TTL: oucm.config.L1DefaultTTL,
		L2TTL: oucm.config.L2DefaultTTL,
	}
	for _, option := range options {
		option(opts)
	}

	// Determine optimal TTLs based on key patterns
	if opts.L1TTL == oucm.config.L1DefaultTTL {
		opts.L1TTL = oucm.calculateOptimalL1TTL(key)
	}
	if opts.L2TTL == oucm.config.L2DefaultTTL {
		opts.L2TTL = oucm.calculateOptimalL2TTL(key)
	}

	// Store in L1 (Ristretto) with optimized TTL
	if !oucm.ristretto.Set(key, value, opts.L1TTL) {
		fmt.Printf("⚠️ Failed to store in L1 cache for key: %s\n", key)
	}

	// Store in L2 (Optimized Redis) with longer TTL
	if err := oucm.redis.SafeSet(key, value, opts.L2TTL); err != nil {
		return fmt.Errorf("failed to store in optimized L2 cache: %v", err)
	}

	return nil
}

// CacheSetOptions holds options for cache set operations
type CacheSetOptions struct {
	L1TTL time.Duration
	L2TTL time.Duration
}

// CacheSetOption is a function type for setting cache options
type CacheSetOption func(*CacheSetOptions)

// WithL1TTL sets custom L1 TTL
func WithL1TTL(ttl time.Duration) CacheSetOption {
	return func(opts *CacheSetOptions) {
		opts.L1TTL = ttl
	}
}

// WithL2TTL sets custom L2 TTL
func WithL2TTL(ttl time.Duration) CacheSetOption {
	return func(opts *CacheSetOptions) {
		opts.L2TTL = ttl
	}
}

// WithHotData configures for frequently accessed data
func WithHotData() CacheSetOption {
	return func(opts *CacheSetOptions) {
		opts.L1TTL = 15 * time.Minute // Keep hot data longer in L1
		opts.L2TTL = 6 * time.Hour    // Extended L2 TTL for hot data
	}
}

// WithColdData configures for rarely accessed data
func WithColdData() CacheSetOption {
	return func(opts *CacheSetOptions) {
		opts.L1TTL = 2 * time.Minute // Shorter L1 TTL for cold data
		opts.L2TTL = 24 * time.Hour  // Standard L2 TTL for persistence
	}
}

// SmartDelete removes a value from both cache levels with optimized cleanup
func (oucm *OptimizedUnifiedCacheManager) SmartDelete(key string) error {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_delete")()

	// Remove from L1 (always succeeds)
	oucm.ristretto.Delete(key)

	// Remove from L2 with error handling
	if err := oucm.redis.SafeDelete(key); err != nil {
		return fmt.Errorf("failed to delete from optimized L2 cache: %v", err)
	}

	return nil
}

// SmartBulkDelete removes multiple keys efficiently using pipeline
func (oucm *OptimizedUnifiedCacheManager) SmartBulkDelete(keys []string) error {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_bulk_delete")()

	// Remove from L1 (batch operation)
	for _, key := range keys {
		oucm.ristretto.Delete(key)
	}

	// Remove from L2 using Redis pipeline for efficiency
	if oucm.redis.client != nil {
		pipe := oucm.redis.client.Pipeline()
		for _, key := range keys {
			pipe.Del(oucm.redis.ctx, key)
		}
		_, err := pipe.Exec(oucm.redis.ctx)
		if err != nil {
			return fmt.Errorf("failed bulk delete from L2 cache: %v", err)
		}
	}

	return nil
}

// calculateOptimalL1TTL determines intelligent L1 TTL based on access patterns
func (oucm *OptimizedUnifiedCacheManager) calculateOptimalL1TTL(key string) time.Duration {
	// High-frequency patterns get longer L1 TTL
	switch {
	case strings.Contains(key, "trending"):
		return 2 * time.Minute // Fast-changing trending data
	case strings.Contains(key, "articles:page:"):
		return 10 * time.Minute // Hot pagination data
	case strings.Contains(key, "categories:"):
		return 15 * time.Minute // Semi-static category data
	case strings.Contains(key, "tags:"):
		return 12 * time.Minute // Semi-static tag data
	case strings.Contains(key, "article:") && !strings.Contains(key, "articles:"):
		return 20 * time.Minute // Individual articles (stable content)
	case strings.Contains(key, "user:"):
		return 8 * time.Minute // User data (moderately dynamic)
	case strings.Contains(key, "stats:"):
		return 5 * time.Minute // Statistics (updates periodically)
	default:
		return oucm.config.L1DefaultTTL // Default 5 minutes
	}
}

// calculateOptimalL2TTL determines intelligent L2 TTL based on data stability
func (oucm *OptimizedUnifiedCacheManager) calculateOptimalL2TTL(key string) time.Duration {
	// Stable data gets longer L2 TTL
	switch {
	case strings.Contains(key, "article:") && !strings.Contains(key, "articles:"):
		return 24 * time.Hour // Articles rarely change once published
	case strings.Contains(key, "categories:"), strings.Contains(key, "tags:"):
		return 12 * time.Hour // Semi-static metadata
	case strings.Contains(key, "user:profile:"):
		return 6 * time.Hour // User profiles change occasionally
	case strings.Contains(key, "trending"), strings.Contains(key, "stats:"):
		return 30 * time.Minute // Dynamic data needs shorter L2 TTL
	case strings.Contains(key, "articles:page:"):
		return 2 * time.Hour // Pagination data changes with new content
	default:
		return oucm.config.L2DefaultTTL // Default 1 hour
	}
}

// GetHealthStatus returns comprehensive cache health information
func (oucm *OptimizedUnifiedCacheManager) GetHealthStatus() *CacheHealthStatus {
	oucm.healthMonitor.mutex.RLock()
	defer oucm.healthMonitor.mutex.RUnlock()

	// Get Redis connection health
	redisHealth := oucm.redis.GetConnectionHealth()

	return &CacheHealthStatus{
		OverallHealthy:   oucm.healthMonitor.L1Healthy && oucm.healthMonitor.L2Healthy,
		L1Healthy:        oucm.healthMonitor.L1Healthy,
		L2Healthy:        oucm.healthMonitor.L2Healthy,
		L1HitRate:        oucm.healthMonitor.L1HitRate,
		L2HitRate:        oucm.healthMonitor.L2HitRate,
		OverallHitRate:   oucm.healthMonitor.OverallHitRate,
		AvgLatencyL1:     oucm.healthMonitor.AvgLatencyL1,
		AvgLatencyL2:     oucm.healthMonitor.AvgLatencyL2,
		RequestCountL1:   oucm.healthMonitor.L1RequestCount,
		RequestCountL2:   oucm.healthMonitor.L2RequestCount,
		SingleflightHits: oucm.healthMonitor.SingleflightHits,
		RedisHealth:      redisHealth,
		LastHealthCheck:  oucm.healthMonitor.LastHealthCheck,
	}
}

// CacheHealthStatus represents comprehensive cache health information
type CacheHealthStatus struct {
	OverallHealthy   bool              `json:"overall_healthy"`
	L1Healthy        bool              `json:"l1_healthy"`
	L2Healthy        bool              `json:"l2_healthy"`
	L1HitRate        float64           `json:"l1_hit_rate"`
	L2HitRate        float64           `json:"l2_hit_rate"`
	OverallHitRate   float64           `json:"overall_hit_rate"`
	AvgLatencyL1     time.Duration     `json:"avg_latency_l1"`
	AvgLatencyL2     time.Duration     `json:"avg_latency_l2"`
	RequestCountL1   int64             `json:"request_count_l1"`
	RequestCountL2   int64             `json:"request_count_l2"`
	SingleflightHits int64             `json:"singleflight_hits"`
	RedisHealth      *ConnectionHealth `json:"redis_health"`
	LastHealthCheck  time.Time         `json:"last_health_check"`
}

// updateL1Metrics updates L1 cache performance metrics
func (oucm *OptimizedUnifiedCacheManager) updateL1Metrics(hit bool, latency time.Duration) {
	oucm.healthMonitor.mutex.Lock()
	defer oucm.healthMonitor.mutex.Unlock()

	oucm.healthMonitor.L1RequestCount++
	if hit {
		oucm.healthMonitor.L1HitCount++
	}

	// Update L1 hit rate
	oucm.healthMonitor.L1HitRate = float64(oucm.healthMonitor.L1HitCount) / float64(oucm.healthMonitor.L1RequestCount)

	// Update overall hit rate (critical fix for analytics)
	totalHits := oucm.healthMonitor.L1HitCount + oucm.healthMonitor.L2HitCount
	totalRequests := oucm.healthMonitor.L1RequestCount + oucm.healthMonitor.L2RequestCount
	if totalRequests > 0 {
		oucm.healthMonitor.OverallHitRate = float64(totalHits) / float64(totalRequests)
	}

	// Update L1 average latency
	if oucm.healthMonitor.AvgLatencyL1 == 0 {
		oucm.healthMonitor.AvgLatencyL1 = latency
	} else {
		oucm.healthMonitor.AvgLatencyL1 = time.Duration(
			0.9*float64(oucm.healthMonitor.AvgLatencyL1) + 0.1*float64(latency),
		)
	}
}

// updateL2Metrics updates L2 cache performance metrics
func (oucm *OptimizedUnifiedCacheManager) updateL2Metrics(hit bool, latency time.Duration) {
	oucm.healthMonitor.mutex.Lock()
	defer oucm.healthMonitor.mutex.Unlock()

	oucm.healthMonitor.L2RequestCount++
	if hit {
		oucm.healthMonitor.L2HitCount++
	}

	// Update L2 hit rate
	oucm.healthMonitor.L2HitRate = float64(oucm.healthMonitor.L2HitCount) / float64(oucm.healthMonitor.L2RequestCount)

	// Update overall hit rate (with safety check)
	totalHits := oucm.healthMonitor.L1HitCount + oucm.healthMonitor.L2HitCount
	totalRequests := oucm.healthMonitor.L1RequestCount + oucm.healthMonitor.L2RequestCount
	if totalRequests > 0 {
		oucm.healthMonitor.OverallHitRate = float64(totalHits) / float64(totalRequests)
	}

	// Update L2 average latency
	if oucm.healthMonitor.AvgLatencyL2 == 0 {
		oucm.healthMonitor.AvgLatencyL2 = latency
	} else {
		oucm.healthMonitor.AvgLatencyL2 = time.Duration(
			0.9*float64(oucm.healthMonitor.AvgLatencyL2) + 0.1*float64(latency),
		)
	}
}

// updateSingleflightMetrics updates singleflight performance metrics
func (oucm *OptimizedUnifiedCacheManager) updateSingleflightMetrics() {
	oucm.healthMonitor.mutex.Lock()
	defer oucm.healthMonitor.mutex.Unlock()

	oucm.healthMonitor.SingleflightHits++
}

// healthMonitorLoop continuously monitors cache health
func (oucm *OptimizedUnifiedCacheManager) healthMonitorLoop() {
	ticker := time.NewTicker(oucm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			oucm.performHealthCheck()
		}
	}
}

// performHealthCheck executes comprehensive cache health check
func (oucm *OptimizedUnifiedCacheManager) performHealthCheck() {
	oucm.healthMonitor.mutex.Lock()
	defer oucm.healthMonitor.mutex.Unlock()

	oucm.healthMonitor.LastHealthCheck = time.Now()

	// Check L1 (Ristretto) health - always healthy if initialized
	oucm.healthMonitor.L1Healthy = oucm.ristretto != nil

	// Check L2 (Redis) health
	if oucm.redis != nil {
		redisHealth := oucm.redis.GetConnectionHealth()
		oucm.healthMonitor.L2Healthy = redisHealth.IsHealthy
	} else {
		oucm.healthMonitor.L2Healthy = false
	}

	// Log health status changes
	if !oucm.healthMonitor.L1Healthy || !oucm.healthMonitor.L2Healthy {
		fmt.Printf("⚠️ Cache health issue detected - L1: %v, L2: %v\n",
			oucm.healthMonitor.L1Healthy, oucm.healthMonitor.L2Healthy)
	}
}

// GetString retrieves a string value using optimized cache hierarchy (compatibility method)
func (oucm *OptimizedUnifiedCacheManager) GetString(key string) (string, bool) {
	value, found := oucm.SmartGet(key)
	if !found {
		return "", false
	}

	// Handle different value types
	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		// Try to marshal and unmarshal for consistency
		if bytes, err := json.Marshal(v); err == nil {
			var result string
			if err := json.Unmarshal(bytes, &result); err == nil {
				return result, true
			}
		}
	}

	return "", false
}

// Get retrieves a value using optimized cache hierarchy (compatibility method)
func (oucm *OptimizedUnifiedCacheManager) Get(key string) (interface{}, bool) {
	return oucm.SmartGet(key)
}

// Set stores a value in both cache levels (compatibility method)
func (oucm *OptimizedUnifiedCacheManager) Set(key string, value interface{}, l1TTL, l2TTL time.Duration) error {
	return oucm.SmartSet(key, value, WithL1TTL(l1TTL), WithL2TTL(l2TTL))
}

// Delete removes a value from both cache levels (compatibility method)
func (oucm *OptimizedUnifiedCacheManager) Delete(key string) error {
	return oucm.SmartDelete(key)
}

// Clear clears both L1 and L2 cache levels for admin operations
func (oucm *OptimizedUnifiedCacheManager) Clear() error {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_clear")()

	// Clear L1 (Ristretto)
	if oucm.ristretto != nil {
		oucm.ristretto.Clear()
		fmt.Println("✅ L1 cache (Ristretto) cleared")
	}

	// Clear L2 (Optimized Redis)
	if oucm.redis != nil {
		// Use FlushDB command to clear the current Redis database
		if err := oucm.redis.client.FlushDB(oucm.redis.ctx).Err(); err != nil {
			return fmt.Errorf("failed to clear L2 cache (Redis): %v", err)
		}
		fmt.Println("✅ L2 cache (Redis) cleared")
	}

	// Reset health monitor metrics
	oucm.healthMonitor.mutex.Lock()
	oucm.healthMonitor.L1RequestCount = 0
	oucm.healthMonitor.L2RequestCount = 0
	oucm.healthMonitor.L1HitCount = 0
	oucm.healthMonitor.L2HitCount = 0
	oucm.healthMonitor.L1HitRate = 0
	oucm.healthMonitor.L2HitRate = 0
	oucm.healthMonitor.OverallHitRate = 0
	oucm.healthMonitor.SingleflightHits = 0
	oucm.healthMonitor.mutex.Unlock()

	fmt.Println("✅ Optimized unified cache cleared successfully")
	return nil
}

// PreloadPopularContent warms the cache with frequently accessed content
func (oucm *OptimizedUnifiedCacheManager) PreloadPopularContent() error {
	defer metrics.TrackDatabaseOperation("optimized_unified_cache_preload")()

	// Popular cache keys for warming
	popularKeys := []struct {
		key   string
		value string
		l1TTL time.Duration
		l2TTL time.Duration
	}{
		{"articles:list:recent", `{"preloaded": true, "type": "recent_articles"}`, 10 * time.Minute, 1 * time.Hour},
		{"articles:list:popular", `{"preloaded": true, "type": "popular_articles"}`, 15 * time.Minute, 2 * time.Hour},
		{"site:config", `{"preloaded": true, "type": "site_configuration"}`, 30 * time.Minute, 6 * time.Hour},
		{"categories:list", `{"preloaded": true, "type": "categories"}`, 20 * time.Minute, 4 * time.Hour},
	}

	preloadedCount := 0
	for _, item := range popularKeys {
		if err := oucm.SmartSet(item.key, item.value, WithL1TTL(item.l1TTL), WithL2TTL(item.l2TTL)); err != nil {
			fmt.Printf("⚠️ Failed to preload cache key %s: %v\n", item.key, err)
		} else {
			preloadedCount++
		}
	}

	fmt.Printf("✅ Cache preloaded successfully: %d/%d keys\n", preloadedCount, len(popularKeys))
	return nil
}

// GetCacheSize returns approximate cache size information
func (oucm *OptimizedUnifiedCacheManager) GetCacheSize() map[string]interface{} {
	oucm.healthMonitor.mutex.RLock()
	defer oucm.healthMonitor.mutex.RUnlock()

	// Get Ristretto stats for L1 size info
	ristrettoStats := oucm.ristretto.Stats()

	return map[string]interface{}{
		"l1_cache": map[string]interface{}{
			"keys_count":   ristrettoStats["keys_added"].(uint64) - ristrettoStats["keys_evicted"].(uint64),
			"memory_usage": ristrettoStats["cost_added"].(uint64) - ristrettoStats["cost_evicted"].(uint64),
			"keys_added":   ristrettoStats["keys_added"],
			"keys_evicted": ristrettoStats["keys_evicted"],
		},
		"l2_cache": map[string]interface{}{
			"connection_healthy": oucm.healthMonitor.L2Healthy,
			"request_count":      oucm.healthMonitor.L2RequestCount,
			"hit_count":          oucm.healthMonitor.L2HitCount,
		},
		"overall": map[string]interface{}{
			"total_requests": oucm.healthMonitor.L1RequestCount + oucm.healthMonitor.L2RequestCount,
			"total_hits":     oucm.healthMonitor.L1HitCount + oucm.healthMonitor.L2HitCount,
			"hit_rate":       oucm.healthMonitor.OverallHitRate,
		},
	}
}
