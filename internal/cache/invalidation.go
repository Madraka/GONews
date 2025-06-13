package cache

import (
	"fmt"
	"strings"
	"time"

	"news/internal/metrics"
)

// CacheInvalidator handles cache invalidation strategies
type CacheInvalidator struct {
	unified *UnifiedCacheManager
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator() *CacheInvalidator {
	return &CacheInvalidator{
		unified: GetUnifiedCache(),
	}
}

// InvalidateArticle invalidates all cache entries related to a specific article
func (ci *CacheInvalidator) InvalidateArticle(articleID int64) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_article")()

	// Patterns to invalidate
	patterns := []string{
		fmt.Sprintf("articles:%d", articleID),            // Individual article
		fmt.Sprintf("articles:%d:*", articleID),          // Article with any suffix
		fmt.Sprintf("article:%d", articleID),             // Article cache key format
		fmt.Sprintf("article:%d:*", articleID),           // Article with any suffix (including :with_blocks)
		fmt.Sprintf("article:%d:with_blocks", articleID), // Explicit with_blocks cache key
		"articles:list:*",                                // All article lists
		"articles:pagination:*",                          // All paginated lists
		"articles:featured:*",                            // Featured articles
		"articles:breaking:*",                            // Breaking news
	}

	var lastError error
	for _, pattern := range patterns {
		if err := ci.unified.DeletePattern(pattern); err != nil {
			lastError = err
			fmt.Printf("Warning: Failed to invalidate pattern %s: %v\n", pattern, err)
		} else {
			fmt.Printf("DEBUG: Successfully invalidated pattern: %s\n", pattern)
		}
	}

	return lastError
}

// InvalidateCategory invalidates cache entries for a category
func (ci *CacheInvalidator) InvalidateCategory(categoryID int64) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_category")()

	patterns := []string{
		fmt.Sprintf("categories:%d", categoryID),
		fmt.Sprintf("categories:%d:*", categoryID),
		"categories:list:*",
		"articles:category:*", // Articles filtered by category
	}

	return ci.invalidatePatterns(patterns)
}

// InvalidateTag invalidates cache entries for a tag
func (ci *CacheInvalidator) InvalidateTag(tagID int64) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_tag")()

	patterns := []string{
		fmt.Sprintf("tags:%d", tagID),
		fmt.Sprintf("tags:%d:*", tagID),
		"tags:list:*",
		"articles:tag:*", // Articles filtered by tag
	}

	return ci.invalidatePatterns(patterns)
}

// InvalidateArticleLists invalidates all article list caches (pagination, featured, etc.)
func (ci *CacheInvalidator) InvalidateArticleLists() error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_article_lists")()

	patterns := []string{
		"articles:list:*",       // All article lists
		"articles:page:*",       // All paginated lists
		"articles:pagination:*", // All pagination caches
		"articles:featured:*",   // Featured articles
		"articles:breaking:*",   // Breaking news
		"articles:trending:*",   // Trending articles
		"articles:popular:*",    // Popular articles
		"all_articles",          // Legacy global cache
	}

	return ci.invalidatePatterns(patterns)
}

// InvalidateUser invalidates cache entries for a user
func (ci *CacheInvalidator) InvalidateUser(userID int64) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_user")()

	patterns := []string{
		fmt.Sprintf("users:%d", userID),
		fmt.Sprintf("users:%d:*", userID),
		"articles:author:*", // Articles by author
	}

	return ci.invalidatePatterns(patterns)
}

// InvalidateSearch invalidates search-related cache entries
func (ci *CacheInvalidator) InvalidateSearch() error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_search")()

	patterns := []string{
		"search:*",
		"articles:search:*",
	}

	return ci.invalidatePatterns(patterns)
}

// InvalidateAll invalidates all cache entries (use with caution)
func (ci *CacheInvalidator) InvalidateAll() error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_all")()

	// Clear L1 completely
	ci.unified.ristretto.Clear()

	// Clear L2 patterns
	patterns := []string{"*"}
	return ci.invalidatePatterns(patterns)
}

// BulkInvalidate invalidates multiple patterns efficiently
func (ci *CacheInvalidator) BulkInvalidate(patterns []string) error {
	defer metrics.TrackDatabaseOperation("cache_bulk_invalidate")()
	return ci.invalidatePatterns(patterns)
}

// InvalidateByPrefix invalidates all keys with a specific prefix
func (ci *CacheInvalidator) InvalidateByPrefix(prefix string) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_prefix")()

	pattern := prefix
	if !strings.HasSuffix(prefix, "*") {
		pattern = prefix + "*"
	}

	return ci.unified.DeletePattern(pattern)
}

