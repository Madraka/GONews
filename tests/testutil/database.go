package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB holds test database configuration
type TestDB struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

// SetupTestDB creates and returns a test database connection
func SetupTestDB(t *testing.T) *TestDB {
	dsn := GetTestDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Silent mode for tests
				Colorful:      false,         // Disable color
			},
		),
	})
	require.NoError(t, err, "Failed to connect to test database")

	sqlDB, err := db.DB()
	require.NoError(t, err, "Failed to get underlying sql.DB")

	// Test the connection
	err = sqlDB.Ping()
	require.NoError(t, err, "Failed to ping test database")

	return &TestDB{
		DB:    db,
		sqlDB: sqlDB,
	}
}

// Close closes the database connection
func (tdb *TestDB) Close() error {
	if tdb.sqlDB != nil {
		return tdb.sqlDB.Close()
	}
	return nil
}

// Cleanup performs database cleanup for tests
func (tdb *TestDB) Cleanup(t *testing.T) {
	// Clear test data - in reverse order of dependencies
	tables := []string{
		"translations",
		"user_roles",
		"articles",
		"categories",
		"users",
	}

	for _, table := range tables {
		result := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if result.Error != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, result.Error)
		}
	}
}

// GetTestDSN returns the test database connection string
func GetTestDSN() string {
	// Check if DATABASE_URL is set (for Docker testing)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return databaseURL
	}

	// Fallback to environment variables or defaults
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5434")             // Docker test port
	user := getEnv("TEST_DB_USER", "testuser")         // Docker test user
	password := getEnv("TEST_DB_PASSWORD", "testpass") // Docker test password
	dbname := getEnv("TEST_DB_NAME", "newsapi_test")   // Docker test database
	sslmode := getEnv("TEST_DB_SSLMODE", "disable")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
