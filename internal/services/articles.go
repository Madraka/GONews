package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"news/internal/cache"
	"news/internal/database"
	"news/internal/json"
	"news/internal/models"
	"news/internal/repositories"
	"news/internal/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

var (
	// Global cache invalidator instance
	cacheInvalidator *cache.CacheInvalidator
)

// init initializes the cache invalidator
func init() {
	cacheInvalidator = cache.NewCacheInvalidator()
}

var (
	// Define specific error types for better error handling
	ErrNotFound       = errors.New("article not found")
	ErrDatabaseError  = errors.New("database operation failed")
	ErrCacheOperation = errors.New("cache operation failed")
	ErrValidation     = errors.New("validation error")
)

const (
	articleCacheDuration = 30 * time.Minute
	articleKeyPrefix     = "article:"
	maxCacheRetries      = 3
)

// GetArticlesWithPagination retrieves articles with pagination and optional filtering
func GetArticlesWithPagination(offset, limit int, category string) ([]models.Article, int, error) {
	// Create cache key for pagination
	cacheKey := fmt.Sprintf("articles:page:%d:limit:%d:category:%s", offset/limit+1, limit, category)

	// Use migration cache manager for intelligent cache handling
	cacheManager := cache.GetMigrationCacheManager()

	// Try to get from cache (optimized first, then standard fallback)
	if cachedData, found := cacheManager.SmartGet(cacheKey); found {
		var result struct {
			Articles []models.Article `json:"articles"`
			Total    int              `json:"total"`
		}
		if err := json.UnmarshalForCache([]byte(cachedData), &result); err == nil {
			log.Printf("Retrieved articles page from cache (offset: %d, limit: %d, category: %s)", offset, limit, category)
			return result.Articles, result.Total, nil
		} else {
			log.Printf("Failed to unmarshal cached articles: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for articles page, fetching from database (offset: %d, limit: %d, category: %s)", offset, limit, category)
	articles, total, err := repositories.FetchArticlesWithPagination(offset, limit, category)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result using intelligent cache management
	result := struct {
		Articles []models.Article `json:"articles"`
		Total    int              `json:"total"`
	}{
		Articles: articles,
		Total:    total,
	}

	if cacheData, err := json.MarshalForCache(result); err == nil {
		// Store using migration cache manager (optimized first, standard fallback)
		l1TTL := 2 * time.Minute // Hot data in L1 for 2 minutes
		l2TTL := 5 * time.Minute // Persistent data in L2 for 5 minutes

		if err := cacheManager.SmartSet(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache articles page: %v", err)
		} else {
			log.Printf("Successfully cached articles page with intelligent cache management (L1: %v, L2: %v)", l1TTL, l2TTL)
		}
	}

	return articles, total, nil
}

// GetArticlesWithPaginationContext retrieves articles with pagination and tracing
func GetArticlesWithPaginationContext(ctx context.Context, offset, limit int, category string) ([]models.Article, int, error) {
	_, span := tracing.StartSpan(ctx, "GetArticlesWithPagination")
	defer span.End()

	// Add attributes for better observability
	span.SetAttributes(
		attribute.Int("pagination.offset", offset),
		attribute.Int("pagination.limit", limit),
		attribute.String("filter.category", category),
	)

	articles, total, err := repositories.FetchArticlesWithPagination(offset, limit, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, 0, err
	}

	// Add result attributes
	span.SetAttributes(
		attribute.Int("result.count", len(articles)),
		attribute.Int("result.total", total),
	)

	return articles, total, nil
}

// GetArticleById retrieves a single article by ID from cache or database
func GetArticleById(id string) (models.Article, error) {
	// Try to get from unified cache first (L1: Ristretto -> L2: Redis)
	cacheKey := articleKeyPrefix + id
	unifiedCache := cache.GetUnifiedCache()

	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var article models.Article
		if err := json.UnmarshalForCache([]byte(cachedData), &article); err == nil {
			log.Printf("Retrieved article %s from unified cache", id)
			return article, nil
		} else {
			log.Printf("Failed to unmarshal cached article: %v", err)
		}
	}

	// If not in cache or error, get from database
	log.Printf("Cache miss for article %s, fetching from database", id)
	article, err := repositories.GetArticleByID(id)
	if err != nil {
		log.Printf("Article with id %s not found: %v", id, err)
		return models.Article{}, ErrNotFound
	}

	// Cache the result in both L1 and L2
	if cacheData, err := json.MarshalForCache(article); err == nil {
		// L1 cache (Ristretto): 10 minutes for individual articles
		// L2 cache (Redis): 30 minutes for persistence
		l1TTL := 10 * time.Minute
		l2TTL := articleCacheDuration // 30 minutes

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache article %s in unified cache: %v", id, err)
		} else {
			log.Printf("Cached article %s in unified cache (L1: %v, L2: %v)", id, l1TTL, l2TTL)
		}
	}

	return article, nil
}

