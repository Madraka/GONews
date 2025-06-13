package organized

import (
	"fmt"
	"news/internal/database"
	"news/internal/models"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// SeedUsers seeds essential users
// Priority: 2 - Depends on: system_settings
func SeedUsers(db *sqlx.DB) error {
	fmt.Println("👥 [02] Seeding users...")

	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️  Users already exist (%d found), skipping...\n", count)
		return nil
	}

	// Hash password for all users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users := []models.User{
		{
			Username:   "admin",
			Email:      "admin@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "System",
			LastName:   "Administrator",
			Role:       "admin",
			Status:     "active",
			IsVerified: true,
			Bio:        "System administrator with full access to all features",
			Website:    "https://newsapi.dev",
			Location:   "Global",
		},
		{
			Username:   "editor",
			Email:      "editor@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "Chief",
			LastName:   "Editor",
			Role:       "editor",
			Status:     "active",
			IsVerified: true,
			Bio:        "Chief editor responsible for content quality and editorial standards",
			Website:    "https://newsapi.dev/team/editor",
			Location:   "New York, USA",
		},
		{
			Username:   "writer1",
			Email:      "writer1@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "Tech",
			LastName:   "Writer",
			Role:       "writer",
			Status:     "active",
			IsVerified: true,
			Bio:        "Technology journalist specializing in AI, blockchain, and emerging tech",
			Website:    "https://newsapi.dev/authors/tech-writer",
			Location:   "San Francisco, USA",
		},
		{
			Username:   "writer2",
			Email:      "writer2@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "Political",
			LastName:   "Correspondent",
			Role:       "writer",
			Status:     "active",
			IsVerified: true,
			Bio:        "Political correspondent covering domestic and international affairs",
			Website:    "https://newsapi.dev/authors/political-correspondent",
			Location:   "Washington DC, USA",
		},
		{
			Username:   "writer3",
			Email:      "writer3@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "Sports",
			LastName:   "Reporter",
			Role:       "writer",
			Status:     "active",
			IsVerified: true,
			Bio:        "Sports reporter covering major leagues and international competitions",
			Website:    "https://newsapi.dev/authors/sports-reporter",
			Location:   "London, UK",
		},
		{
			Username:   "guest",
			Email:      "guest@newsapi.dev",
			Password:   string(hashedPassword),
			FirstName:  "Guest",
			LastName:   "User",
			Role:       "user",
			Status:     "active",
			IsVerified: true,
			Bio:        "Guest user account for testing and demonstration purposes",
			Location:   "Global",
		},
	}

	for _, user := range users {
		if err := database.DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
	}

	fmt.Printf("✅ [02] Created %d users\n", len(users))
	return nil
}
