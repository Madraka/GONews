package repositories

import (
	"news/internal/database"
	"news/internal/metrics"
	"news/internal/models"
)

// FetchArticlesWithPagination retrieves articles with pagination and optional filtering
func FetchArticlesWithPagination(offset, limit int, category string) ([]models.Article, int, error) {
	// Track database operation
	defer metrics.TrackDatabaseOperation("fetch_articles_with_pagination")()

	var articles []models.Article
	var total int64

	// Build the query for articles table with published status
	// Using optimized index: idx_articles_status_created_at
	query := database.DB.Model(&models.Article{}).Where("status = ?", "published")

	// Add category filter if provided
	if category != "" {
		// Use subquery instead of JOIN for better performance
		query = query.Where("id IN (?)",
			database.DB.Model(&models.Article{}).Select("articles.id").
				Joins("JOIN article_categories ON articles.id = article_categories.article_id").
				Joins("JOIN categories ON article_categories.category_id = categories.id").
				Where("categories.slug = ? OR categories.name = ?", category, category))
	}

	// Count total matching records (for pagination info)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Execute the query with pagination - optimized for empty data
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	// Skip relations loading if no articles found (major optimization for empty data)
	if len(articles) == 0 {
		return articles, int(total), nil
	}

	// Load relations efficiently using IN queries instead of preload
	articleIDs := getArticleIDs(articles)

	// Use GORM preload but only on the found articles
	if err := database.DB.Preload("Author").Preload("Categories").Preload("Tags").
		Where("id IN ?", articleIDs).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, int(total), nil
}

// Helper function to extract article IDs
func getArticleIDs(articles []models.Article) []uint {
	ids := make([]uint, len(articles))
	for i, article := range articles {
		ids[i] = article.ID
	}
	return ids
}

// GetArticleByID retrieves a single article by ID
func GetArticleByID(id string) (models.Article, error) {
	// Track database operation
	defer metrics.TrackDatabaseOperation("get_article_by_id")()

	var article models.Article
	if err := database.DB.Preload("Author").Preload("Categories").Preload("Tags").
		Where("id = ? AND status = ?", id, "published").First(&article).Error; err != nil {
		return models.Article{}, err
	}

	return article, nil
}

// InsertArticle creates a new article
func InsertArticle(article models.Article) (models.Article, error) {
	// Track database operation
	defer metrics.TrackDatabaseOperation("insert_article")()

	if err := database.DB.Create(&article).Error; err != nil {
		return models.Article{}, err
	}
	return article, nil
}

// UpdateArticle updates an existing article
func UpdateArticle(article models.Article) error {
	// Track database operation
	defer metrics.TrackDatabaseOperation("update_article")()

	if err := database.DB.Save(&article).Error; err != nil {
		return err
	}
	return nil
}

// DeleteArticleByID deletes an article by ID
func DeleteArticleByID(id string) error {
	// Track database operation
	defer metrics.TrackDatabaseOperation("delete_article")()

	if err := database.DB.Delete(&models.Article{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// FetchAllArticles retrieves all published articles
func FetchAllArticles() ([]models.Article, error) {
	// Track database operation
	defer metrics.TrackDatabaseOperation("fetch_all_articles")()

	var articles []models.Article
	if err := database.DB.Where("status = ?", "published").
		Preload("Author").Preload("Categories").Preload("Tags").
		Order("created_at DESC").Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