// GetArticleByIdWithContext retrieves a single article by ID with tracing
func GetArticleByIdWithContext(ctx context.Context, id string) (models.Article, error) {
	ctx, span := tracing.StartSpan(ctx, "GetArticleById")
	defer span.End()

	// Add attributes to span for better observability
	span.SetAttributes(attribute.String("article.id", id))

	// Try to get from optimized cache first
	cacheKey := articleKeyPrefix + id
	optimizedCache := cache.GetOptimizedUnifiedCache()
	if optimizedCache != nil {
		if cachedValue, found := optimizedCache.GetString(cacheKey); found {
			var article models.Article
			if err := json.UnmarshalForCache([]byte(cachedValue), &article); err == nil {
				log.Printf("Retrieved article %s from optimized cache", id)
				span.SetAttributes(
					attribute.Bool("cache.hit", true),
					attribute.Bool("optimized_cache.hit", true),
					attribute.String("article.title", article.Title),
					attribute.String("article.status", article.Status),
				)
				return article, nil
			} else {
				log.Printf("Failed to unmarshal cached article: %v", err)
				span.SetAttributes(attribute.Bool("cache.unmarshal_error", true))
			}
		}
	}

	// Fallback to standard unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cachedValue, found := unifiedCache.GetString(cacheKey); found {
		var article models.Article
		if err := json.UnmarshalForCache([]byte(cachedValue), &article); err == nil {
			log.Printf("Retrieved article %s from unified cache", id)
			span.SetAttributes(
				attribute.Bool("cache.hit", true),
				attribute.Bool("unified_cache.hit", true),
				attribute.String("article.title", article.Title),
				attribute.String("article.status", article.Status),
			)
			return article, nil
		} else {
			log.Printf("Failed to unmarshal cached article: %v", err)
			span.SetAttributes(attribute.Bool("cache.unmarshal_error", true))
		}
	}

	// If not in cache or error, get from database
	span.SetAttributes(attribute.Bool("cache.hit", false))
	log.Printf("Cache miss for article %s, fetching from database", id)

	_, dbSpan := tracing.StartSpan(ctx, "GetArticleById.Database")
	article, err := repositories.GetArticleByID(id)
	if err != nil {
		log.Printf("Article with id %s not found: %v", id, err)
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		span.RecordError(ErrNotFound)
		span.SetStatus(codes.Error, "Article not found")
		dbSpan.End()
		return models.Article{}, ErrNotFound
	}
	dbSpan.SetAttributes(
		attribute.String("article.title", article.Title),
		attribute.String("article.status", article.Status),
	)
	dbSpan.End()

	// Cache the result with optimized cache first, then fallback
	if cacheData, err := json.MarshalForCache(article); err == nil {
		// Try optimized cache first
		optimizedCache := cache.GetOptimizedUnifiedCache()
		if optimizedCache != nil {
			l1TTL := 10 * time.Minute
			l2TTL := articleCacheDuration // 30 minutes
			if err := optimizedCache.SmartSet(cacheKey, string(cacheData), cache.WithL1TTL(l1TTL), cache.WithL2TTL(l2TTL)); err != nil {
				log.Printf("Warning: Failed to cache article %s in optimized cache: %v", id, err)
				span.SetAttributes(attribute.Bool("optimized_cache.store_error", true))
			} else {
				log.Printf("Cached article %s in optimized cache (L1: %v, L2: %v)", id, l1TTL, l2TTL)
				span.SetAttributes(attribute.Bool("optimized_cache.stored", true))
				span.SetAttributes(
					attribute.String("article.title", article.Title),
					attribute.String("article.status", article.Status),
				)
				return article, nil
			}
		}

		// Fallback to retry cache for compatibility
		if err := cache.RetryCache(cacheKey, string(cacheData), articleCacheDuration, maxCacheRetries); err != nil {
			log.Printf("Warning: Failed to cache article %s after retries: %v", id, err)
			span.SetAttributes(attribute.Bool("cache.store_error", true))
		} else {
			span.SetAttributes(attribute.Bool("cache.stored", true))
		}
	}

	span.SetAttributes(
		attribute.String("article.title", article.Title),
		attribute.String("article.status", article.Status),
	)

	return article, nil
}

