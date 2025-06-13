package organized

import (
	"fmt"
	"news/internal/database"
	"news/internal/models"

	"github.com/jmoiron/sqlx"
)

// SeedTags seeds essential tags for content organization
// Priority: 4 - Depends on: users
func SeedTags(db *sqlx.DB) error {
	fmt.Println("🏷️  [04] Seeding tags...")

	var count int64
	database.DB.Model(&models.Tag{}).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️  Tags already exist (%d found), skipping...\n", count)
		return nil
	}

	tags := []models.Tag{
		// Breaking and Urgent Tags
		{Name: "breaking", Slug: "breaking", Description: "Breaking news and urgent updates", Color: "#dc3545"},
		{Name: "urgent", Slug: "urgent", Description: "Urgent news requiring immediate attention", Color: "#dc3545"},
		{Name: "developing", Slug: "developing", Description: "Developing stories with ongoing updates", Color: "#fd7e14"},
		{Name: "alert", Slug: "alert", Description: "News alerts and important announcements", Color: "#ff6b6b"},

		// Content Type Tags
		{Name: "analysis", Slug: "analysis", Description: "In-depth analysis and commentary", Color: "#6f42c1"},
		{Name: "exclusive", Slug: "exclusive", Description: "Exclusive reports and investigations", Color: "#fd7e14"},
		{Name: "interview", Slug: "interview", Description: "Interviews with newsmakers", Color: "#20c997"},
		{Name: "investigation", Slug: "investigation", Description: "Investigative journalism", Color: "#e83e8c"},
		{Name: "opinion", Slug: "opinion", Description: "Opinion pieces and editorials", Color: "#6610f2"},
		{Name: "review", Slug: "review", Description: "Reviews and evaluations", Color: "#17a2b8"},
		{Name: "special-report", Slug: "special-report", Description: "Special reports and long-form journalism", Color: "#6f42c1"},

		// Media and Format Tags
		{Name: "video", Slug: "video", Description: "Video content and multimedia reports", Color: "#007bff"},
		{Name: "live", Slug: "live", Description: "Live coverage and real-time updates", Color: "#28a745"},
		{Name: "podcast", Slug: "podcast", Description: "Podcast episodes and audio content", Color: "#795548"},
		{Name: "gallery", Slug: "gallery", Description: "Photo galleries and image collections", Color: "#ff9800"},
		{Name: "infographic", Slug: "infographic", Description: "Infographics and data visualizations", Color: "#9c27b0"},

		// Geographic Tags
		{Name: "international", Slug: "international", Description: "International and global news", Color: "#495057"},
		{Name: "local", Slug: "local", Description: "Local and community news", Color: "#6c757d"},
		{Name: "national", Slug: "national", Description: "National news and domestic affairs", Color: "#343a40"},
		{Name: "regional", Slug: "regional", Description: "Regional news and developments", Color: "#868e96"},

		// Popularity and Engagement Tags
		{Name: "trending", Slug: "trending", Description: "Trending topics and viral news", Color: "#17a2b8"},
		{Name: "featured", Slug: "featured", Description: "Featured stories and highlights", Color: "#ffc107"},
		{Name: "top-story", Slug: "top-story", Description: "Top stories of the day", Color: "#fd7e14"},
		{Name: "most-read", Slug: "most-read", Description: "Most popular and widely read articles", Color: "#28a745"},
		{Name: "viral", Slug: "viral", Description: "Viral content and social media buzz", Color: "#e91e63"},

		// Update and Status Tags
		{Name: "update", Slug: "update", Description: "Updates to ongoing stories", Color: "#17a2b8"},
		{Name: "correction", Slug: "correction", Description: "Corrections and clarifications", Color: "#6c757d"},
		{Name: "follow-up", Slug: "follow-up", Description: "Follow-up stories and continued coverage", Color: "#495057"},
		{Name: "recap", Slug: "recap", Description: "Story recaps and summaries", Color: "#868e96"},

		// Seasonal and Event Tags
		{Name: "election", Slug: "election", Description: "Election coverage and political campaigns", Color: "#dc3545"},
		{Name: "crisis", Slug: "crisis", Description: "Crisis coverage and emergency situations", Color: "#dc3545"},
		{Name: "achievement", Slug: "achievement", Description: "Achievements and success stories", Color: "#28a745"},
		{Name: "milestone", Slug: "milestone", Description: "Important milestones and anniversaries", Color: "#6610f2"},
		{Name: "innovation", Slug: "innovation", Description: "Innovation and technological breakthroughs", Color: "#007bff"},

		// Audience and Interest Tags
		{Name: "youth", Slug: "youth", Description: "News relevant to young audiences", Color: "#e91e63"},
		{Name: "seniors", Slug: "seniors", Description: "News relevant to senior citizens", Color: "#795548"},
		{Name: "women", Slug: "women", Description: "Women-focused news and gender issues", Color: "#e91e63"},
		{Name: "family", Slug: "family", Description: "Family-oriented news and parenting", Color: "#ff9800"},
		{Name: "student", Slug: "student", Description: "Student life and educational matters", Color: "#3f51b5"},

		// Financial and Economic Tags
		{Name: "market", Slug: "market", Description: "Stock market and financial news", Color: "#4caf50"},
		{Name: "economy", Slug: "economy", Description: "Economic indicators and financial analysis", Color: "#2196f3"},
		{Name: "startup", Slug: "startup", Description: "Startup news and entrepreneurship", Color: "#ff5722"},
		{Name: "crypto", Slug: "crypto", Description: "Cryptocurrency and blockchain news", Color: "#ffc107"},

		// Technology Specific Tags
		{Name: "ai", Slug: "ai", Description: "Artificial Intelligence developments", Color: "#9c27b0"},
		{Name: "cybersecurity", Slug: "cybersecurity", Description: "Cybersecurity threats and solutions", Color: "#f44336"},
		{Name: "mobile", Slug: "mobile", Description: "Mobile technology and apps", Color: "#03dac6"},
		{Name: "gadgets", Slug: "gadgets", Description: "New gadgets and consumer electronics", Color: "#ff6f00"},
	}

	for _, tag := range tags {
		if err := database.DB.Create(&tag).Error; err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tag.Name, err)
		}
	}

	fmt.Printf("✅ [04] Created %d tags\n", len(tags))
	return nil
}
