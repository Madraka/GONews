package services

import (
	"context"
	"fmt"
	"log"
	"news/internal/models"
	"news/internal/tracing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"gorm.io/gorm"
)

// RecommendationService handles article recommendation logic
type RecommendationService struct {
	db *gorm.DB
}

// NewRecommendationService creates a new recommendation service
func NewRecommendationService(db *gorm.DB) *RecommendationService {
	return &RecommendationService{db: db}
}

// GetPersonalizedRecommendations returns recommendations based on user's reading history and preferences
func (rs *RecommendationService) GetPersonalizedRecommendations(userID uint, limit int) ([]models.Article, error) {
	return rs.GetPersonalizedRecommendationsWithContext(context.Background(), userID, limit)
}

// GetPersonalizedRecommendationsWithContext returns recommendations with tracing context
func (rs *RecommendationService) GetPersonalizedRecommendationsWithContext(ctx context.Context, userID uint, limit int) ([]models.Article, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "RecommendationService.GetPersonalizedRecommendations",
		attribute.Int("user.id", int(userID)),
		attribute.Int("limit", limit))
	defer span.End()

	var articles []models.Article

	// Complex recommendation query that considers:
	// 1. User's category preferences (based on reading history)
	// 2. Articles they haven't read yet
	// 3. Popular articles from categories they like
	// 4. Recent articles to keep recommendations fresh
	query := `
		SELECT DISTINCT a.* FROM articles a
		JOIN categories c ON a.category_id = c.id
		WHERE a.status = 'published' 
		AND a.id NOT IN (
			-- Exclude articles already read by user
			SELECT DISTINCT article_id FROM user_article_interactions 
			WHERE user_id = ? AND interaction_type IN ('view', 'bookmark')
		)
		AND (
			-- Articles from categories user has interacted with
			c.id IN (
				SELECT DISTINCT cat.id FROM categories cat
				JOIN articles art ON cat.id = art.category_id
				JOIN user_article_interactions uai ON art.id = uai.article_id
				WHERE uai.user_id = ? AND uai.interaction_type IN ('view', 'bookmark', 'upvote')
			)
			-- OR popular articles from last 7 days
			OR a.id IN (
				SELECT article_id FROM user_article_interactions
				WHERE interaction_type = 'upvote' 
				AND created_at > ?
				GROUP BY article_id
				HAVING COUNT(*) >= 2
			)
		)
		ORDER BY 
			-- Prioritize articles from preferred categories
			CASE WHEN c.id IN (
				SELECT DISTINCT cat.id FROM categories cat
				JOIN articles art ON cat.id = art.category_id
				JOIN user_article_interactions uai ON art.id = uai.article_id
				WHERE uai.user_id = ? AND uai.interaction_type IN ('view', 'bookmark', 'upvote')
			) THEN 1 ELSE 2 END,
			-- Then by recency
			a.published_at DESC
		LIMIT ?
	`

	weekAgo := time.Now().AddDate(0, 0, -7)

	// Trace the database query
	ctx, dbSpan := tracing.StartSpanWithAttributes(ctx, "DB.RecommendationQuery",
		attribute.String("db.type", "raw_sql"),
		attribute.String("db.operation", "recommendation_query"),
		attribute.Int("user.id", int(userID)),
		attribute.String("db.timeframe", "7days"))

	err := rs.db.Raw(query, userID, userID, weekAgo, userID, limit).Scan(&articles).Error

	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
		dbSpan.End()

		log.Printf("Error getting personalized recommendations: %v", err)
		span.AddEvent("Falling back to popular recommendations")

		// Fallback to popular articles if personalized query fails
		return rs.GetPopularRecommendationsWithContext(ctx, limit)
	}

	dbSpan.SetAttributes(attribute.Int("result.count", len(articles)))
	dbSpan.End()

	// If we don't have enough personalized recommendations, fill with popular ones
	if len(articles) < limit {
		remaining := limit - len(articles)

		span.AddEvent("Filling with popular recommendations")
		span.SetAttributes(
			attribute.Int("current_count", len(articles)),
			attribute.Int("requested_limit", limit),
			attribute.Int("remaining_needed", remaining))

		popularArticles, err := rs.GetPopularRecommendationsWithContext(ctx, remaining)
		if err == nil {
			// Remove duplicates and add popular articles
			existingIDs := make(map[uint]bool)
			for _, article := range articles {
				existingIDs[article.ID] = true
			}

			for _, popular := range popularArticles {
				if !existingIDs[popular.ID] && len(articles) < limit {
					articles = append(articles, popular)
				}
			}

			span.SetAttributes(attribute.Int("final_count", len(articles)))
		} else {
			span.RecordError(err)
			span.AddEvent("Failed to get popular recommendations")
		}
	}

	return articles, nil
}

// GetPopularRecommendations returns trending/popular articles
func (rs *RecommendationService) GetPopularRecommendations(limit int) ([]models.Article, error) {
	return rs.GetPopularRecommendationsWithContext(context.Background(), limit)
}