// CreateArticle creates a new article with validation and cache invalidation
func CreateArticle(article models.Article) (models.Article, error) {
	// Validate article before creating
	if err := validateArticle(article); err != nil {
		return models.Article{}, err
	}

	// Generate slug from title
	article.Slug = repositories.GenerateSlug(article.Title)

	createdArticle, err := repositories.InsertArticle(article)
	if err != nil {
		log.Printf("Error creating article: %v", err)
		return models.Article{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Use unified cache invalidation system
	if cacheInvalidator != nil {
		// Invalidate all article lists and pagination caches
		if err := cacheInvalidator.InvalidateArticleLists(); err != nil {
			log.Printf("Warning: Failed to invalidate article lists cache after creation: %v", err)
		}

		// If article has categories, invalidate related category caches
		if len(createdArticle.Categories) > 0 {
			for _, category := range createdArticle.Categories {
				if err := cacheInvalidator.InvalidateCategory(int64(category.ID)); err != nil {
					log.Printf("Warning: Failed to invalidate category cache after article creation: %v", err)
				}
			}
		}

		log.Printf("Successfully invalidated caches after creating article %d", createdArticle.ID)
	}

	// Cache the new article in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.MarshalForCache(createdArticle); err == nil {
		cacheKey := articleKeyPrefix + strconv.FormatUint(uint64(createdArticle.ID), 10)
		l1TTL := 10 * time.Minute
		l2TTL := articleCacheDuration // 30 minutes

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache new article in unified cache: %v", err)
		} else {
			log.Printf("Cached new article %d in unified cache", createdArticle.ID)
		}
	}

	return createdArticle, nil
}

// UpdateArticle updates an existing article with cache invalidation
func UpdateArticle(id string, updatedArticle models.Article) (models.Article, error) {
	// Validate article before updating
	if err := validateArticle(updatedArticle); err != nil {
		return models.Article{}, err
	}

	existingArticle, err := repositories.GetArticleByID(id)
	if err != nil {
		log.Printf("Article with id %s not found: %v", id, err)
		return models.Article{}, ErrNotFound
	}

	// Update fields
	existingArticle.Title = updatedArticle.Title
	existingArticle.Content = updatedArticle.Content
	existingArticle.FeaturedImage = updatedArticle.FeaturedImage
	existingArticle.Status = updatedArticle.Status
	existingArticle.MetaTitle = updatedArticle.MetaTitle
	existingArticle.MetaDesc = updatedArticle.MetaDesc
	existingArticle.UpdatedAt = time.Now()

	if err := repositories.UpdateArticle(existingArticle); err != nil {
		log.Printf("Error updating article: %v", err)
		return models.Article{}, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Use unified cache invalidation system
	if cacheInvalidator != nil {
		// Invalidate the specific article
		articleIDInt, _ := strconv.ParseInt(id, 10, 64)
		if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
			log.Printf("Warning: Failed to invalidate article cache after update: %v", err)
		}

		// Invalidate article lists (in case title, status, or other list-affecting fields changed)
		if err := cacheInvalidator.InvalidateArticleLists(); err != nil {
			log.Printf("Warning: Failed to invalidate article lists cache after update: %v", err)
		}

		// If article has categories, invalidate related category caches
		if len(existingArticle.Categories) > 0 {
			for _, category := range existingArticle.Categories {
				if err := cacheInvalidator.InvalidateCategory(int64(category.ID)); err != nil {
					log.Printf("Warning: Failed to invalidate category cache after article update: %v", err)
				}
			}
		}

		log.Printf("Successfully invalidated caches after updating article %s", id)
	}

	// Cache the updated article in unified cache
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.MarshalForCache(existingArticle); err == nil {
		cacheKey := articleKeyPrefix + id
		l1TTL := 10 * time.Minute
		l2TTL := articleCacheDuration // 30 minutes

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache updated article in unified cache: %v", err)
		} else {
			log.Printf("Cached updated article %s in unified cache", id)
		}
	}

	return existingArticle, nil
}

