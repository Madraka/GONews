package organized

import (
	"fmt"
	"news/internal/database"
	"news/internal/models"

	"github.com/jmoiron/sqlx"
)

// SeedSystemSettings seeds essential system settings
// Priority: 1 (Highest) - No dependencies
func SeedSystemSettings(db *sqlx.DB) error {
	fmt.Println("⚙️  [01] Seeding system settings...")

	var count int64
	database.DB.Model(&models.Setting{}).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️  System settings already exist (%d found), skipping...\n", count)
		return nil
	}

	settings := []models.Setting{
		// Site Configuration
		{Key: "site_name", Value: "News API", Type: "string", Description: "Site name", Group: "general", IsPublic: true},
		{Key: "site_description", Value: "Modern news API with multi-language support", Type: "string", Description: "Site description", Group: "general", IsPublic: true},
		{Key: "site_url", Value: "https://newsapi.dev", Type: "string", Description: "Site URL", Group: "general", IsPublic: true},
		{Key: "site_logo", Value: "/images/logo.png", Type: "string", Description: "Site logo path", Group: "general", IsPublic: true},
		{Key: "site_favicon", Value: "/images/favicon.ico", Type: "string", Description: "Site favicon path", Group: "general", IsPublic: true},

		// Content Settings
		{Key: "default_language", Value: "tr", Type: "string", Description: "Default language code", Group: "content", IsPublic: true},
		{Key: "supported_languages", Value: "tr,en,es,fr,de", Type: "string", Description: "Comma-separated supported languages", Group: "content", IsPublic: true},
		{Key: "articles_per_page", Value: "10", Type: "integer", Description: "Default articles per page", Group: "content", IsPublic: true},
		{Key: "auto_publish", Value: "false", Type: "boolean", Description: "Auto-publish articles", Group: "content", IsPublic: false},
		{Key: "enable_comments", Value: "true", Type: "boolean", Description: "Enable article comments", Group: "content", IsPublic: true},

		// Media Settings
		{Key: "max_upload_size", Value: "10485760", Type: "integer", Description: "Max upload size in bytes (10MB)", Group: "media", IsPublic: false},
		{Key: "allowed_file_types", Value: "jpg,jpeg,png,gif,webp,pdf,doc,docx", Type: "string", Description: "Allowed file types", Group: "media", IsPublic: false},
		{Key: "image_quality", Value: "85", Type: "integer", Description: "Image compression quality", Group: "media", IsPublic: false},

		// SEO Settings
		{Key: "seo_title_suffix", Value: " | News API", Type: "string", Description: "SEO title suffix", Group: "seo", IsPublic: true},
		{Key: "default_meta_description", Value: "Latest news and updates from News API", Type: "string", Description: "Default meta description", Group: "seo", IsPublic: true},
		{Key: "robots_txt", Value: "User-agent: *\nDisallow: /admin/\nSitemap: /sitemap.xml", Type: "text", Description: "Robots.txt content", Group: "seo", IsPublic: true},

		// Social Media
		{Key: "twitter_handle", Value: "@newsapi", Type: "string", Description: "Twitter handle", Group: "social", IsPublic: true},
		{Key: "facebook_page", Value: "https://facebook.com/newsapi", Type: "string", Description: "Facebook page URL", Group: "social", IsPublic: true},
		{Key: "instagram_handle", Value: "@newsapi", Type: "string", Description: "Instagram handle", Group: "social", IsPublic: true},

		// API Settings
		{Key: "api_rate_limit", Value: "1000", Type: "integer", Description: "API rate limit per hour", Group: "api", IsPublic: false},
		{Key: "api_cache_ttl", Value: "300", Type: "integer", Description: "API cache TTL in seconds", Group: "api", IsPublic: false},
		{Key: "enable_api_docs", Value: "true", Type: "boolean", Description: "Enable API documentation", Group: "api", IsPublic: true},

		// Email Settings
		{Key: "smtp_host", Value: "localhost", Type: "string", Description: "SMTP host", Group: "email", IsPublic: false},
		{Key: "smtp_port", Value: "587", Type: "integer", Description: "SMTP port", Group: "email", IsPublic: false},
		{Key: "from_email", Value: "noreply@newsapi.dev", Type: "string", Description: "From email address", Group: "email", IsPublic: false},
		{Key: "from_name", Value: "News API", Type: "string", Description: "From name", Group: "email", IsPublic: false},

		// Security Settings
		{Key: "jwt_expiry_hours", Value: "24", Type: "integer", Description: "JWT token expiry in hours", Group: "security", IsPublic: false},
		{Key: "password_min_length", Value: "8", Type: "integer", Description: "Minimum password length", Group: "security", IsPublic: false},
		{Key: "enable_2fa", Value: "false", Type: "boolean", Description: "Enable 2FA", Group: "security", IsPublic: false},
		{Key: "session_timeout", Value: "3600", Type: "integer", Description: "Session timeout in seconds", Group: "security", IsPublic: false},

		// Analytics Settings
		{Key: "google_analytics_id", Value: "", Type: "string", Description: "Google Analytics ID", Group: "analytics", IsPublic: true},
		{Key: "enable_analytics", Value: "true", Type: "boolean", Description: "Enable analytics", Group: "analytics", IsPublic: true},

		// Maintenance
		{Key: "maintenance_mode", Value: "false", Type: "boolean", Description: "Maintenance mode", Group: "maintenance", IsPublic: true},
		{Key: "maintenance_message", Value: "Site is under maintenance. Please check back later.", Type: "text", Description: "Maintenance message", Group: "maintenance", IsPublic: true},
	}

	for _, setting := range settings {
		if err := database.DB.Create(&setting).Error; err != nil {
			return fmt.Errorf("failed to create setting %s: %w", setting.Key, err)
		}
	}

	fmt.Printf("✅ [01] Created %d system settings\n", len(settings))
	return nil
}
