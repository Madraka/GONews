package config

import (
	"os"
	"strconv"
)

// QueueConfig holds configuration for Redis queues
type QueueConfig struct {
	// Redis connection settings
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Worker settings
	TranslationWorkers int
	VideoWorkers       int
	AgentWorkers       int
	GeneralWorkers     int

	// Queue settings
	MaxRetries        int
	RetryDelay        int // seconds
	JobTimeout        int // seconds
	DeadLetterEnabled bool

	// Processing settings
	BatchSize       int
	ProcessInterval int // seconds
	HealthCheckPort int
}

// GetQueueConfig returns queue configuration from environment variables
func GetQueueConfig() *QueueConfig {
	return &QueueConfig{
		// Redis settings
		RedisAddr:     getEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnvString("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Worker counts
		TranslationWorkers: getEnvInt("QUEUE_TRANSLATION_WORKERS", 3),
		VideoWorkers:       getEnvInt("QUEUE_VIDEO_WORKERS", 2),
		AgentWorkers:       getEnvInt("QUEUE_AGENT_WORKERS", 2),
		GeneralWorkers:     getEnvInt("QUEUE_GENERAL_WORKERS", 3),

		// Queue settings
		MaxRetries:        getEnvInt("QUEUE_MAX_RETRIES", 3),
		RetryDelay:        getEnvInt("QUEUE_RETRY_DELAY", 60),
		JobTimeout:        getEnvInt("QUEUE_JOB_TIMEOUT", 300),
		DeadLetterEnabled: getEnvBool("QUEUE_DEAD_LETTER_ENABLED", true),

		// Processing settings
		BatchSize:       getEnvInt("QUEUE_BATCH_SIZE", 10),
		ProcessInterval: getEnvInt("QUEUE_PROCESS_INTERVAL", 5),
		HealthCheckPort: getEnvInt("QUEUE_HEALTH_PORT", 8081),
	}
}

// Helper functions to get environment variables with defaults
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
