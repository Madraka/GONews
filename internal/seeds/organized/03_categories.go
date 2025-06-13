package organized

import (
	"fmt"
	"news/internal/database"
	"news/internal/models"

	"github.com/jmoiron/sqlx"
)

// SeedCategories seeds essential news categories
// Priority: 3 - Depends on: users
func SeedCategories(db *sqlx.DB) error {
	fmt.Println("📂 [03] Seeding categories...")

	var count int64
	database.DB.Model(&models.Category{}).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️  Categories already exist (%d found), skipping...\n", count)
		return nil
	}

	categories := []models.Category{
		// Primary News Categories
		{
			Name:        "Politics",
			Slug:        "politics",
			Description: "Government news, policy changes, elections, and political analysis from around the world",
			Color:       "#dc3545",
			Icon:        "🏛️",
			SortOrder:   1,
			IsActive:    true,
		},
		{
			Name:        "Technology",
			Slug:        "technology",
			Description: "Latest tech news, product launches, software updates, and innovation in the digital world",
			Color:       "#007bff",
			Icon:        "💻",
			SortOrder:   2,
			IsActive:    true,
		},
		{
			Name:        "Business",
			Slug:        "business",
			Description: "Market news, corporate updates, economic analysis, and financial insights for informed decisions",
			Color:       "#6f42c1",
			Icon:        "📈",
			SortOrder:   3,
			IsActive:    true,
		},
		{
			Name:        "Sports",
			Slug:        "sports",
			Description: "Comprehensive sports coverage including scores, highlights, player news, and championship updates",
			Color:       "#28a745",
			Icon:        "⚽",
			SortOrder:   4,
			IsActive:    true,
		},
		{
			Name:        "Entertainment",
			Slug:        "entertainment",
			Description: "Movies, TV shows, celebrity news, music releases, and entertainment industry updates",
			Color:       "#fd7e14",
			Icon:        "🎬",
			SortOrder:   5,
			IsActive:    true,
		},
		{
			Name:        "Health",
			Slug:        "health",
			Description: "Medical breakthroughs, health tips, pandemic updates, and wellness advice from experts",
			Color:       "#20c997",
			Icon:        "🏥",
			SortOrder:   6,
			IsActive:    true,
		},
		{
			Name:        "Science",
			Slug:        "science",
			Description: "Scientific discoveries, research findings, space exploration, and environmental studies",
			Color:       "#6610f2",
			Icon:        "🔬",
			SortOrder:   7,
			IsActive:    true,
		},
		{
			Name:        "World News",
			Slug:        "world",
			Description: "International news, global events, foreign affairs, and cross-border developments",
			Color:       "#e83e8c",
			Icon:        "🌍",
			SortOrder:   8,
			IsActive:    true,
		},
		{
			Name:        "Local News",
			Slug:        "local",
			Description: "Community news, local events, city government, and regional developments",
			Color:       "#17a2b8",
			Icon:        "🏘️",
			SortOrder:   9,
			IsActive:    true,
		},
		{
			Name:        "Breaking News",
			Slug:        "breaking",
			Description: "Urgent news updates, emergency alerts, and developing stories that require immediate attention",
			Color:       "#ff6b6b",
			Icon:        "🚨",
			SortOrder:   10,
			IsActive:    true,
		},

		// Secondary/Specialized Categories
		{
			Name:        "Education",
			Slug:        "education",
			Description: "Educational news, university updates, research developments, and academic insights",
			Color:       "#795548",
			Icon:        "🎓",
			SortOrder:   11,
			IsActive:    true,
		},
		{
			Name:        "Environment",
			Slug:        "environment",
			Description: "Climate change, sustainability, conservation efforts, and environmental policy news",
			Color:       "#4caf50",
			Icon:        "🌱",
			SortOrder:   12,
			IsActive:    true,
		},
		{
			Name:        "Travel",
			Slug:        "travel",
			Description: "Travel news, destination guides, tourism industry updates, and cultural experiences",
			Color:       "#ff9800",
			Icon:        "✈️",
			SortOrder:   13,
			IsActive:    true,
		},
		{
			Name:        "Lifestyle",
			Slug:        "lifestyle",
			Description: "Fashion, food, home, relationships, and personal development stories",
			Color:       "#e91e63",
			Icon:        "💄",
			SortOrder:   14,
			IsActive:    true,
		},
		{
			Name:        "Opinion",
			Slug:        "opinion",
			Description: "Editorial pieces, opinion columns, analysis, and commentary from experts",
			Color:       "#9c27b0",
			Icon:        "💭",
			SortOrder:   15,
			IsActive:    true,
		},
	}

	for _, category := range categories {
		if err := database.DB.Create(&category).Error; err != nil {
			return fmt.Errorf("failed to create category %s: %w", category.Name, err)
		}
	}

	fmt.Printf("✅ [03] Created %d categories\n", len(categories))
	return nil
}
