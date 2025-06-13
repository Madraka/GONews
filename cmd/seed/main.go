package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"news/internal/database"
	"news/internal/seeds"
)

func main() {
	// Parse command line flags
	envFlag := flag.String("env", "dev", "Environment (dev, test, prod)")
	flag.Parse()

	env := *envFlag
	fmt.Printf("🌱 Starting organized database seeding for %s environment...\n", env)
	fmt.Println("📋 This will create: System Settings, Users, Categories, Tags, Articles, Templates, Pages, Menus, and Translations")

	// Set environment-specific database URL
	switch env {
	case "dev":
		// For development, use PostgreSQL if available, otherwise SQLite
		if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
			fmt.Printf("📡 Using existing DATABASE_URL for %s environment\n", env)
			// Use existing DATABASE_URL
		} else {
			if err := os.Setenv("DATABASE_URL", "postgres://devuser:devpass@localhost:5433/newsapi_dev?sslmode=disable"); err != nil {
				log.Printf("Warning: Failed to set DATABASE_URL: %v", err)
			}
			fmt.Printf("📡 Set default DATABASE_URL for %s environment (Docker port 5433)\n", env)
		}
	case "test":
		if err := os.Setenv("DATABASE_URL", "postgres://newsuser:newspass@localhost:5432/newsdb_test?sslmode=disable"); err != nil {
			log.Printf("Warning: Failed to set DATABASE_URL: %v", err)
		}
		fmt.Printf("📡 Set DATABASE_URL for %s environment\n", env)
	case "prod":
		if err := os.Setenv("DATABASE_URL", "postgres://produser:prodpass@prod_db:5432/newsdb_prod?sslmode=disable"); err != nil {
			log.Printf("Warning: Failed to set DATABASE_URL: %v", err)
		}
		fmt.Printf("📡 Set DATABASE_URL for %s environment\n", env)
	default:
		log.Fatalf("❌ Unknown environment: %s", env)
	}

	// Connect to database
	fmt.Println("📡 Connecting to database...")
	database.Connect()

	// Run GORM AutoMigrate
	fmt.Println("🔄 Running GORM AutoMigrate...")
	database.AutoMigrate()

	// Run organized seeds
	fmt.Println("🌱 Running organized database seeds...")
	if err := seeds.RunAllSeeds(env); err != nil {
		log.Fatalf("❌ Failed to run seeds: %v", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
}
