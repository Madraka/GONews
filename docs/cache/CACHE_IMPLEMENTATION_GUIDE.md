# 🎯 Cache Implementation Quick Start Guide

**For Developers** | **Production Ready** | **Copy-Paste Examples**

---

## 🚀 **Quick Start (5 Minutes)**

### **1. Basic Cache Operations**

```go
package main

import (
    "time"
    "news/internal/cache"
)

func main() {
    // Get cache manager (handles optimized + fallback automatically)
    cacheManager := cache.GetMigrationCacheManager()
    
    // ✅ RECOMMENDED: Smart cache operations
    articles, found := cacheManager.SmartGet("articles:recent")
    if !found {
        // Cache miss - fetch from database
        articles = fetchArticlesFromDB()
        
        // Cache with intelligent TTL
        cacheManager.SmartSet("articles:recent", articles, 
            5*time.Minute,  // L1 TTL (in-memory)
            15*time.Minute) // L2 TTL (Redis)
    }
}
```

### **2. Service Integration Pattern**

```go
// Example: Article Service with Cache
type ArticleService struct {
    db    *gorm.DB
    cache *cache.CacheManager
}

func NewArticleService(db *gorm.DB) *ArticleService {
    return &ArticleService{
        db:    db,
        cache: cache.GetMigrationCacheManager(),
    }
}

func (s *ArticleService) GetArticles(page, limit int, category string) ([]Article, error) {
    // 1. Generate cache key
    cacheKey := fmt.Sprintf("articles:page:%d:limit:%d:category:%s", 
        page, limit, category)
    
    // 2. Try cache first
    if cachedData, found := s.cache.SmartGet(cacheKey); found {
        var articles []Article
        if err := json.Unmarshal([]byte(cachedData), &articles); err == nil {
            return articles, nil
        }
    }
    
    // 3. Cache miss - fetch from database
    var articles []Article
    offset := page * limit
    
    query := s.db.Offset(offset).Limit(limit)
    if category != "" {
        query = query.Where("category = ?", category)
    }
    
    if err := query.Find(&articles).Error; err != nil {
        return nil, err
    }
    
    // 4. Cache the result
    if data, err := json.Marshal(articles); err == nil {
        s.cache.SmartSet(cacheKey, string(data), 
            5*time.Minute,  // L1: 5 minutes for frequent access
            1*time.Hour)    // L2: 1 hour for persistence
    }
    
    return articles, nil
}
```

---

## 🔧 **Advanced Usage Patterns**

### **Hot Data Optimization**

```go
// For frequently accessed data (trending articles, popular categories)
optimizedCache := cache.GetOptimizedUnifiedCache()

// Use hot data configuration
err := optimizedCache.SmartSet("trending:articles", data,
    cache.WithHotData()) // 15min L1, 6hour L2

// Or custom TTL
err = optimizedCache.SmartSet("popular:content", data,
    cache.WithL1TTL(15*time.Minute),
    cache.WithL2TTL(6*time.Hour))
```

### **Cold Data Optimization**

```go
// For rarely accessed data (archive content, old articles)
err := optimizedCache.SmartSet("archive:article:123", data,
    cache.WithColdData()) // 2min L1, 24hour L2
```

### **Bulk Operations**

```go
// Efficient bulk deletion
keys := []string{
    "articles:page:1",
    "articles:page:2", 
    "articles:category:tech",
}
err := optimizedCache.SmartBulkDelete(keys)
```

### **Cache Invalidation Patterns**

```go
// Pattern 1: Tag-based invalidation
func (s *ArticleService) CreateArticle(article Article) error {
    // 1. Save to database
    if err := s.db.Create(&article).Error; err != nil {
        return err
    }
    
    // 2. Invalidate related cache keys
    keysToInvalidate := []string{
        "articles:recent",
        fmt.Sprintf("articles:category:%s", article.Category),
        "articles:page:1", // First page likely affected
    }
    
    for _, key := range keysToInvalidate {
        s.cache.SmartDelete(key)
    }
    
    return nil
}

// Pattern 2: Proactive cache warming
func (s *ArticleService) UpdateArticle(id uint, updates Article) error {
    // 1. Update database
    if err := s.db.Model(&Article{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        return err
    }
    
    // 2. Immediately warm cache with new data
    cacheKey := fmt.Sprintf("article:%d", id)
    
    var updatedArticle Article
    if err := s.db.First(&updatedArticle, id).Error; err == nil {
        if data, err := json.Marshal(updatedArticle); err == nil {
            s.cache.SmartSet(cacheKey, string(data), 
                10*time.Minute, 
                2*time.Hour)
        }
    }
    
    return nil
}
```

---

## 📊 **Monitoring Integration**

### **Health Checks in Handlers**

```go
// Health endpoint with cache status
func HealthHandler(c *gin.Context) {
    cacheManager := cache.GetMigrationCacheManager()
    health := cacheManager.GetCacheHealth()
    
    status := "healthy"
    if !health["overall_healthy"].(bool) {
        status = "degraded"
        c.Status(http.StatusServiceUnavailable)
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": status,
        "cache": health,
        "timestamp": time.Now(),
    })
}
```