// DeleteArticle deletes an article by ID with cache invalidation
func DeleteArticle(id string) error {
	// First get the article to check its categories before deletion
	existingArticle, err := repositories.GetArticleByID(id)
	if err != nil {
		log.Printf("Article with id %s not found: %v", id, err)
		return ErrNotFound
	}

	err = repositories.DeleteArticleByID(id)
	if err != nil {
		log.Printf("Error deleting article: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Use unified cache invalidation system
	if cacheInvalidator != nil {
		// Invalidate the specific article
		articleIDInt, _ := strconv.ParseInt(id, 10, 64)
		if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
			log.Printf("Warning: Failed to invalidate article cache after deletion: %v", err)
		}

		// Invalidate article lists
		if err := cacheInvalidator.InvalidateArticleLists(); err != nil {
			log.Printf("Warning: Failed to invalidate article lists cache after deletion: %v", err)
		}

		// If article had categories, invalidate related category caches
		if len(existingArticle.Categories) > 0 {
			for _, category := range existingArticle.Categories {
				if err := cacheInvalidator.InvalidateCategory(int64(category.ID)); err != nil {
					log.Printf("Warning: Failed to invalidate category cache after article deletion: %v", err)
				}
			}
		}

		log.Printf("Successfully invalidated caches after deleting article %s", id)
	}

	return nil
}

// validateArticle performs validation checks on articles
func validateArticle(article models.Article) error {
	if len(article.Title) < 3 {
		return fmt.Errorf("%w: title must be at least 3 characters", ErrValidation)
	}

	if len(article.Content) < 10 {
		return fmt.Errorf("%w: content must be at least 10 characters", ErrValidation)
	}

	if article.FeaturedImage != "" {
		if _, err := url.ParseRequestURI(article.FeaturedImage); err != nil {
			return fmt.Errorf("%w: invalid featured image URL", ErrValidation)
		}
	}

	return nil
}

// Content Blocks Services - Modern content management

// GetArticleWithBlocks retrieves an article with its content blocks
func GetArticleWithBlocks(id string) (models.Article, error) {
	// Create a special cache key for articles with blocks
	cacheKey := fmt.Sprintf("article:%s:with_blocks", id)
	unifiedCache := cache.GetUnifiedCache()

	// Try to get the complete article with blocks from cache
	if cachedData, found := unifiedCache.GetString(cacheKey); found {
		var article models.Article
		if err := json.UnmarshalForCache([]byte(cachedData), &article); err == nil {
			log.Printf("Retrieved article %s with blocks from cache", id)
			return article, nil
		} else {
			log.Printf("Failed to unmarshal cached article with blocks: %v", err)
		}
	}

	// Cache miss - get from database
	log.Printf("Cache miss for article %s with blocks, fetching from database", id)
	article, err := repositories.GetArticleByID(id)
	if err != nil {
		log.Printf("Article with id %s not found: %v", id, err)
		return models.Article{}, ErrNotFound
	}

	// Load content blocks if article uses them
	if article.IsUsingBlocks() {
		blocks, err := repositories.ArticleContentBlockRepo.GetVisibleBlocksByArticleID(article.ID)
		if err != nil {
			log.Printf("Warning: Failed to load content blocks for article %d: %v", article.ID, err)
		} else {
			article.ContentBlocks = blocks
		}
	}

	// Cache the complete article with blocks
	if cacheData, err := json.MarshalForCache(article); err == nil {
		l1TTL := 10 * time.Minute
		l2TTL := articleCacheDuration // 30 minutes

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache article %s with blocks: %v", id, err)
		} else {
			log.Printf("Cached article %s with blocks (L1: %v, L2: %v)", id, l1TTL, l2TTL)
		}
	}

	return article, nil
}

