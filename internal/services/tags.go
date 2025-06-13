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
	// Global cache invalidator instance for tags
	tagCacheInvalidator *cache.CacheInvalidator
)

// init initializes the cache invalidator for tags
func init() {
	tagCacheInvalidator = cache.NewCacheInvalidator()
}

const (
	tagCacheDuration = 15 * time.Minute
	tagKeyPrefix     = "tag:"
	tagsListKey      = "tags:list"
)

// GetTagsWithCache retrieves all tags with unified cache
func GetTagsWithCache(sort string, limit int) ([]models.Tag, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("%s:sort:%s:limit:%d", tagsListKey, sort, limit)

	// Try to get from unified cache first (L1: Ristretto -> L2: Redis)
	unifiedCache := cache.GetUnifiedCache()
	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var tags []models.Tag
		if err := json.Unmarshal([]byte(cachedData), &tags); err == nil {
			log.Printf("Retrieved tags from unified cache (sort: %s, limit: %d)", sort, limit)
			return tags, nil
		} else {
			log.Printf("Failed to unmarshal cached tags: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for tags, fetching from database (sort: %s, limit: %d)", sort, limit)

	var tags []models.Tag
	query := database.DB.Limit(limit)

	switch sort {
	case "usage_count":
		query = query.Order("usage_count DESC")
	default:
		query = query.Order("name ASC")
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	// Cache the result in both L1 and L2
	if cacheData, err := json.Marshal(tags); err == nil {
		// L1 cache (Ristretto): 3 minutes for hot data
		// L2 cache (Redis): 15 minutes for persistence
		l1TTL := 3 * time.Minute
		l2TTL := tagCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache tags in unified cache: %v", err)
		} else {
			log.Printf("Cached tags in unified cache (L1: %v, L2: %v)", l1TTL, l2TTL)
		}
	}

	return tags, nil
}

// GetTagBySlugWithCache retrieves a tag by slug with unified cache
func GetTagBySlugWithCache(slug string) (models.Tag, error) {
	// Create cache key
	cacheKey := tagKeyPrefix + slug

	// Try to get from unified cache first
	unifiedCache := cache.GetUnifiedCache()
	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var tag models.Tag
		if err := json.Unmarshal([]byte(cachedData), &tag); err == nil {
			log.Printf("Retrieved tag %s from unified cache", slug)
			return tag, nil
		} else {
			log.Printf("Failed to unmarshal cached tag: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for tag %s, fetching from database", slug)

	var tag models.Tag
	if err := database.DB.Where("slug = ?", slug).
		Preload("Articles", "status = ?", "published").
		First(&tag).Error; err != nil {
		return models.Tag{}, fmt.Errorf("tag not found: %w", err)
	}

	// Cache the result in both L1 and L2
	if cacheData, err := json.Marshal(tag); err == nil {
		// L1 cache (Ristretto): 8 minutes for individual tags
		// L2 cache (Redis): 15 minutes for persistence
		l1TTL := 8 * time.Minute
		l2TTL := tagCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache tag %s in unified cache: %v", slug, err)
		} else {
			log.Printf("Cached tag %s in unified cache (L1: %v, L2: %v)", slug, l1TTL, l2TTL)
		}
	}

	return tag, nil
}

// CreateTagWithCache creates a new tag with cache invalidation
func CreateTagWithCache(tag models.Tag) (models.Tag, error) {
	// Generate slug if not provided
	if tag.Slug == "" {
		tag.Slug = generateTagSlug(tag.Name)
	}

	if err := database.DB.Create(&tag).Error; err != nil {
		return models.Tag{}, fmt.Errorf("failed to create tag: %w", err)
	}

	// Use unified cache invalidation system
	if tagCacheInvalidator != nil {
		// Invalidate all tag lists
		if err := tagCacheInvalidator.InvalidateByPrefix("tags:list"); err != nil {
			log.Printf("Warning: Failed to invalidate tag lists cache after creation: %v", err)
		}

		log.Printf("Successfully invalidated tag caches after creating tag %d", tag.ID)
	}

	// Cache the new tag in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.Marshal(tag); err == nil {
		cacheKey := tagKeyPrefix + tag.Slug
		l1TTL := 8 * time.Minute
		l2TTL := tagCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache new tag in unified cache: %v", err)
		} else {
			log.Printf("Cached new tag %s in unified cache", tag.Slug)
		}
	}

	return tag, nil
}

// UpdateTagWithCache updates an existing tag with cache invalidation
func UpdateTagWithCache(id string, updateData models.Tag) (models.Tag, error) {
	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return models.Tag{}, fmt.Errorf("tag not found: %w", err)
	}

	// Store old slug for cache invalidation
	oldSlug := tag.Slug

	// Update fields
	if updateData.Name != "" {
		tag.Name = updateData.Name
	}
	if updateData.Description != "" {
		tag.Description = updateData.Description
	}
	if updateData.Color != "" {
		tag.Color = updateData.Color
	}

	if err := database.DB.Save(&tag).Error; err != nil {
		return models.Tag{}, fmt.Errorf("failed to update tag: %w", err)
	}

	// Use unified cache invalidation system
	if tagCacheInvalidator != nil {
		// Invalidate the specific tag (old slug)
		if err := tagCacheInvalidator.InvalidateByPrefix(tagKeyPrefix + oldSlug); err != nil {
			log.Printf("Warning: Failed to invalidate old tag cache after update: %v", err)
		}

		// Invalidate tag lists
		if err := tagCacheInvalidator.InvalidateByPrefix("tags:list"); err != nil {
			log.Printf("Warning: Failed to invalidate tag lists cache after update: %v", err)
		}

		log.Printf("Successfully invalidated tag caches after updating tag %s", id)
	}

	// Cache the updated tag in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.Marshal(tag); err == nil {
		cacheKey := tagKeyPrefix + tag.Slug
		l1TTL := 8 * time.Minute
		l2TTL := tagCacheDuration

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache updated tag in unified cache: %v", err)
		} else {
			log.Printf("Cached updated tag %s in unified cache", tag.Slug)
		}
	}

	return tag, nil
}

// DeleteTagWithCache deletes a tag with cache invalidation
func DeleteTagWithCache(id string) error {
	// First get the tag to check its slug before deletion
	var tag models.Tag
	if err := database.DB.First(&tag, id).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	if err := database.DB.Delete(&tag).Error; err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	// Use unified cache invalidation system
	if tagCacheInvalidator != nil {
		// Invalidate the specific tag
		if err := tagCacheInvalidator.InvalidateByPrefix(tagKeyPrefix + tag.Slug); err != nil {
			log.Printf("Warning: Failed to invalidate tag cache after deletion: %v", err)
		}

		// Invalidate tag lists
		if err := tagCacheInvalidator.InvalidateByPrefix("tags:list"); err != nil {
			log.Printf("Warning: Failed to invalidate tag lists cache after deletion: %v", err)
		}

		log.Printf("Successfully invalidated tag caches after deleting tag %s", id)
	}

	return nil
}

// Helper function to generate a slug from a string
func generateTagSlug(s string) string {
	// Simple slug generation - in production, consider using a proper slug library
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}