// GetPopularRecommendationsWithContext returns trending/popular articles with tracing context
func (rs *RecommendationService) GetPopularRecommendationsWithContext(ctx context.Context, limit int) ([]models.Article, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "RecommendationService.GetPopularRecommendations",
		attribute.Int("limit", limit))
	defer span.End()

	var articles []models.Article

	// Get articles with most interactions in the last 7 days
	query := `
		SELECT a.*, COUNT(uai.id) as interaction_count
		FROM articles a
		LEFT JOIN user_article_interactions uai ON a.id = uai.article_id 
			AND uai.created_at > ? 
			AND uai.interaction_type IN ('view', 'upvote', 'bookmark')
		WHERE a.status = 'published'
		GROUP BY a.id
		ORDER BY interaction_count DESC, a.published_at DESC
		LIMIT ?
	`

	weekAgo := time.Now().AddDate(0, 0, -7)

	// Trace the database query
	_, dbSpan := tracing.StartSpanWithAttributes(ctx, "DB.PopularArticlesQuery",
		attribute.String("db.type", "raw_sql"),
		attribute.String("db.operation", "popular_articles_query"),
		attribute.String("db.timeframe", "7days"),
		attribute.Int("limit", limit))

	err := rs.db.Raw(query, weekAgo, limit).Scan(&articles).Error

	if err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, err.Error())
	} else {
		dbSpan.SetAttributes(attribute.Int("result.count", len(articles)))
	}
	dbSpan.End()
	if err != nil {
		log.Printf("Error getting popular recommendations: %v", err)
		// Final fallback to just recent articles
		return rs.GetRecentRecommendations(limit)
	}

	return articles, nil
}

// GetRecentRecommendations returns the most recently published articles
func (rs *RecommendationService) GetRecentRecommendations(limit int) ([]models.Article, error) {
	var articles []models.Article

	err := rs.db.Where("status = ?", "published").
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// GetSimilarArticles returns articles similar to a given article based on category and tags
func (rs *RecommendationService) GetSimilarArticles(articleID uint, limit int) ([]models.Article, error) {
	var articles []models.Article

	// First, get the source article to understand its properties
	var sourceArticle models.Article
	err := rs.db.First(&sourceArticle, articleID).Error
	if err != nil {
		return nil, fmt.Errorf("source article not found: %w", err)
	}

	// Find similar articles based on:
	// 1. Same category
	// 2. Exclude the source article itself
	// 3. Order by recency and popularity
	query := `
		SELECT a.*, COUNT(uai.id) as interaction_count
		FROM articles a
		LEFT JOIN user_article_interactions uai ON a.id = uai.article_id 
			AND uai.interaction_type IN ('view', 'upvote', 'bookmark')
		WHERE a.status = 'published'
		AND a.id != ?
		AND (
			a.category_id = ? 
			OR a.title LIKE ?
		)
		GROUP BY a.id
		ORDER BY 
			CASE WHEN a.category_id = ? THEN 1 ELSE 2 END,
			interaction_count DESC, 
			a.published_at DESC
		LIMIT ?
	`

	// Simple keyword matching for title similarity
	titlePattern := fmt.Sprintf("%%%s%%", extractKeywords(sourceArticle.Title))

	err = rs.db.Raw(query, articleID, sourceArticle.Categories[0].ID, titlePattern, sourceArticle.Categories[0].ID, limit).Scan(&articles).Error
	if err != nil {
		log.Printf("Error getting similar articles: %v", err)
		// Fallback to just same category articles
		err = rs.db.Where("category_id = ? AND id != ? AND status = ?",
			sourceArticle.Categories[0].ID, articleID, "published").
			Order("published_at DESC").
			Limit(limit).
			Find(&articles).Error
	}

	return articles, err
}

// GetCategoryRecommendations returns popular articles from a specific category
func (rs *RecommendationService) GetCategoryRecommendations(categoryID uint, limit int, excludeArticleIDs []uint) ([]models.Article, error) {
	var articles []models.Article

	query := rs.db.Where("category_id = ? AND status = ?", categoryID, "published")

	if len(excludeArticleIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeArticleIDs)
	}

	err := query.Order("published_at DESC").Limit(limit).Find(&articles).Error
	return articles, err
}

// extractKeywords extracts simple keywords from a title for similarity matching
func extractKeywords(title string) string {
	// This is a simple implementation - in production, you might want to use
	// more sophisticated NLP techniques or external services
	if len(title) > 20 {
		return title[:20] // Take first 20 characters as a simple keyword
	}
	return title
}

// GetUserReadingHistory returns articles the user has interacted with
func (rs *RecommendationService) GetUserReadingHistory(userID uint, limit int) ([]models.Article, error) {
	var articles []models.Article

	query := `
		SELECT DISTINCT ON (a.id) a.*, uai.created_at as last_interaction
		FROM articles a
		JOIN user_article_interactions uai ON a.id = uai.article_id
		WHERE uai.user_id = ?
		AND uai.interaction_type IN ('view', 'bookmark')
		ORDER BY a.id, uai.created_at DESC
		LIMIT ?
	`

	err := rs.db.Raw(query, userID, limit).Scan(&articles).Error
	return articles, err
}