// CreateArticleWithBlocks creates a new article with content blocks
func CreateArticleWithBlocks(article models.Article, blocks []models.ArticleContentBlock) (models.Article, error) {
	// Validate article
	if err := validateArticle(article); err != nil {
		return models.Article{}, err
	}

	// Set article to use blocks
	article.ContentType = "blocks"
	article.HasBlocks = true
	article.BlocksVersion = 1

	// Generate slug from title
	article.Slug = repositories.GenerateSlug(article.Title)

	var createdArticle models.Article

	// Start transaction
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Create article
		var err error
		createdArticle, err = repositories.InsertArticle(article)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDatabaseError, err)
		}

		// Create content blocks
		if len(blocks) > 0 {
			for i := range blocks {
				blocks[i].ArticleID = createdArticle.ID
				blocks[i].Position = i + 1
			}

			if err := repositories.ArticleContentBlockRepo.BulkCreateBlocks(blocks); err != nil {
				return fmt.Errorf("failed to create content blocks: %v", err)
			}

			// Update content from blocks for backward compatibility
			createdArticle.ContentBlocks = blocks
			createdArticle.UpdateContentFromBlocks()

			// Update the article with generated content
			if err := repositories.UpdateArticle(createdArticle); err != nil {
				log.Printf("Warning: Failed to update article content from blocks: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return models.Article{}, err
	}

	// Invalidate caches
	if cacheInvalidator != nil {
		if err := cacheInvalidator.InvalidateArticleLists(); err != nil {
			log.Printf("Warning: Failed to invalidate article lists cache: %v", err)
		}

		if len(createdArticle.Categories) > 0 {
			for _, category := range createdArticle.Categories {
				if err := cacheInvalidator.InvalidateCategory(int64(category.ID)); err != nil {
					log.Printf("Warning: Failed to invalidate category cache: %v", err)
				}
			}
		}
	}

	// Cache the new article
	cacheNewArticle(createdArticle)

	return createdArticle, nil
}

// UpdateArticleBlocks updates article content blocks
func UpdateArticleBlocks(articleID string, blocks []models.ArticleContentBlock) error {
	// Get article
	article, err := GetArticleById(articleID)
	if err != nil {
		return err
	}

	// Ensure article uses blocks
	if !article.IsUsingBlocks() {
		return fmt.Errorf("article %s does not use content blocks", articleID)
	}

	// Start transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete existing blocks
		if err := repositories.ArticleContentBlockRepo.DeleteBlocksByArticleID(article.ID); err != nil {
			return fmt.Errorf("failed to delete existing blocks: %v", err)
		}

		// Create new blocks
		if len(blocks) > 0 {
			for i := range blocks {
				blocks[i].ArticleID = article.ID
				blocks[i].Position = i + 1
			}

			if err := repositories.ArticleContentBlockRepo.BulkCreateBlocks(blocks); err != nil {
				return fmt.Errorf("failed to create content blocks: %v", err)
			}

			// Update content from blocks for backward compatibility
			article.ContentBlocks = blocks
			article.UpdateContentFromBlocks()

			// Update the article with generated content
			if err := repositories.UpdateArticle(article); err != nil {
				log.Printf("Warning: Failed to update article content from blocks: %v", err)
			}
		}

		// Invalidate article cache
		if cacheInvalidator != nil {
			articleIDInt, _ := strconv.ParseInt(articleID, 10, 64)
			if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
				log.Printf("Warning: Failed to invalidate article cache: %v", err)
			}
		}

		return nil
	})
}

// MigrateArticleToBlocks converts a legacy article to use content blocks
func MigrateArticleToBlocks(articleID string) (models.Article, error) {
	// Get article
	article, err := GetArticleById(articleID)
	if err != nil {
		return models.Article{}, err
	}

	// Skip if already using blocks
	if article.IsUsingBlocks() {
		return article, nil
	}

	// Migrate to blocks
	if err := repositories.ArticleContentBlockRepo.MigrateArticleToBlocks(article.ID); err != nil {
		return models.Article{}, fmt.Errorf("failed to migrate article to blocks: %v", err)
	}

	// Get updated article with blocks
	updatedArticle, err := GetArticleWithBlocks(articleID)
	if err != nil {
		return models.Article{}, err
	}

	// Invalidate article cache
	if cacheInvalidator != nil {
		articleIDInt, _ := strconv.ParseInt(articleID, 10, 64)
		if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
			log.Printf("Warning: Failed to invalidate article cache: %v", err)
		}
	}

	return updatedArticle, nil
}

// AddContentBlock adds a single content block to an article
func AddContentBlock(articleID string, block models.ArticleContentBlock) (*models.ArticleContentBlock, error) {
	// Get article
	article, err := GetArticleById(articleID)
	if err != nil {
		return nil, err
	}

	// Get next position
	lastPosition, err := repositories.ArticleContentBlockRepo.GetLastPositionForArticle(article.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get last position: %v", err)
	}

	// Set block properties
	block.ArticleID = article.ID
	block.Position = lastPosition + 1

	// Create block
	createdBlock, err := repositories.ArticleContentBlockRepo.CreateBlock(&block)
	if err != nil {
		return nil, fmt.Errorf("failed to create content block: %v", err)
	}

	// Update article content from blocks if using blocks
	if article.IsUsingBlocks() {
		blocks, err := repositories.ArticleContentBlockRepo.GetVisibleBlocksByArticleID(article.ID)
		if err == nil {
			article.ContentBlocks = blocks
			article.UpdateContentFromBlocks()
			if err := repositories.UpdateArticle(article); err != nil {
				log.Printf("Warning: Failed to update article with blocks: %v", err)
			}
		}
	}

	// Invalidate article cache
	if cacheInvalidator != nil {
		articleIDInt, _ := strconv.ParseInt(articleID, 10, 64)
		log.Printf("DEBUG: [ADD BLOCK CACHE INVALIDATION] Invalidating cache for article %s (ID: %d)", articleID, articleIDInt)
		if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
			log.Printf("Warning: Failed to invalidate article cache: %v", err)
		} else {
			log.Printf("DEBUG: [ADD BLOCK CACHE INVALIDATION] Successfully invalidated cache for article %s", articleID)
		}
	} else {
		log.Printf("DEBUG: [ADD BLOCK CACHE INVALIDATION] Cache invalidator is nil!")
	}

	return createdBlock, nil
}