### **Performance Monitoring**

```go
// Middleware for cache performance tracking
func CachePerformanceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Process request
        c.Next()
        
        // Track cache performance
        if c.Request.Method == "GET" {
            latency := time.Since(start)
            
            // Log slow requests
            if latency > 100*time.Millisecond {
                log.Printf("Slow request: %s %s took %v", 
                    c.Request.Method, 
                    c.Request.URL.Path, 
                    latency)
                
                // Check cache health
                analytics := cache.GetMigrationCacheManager().GetCacheEfficiency()
                if hitRate := analytics["overall_hit_rate"].(float64); hitRate < 0.8 {
                    log.Printf("Low cache hit rate: %.2f%%", hitRate*100)
                }
            }
        }
    }
}
```

---

## 🔧 **Configuration Patterns**

### **Environment-Based Configuration**

```go
// config/cache.go
type CacheConfig struct {
    L1TTL           time.Duration
    L2TTL           time.Duration
    EnableOptimized bool
    PoolSize        int
}

func GetCacheConfig() *CacheConfig {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "production":
        return &CacheConfig{
            L1TTL:           10 * time.Minute,
            L2TTL:           2 * time.Hour,
            EnableOptimized: true,
            PoolSize:        50,
        }
    case "staging":
        return &CacheConfig{
            L1TTL:           5 * time.Minute,
            L2TTL:           1 * time.Hour,
            EnableOptimized: true,
            PoolSize:        25,
        }
    default: // development
        return &CacheConfig{
            L1TTL:           2 * time.Minute,
            L2TTL:           15 * time.Minute,
            EnableOptimized: false,
            PoolSize:        10,
        }
    }
}
```

### **TTL Strategy by Data Type**

```go
// Cache TTL constants
const (
    // Hot data (frequently accessed)
    HotDataL1TTL = 15 * time.Minute
    HotDataL2TTL = 6 * time.Hour
    
    // Warm data (moderately accessed)
    WarmDataL1TTL = 5 * time.Minute
    WarmDataL2TTL = 1 * time.Hour
    
    // Cold data (rarely accessed)
    ColdDataL1TTL = 2 * time.Minute
    ColdDataL2TTL = 24 * time.Hour
    
    // Session data
    SessionL1TTL = 30 * time.Minute
    SessionL2TTL = 8 * time.Hour
    
    // Static content
    StaticL1TTL = 1 * time.Hour
    StaticL2TTL = 7 * 24 * time.Hour // 1 week
)

// TTL selector helper
func GetTTLForDataType(dataType string) (time.Duration, time.Duration) {
    switch dataType {
    case "trending", "popular", "recent":
        return HotDataL1TTL, HotDataL2TTL
    case "articles", "categories", "tags":
        return WarmDataL1TTL, WarmDataL2TTL
    case "archive", "old":
        return ColdDataL1TTL, ColdDataL2TTL
    case "session", "user":
        return SessionL1TTL, SessionL2TTL
    case "static", "config":
        return StaticL1TTL, StaticL2TTL
    default:
        return WarmDataL1TTL, WarmDataL2TTL
    }
}
```

---

## 🚨 **Error Handling Patterns**

### **Graceful Degradation**

```go
func (s *ArticleService) GetArticleWithFallback(id uint) (*Article, error) {
    cacheKey := fmt.Sprintf("article:%d", id)
    
    // Try optimized cache first
    cacheManager := cache.GetMigrationCacheManager()
    
    if data, found := cacheManager.SmartGet(cacheKey); found {
        var article Article
        if err := json.Unmarshal([]byte(data), &article); err == nil {
            return &article, nil
        }
        // Cache data corrupted, continue to database
        log.Printf("Cache data corrupted for key: %s", cacheKey)
    }
    
    // Fallback to database
    var article Article
    if err := s.db.First(&article, id).Error; err != nil {
        return nil, err
    }
    
    // Try to cache the result (best effort)
    if data, err := json.Marshal(article); err == nil {
        l1TTL, l2TTL := GetTTLForDataType("articles")
        if err := cacheManager.SmartSet(cacheKey, string(data), l1TTL, l2TTL); err != nil {
            log.Printf("Failed to cache article %d: %v", id, err)
            // Don't fail the request, just log the error
        }
    }
    
    return &article, nil
}
```

### **Circuit Breaker Pattern**

```go
// Use built-in circuit breaker protection
func SafeCacheOperation(key string, operation func() (interface{}, error)) interface{} {
    // The optimized cache has built-in circuit breaker protection
    optimizedCache := cache.GetOptimizedUnifiedCache()
    
    // Safe operations with automatic circuit breaker
    if value, found := optimizedCache.SmartGet(key); found {
        return value
    }
    
    // Circuit breaker will handle Redis failures automatically
    if data, err := operation(); err == nil {
        optimizedCache.SmartSet(key, data, 5*time.Minute, 1*time.Hour)
        return data
    }
    
    return nil
}
```

---

## 📈 **Performance Testing**

### **Cache Warming Scripts**