// InvalidateBulkArticles invalidates cache for multiple articles efficiently
func (ci *CacheInvalidator) InvalidateBulkArticles(articleIDs []int64) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_bulk_articles")()

	// Batch patterns to minimize cache operations
	patterns := make(map[string]bool)

	for _, articleID := range articleIDs {
		patterns[fmt.Sprintf("articles:%d", articleID)] = true
		patterns[fmt.Sprintf("articles:%d:*", articleID)] = true
	}

	// Global patterns (only once for all articles)
	globalPatterns := []string{
		"articles:list:*",
		"articles:pagination:*",
		"articles:featured:*",
		"articles:trending:*",
	}

	for _, pattern := range globalPatterns {
		patterns[pattern] = true
	}

	// Convert to slice for invalidation
	var allPatterns []string
	for pattern := range patterns {
		allPatterns = append(allPatterns, pattern)
	}

	return ci.invalidatePatterns(allPatterns)
}

// InvalidateSmartPattern uses pattern analysis for optimal invalidation
func (ci *CacheInvalidator) InvalidateSmartPattern(contentType string, contentID int64, operation string) error {
	defer metrics.TrackDatabaseOperation("cache_invalidate_smart")()

	patterns := ci.generateSmartPatterns(contentType, contentID, operation)
	return ci.invalidatePatterns(patterns)
}

// generateSmartPatterns creates optimized invalidation patterns based on content type and operation
func (ci *CacheInvalidator) generateSmartPatterns(contentType string, contentID int64, operation string) []string {
	var patterns []string

	switch contentType {
	case "article":
		patterns = append(patterns, fmt.Sprintf("articles:%d", contentID))

		if operation == "create" {
			// New article affects pagination and lists
			patterns = append(patterns,
				"articles:list:*",
				"articles:page:1:*", // First pages most affected
			)
		} else if operation == "update" {
			// Updated article affects individual cache and related lists
			patterns = append(patterns,
				fmt.Sprintf("articles:%d:*", contentID),
				"articles:featured:*",
				"articles:trending:*",
			)
		} else if operation == "delete" {
			// Deleted article requires comprehensive cleanup
			patterns = append(patterns,
				fmt.Sprintf("articles:%d:*", contentID),
				"articles:list:*",
				"articles:pagination:*",
			)
		}

	case "category":
		patterns = append(patterns,
			fmt.Sprintf("categories:%d", contentID),
			"categories:list:*",
		)

		if operation != "update" {
			// Category changes affect article listings
			patterns = append(patterns, "articles:category:*")
		}

	case "tag":
		patterns = append(patterns,
			fmt.Sprintf("tags:%d", contentID),
			"tags:list:*",
		)

		if operation != "update" {
			patterns = append(patterns, "articles:tag:*")
		}
	}

	return patterns
}

// ScheduledInvalidation performs background cache cleanup
func (ci *CacheInvalidator) ScheduledInvalidation() error {
	defer metrics.TrackDatabaseOperation("cache_scheduled_cleanup")()

	// Clean expired entries and optimize cache
	stalePatterns := []string{
		"articles:temp:*",   // Temporary cache entries
		"search:old:*",      // Old search results
		"analytics:daily:*", // Expired analytics
	}

	return ci.invalidatePatterns(stalePatterns)
}

// GetInvalidationStats returns invalidation statistics
func (ci *CacheInvalidator) GetInvalidationStats() map[string]interface{} {
	return map[string]interface{}{
		"invalidator_active": true,
		"cache_health":       ci.unified.Health(),
		"last_updated":       time.Now(),
	}
}

// GetInvalidationMetrics returns invalidation performance metrics
func (ci *CacheInvalidator) GetInvalidationMetrics() map[string]interface{} {
	return map[string]interface{}{
		"invalidator_active": true,
		"cache_health":       ci.unified.Health(),
		"last_updated":       time.Now(),
		"optimization_level": "smart_patterns",
		"bulk_operations":    "supported",
		"scheduled_cleanup":  "enabled",
	}
}

// invalidatePatterns is a helper function to invalidate multiple patterns
func (ci *CacheInvalidator) invalidatePatterns(patterns []string) error {
	var lastError error
	for _, pattern := range patterns {
		if err := ci.unified.DeletePattern(pattern); err != nil {
			lastError = err
			fmt.Printf("Warning: Failed to invalidate pattern %s: %v\n", pattern, err)
		}
	}
	return lastError
}
