package main

import (
	"log"
	"os"

	"news/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Check if DATABASE_URL is set
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Println("Starting database migration...")

	// Connect to database
	database.Connect()

	// Run migrations
	database.AutoMigrate()

	log.Println("Migration completed successfully!")
}
