package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"news/internal/metrics"

	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
)

// OptimizedRedisClient provides enterprise-grade Redis client with advanced optimizations
type OptimizedRedisClient struct {
	client       *redis.Client
	ctx          context.Context
	singleflight *singleflight.Group // Prevents duplicate cache requests
	connMonitor  *ConnectionMonitor
	config       *RedisConfig
}

// RedisConfig holds comprehensive Redis configuration
type RedisConfig struct {
	// Connection settings
	Addr     string
	Password string
	DB       int

	// Pool settings for high concurrency
	PoolSize        int           // Maximum number of socket connections
	MinIdleConns    int           // Minimum number of idle connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime
	ConnMaxIdleTime time.Duration // Maximum idle time before closing

	// Timeout settings for latency control
	DialTimeout  time.Duration // Connection dial timeout
	ReadTimeout  time.Duration // Socket read timeout
	WriteTimeout time.Duration // Socket write timeout
	PoolTimeout  time.Duration // Pool timeout when all connections busy

	// Retry and resilience settings
	MaxRetries      int           // Maximum retry attempts
	MinRetryBackoff time.Duration // Minimum retry backoff
	MaxRetryBackoff time.Duration // Maximum retry backoff

	// Health monitoring
	HealthCheckInterval time.Duration
}

// ConnectionMonitor tracks Redis connection health and performance
type ConnectionMonitor struct {
	isHealthy        bool
	lastHealthCheck  time.Time
	consecutiveFails int
	mutex            sync.RWMutex

	// Performance metrics
	connectionTime time.Duration
	lastLatency    time.Duration
	avgLatency     time.Duration
	requestCount   int64
}

var (
	optimizedClient *OptimizedRedisClient
	clientOnce      sync.Once
)

// GetOptimalRedisConfig returns enterprise-grade Redis configuration
func GetOptimalRedisConfig() *RedisConfig {
	return &RedisConfig{
		// Connection pool optimization for high concurrency
		PoolSize:        getEnvInt("REDIS_POOL_SIZE", 50),                      // 50 connections for high load
		MinIdleConns:    getEnvInt("REDIS_MIN_IDLE", 10),                       // Always keep 10 connections ready
		MaxIdleConns:    getEnvInt("REDIS_MAX_IDLE", 20),                       // Up to 20 idle connections
		ConnMaxLifetime: getEnvDuration("REDIS_CONN_LIFETIME", 30*time.Minute), // Recycle connections every 30 min
		ConnMaxIdleTime: getEnvDuration("REDIS_CONN_IDLE_TIME", 5*time.Minute), // Close idle after 5 min

		// Aggressive timeout optimization for low latency
		DialTimeout:  getEnvDuration("REDIS_DIAL_TIMEOUT", 2*time.Second),        // Fast connection establishment
		ReadTimeout:  getEnvDuration("REDIS_READ_TIMEOUT", 1*time.Second),        // Quick read operations
		WriteTimeout: getEnvDuration("REDIS_WRITE_TIMEOUT", 1*time.Second),       // Quick write operations
		PoolTimeout:  getEnvDuration("REDIS_POOL_TIMEOUT", 500*time.Millisecond), // Fast pool acquisition

		// Smart retry strategy with exponential backoff
		MaxRetries:      getEnvInt("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff: getEnvDuration("REDIS_MIN_BACKOFF", 8*time.Millisecond),   // Start small
		MaxRetryBackoff: getEnvDuration("REDIS_MAX_BACKOFF", 512*time.Millisecond), // Cap at 512ms

		// Health monitoring for proactive issue detection
		HealthCheckInterval: getEnvDuration("REDIS_HEALTH_INTERVAL", 30*time.Second),
	}
}

// InitOptimizedRedis initializes the optimized Redis client
func InitOptimizedRedis() error {
	var initErr error

	clientOnce.Do(func() {
		if inTestMode {
			optimizedClient = &OptimizedRedisClient{
				client:       nil,
				ctx:          context.Background(),
				singleflight: &singleflight.Group{},
				connMonitor:  &ConnectionMonitor{isHealthy: true},
			}
			return
		}

		config := GetOptimalRedisConfig()

		// Build Redis address
		redisAddr := os.Getenv("REDIS_URL")

		// Parse Redis URL if it contains protocol
		if strings.HasPrefix(redisAddr, "redis://") {
			// Extract host:port from redis://host:port/db
			redisAddr = strings.TrimPrefix(redisAddr, "redis://")
			if strings.Contains(redisAddr, "/") {
				// Remove database suffix (e.g., "/0")
				redisAddr = strings.Split(redisAddr, "/")[0]
			}
		}

		if redisAddr == "" {
			redisHost := os.Getenv("REDIS_HOST")
			redisPort := os.Getenv("REDIS_PORT")
			if redisHost != "" && redisPort != "" {
				redisAddr = redisHost + ":" + redisPort
			} else {
				redisAddr = "localhost:6379"
			}
		}

		config.Addr = redisAddr
		config.Password = os.Getenv("REDIS_PASSWORD")
		config.DB = getEnvInt("REDIS_DB", 0)

		// Create optimized Redis client with comprehensive configuration
		client := redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,

			// Connection pool optimization
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxConnAge:   config.ConnMaxLifetime,
			IdleTimeout:  config.ConnMaxIdleTime,

			// Timeout optimization for low latency
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			PoolTimeout:  config.PoolTimeout,

			// Retry configuration with exponential backoff
			MaxRetries:      config.MaxRetries,
			MinRetryBackoff: config.MinRetryBackoff,
			MaxRetryBackoff: config.MaxRetryBackoff,

			// Enable keep-alive for connection persistence
			OnConnect: func(ctx context.Context, cn *redis.Conn) error {
				fmt.Printf("New Redis connection established: %s\n", cn.String())
				return nil
			},
		})

		// Initialize connection monitor
		monitor := &ConnectionMonitor{
			isHealthy:       true,
			lastHealthCheck: time.Now(),
		}

		optimizedClient = &OptimizedRedisClient{
			client:       client,
			ctx:          context.Background(),
			singleflight: &singleflight.Group{},
			connMonitor:  monitor,
			config:       config,
		}

		// Test initial connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		_, err := client.Ping(ctx).Result()
		connectionTime := time.Since(start)

		if err != nil {
			initErr = fmt.Errorf("failed to connect to Redis: %v", err)
			return
		}

		monitor.connectionTime = connectionTime
		fmt.Printf("✅ Optimized Redis client initialized successfully (connection time: %v)\n", connectionTime)

		// Start health monitoring goroutine
		go optimizedClient.healthMonitorLoop()
	})

	return initErr
}