```go
// scripts/warm_cache.go
func WarmProductionCache() error {
    cacheManager := cache.GetMigrationCacheManager()
    
    // 1. Warm popular articles
    popularKeys := []string{
        "articles:recent:limit:20",
        "articles:trending:limit:10", 
        "articles:popular:limit:15",
    }
    
    for _, key := range popularKeys {
        if _, found := cacheManager.SmartGet(key); !found {
            // Fetch and cache
            switch {
            case strings.Contains(key, "recent"):
                articles, _ := articleService.GetRecent(20)
                cacheData, _ := json.Marshal(articles)
                cacheManager.SmartSet(key, string(cacheData), HotDataL1TTL, HotDataL2TTL)
                
            case strings.Contains(key, "trending"):
                articles, _ := articleService.GetTrending(10)
                cacheData, _ := json.Marshal(articles)
                cacheManager.SmartSet(key, string(cacheData), HotDataL1TTL, HotDataL2TTL)
            }
        }
    }
    
    // 2. Warm categories and tags
    categories, _ := categoryService.GetAll()
    cacheData, _ := json.Marshal(categories)
    cacheManager.SmartSet("categories:all", string(cacheData), WarmDataL1TTL, WarmDataL2TTL)
    
    return nil
}
```

### **Load Testing with Cache**

```bash
#!/bin/bash
# scripts/load_test_cache.sh

echo "🚀 Starting Cache Load Test"

# Warm cache first
curl -X POST http://localhost:8081/api/cache/preload

# Run concurrent requests
echo "📊 Testing concurrent load..."
for i in {1..100}; do
    (
        curl -s "http://localhost:8081/api/articles?page=1&limit=20" > /dev/null
        curl -s "http://localhost:8081/api/articles?page=2&limit=20" > /dev/null
        curl -s "http://localhost:8081/api/categories" > /dev/null
    ) &
    
    # Limit concurrent connections
    if (( i % 20 == 0 )); then
        wait
    fi
done

wait

# Check final performance
echo "📈 Final Cache Performance:"
curl -s http://localhost:8081/api/cache/analytics | jq '.analytics.performance_metrics'
```

---

## 🎯 **Best Practices Checklist**

### **✅ Implementation Checklist**

- [ ] Use `cache.GetMigrationCacheManager()` for all cache operations
- [ ] Implement cache key naming conventions (`service:operation:params`)
- [ ] Set appropriate TTL values based on data access patterns
- [ ] Handle cache misses gracefully with database fallback
- [ ] Implement cache invalidation on data updates
- [ ] Add cache performance monitoring to critical paths
- [ ] Use bulk operations for multiple cache operations
- [ ] Test cache warming strategies
- [ ] Monitor cache hit rates and latency
- [ ] Document cache keys and TTL strategies

### **🚨 Common Pitfalls to Avoid**

- ❌ Caching large objects (>1MB) - breaks down into smaller chunks
- ❌ Using cache for critical consistency requirements - use database
- ❌ Ignoring cache failures - always implement fallbacks
- ❌ Hard-coding TTL values - use configuration
- ❌ Not monitoring cache performance - setup alerts
- ❌ Caching user-specific data with shared keys - ensure key isolation
- ❌ Not warming cache after deployments - implement warming strategies

---

## 🔧 **Development Tools**

### **Cache Debug Utility**

```go
// tools/cache_debug.go
func DebugCacheKey(key string) {
    cacheManager := cache.GetMigrationCacheManager()
    
    // Check if key exists in both cache systems
    optimizedCache := cache.GetOptimizedUnifiedCache()
    
    // L1 check
    if _, found := optimizedCache.Get(key); found {
        fmt.Printf("✅ Found in L1 (Ristretto): %s\n", key)
    } else {
        fmt.Printf("❌ Not found in L1: %s\n", key)
    }
    
    // Health status
    health := optimizedCache.GetHealthStatus()
    fmt.Printf("📊 L1 Hit Rate: %.2f%%\n", health.L1HitRate*100)
    fmt.Printf("📊 Overall Hit Rate: %.2f%%\n", health.OverallHitRate*100)
    fmt.Printf("⚡ L1 Latency: %v\n", health.AvgLatencyL1)
}
```

### **Cache Analytics Dashboard** 

```bash
# Quick cache status
alias cache-status='curl -s http://localhost:8081/api/cache/analytics | jq ".analytics.performance_metrics"'

# Cache health check
alias cache-health='curl -s http://localhost:8081/api/cache/health | jq ".overall_healthy, .status"'

# Cache warm
alias cache-warm='curl -X POST http://localhost:8081/api/cache/preload'

# Performance benchmark
alias cache-bench='./scripts/performance/cache_performance_benchmark.sh'
```

---

**🔗 Related**: [Cache Architecture Guide](./CACHE_ARCHITECTURE_GUIDE.md) | [Performance Benchmarks](./PERFORMANCE_BENCHMARKS.md)

**📞 Support**: For implementation questions, check the comprehensive [Cache Architecture Guide](./CACHE_ARCHITECTURE_GUIDE.md) or contact the development team.
