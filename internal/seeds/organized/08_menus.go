package organized

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedMenus creates navigation menus for the website
func SeedMenus(db *sqlx.DB) error {
	fmt.Println("🧭 Seeding menus...")

	// Main Navigation Menu
	mainMenuItems := []map[string]interface{}{
		{
			"title":       "Home",
			"url":         "/",
			"category_id": nil,
			"sort_order":  1,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Politics",
			"url":         "/category/politics",
			"category_id": 1, // Politics category
			"sort_order":  2,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Technology",
			"url":         "/category/technology",
			"category_id": 2, // Technology category
			"sort_order":  3,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Business",
			"url":         "/category/business",
			"category_id": 3, // Business category
			"sort_order":  4,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Sports",
			"url":         "/category/sports",
			"category_id": 4, // Sports category
			"sort_order":  5,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Health",
			"url":         "/category/health",
			"category_id": 5, // Health category
			"sort_order":  6,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "More",
			"url":         "#",
			"category_id": nil,
			"sort_order":  7,
			"is_active":   true,
			"parent_id":   nil,
		},
	}

	// More submenu items
	moreSubmenuItems := []map[string]interface{}{
		{
			"title":       "Entertainment",
			"url":         "/category/entertainment",
			"category_id": 6, // Entertainment category
			"sort_order":  1,
			"is_active":   true,
			"parent_id":   7, // More menu item
		},
		{
			"title":       "Education",
			"url":         "/category/education",
			"category_id": 7, // Education category
			"sort_order":  2,
			"is_active":   true,
			"parent_id":   7,
		},
		{
			"title":       "Environment",
			"url":         "/category/environment",
			"category_id": 8, // Environment category
			"sort_order":  3,
			"is_active":   true,
			"parent_id":   7,
		},
		{
			"title":       "Travel",
			"url":         "/category/travel",
			"category_id": 9, // Travel category
			"sort_order":  4,
			"is_active":   true,
			"parent_id":   7,
		},
		{
			"title":       "Lifestyle",
			"url":         "/category/lifestyle",
			"category_id": 10, // Lifestyle category
			"sort_order":  5,
			"is_active":   true,
			"parent_id":   7,
		},
		{
			"title":       "Opinion",
			"url":         "/category/opinion",
			"category_id": 11, // Opinion category
			"sort_order":  6,
			"is_active":   true,
			"parent_id":   7,
		},
		{
			"title":       "Archive",
			"url":         "/archive",
			"category_id": nil,
			"sort_order":  7,
			"is_active":   true,
			"parent_id":   7,
		},
	}

	// Footer Menu Items
	footerMenuItems := []map[string]interface{}{
		{
			"title":       "About Us",
			"url":         "/about",
			"category_id": nil,
			"sort_order":  1,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Contact",
			"url":         "/contact",
			"category_id": nil,
			"sort_order":  2,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Privacy Policy",
			"url":         "/privacy",
			"category_id": nil,
			"sort_order":  3,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Terms of Service",
			"url":         "/terms",
			"category_id": nil,
			"sort_order":  4,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "FAQ",
			"url":         "/faq",
			"category_id": nil,
			"sort_order":  5,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Editorial Guidelines",
			"url":         "/editorial-guidelines",
			"category_id": nil,
			"sort_order":  6,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Advertise",
			"url":         "/advertise",
			"category_id": nil,
			"sort_order":  7,
			"is_active":   true,
			"parent_id":   nil,
		},
		{
			"title":       "Newsletter",
			"url":         "/newsletter",
			"category_id": nil,
			"sort_order":  8,
			"is_active":   true,
			"parent_id":   nil,
		},
	}

	// Social Media Menu Items
	socialMenuItems := []map[string]interface{}{
		{
			"title":       "Facebook",
			"url":         "https://facebook.com/newssite",
			"category_id": nil,
			"sort_order":  1,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "facebook",
			"target":      "_blank",
		},
		{
			"title":       "Twitter",
			"url":         "https://twitter.com/newssite",
			"category_id": nil,
			"sort_order":  2,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "twitter",
			"target":      "_blank",
		},
		{
			"title":       "Instagram",
			"url":         "https://instagram.com/newssite",
			"category_id": nil,
			"sort_order":  3,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "instagram",
			"target":      "_blank",
		},
		{
			"title":       "LinkedIn",
			"url":         "https://linkedin.com/company/newssite",
			"category_id": nil,
			"sort_order":  4,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "linkedin",
			"target":      "_blank",
		},
		{
			"title":       "YouTube",
			"url":         "https://youtube.com/newssite",
			"category_id": nil,
			"sort_order":  5,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "youtube",
			"target":      "_blank",
		},
		{
			"title":       "RSS Feed",
			"url":         "/rss.xml",
			"category_id": nil,
			"sort_order":  6,
			"is_active":   true,
			"parent_id":   nil,
			"icon":        "rss",
			"target":      "_self",
		},
	}

	// Create menus
	menus := []map[string]interface{}{
		{
			"name":      "Main Navigation",
			"slug":      "main-nav",
			"location":  "header",
			"is_active": true,
		},
		{
			"name":      "Footer Links",
			"slug":      "footer-links",
			"location":  "footer",
			"is_active": true,
		},
		{
			"name":      "Social Media",
			"slug":      "social-media",
			"location":  "social",
			"is_active": true,
		},
	}

	var menuIDs []int

	// Check and insert menus
	for _, menu := range menus {
		var menuID int
		// First check if menu already exists
		err := db.Get(&menuID, "SELECT id FROM menus WHERE slug = $1", menu["slug"])
		if err == nil {
			// Menu already exists, use existing ID
			menuIDs = append(menuIDs, menuID)
			fmt.Printf("   ○ Menu already exists: %s (ID: %d)\n", menu["name"], menuID)
			continue
		}

		// Menu doesn't exist, create it
		query := `
			INSERT INTO menus (
				name, slug, location, is_active, created_at, updated_at
			) VALUES (
				:name, :slug, :location, :is_active, NOW(), NOW()
			) RETURNING id`

		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("error preparing menu query: %v", err)
		}
		defer func() {
			if closeErr := stmt.Close(); closeErr != nil {
				fmt.Printf("Warning: Error closing statement: %v\n", closeErr)
			}
		}()

		err = stmt.Get(&menuID, menu)
		if err != nil {
			return fmt.Errorf("error inserting menu '%s': %v", menu["name"], err)
		}

		menuIDs = append(menuIDs, menuID)
		fmt.Printf("   ✓ Created menu: %s (ID: %d)\n", menu["name"], menuID)
	}

	// Insert menu items for Main Navigation
	allMainItems := append(mainMenuItems, moreSubmenuItems...)
	err := insertMenuItems(db, menuIDs[0], allMainItems, "main navigation")
	if err != nil {
		return err
	}

	// Insert menu items for Footer Links
	err = insertMenuItems(db, menuIDs[1], footerMenuItems, "footer links")
	if err != nil {
		return err
	}

	// Insert menu items for Social Media
	err = insertMenuItems(db, menuIDs[2], socialMenuItems, "social media")
	if err != nil {
		return err
	}

	fmt.Printf("✅ Successfully seeded %d menus with navigation items\n", len(menus))
	return nil
}