// GetOptimizedRedisClient returns the singleton optimized Redis client
func GetOptimizedRedisClient() *OptimizedRedisClient {
	if optimizedClient == nil {
		if err := InitOptimizedRedis(); err != nil {
			panic(fmt.Sprintf("Failed to initialize optimized Redis client: %v", err))
		}
	}
	return optimizedClient
}

// SafeGet implements singleflight pattern to prevent duplicate cache requests
func (orc *OptimizedRedisClient) SafeGet(key string) (string, error) {
	if inTestMode {
		return "", fmt.Errorf("test mode")
	}

	// Use singleflight to prevent duplicate requests for the same key
	value, err, shared := orc.singleflight.Do(key, func() (interface{}, error) {
		start := time.Now()

		result, err := orc.client.Get(orc.ctx, key).Result()
		latency := time.Since(start)

		// Update performance metrics
		orc.updateLatencyMetrics(latency)

		if err == redis.Nil {
			return "", fmt.Errorf("key not found")
		}

		return result, err
	})

	if shared {
		metrics.IncrementCounter("redis_singleflight_shared_requests")
	}

	if err != nil {
		return "", err
	}

	return value.(string), nil
}

// SafeSet implements optimized cache set with monitoring
func (orc *OptimizedRedisClient) SafeSet(key string, value interface{}, expiration time.Duration) error {
	if inTestMode {
		return nil
	}

	start := time.Now()

	// Use pipeline for better performance when setting multiple operations
	pipe := orc.client.Pipeline()
	pipe.Set(orc.ctx, key, value, expiration)

	_, err := pipe.Exec(orc.ctx)
	latency := time.Since(start)

	orc.updateLatencyMetrics(latency)

	if err != nil {
		orc.handleConnectionError(err)
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

// SafeDelete implements optimized cache deletion
func (orc *OptimizedRedisClient) SafeDelete(key string) error {
	if inTestMode {
		return nil
	}

	start := time.Now()
	err := orc.client.Del(orc.ctx, key).Err()
	latency := time.Since(start)

	orc.updateLatencyMetrics(latency)

	if err != nil {
		orc.handleConnectionError(err)
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	return nil
}

// GetConnectionHealth returns current Redis connection health status
func (orc *OptimizedRedisClient) GetConnectionHealth() *ConnectionHealth {
	orc.connMonitor.mutex.RLock()
	defer orc.connMonitor.mutex.RUnlock()

	return &ConnectionHealth{
		IsHealthy:        orc.connMonitor.isHealthy,
		LastHealthCheck:  orc.connMonitor.lastHealthCheck,
		ConsecutiveFails: orc.connMonitor.consecutiveFails,
		ConnectionTime:   orc.connMonitor.connectionTime,
		LastLatency:      orc.connMonitor.lastLatency,
		AvgLatency:       orc.connMonitor.avgLatency,
		RequestCount:     orc.connMonitor.requestCount,
	}
}

// ConnectionHealth represents Redis connection health status
type ConnectionHealth struct {
	IsHealthy        bool          `json:"is_healthy"`
	LastHealthCheck  time.Time     `json:"last_health_check"`
	ConsecutiveFails int           `json:"consecutive_fails"`
	ConnectionTime   time.Duration `json:"connection_time"`
	LastLatency      time.Duration `json:"last_latency"`
	AvgLatency       time.Duration `json:"avg_latency"`
	RequestCount     int64         `json:"request_count"`
}

// healthMonitorLoop continuously monitors Redis connection health
func (orc *OptimizedRedisClient) healthMonitorLoop() {
	ticker := time.NewTicker(orc.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			orc.performHealthCheck()
		case <-orc.ctx.Done():
			return
		}
	}
}

// performHealthCheck executes Redis health check with timeout
func (orc *OptimizedRedisClient) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	_, err := orc.client.Ping(ctx).Result()
	latency := time.Since(start)

	orc.connMonitor.mutex.Lock()
	defer orc.connMonitor.mutex.Unlock()

	orc.connMonitor.lastHealthCheck = time.Now()

	if err != nil {
		orc.connMonitor.consecutiveFails++
		if orc.connMonitor.consecutiveFails >= 3 {
			orc.connMonitor.isHealthy = false
			fmt.Printf("⚠️ Redis health check failed (consecutive: %d): %v\n",
				orc.connMonitor.consecutiveFails, err)
		}
	} else {
		orc.connMonitor.consecutiveFails = 0
		orc.connMonitor.isHealthy = true
		orc.connMonitor.lastLatency = latency

		// Log if health was restored
		if !orc.connMonitor.isHealthy {
			fmt.Printf("✅ Redis health restored (latency: %v)\n", latency)
		}
	}
}

// updateLatencyMetrics updates performance tracking metrics
func (orc *OptimizedRedisClient) updateLatencyMetrics(latency time.Duration) {
	orc.connMonitor.mutex.Lock()
	defer orc.connMonitor.mutex.Unlock()

	orc.connMonitor.requestCount++
	orc.connMonitor.lastLatency = latency

	// Calculate rolling average latency
	if orc.connMonitor.avgLatency == 0 {
		orc.connMonitor.avgLatency = latency
	} else {
		// Exponential moving average with 0.1 alpha
		orc.connMonitor.avgLatency = time.Duration(
			0.9*float64(orc.connMonitor.avgLatency) + 0.1*float64(latency),
		)
	}

	// Alert on high latency
	if latency > 100*time.Millisecond {
		fmt.Printf("⚠️ High Redis latency detected: %v (avg: %v)\n",
			latency, orc.connMonitor.avgLatency)
	}
}

// handleConnectionError handles Redis connection errors
func (orc *OptimizedRedisClient) handleConnectionError(err error) {
	if IsConnectionError(err) {
		orc.connMonitor.mutex.Lock()
		orc.connMonitor.consecutiveFails++
		if orc.connMonitor.consecutiveFails >= 3 {
			orc.connMonitor.isHealthy = false
		}
		orc.connMonitor.mutex.Unlock()

		fmt.Printf("🔴 Redis connection error detected: %v\n", err)
		metrics.IncrementCounter("redis_connection_errors")
	}
}

// GetPoolStats returns Redis connection pool statistics
func (orc *OptimizedRedisClient) GetPoolStats() *redis.PoolStats {
	if orc.client == nil {
		return nil
	}
	return orc.client.PoolStats()
}

// Helper functions for environment variable parsing
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