// UpdateContentBlock updates a specific content block
func UpdateContentBlock(blockID uint, updateData map[string]interface{}) (*models.ArticleContentBlock, error) {
	// Update block
	updatedBlock, err := repositories.ArticleContentBlockRepo.UpdateBlock(blockID, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update content block: %v", err)
	}

	// Update article content from blocks if using blocks
	article, err := GetArticleById(strconv.FormatUint(uint64(updatedBlock.ArticleID), 10))
	if err == nil && article.IsUsingBlocks() {
		blocks, err := repositories.ArticleContentBlockRepo.GetVisibleBlocksByArticleID(article.ID)
		if err == nil {
			article.ContentBlocks = blocks
			article.UpdateContentFromBlocks()
			if err := repositories.UpdateArticle(article); err != nil {
				log.Printf("Warning: Failed to update article after block update: %v", err)
			}
		}
	}

	// Invalidate article cache
	if cacheInvalidator != nil {
		if err := cacheInvalidator.InvalidateArticle(int64(updatedBlock.ArticleID)); err != nil {
			log.Printf("Warning: Failed to invalidate article cache: %v", err)
		}
	}

	return updatedBlock, nil
}

// DeleteContentBlock deletes a content block
func DeleteContentBlock(blockID uint) error {
	// Get block to find article ID
	block, err := repositories.ArticleContentBlockRepo.GetBlockByID(blockID)
	if err != nil {
		return fmt.Errorf("content block not found: %v", err)
	}

	articleID := block.ArticleID

	// Delete block
	if err := repositories.ArticleContentBlockRepo.DeleteBlock(blockID); err != nil {
		return fmt.Errorf("failed to delete content block: %v", err)
	}

	// Update article content from remaining blocks
	article, err := GetArticleById(strconv.FormatUint(uint64(articleID), 10))
	if err == nil && article.IsUsingBlocks() {
		blocks, err := repositories.ArticleContentBlockRepo.GetVisibleBlocksByArticleID(article.ID)
		if err == nil {
			article.ContentBlocks = blocks
			article.UpdateContentFromBlocks()
			if err := repositories.UpdateArticle(article); err != nil {
				log.Printf("Warning: Failed to update article after block deletion: %v", err)
			}
		}
	}

	// Invalidate article cache
	if cacheInvalidator != nil {
		if err := cacheInvalidator.InvalidateArticle(int64(articleID)); err != nil {
			log.Printf("Warning: Failed to invalidate article cache: %v", err)
		}
	}

	return nil
}

// ReorderContentBlocks reorders content blocks for an article
func ReorderContentBlocks(articleID string, blockPositions map[uint]int) error {
	// Get article
	article, err := GetArticleById(articleID)
	if err != nil {
		return err
	}

	// Reorder blocks
	if err := repositories.ArticleContentBlockRepo.ReorderBlocks(article.ID, blockPositions); err != nil {
		return fmt.Errorf("failed to reorder content blocks: %v", err)
	}

	// Update article content from blocks
	if article.IsUsingBlocks() {
		blocks, err := repositories.ArticleContentBlockRepo.GetVisibleBlocksByArticleID(article.ID)
		if err == nil {
			article.ContentBlocks = blocks
			article.UpdateContentFromBlocks()
			if err := repositories.UpdateArticle(article); err != nil {
				log.Printf("Warning: Failed to update article after content block sync: %v", err)
			}
		}
	}

	// Invalidate article cache
	if cacheInvalidator != nil {
		articleIDInt, _ := strconv.ParseInt(articleID, 10, 64)
		if err := cacheInvalidator.InvalidateArticle(articleIDInt); err != nil {
			log.Printf("Warning: Failed to invalidate article cache: %v", err)
		}
	}

	return nil
}

