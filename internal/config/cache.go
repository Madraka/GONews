package config

import (
	"os"
	"strconv"
	"time"
)

// CacheConfig holds comprehensive cache configuration
type CacheConfig struct {
	// L1 (Ristretto) Cache Settings
	L1DefaultTTL   time.Duration
	L1MaxCostRatio float64

	// L2 (Redis) Cache Settings
	L2DefaultTTL time.Duration
	L2LongTTL    time.Duration

	// Singleflight Settings
	EnableSingleflight bool
	SingleflightTTL    time.Duration

	// Circuit Breaker Settings
	CircuitBreakerFailureThreshold int
	CircuitBreakerSuccessThreshold int
	CircuitBreakerTimeout          time.Duration

	// Health Monitoring Settings
	HealthCheckInterval time.Duration
	MaxFailureRate      float64
}

// GetCacheConfig returns cache configuration from environment variables
func GetCacheConfig() *CacheConfig {
	return &CacheConfig{
		// L1 Cache (Ristretto) configuration
		L1DefaultTTL:   getEnvDuration("CACHE_L1_DEFAULT_TTL", 5*time.Minute),
		L1MaxCostRatio: getEnvFloat("CACHE_L1_MAX_COST_RATIO", 0.8),

		// L2 Cache (Redis) configuration
		L2DefaultTTL: getEnvDuration("CACHE_L2_DEFAULT_TTL", 1*time.Hour),
		L2LongTTL:    getEnvDuration("CACHE_L2_LONG_TTL", 24*time.Hour),

		// Singleflight configuration
		EnableSingleflight: getEnvBool("CACHE_ENABLE_SINGLEFLIGHT", true),
		SingleflightTTL:    getEnvDuration("CACHE_SINGLEFLIGHT_TTL", 10*time.Second),

		// Circuit Breaker configuration
		CircuitBreakerFailureThreshold: getEnvInt("CACHE_CIRCUIT_BREAKER_FAILURE_THRESHOLD", 5),
		CircuitBreakerSuccessThreshold: getEnvInt("CACHE_CIRCUIT_BREAKER_SUCCESS_THRESHOLD", 3),
		CircuitBreakerTimeout:          getEnvDuration("CACHE_CIRCUIT_BREAKER_TIMEOUT", 1*time.Second),

		// Health monitoring configuration
		HealthCheckInterval: getEnvDuration("CACHE_HEALTH_CHECK_INTERVAL", 30*time.Second),
		MaxFailureRate:      getEnvFloat("CACHE_MAX_FAILURE_RATE", 0.05),
	}
}

// Helper functions to get environment variables with defaults
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
