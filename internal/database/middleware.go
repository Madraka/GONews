package database

import (
	"log"
	"time"

	"news/internal/metrics"

	"gorm.io/gorm"
)

// MetricsMiddleware implements gorm callbacks for tracking metrics
type MetricsMiddleware struct{}

// Register registers callbacks for database operations
func (m *MetricsMiddleware) Register(db *gorm.DB) {
	// Register callbacks for different operations
	callback := db.Callback()

	if err := callback.Create().Before("gorm:create").Register("metrics:before_create", m.beforeCreate); err != nil {
		log.Printf("Failed to register create before callback: %v", err)
	}
	if err := callback.Create().After("gorm:create").Register("metrics:after_create", m.afterCreate); err != nil {
		log.Printf("Failed to register create after callback: %v", err)
	}

	if err := callback.Query().Before("gorm:query").Register("metrics:before_query", m.beforeQuery); err != nil {
		log.Printf("Failed to register query before callback: %v", err)
	}
	if err := callback.Query().After("gorm:query").Register("metrics:after_query", m.afterQuery); err != nil {
		log.Printf("Failed to register query after callback: %v", err)
	}

	if err := callback.Update().Before("gorm:update").Register("metrics:before_update", m.beforeUpdate); err != nil {
		log.Printf("Failed to register update before callback: %v", err)
	}
	if err := callback.Update().After("gorm:update").Register("metrics:after_update", m.afterUpdate); err != nil {
		log.Printf("Failed to register update after callback: %v", err)
	}

	if err := callback.Delete().Before("gorm:delete").Register("metrics:before_delete", m.beforeDelete); err != nil {
		log.Printf("Failed to register delete before callback: %v", err)
	}
	if err := callback.Delete().After("gorm:delete").Register("metrics:after_delete", m.afterDelete); err != nil {
		log.Printf("Failed to register delete after callback: %v", err)
	}
}

// Helper function to store start time in context
func (m *MetricsMiddleware) startTimer(db *gorm.DB, operation string) {
	db.InstanceSet("metrics:start_time", time.Now())
	db.InstanceSet("metrics:operation", operation)
}

// Helper function to calculate duration and report metrics
func (m *MetricsMiddleware) reportDuration(db *gorm.DB) {
	if start, ok := db.InstanceGet("metrics:start_time"); ok {
		operation, _ := db.InstanceGet("metrics:operation")
		duration := time.Since(start.(time.Time)).Seconds()
		metrics.DatabaseOperationDuration.WithLabelValues(operation.(string)).Observe(duration)
	}
}

// GORM callback implementations
func (m *MetricsMiddleware) beforeCreate(db *gorm.DB) {
	m.startTimer(db, "create")
}

func (m *MetricsMiddleware) afterCreate(db *gorm.DB) {
	m.reportDuration(db)
}

func (m *MetricsMiddleware) beforeQuery(db *gorm.DB) {
	m.startTimer(db, "query")
}

func (m *MetricsMiddleware) afterQuery(db *gorm.DB) {
	m.reportDuration(db)
}

func (m *MetricsMiddleware) beforeUpdate(db *gorm.DB) {
	m.startTimer(db, "update")
}

func (m *MetricsMiddleware) afterUpdate(db *gorm.DB) {
	m.reportDuration(db)
}

func (m *MetricsMiddleware) beforeDelete(db *gorm.DB) {
	m.startTimer(db, "delete")
}

func (m *MetricsMiddleware) afterDelete(db *gorm.DB) {
	m.reportDuration(db)
}

// SetupMetricsMiddleware registers metrics middleware with the database
func SetupMetricsMiddleware(db *gorm.DB) {
	middleware := &MetricsMiddleware{}
	middleware.Register(db)
}
