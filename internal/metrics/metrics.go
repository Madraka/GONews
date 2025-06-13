package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestDuration tracks HTTP request duration
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "news_api_http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 2, 5, 10},
	}, []string{"path", "method", "status"})

	// RequestsTotal tracks total number of HTTP requests
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "news_api_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"path", "method", "status"})

	// RateLimitExceeded tracks rate limit exceeded events
	RateLimitExceeded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "news_api_rate_limit_exceeded_total",
		Help: "Total number of rate limit exceeded events",
	}, []string{"path", "ip"})

	// CacheHitTotal tracks cache hits
	CacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "news_api_cache_hit_total",
		Help: "Total number of cache hits",
	}, []string{"key"})

	// CacheMissTotal tracks cache misses
	CacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "news_api_cache_miss_total",
		Help: "Total number of cache misses",
	}, []string{"key"})

	// ActiveConnections tracks active connections
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "news_api_active_connections",
		Help: "Current number of active connections",
	})

	// DatabaseOperationDuration tracks database operation duration
	DatabaseOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "news_api_database_operation_duration_seconds",
		Help:    "Duration of database operations in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
	}, []string{"operation"})
)

// PrometheusMiddleware collects metrics for HTTP requests
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Track active connections
		ActiveConnections.Inc()
		defer ActiveConnections.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		// Record request duration and count
		RequestDuration.WithLabelValues(path, method, strconv.Itoa(statusCode)).Observe(duration)
		RequestsTotal.WithLabelValues(path, method, strconv.Itoa(statusCode)).Inc()
	}
}

// SetupMetrics sets up Prometheus metrics endpoint
func SetupMetrics(router *gin.Engine) {
	// Add prometheus metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		h := promhttp.Handler()
		h.ServeHTTP(c.Writer, c.Request)
	})
}

// TrackDatabaseOperation tracks the duration of a database operation
func TrackDatabaseOperation(operation string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		DatabaseOperationDuration.WithLabelValues(operation).Observe(duration)
	}
}

// TrackCacheHit records a cache hit
func TrackCacheHit(key string) {
	CacheHitTotal.WithLabelValues(key).Inc()
}

// TrackCacheMiss records a cache miss
func TrackCacheMiss(key string) {
	CacheMissTotal.WithLabelValues(key).Inc()
}

// IncrementCacheHit records a cache hit with cache layer information
func IncrementCacheHit(layer, key string) {
	// Use existing TrackCacheHit but add layer info in key
	TrackCacheHit(layer + ":" + key)
}

// IncrementCacheMiss records a cache miss with cache layer information
func IncrementCacheMiss(layer, key string) {
	// Use existing TrackCacheMiss but add layer info in key
	TrackCacheMiss(layer + ":" + key)
}

// TrackRateLimitExceeded records a rate limit exceeded event
func TrackRateLimitExceeded(path, ip string) {
	RateLimitExceeded.WithLabelValues(path, ip).Inc()
}

// IncrementCounter increments a general purpose counter metric
func IncrementCounter(counterName string) {
	// Use existing CacheHitTotal as a general counter for circuit breaker events
	CacheHitTotal.WithLabelValues(counterName).Inc()
}

// TrackCacheSet records a cache set operation
func TrackCacheSet(key string) {
	// Track cache set operations using cache hit counter
	CacheHitTotal.WithLabelValues("cache_set:" + key).Inc()
}

// TrackCacheDelete records a cache delete operation
func TrackCacheDelete(key string) {
	// Track cache delete operations using cache hit counter
	CacheHitTotal.WithLabelValues("cache_delete:" + key).Inc()
}
