package organized

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedPages creates sample pages using page templates
func SeedPages(db *sqlx.DB) error {
	fmt.Println("📑 Seeding pages...")

	// Check if pages already exist
	var existingCount int
	err := db.Get(&existingCount, "SELECT COUNT(*) FROM pages")
	if err != nil {
		return fmt.Errorf("error checking existing pages: %v", err)
	}

	if existingCount > 0 {
		fmt.Printf("   ⚠️  Pages already exist (%d found), skipping seeding\n", existingCount)
		return nil
	}

	pages := []map[string]interface{}{
		{
			"title":       "Home",
			"slug":        "home",
			"template":    "Landing Page",
			"status":      "published",
			"is_homepage": true,
			"meta_desc":   "Stay informed with the latest news, breaking stories, and in-depth analysis from around the world.",
			"author_id":   1, // Admin
		},
		{
			"title":       "About Us",
			"slug":        "about",
			"template":    "About Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Learn about our mission, team, and commitment to delivering quality journalism and news coverage.",
			"author_id":   1, // Admin
		},
		{
			"title":       "Contact Us",
			"slug":        "contact",
			"template":    "Contact Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Get in touch with our editorial team. Contact information, office location, and inquiry form.",
			"author_id":   2, // Editor
		},
		{
			"title":       "News Archive",
			"slug":        "archive",
			"template":    "News Archive",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Browse our complete archive of news articles, stories, and reports organized by date and category.",
			"author_id":   2, // Editor
		},
		{
			"title":       "Privacy Policy",
			"slug":        "privacy",
			"template":    "Privacy Policy",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Our privacy policy explains how we collect, use, and protect your personal information.",
			"author_id":   1, // Admin
		},
		{
			"title":       "Terms of Service",
			"slug":        "terms",
			"template":    "Terms of Service",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Terms and conditions for using our news website and services.",
			"author_id":   1, // Admin
		},
		{
			"title":       "Frequently Asked Questions",
			"slug":        "faq",
			"template":    "FAQ Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Find answers to frequently asked questions about our news service, subscriptions, and policies.",
			"author_id":   2, // Editor
		},
		{
			"title":       "Editorial Guidelines",
			"slug":        "editorial-guidelines",
			"template":    "Default Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Our editorial standards and guidelines for journalism, fact-checking, and news reporting.",
			"author_id":   2, // Editor
		},
		{
			"title":       "Advertise With Us",
			"slug":        "advertise",
			"template":    "Default Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Advertising opportunities and media kit for businesses looking to reach our audience.",
			"author_id":   1, // Admin
		},
		{
			"title":       "Newsletter Subscription",
			"slug":        "newsletter",
			"template":    "Default Page",
			"status":      "published",
			"is_homepage": false,
			"meta_desc":   "Subscribe to our newsletter for daily news updates and exclusive content delivered to your inbox.",
			"author_id":   2, // Editor
		},
	}

	for _, page := range pages {
		query := `
			INSERT INTO pages (
				title, slug, template, status, is_homepage, 
				meta_desc, author_id, created_at, updated_at, published_at
			) VALUES (
				:title, :slug, :template, :status, :is_homepage,
				:meta_desc, :author_id, NOW(), NOW(), NOW()
			) RETURNING id`

		var pageID int
		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("error preparing page query: %v", err)
		}
		defer func() {
			if closeErr := stmt.Close(); closeErr != nil {
				fmt.Printf("Warning: Error closing statement: %v\n", closeErr)
			}
		}()

		err = stmt.Get(&pageID, page)
		if err != nil {
			return fmt.Errorf("error inserting page '%s': %v", page["title"], err)
		}

		fmt.Printf("   ✓ Created page: %s (ID: %d)\n", page["title"], pageID)
	}

	fmt.Printf("✅ Successfully seeded %d pages\n", len(pages))
	return nil
}