// GetContentBlocksByType retrieves content blocks of a specific type for an article
func GetContentBlocksByType(articleID string, blockType string) ([]models.ArticleContentBlock, error) {
	// Get article
	article, err := GetArticleById(articleID)
	if err != nil {
		return nil, err
	}

	// Get blocks by type
	blocks, err := repositories.ArticleContentBlockRepo.GetBlocksByType(article.ID, blockType)
	if err != nil {
		return nil, fmt.Errorf("failed to get content blocks by type: %v", err)
	}

	return blocks, nil
}

// Helper function to cache new articles
func cacheNewArticle(article models.Article) {
	unifiedCache := cache.GetUnifiedCache()
	if cacheData, err := json.MarshalForCache(article); err == nil {
		cacheKey := articleKeyPrefix + strconv.FormatUint(uint64(article.ID), 10)
		l1TTL := 10 * time.Minute
		l2TTL := articleCacheDuration // 30 minutes

		if err := unifiedCache.Set(cacheKey, string(cacheData), l1TTL, l2TTL); err != nil {
			log.Printf("Warning: Failed to cache new article in unified cache: %v", err)
		} else {
			log.Printf("Cached new article %d in unified cache", article.ID)
		}
	}
}

// Legacy News service functions for backward compatibility
// These functions are wrappers around Article functions