// insertMenuItems is a helper function to insert menu items
func insertMenuItems(db *sqlx.DB, menuID int, items []map[string]interface{}, menuType string) error {
	// First check if menu items already exist for this menu
	var existingCount int
	err := db.Get(&existingCount, "SELECT COUNT(*) FROM menu_items WHERE menu_id = $1", menuID)
	if err != nil {
		return fmt.Errorf("error checking existing menu items: %v", err)
	}

	if existingCount > 0 {
		fmt.Printf("     ○ Menu items already exist for %s (%d items), skipping\n", menuType, existingCount)
		return nil
	}

	for _, item := range items {
		// Add menu_id to the item
		item["menu_id"] = menuID

		// Set default values for optional fields if they don't exist
		if _, exists := item["icon"]; !exists {
			item["icon"] = nil
		}
		if _, exists := item["target"]; !exists {
			item["target"] = "_self"
		}

		query := `
			INSERT INTO menu_items (
				menu_id, title, url, category_id, sort_order, is_active, 
				parent_id, icon, target, created_at, updated_at
			) VALUES (
				:menu_id, :title, :url, :category_id, :sort_order, :is_active,
				:parent_id, :icon, :target, NOW(), NOW()
			) RETURNING id`

		var itemID int
		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("error preparing menu item query: %v", err)
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Printf("Warning: Failed to close prepared statement: %v", err)
			}
		}()

		err = stmt.Get(&itemID, item)
		if err != nil {
			return fmt.Errorf("error inserting menu item '%s' for %s: %v", item["title"], menuType, err)
		}

		// Update parent_id for submenu items (they reference the item ID, not the original parent_id)
		if item["parent_id"] != nil && item["parent_id"] == 7 {
			// Find the "More" menu item ID and update submenu items
			var moreItemID int
			err = db.Get(&moreItemID, `
				SELECT id FROM menu_items 
				WHERE menu_id = $1 AND title = 'More' AND parent_id IS NULL
			`, menuID)
			if err == nil {
				_, err = db.Exec(`
					UPDATE menu_items 
					SET parent_id = $1 
					WHERE id = $2
				`, moreItemID, itemID)
				if err != nil {
					return fmt.Errorf("error updating parent_id for submenu item: %v", err)
				}
			}
		}

		fmt.Printf("     • Added item: %s\n", item["title"])
	}

	return nil
}
