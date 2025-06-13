package database

import (
	"log"
	"os"
	"time"

	"news/internal/metrics"
	"news/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

var DB *gorm.DB

// MigrationMode represents the migration strategy to use
type MigrationMode string

const (
	// MigrationModeAuto uses GORM's built-in migration (development only)
	MigrationModeAuto MigrationMode = "auto"
	// MigrationModeAtlas uses Atlas for migrations (production recommended)
	MigrationModeAtlas MigrationMode = "atlas"
	// MigrationModeNone disables migrations entirely
	MigrationModeNone MigrationMode = "none"
)

func Connect() {
	// Track database connection time
	defer metrics.TrackDatabaseOperation("connect")()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	// Create a custom GORM logger for metrics
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true, // Enable prepared statement caching for better performance
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Add metrics tracking to GORM
	if err := db.Use(tracing.NewPlugin()); err != nil {
		log.Printf("Failed to add tracing plugin: %v", err)
	}

	// Register metrics middleware
	SetupMetricsMiddleware(db)

	// Configure connection pool for better performance under load
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get underlying sql.DB: %v", err)
	} else {
		// Connection pool settings optimized for load testing
		sqlDB.SetMaxIdleConns(25)                  // Increased from default 2
		sqlDB.SetMaxOpenConns(100)                 // Increased from default unlimited
		sqlDB.SetConnMaxLifetime(time.Hour)        // Close connections after 1 hour
		sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 minutes
		log.Println("Database connection pool optimized for high load")
	}

	DB = db

	log.Println("Database connection established.")

	// Print migration configuration info
	PrintMigrationInfo()
}

// GetMigrationMode returns the configured migration mode
func GetMigrationMode() MigrationMode {
	mode := os.Getenv("DB_MIGRATION_MODE")

	switch mode {
	case "auto":
		return MigrationModeAuto
	case "atlas":
		return MigrationModeAtlas
	case "none":
		return MigrationModeNone
	default:
		// Default to AutoMigrate in development, Atlas in production
		env := os.Getenv("APP_ENV")
		if env == "development" || env == "dev" {
			return MigrationModeAuto
		}
		return MigrationModeAtlas
	}
}

// RunMigrations executes database migrations based on the configured mode
func RunMigrations() {
	mode := GetMigrationMode()

	switch mode {
	case MigrationModeAuto:
		log.Println("🔄 Running GORM AutoMigrate (Development Mode)")
		AutoMigrateModels()
	case MigrationModeAtlas:
		log.Println("🎯 Atlas migrations enabled")
		log.Println("💡 For development: atlas migrate apply --env dev")
		log.Println("💡 For production: atlas migrate apply --env prod")
		log.Println("💡 Check status: atlas migrate status --env dev")
		log.Println("💡 Create new migration: atlas migrate diff --env dev")
		// TODO: Implement automatic Atlas migration execution
		// For now, Atlas migrations must be run manually
	case MigrationModeNone:
		log.Println("⏭️  Database migrations disabled")
	default:
		log.Printf("⚠️  Unknown migration mode: %s, defaulting to Atlas", mode)
	}
}

// AutoMigrateModels runs GORM AutoMigrate for all models (development only)
func AutoMigrateModels() {
	if DB == nil {
		log.Fatal("Database connection not initialized")
	}

	log.Println("Running GORM AutoMigrate for all models...")

	err := DB.AutoMigrate(
		// Core models
		&models.User{},
		&models.Article{},
		&models.Category{},
		&models.Tag{},

		// Content & Interaction models
		&models.Comment{},
		&models.Vote{},
		&models.Bookmark{},
		&models.Follow{},
		&models.Subscription{},
		&models.UserArticleInteraction{},

		// Article related models
		&models.ArticleTranslation{},
		&models.ArticleContentBlock{},

		// System models
		&models.Newsletter{},
		&models.Notification{},
		&models.Menu{},
		&models.MenuItem{},
		&models.Setting{},
		&models.Media{},

		// Breaking news & Live news models
		&models.BreakingNewsBanner{},
		&models.LiveNewsStream{},
		&models.LiveNewsUpdate{},
		&models.NewsStory{},
		&models.StoryGroup{},
		&models.StoryGroupItem{},
		&models.StoryView{},

		// AI & Content moderation models
		&models.ContentSuggestion{},
		&models.ModerationResult{},
		&models.ContentAnalysis{},
		&models.AgentTask{},
		&models.AIUsageStats{},

		// Security models
		&models.UserSession{},
		&models.LoginAttempt{},
		&models.SecurityEvent{},
		&models.UserTOTP{},

		// Translation models
		&models.Translation{},
		&models.TranslationQueue{},
		&models.CategoryTranslation{},
		&models.TagTranslation{},
		&models.MenuTranslation{},
		&models.MenuItemTranslation{},
		&models.NotificationTranslation{},
		&models.SettingTranslation{},

		// Article analytics models
		&models.RelatedArticle{},

		// Video models
		&models.Video{},
		&models.VideoComment{},
		&models.VideoVote{},
		&models.VideoCommentVote{},
		&models.VideoView{},
		&models.VideoPlaylist{},
		&models.VideoPlaylistItem{},
		&models.VideoProcessingJob{},

		// Page System models (Modern CMS)
		&models.Page{},
		&models.PageContentBlock{},
		&models.PageTemplate{},
	)

	if err != nil {
		log.Fatalf("Failed to run GORM AutoMigrate: %v", err)
	}

	log.Println("GORM AutoMigrate completed successfully with all indexes.")
}

// AutoMigrate is the main migration function - now supports both GORM and Atlas
func AutoMigrate() {
	RunMigrations()
}

// Migrate is an alias for AutoMigrate for backward compatibility
func Migrate() {
	RunMigrations()
}

// PrintMigrationInfo displays current migration configuration
func PrintMigrationInfo() {
	mode := GetMigrationMode()

	log.Println("🎯 Database Migration Configuration")
	log.Println("===================================")
	log.Printf("Migration Mode: %s", mode)

	if mode == MigrationModeAtlas {
		log.Println("📋 Atlas Commands:")
		log.Println("Development: atlas migrate apply --env dev")
		log.Println("Production:  atlas migrate apply --env prod")
		log.Println("Status:      atlas migrate status --env dev")
		log.Println("New migration: atlas migrate diff --env dev")
	}

	if mode == MigrationModeAuto {
		log.Println("⚠️  Development Mode: Using GORM AutoMigrate")
		log.Println("💡 For production, set DB_MIGRATION_MODE=atlas")
	}
}