// GetArticlesWithPaginationCached retrieves articles with pagination and returns raw cached JSON
func GetArticlesWithPaginationCached(offset, limit int, category string) (string, error) {
	// Create cache key for pagination
	cacheKey := fmt.Sprintf("articles:page:%d:limit:%d:category:%s:json", offset/limit+1, limit, category)

	// Use migration cache manager for intelligent cache handling
	cacheManager := cache.GetMigrationCacheManager()

	// Try to get cached JSON directly
	if cachedJSON, found := cacheManager.SmartGet(cacheKey); found {
		log.Printf("Retrieved articles page JSON from cache (offset: %d, limit: %d, category: %s)", offset, limit, category)
		return cachedJSON, nil
	}

	// Cache miss - get from database and create JSON
	log.Printf("Cache miss for articles page JSON, fetching from database (offset: %d, limit: %d, category: %s)", offset, limit, category)
	articles, total, err := repositories.FetchArticlesWithPagination(offset, limit, category)
	if err != nil {
		return "", err
	}

	// Calculate pagination info
	page := offset/limit + 1
	totalPages := (total + limit - 1) / limit

	// Create paginated response
	response := models.PaginatedResponse{
		Data:       articles,
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	// Marshal to JSON once and cache it
	jsonData, err := json.MarshalForCache(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal articles response: %v", err)
	}

	// Store using migration cache manager (optimized first, standard fallback)
	l1TTL := 2 * time.Minute // Hot data in L1 for 2 minutes
	l2TTL := 5 * time.Minute // Persistent data in L2 for 5 minutes

	if err := cacheManager.SmartSet(cacheKey, string(jsonData), l1TTL, l2TTL); err != nil {
		log.Printf("Warning: Failed to cache articles page JSON: %v", err)
	} else {
		log.Printf("Successfully cached articles page JSON with intelligent cache management (L1: %v, L2: %v)", l1TTL, l2TTL)
	}

	return string(jsonData), nil
}

// GetArticleByIdCached retrieves a single article by ID and returns raw cached JSON
func GetArticleByIdCached(id string) (string, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("article:%s:json", id)

	// Use migration cache manager for intelligent cache handling
	cacheManager := cache.GetMigrationCacheManager()

	// Try to get cached JSON directly
	if cachedJSON, found := cacheManager.SmartGet(cacheKey); found {
		log.Printf("Retrieved article JSON from cache (ID: %s)", id)
		return cachedJSON, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for article JSON, fetching from database (ID: %s)", id)
	article, err := repositories.GetArticleByID(id)
	if err != nil {
		return "", ErrNotFound
	}

	// Marshal to JSON once and cache it
	jsonData, err := json.MarshalForCache(article)
	if err != nil {
		return "", fmt.Errorf("failed to marshal article response: %v", err)
	}

	// Store using migration cache manager (optimized first, standard fallback)
	l1TTL := 5 * time.Minute  // Hot data in L1 for 5 minutes
	l2TTL := 15 * time.Minute // Persistent data in L2 for 15 minutes

	if err := cacheManager.SmartSet(cacheKey, string(jsonData), l1TTL, l2TTL); err != nil {
		log.Printf("Warning: Failed to cache article JSON: %v", err)
	} else {
		log.Printf("Successfully cached article JSON with intelligent cache management (L1: %v, L2: %v)", l1TTL, l2TTL)
	}

	return string(jsonData), nil
}

// GetArticlesWithPaginationCachedWithRedaction retrieves articles with pagination and returns redacted raw cached JSON
func GetArticlesWithPaginationCachedWithRedaction(offset, limit int, category string) (string, error) {
	// Create cache key for redacted pagination
	cacheKey := fmt.Sprintf("articles:page:%d:limit:%d:category:%s:json:redacted:v3", offset/limit+1, limit, category)

	// Use migration cache manager for intelligent cache handling
	cacheManager := cache.GetMigrationCacheManager()

	// Try to get cached redacted JSON directly
	if cachedJSON, found := cacheManager.SmartGet(cacheKey); found {
		log.Printf("Retrieved redacted articles page JSON from cache (offset: %d, limit: %d, category: %s)", offset, limit, category)
		return cachedJSON, nil
	}

	// Cache miss - get from database and create redacted JSON
	log.Printf("Cache miss for redacted articles page JSON, fetching from database (offset: %d, limit: %d, category: %s)", offset, limit, category)
	articles, total, err := repositories.FetchArticlesWithPagination(offset, limit, category)
	if err != nil {
		return "", err
	}

	// Calculate pagination info
	page := offset/limit + 1
	totalPages := (total + limit - 1) / limit

	// Create paginated response
	response := models.PaginatedResponse{
		Data:       articles,
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	// Marshal to JSON with redaction
	jsonData, err := json.MarshalForCacheWithRedaction(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal redacted articles response: %v", err)
	}

	// Store using migration cache manager (optimized first, standard fallback)
	l1TTL := 2 * time.Minute // Hot data in L1 for 2 minutes
	l2TTL := 5 * time.Minute // Persistent data in L2 for 5 minutes

	if err := cacheManager.SmartSet(cacheKey, string(jsonData), l1TTL, l2TTL); err != nil {
		log.Printf("Warning: Failed to cache redacted articles page JSON: %v", err)
	} else {
		log.Printf("Successfully cached redacted articles page JSON with intelligent cache management (L1: %v, L2: %v)", l1TTL, l2TTL)
	}

	return string(jsonData), nil
}

// GetArticleByIdCachedWithRedaction retrieves a single article by ID and returns redacted raw cached JSON
func GetArticleByIdCachedWithRedaction(id string) (string, error) {
	// Create cache key with redaction - temporarily bypass cache for testing
	cacheKey := fmt.Sprintf("article:%s:json:redacted:v3", id)

	// Use migration cache manager for intelligent cache handling
	cacheManager := cache.GetMigrationCacheManager()

	// Try to get cached redacted JSON directly
	if cachedJSON, found := cacheManager.SmartGet(cacheKey); found {
		log.Printf("Retrieved redacted article JSON from cache (ID: %s)", id)
		return cachedJSON, nil
	}

	// Cache miss - get from database
	log.Printf("Cache miss for redacted article JSON, fetching from database (ID: %s)", id)
	article, err := repositories.GetArticleByID(id)
	if err != nil {
		return "", ErrNotFound
	}

	// Marshal to JSON with redaction
	jsonData, err := json.MarshalForCacheWithRedaction(article)
	if err != nil {
		return "", fmt.Errorf("failed to marshal redacted article response: %v", err)
	}

	// Store using migration cache manager (optimized first, standard fallback)
	l1TTL := 5 * time.Minute  // Hot data in L1 for 5 minutes
	l2TTL := 15 * time.Minute // Persistent data in L2 for 15 minutes

	if err := cacheManager.SmartSet(cacheKey, string(jsonData), l1TTL, l2TTL); err != nil {
		log.Printf("Warning: Failed to cache redacted article JSON: %v", err)
	} else {
		log.Printf("Successfully cached redacted article JSON with intelligent cache management (L1: %v, L2: %v)", l1TTL, l2TTL)
	}

	return string(jsonData), nil
}

// GetArticlesWithPaginationCachedSmart retrieves articles with smart redaction based on environment
func GetArticlesWithPaginationCachedSmart(offset, limit int, category string) (string, error) {
	// Check if redaction is enabled
	if json.IsRedactionEnabled() {
		return GetArticlesWithPaginationCachedWithRedaction(offset, limit, category)
	}
	return GetArticlesWithPaginationCached(offset, limit, category)
}

// GetArticleByIdCachedSmart retrieves a single article with smart redaction based on environment
func GetArticleByIdCachedSmart(id string) (string, error) {
	// Check if redaction is enabled
	if json.IsRedactionEnabled() {
		return GetArticleByIdCachedWithRedaction(id)
	}
	return GetArticleByIdCached(id)
}
