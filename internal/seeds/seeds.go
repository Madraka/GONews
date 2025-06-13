package seeds

import (
	"fmt"
	"news/internal/database"
	"news/internal/seeds/organized"

	"github.com/jmoiron/sqlx"
)

// RunAllSeeds runs all seed functions in dependency order
func RunAllSeeds(env string) error {
	fmt.Println("🌱 Starting organized database seeding...")
	fmt.Println("📊 Seed execution order: System Settings → Users → Categories → Tags → Articles → Templates → Pages → Menus → Translations")

	// Get database connection
	db, err := getDatabaseConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Execute seeds in dependency order
	seeds := []struct {
		name string
		fn   func(*sqlx.DB) error
	}{
		{"System Settings", organized.SeedSystemSettings},
		{"Users", organized.SeedUsers},
		{"Categories", organized.SeedCategories},
		{"Tags", organized.SeedTags},
		{"Articles", organized.SeedArticles},
		{"Page Templates", organized.SeedPageTemplates},
		{"Pages", organized.SeedPages},
		{"Menus", organized.SeedMenus},
		{"Translations", organized.SeedTranslations},
	}

	for i, seed := range seeds {
		fmt.Printf("\n[%d/%d] Running %s seed...\n", i+1, len(seeds), seed.name)

		if err := seed.fn(db); err != nil {
			return fmt.Errorf("failed to seed %s: %w", seed.name, err)
		}
	}

	// Run environment-specific additional data
	switch env {
	case "dev":
		if err := seedDevSpecificData(db); err != nil {
			return fmt.Errorf("failed to seed dev-specific data: %w", err)
		}
	case "test":
		if err := seedTestSpecificData(db); err != nil {
			return fmt.Errorf("failed to seed test-specific data: %w", err)
		}
	case "prod":
		if err := seedProdSpecificData(db); err != nil {
			return fmt.Errorf("failed to seed prod-specific data: %w", err)
		}
	}

	fmt.Println("\n✅ All organized seeds completed successfully!")
	fmt.Println("📈 Database now contains: System Settings, Users, Categories, Tags, Articles, Templates, Pages, Menus, and Translations")

	return nil
}

// getDatabaseConnection returns the appropriate database connection
func getDatabaseConnection() (*sqlx.DB, error) {
	// Convert GORM DB to sqlx DB
	sqlDB, err := database.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	db := sqlx.NewDb(sqlDB, "postgres")
	return db, nil
}

// seedDevSpecificData adds development-specific data
func seedDevSpecificData(db *sqlx.DB) error {
	fmt.Println("\n🚀 Adding development-specific data...")

	// Add development-specific settings
	devSettings := []map[string]interface{}{
		{
			"key":         "dev_api_debug",
			"value":       "true",
			"description": "Enable API debugging in development",
			"category":    "development",
			"type":        "boolean",
		},
		{
			"key":         "dev_test_mode",
			"value":       "enabled",
			"description": "Enable test mode features",
			"category":    "development",
			"type":        "string",
		},
	}

	for _, setting := range devSettings {
		_, err := db.NamedExec(`
			INSERT INTO system_settings (key, value, description, category, type, created_at, updated_at)
			VALUES (:key, :value, :description, :category, :type, NOW(), NOW())
			ON CONFLICT (key) DO NOTHING
		`, setting)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not add dev setting %s: %v\n", setting["key"], err)
		}
	}

	fmt.Println("   ✓ Added development-specific settings")
	return nil
}

// seedTestSpecificData adds test-specific data
func seedTestSpecificData(db *sqlx.DB) error {
	fmt.Println("\n🧪 Adding test-specific data...")

	// Add test-specific settings
	testSettings := []map[string]interface{}{
		{
			"key":         "test_mode",
			"value":       "enabled",
			"description": "Enable test mode",
			"category":    "testing",
			"type":        "boolean",
		},
		{
			"key":         "test_db_cleanup",
			"value":       "auto",
			"description": "Automatic test database cleanup",
			"category":    "testing",
			"type":        "string",
		},
	}

	for _, setting := range testSettings {
		_, err := db.NamedExec(`
			INSERT INTO system_settings (key, value, description, category, type, created_at, updated_at)
			VALUES (:key, :value, :description, :category, :type, NOW(), NOW())
			ON CONFLICT (key) DO NOTHING
		`, setting)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not add test setting %s: %v\n", setting["key"], err)
		}
	}

	fmt.Println("   ✓ Added test-specific settings")
	return nil
}

// seedProdSpecificData adds production-specific data
func seedProdSpecificData(db *sqlx.DB) error {
	fmt.Println("\n🚀 Adding production-specific data...")

	// Add production-specific settings
	prodSettings := []map[string]interface{}{
		{
			"key":         "prod_monitoring",
			"value":       "enabled",
			"description": "Enable production monitoring",
			"category":    "production",
			"type":        "boolean",
		},
		{
			"key":         "prod_analytics",
			"value":       "enabled",
			"description": "Enable production analytics",
			"category":    "production",
			"type":        "boolean",
		},
	}

	for _, setting := range prodSettings {
		_, err := db.NamedExec(`
			INSERT INTO system_settings (key, value, description, category, type, created_at, updated_at)
			VALUES (:key, :value, :description, :category, :type, NOW(), NOW())
			ON CONFLICT (key) DO NOTHING
		`, setting)
		if err != nil {
			fmt.Printf("   ⚠️  Warning: Could not add prod setting %s: %v\n", setting["key"], err)
		}
	}

	fmt.Println("   ✓ Added production-specific settings")
	return nil
}

// Legacy seed functions for backward compatibility
// These will delegate to the organized seed functions

// SeedDevData seeds development-specific data (legacy)
func SeedDevData() error {
	fmt.Println("🔄 Legacy SeedDevData called - delegating to organized seeds...")
	return RunAllSeeds("dev")
}

// SeedTestData seeds test-specific data (legacy)
func SeedTestData() error {
	fmt.Println("🔄 Legacy SeedTestData called - delegating to organized seeds...")
	return RunAllSeeds("test")
}

// SeedProdData seeds production-specific data (legacy)
func SeedProdData() error {
	fmt.Println("🔄 Legacy SeedProdData called - delegating to organized seeds...")
	return RunAllSeeds("prod")
}
