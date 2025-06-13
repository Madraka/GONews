package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"news/internal/cache"
	"news/internal/database"
	"news/internal/json"
	"news/internal/models"
)

var (
	// Global cache invalidator instance for categories
	categoryCacheInvalidator *cache.CacheInvalidator
)

// init initializes the cache invalidator for categories
func init() {
	categoryCacheInvalidator = cache.NewCacheInvalidator()
}

const (
	categoryCacheDuration = 20 * time.Minute
	categoryKeyPrefix     = "category:"
	categoriesListKey     = "categories:list"
)

// GetCategoriesWithCache retrieves all categories with unified cache
func GetCategoriesWithCache(hierarchical bool) ([]models.Category, error) {
	// Create cache key
	cacheKey := categoriesListKey
	if hierarchical {
		cacheKey = categoriesListKey + ":hierarchical"
	}

	// Try to get from unified cache first (L1: Ristretto -> L2: Redis)
	unifiedCache := cache.GetUnifiedCache()
	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var categories []models.Category
		if err := json.UnmarshalForCache([]byte(cachedData), &categories); err == nil {
			log.Printf("Retrieved categories from unified cache (hierarchical: %v)", hierarchical)
			return categories, nil
		} else {
			log.Printf("Failed to unmarshal cached categories: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for categories, fetching from database (hierarchical: %v)", hierarchical)

	var categories []models.Category
	query := database.DB.Where("is_active = ?", true).Order("sort_order ASC, name ASC")

	if hierarchical {
		// Get only parent categories and include children
		query = query.Where("parent_id IS NULL").Preload("Children")
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	// Cache the result in both L1 and L2
	if cacheData, err := json.MarshalForCache(categories); err == nil {
		// L1 cache (Ristretto): 5 minutes for hot data
		// L2 cache (Redis): 20 minutes for persistence
		l1TTL := 5 * time.Minute
		l2TTL := categoryCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache categories in unified cache: %v", err)
		} else {
			log.Printf("Cached categories in unified cache (L1: %v, L2: %v)", l1TTL, l2TTL)
		}
	}

	return categories, nil
}

// GetCategoryBySlugWithCache retrieves a category by slug with unified cache
func GetCategoryBySlugWithCache(slug string) (models.Category, error) {
	// Create cache key
	cacheKey := categoryKeyPrefix + slug

	// Try to get from unified cache first
	unifiedCache := cache.GetUnifiedCache()
	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var category models.Category
		if err := json.UnmarshalForCache([]byte(cachedData), &category); err == nil {
			log.Printf("Retrieved category %s from unified cache", slug)
			return category, nil
		} else {
			log.Printf("Failed to unmarshal cached category: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for category %s, fetching from database", slug)

	var category models.Category
	if err := database.DB.Where("slug = ? AND is_active = ?", slug, true).
		Preload("Children").
		Preload("Articles", "status = ?", "published").
		First(&category).Error; err != nil {
		return models.Category{}, fmt.Errorf("category not found: %w", err)
	}

	// Cache the result in both L1 and L2
	if cacheData, err := json.MarshalForCache(category); err == nil {
		// L1 cache (Ristretto): 10 minutes for individual categories
		// L2 cache (Redis): 20 minutes for persistence
		l1TTL := 10 * time.Minute
		l2TTL := categoryCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache category %s in unified cache: %v", slug, err)
		} else {
			log.Printf("Cached category %s in unified cache (L1: %v, L2: %v)", slug, l1TTL, l2TTL)
		}
	}

	return category, nil
}

// CreateCategoryWithCache creates a new category with cache invalidation
func CreateCategoryWithCache(category models.Category) (models.Category, error) {
	// Generate slug if not provided
	if category.Slug == "" {
		category.Slug = generateCategorySlug(category.Name)
	}

	if err := database.DB.Create(&category).Error; err != nil {
		return models.Category{}, fmt.Errorf("failed to create category: %w", err)
	}

	// Use unified cache invalidation system
	if categoryCacheInvalidator != nil {
		// Invalidate all category lists
		if err := categoryCacheInvalidator.InvalidateByPrefix("categories:list"); err != nil {
			log.Printf("Warning: Failed to invalidate category lists cache after creation: %v", err)
		}

		log.Printf("Successfully invalidated category caches after creating category %d", category.ID)
	}

	// Cache the new category in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.MarshalForCache(category); err == nil {
		cacheKey := categoryKeyPrefix + category.Slug
		l1TTL := 10 * time.Minute
		l2TTL := categoryCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache new category in unified cache: %v", err)
		} else {
			log.Printf("Cached new category %s in unified cache", category.Slug)
		}
	}

	return category, nil
}

// UpdateCategoryWithCache updates an existing category with cache invalidation
func UpdateCategoryWithCache(id string, updateData models.Category) (models.Category, error) {
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return models.Category{}, fmt.Errorf("category not found: %w", err)
	}

	// Store old slug for cache invalidation
	oldSlug := category.Slug

	// Update fields
	if updateData.Name != "" {
		category.Name = updateData.Name
	}
	if updateData.Description != "" {
		category.Description = updateData.Description
	}
	if updateData.Color != "" {
		category.Color = updateData.Color
	}
	if updateData.Icon != "" {
		category.Icon = updateData.Icon
	}
	category.IsActive = updateData.IsActive
	category.SortOrder = updateData.SortOrder

	if err := database.DB.Save(&category).Error; err != nil {
		return models.Category{}, fmt.Errorf("failed to update category: %w", err)
	}

	// Use unified cache invalidation system
	if categoryCacheInvalidator != nil {
		// Invalidate the specific category (old slug)
		if err := categoryCacheInvalidator.InvalidateByPrefix(categoryKeyPrefix + oldSlug); err != nil {
			log.Printf("Warning: Failed to invalidate old category cache after update: %v", err)
		}

		// Invalidate category lists
		if err := categoryCacheInvalidator.InvalidateByPrefix("categories:list"); err != nil {
			log.Printf("Warning: Failed to invalidate category lists cache after update: %v", err)
		}

		log.Printf("Successfully invalidated category caches after updating category %s", id)
	}

	// Cache the updated category in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.MarshalForCache(category); err == nil {
		cacheKey := categoryKeyPrefix + category.Slug
		l1TTL := 10 * time.Minute
		l2TTL := categoryCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache updated category in unified cache: %v", err)
		} else {
			log.Printf("Cached updated category %s in unified cache", category.Slug)
		}
	}

	return category, nil
}

// DeleteCategoryWithCache deletes a category with cache invalidation
func DeleteCategoryWithCache(id string) error {
	// First get the category to check its slug before deletion
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Use unified cache invalidation system
	if categoryCacheInvalidator != nil {
		// Invalidate the specific category
		if err := categoryCacheInvalidator.InvalidateByPrefix(categoryKeyPrefix + category.Slug); err != nil {
			log.Printf("Warning: Failed to invalidate category cache after deletion: %v", err)
		}

		// Invalidate category lists
		if err := categoryCacheInvalidator.InvalidateByPrefix("categories:list"); err != nil {
			log.Printf("Warning: Failed to invalidate category lists cache after deletion: %v", err)
		}

		log.Printf("Successfully invalidated category caches after deleting category %s", id)
	}

	return nil
}

// Helper function to generate a slug from a string
func generateCategorySlug(s string) string {
	// Simple slug generation - in production, consider using a proper slug library
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}
