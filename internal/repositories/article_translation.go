package repositories

import (
	"fmt"
	"log"
	"strings"

	"news/internal/database"
	"news/internal/models"

	"gorm.io/gorm"
)

// ArticleTranslationRepository handles database operations for article translations
type ArticleTranslationRepository struct {
	db *gorm.DB
}

// NewArticleTranslationRepository creates a new instance of ArticleTranslationRepository
func NewArticleTranslationRepository(db *gorm.DB) *ArticleTranslationRepository {
	return &ArticleTranslationRepository{db: db}
}

// CreateTranslation creates a new article translation (updated signature for handler compatibility)
func (r *ArticleTranslationRepository) CreateTranslation(translation *models.ArticleTranslation) (*models.ArticleTranslation, error) {
	// Generate slug from title
	translation.Slug = GenerateSlug(translation.Title)

	// Check if translation already exists for this article and language
	var existing models.ArticleTranslation
	err := r.db.Where("article_id = ? AND language = ?", translation.ArticleID, translation.Language).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("translation already exists for article %d in language %s", translation.ArticleID, translation.Language)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err := r.db.Create(translation).Error; err != nil {
		return nil, err
	}

	// Load the created translation with relationships
	var createdTranslation models.ArticleTranslation
	err = r.db.Preload("Article").Preload("Translator").First(&createdTranslation, translation.ID).Error
	if err != nil {
		return nil, err
	}

	return &createdTranslation, nil
}

// UpdateTranslation updates an existing article translation by ID with map data
func (r *ArticleTranslationRepository) UpdateTranslation(translationID uint, updateData map[string]interface{}) (*models.ArticleTranslation, error) {
	// Update slug if title is being changed
	if title, exists := updateData["title"]; exists {
		updateData["slug"] = GenerateSlug(title.(string))
	}

	if err := r.db.Model(&models.ArticleTranslation{}).Where("id = ?", translationID).Updates(updateData).Error; err != nil {
		return nil, err
	}

	// Load and return the updated translation
	var updatedTranslation models.ArticleTranslation
	err := r.db.Preload("Article").Preload("Translator").First(&updatedTranslation, translationID).Error
	if err != nil {
		return nil, err
	}

	return &updatedTranslation, nil
}

// GetTranslationByID retrieves a translation by its ID
func (r *ArticleTranslationRepository) GetTranslationByID(id uint) (*models.ArticleTranslation, error) {
	var translation models.ArticleTranslation
	err := r.db.Preload("Article").Preload("Translator").First(&translation, id).Error
	if err != nil {
		return nil, err
	}
	return &translation, nil
}

// GetTranslationByArticleAndLanguage retrieves a translation by article ID and language
func (r *ArticleTranslationRepository) GetTranslationByArticleAndLanguage(articleID uint, language string) (*models.ArticleTranslation, error) {
	var translation models.ArticleTranslation
	err := r.db.Where("article_id = ? AND language = ?", articleID, language).
		Preload("Article").Preload("Translator").First(&translation).Error
	if err != nil {
		return nil, err
	}
	return &translation, nil
}

// GetTranslationByLanguage retrieves a translation by article ID and language (alias method for handler compatibility)
func (r *ArticleTranslationRepository) GetTranslationByLanguage(articleID uint, language string) (*models.ArticleTranslation, error) {
	return r.GetTranslationByArticleAndLanguage(articleID, language)
}

// GetTranslationsByArticleID retrieves all translations for an article
func (r *ArticleTranslationRepository) GetTranslationsByArticleID(articleID uint) ([]models.ArticleTranslation, error) {
	var translations []models.ArticleTranslation
	err := r.db.Where("article_id = ?", articleID).
		Preload("Translator").
		Order("language ASC").
		Find(&translations).Error
	return translations, err
}

// GetPublishedTranslationsByArticleID retrieves only published translations for an article
func (r *ArticleTranslationRepository) GetPublishedTranslationsByArticleID(articleID uint) ([]models.ArticleTranslation, error) {
	var translations []models.ArticleTranslation
	err := r.db.Where("article_id = ? AND status = ? AND is_active = ?", articleID, "published", true).
		Preload("Translator").
		Order("language ASC").
		Find(&translations).Error
	return translations, err
}

// GetTranslationsByLanguage retrieves all translations in a specific language with pagination
func (r *ArticleTranslationRepository) GetTranslationsByLanguage(language string, offset, limit int) ([]models.ArticleTranslation, int64, error) {
	var translations []models.ArticleTranslation
	var total int64

	// Count total
	if err := r.db.Model(&models.ArticleTranslation{}).
		Where("language = ? AND status = ? AND is_active = ?", language, "published", true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("language = ? AND status = ? AND is_active = ?", language, "published", true).
		Preload("Article").Preload("Article.Author").Preload("Translator").
		Order("updated_at DESC").
		Offset(offset).Limit(limit).
		Find(&translations).Error

	return translations, total, err
}

// DeleteTranslation soft deletes a translation
func (r *ArticleTranslationRepository) DeleteTranslation(id uint) error {
	return r.db.Delete(&models.ArticleTranslation{}, id).Error
}

// GetLocalizedArticle retrieves an article with localized content for a specific language
func (r *ArticleTranslationRepository) GetLocalizedArticle(articleID uint, language string) (*models.LocalizedArticle, error) {
	var article models.Article

	// First get the base article with all necessary relations
	err := r.db.Preload("Author").Preload("Categories").Preload("Tags").
		Preload("ContentBlocks").
		First(&article, articleID).Error
	if err != nil {
		return nil, err
	}

	// Try to get translation for the requested language
	var translation models.ArticleTranslation
	translationExists := false
	err = r.db.Where("article_id = ? AND language = ? AND status = ? AND is_active = ?",
		articleID, language, "published", true).First(&translation).Error

	if err == nil {
		translationExists = true
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Build localized article
	localizedArticle := &models.LocalizedArticle{
		// Base article fields
		ID:            article.ID,
		AuthorID:      article.AuthorID,
		FeaturedImage: article.FeaturedImage,
		Gallery:       article.Gallery,
		Status:        article.Status,
		PublishedAt:   article.PublishedAt,
		ScheduledAt:   article.ScheduledAt,
		Views:         article.Views,
		ReadTime:      article.ReadTime,
		IsBreaking:    article.IsBreaking,
		IsFeatured:    article.IsFeatured,
		IsSticky:      article.IsSticky,
		AllowComments: article.AllowComments,
		Source:        article.Source,
		SourceURL:     article.SourceURL,
		Language:      language,
		CreatedAt:     article.CreatedAt,
		UpdatedAt:     article.UpdatedAt,

		// Relations
		Author:     article.Author,
		Categories: article.Categories,
		Tags:       article.Tags,

		// Translation metadata
		TranslationExists: translationExists,
		OriginalLanguage:  article.Language,
	}

	if translationExists {
		// Use translated content
		localizedArticle.Title = translation.Title
		localizedArticle.Slug = translation.Slug
		localizedArticle.Summary = translation.Summary
		localizedArticle.Content = translation.Content
		localizedArticle.MetaTitle = translation.MetaTitle
		localizedArticle.MetaDescription = translation.MetaDescription
		localizedArticle.TranslationType = translation.TranslationType
		localizedArticle.TranslationQuality = translation.Quality
	} else {
		// Use original content
		localizedArticle.Title = article.Title
		localizedArticle.Slug = article.Slug
		localizedArticle.Summary = article.Summary
		localizedArticle.Content = article.Content
		localizedArticle.MetaTitle = article.MetaTitle
		localizedArticle.MetaDescription = article.MetaDesc
	}

	return localizedArticle, nil
}

// GetLocalizedArticlesPaginated retrieves paginated localized articles for a language
func (r *ArticleTranslationRepository) GetLocalizedArticlesPaginated(language string, page, limit int) ([]models.LocalizedArticle, int64, error) {
	var total int64
	var localizedArticles []models.LocalizedArticle

	offset := (page - 1) * limit
	status := "published" // Default to published articles

	// Count total published articles using the same query structure as the main query
	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM articles a
		LEFT JOIN article_translations t ON a.id = t.article_id 
			AND t.language = ? AND t.status = 'published' AND t.is_active = true
		WHERE a.status = ?
	`

	err := r.db.Raw(countQuery, language, status).Scan(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %w", err)
	}

	// Get articles with their translations
	rows, err := r.db.Raw(`
		SELECT 
			a.id, a.title as original_title, a.slug as original_slug, 
			a.summary as original_summary, a.content as original_content,
			a.meta_title as original_meta_title, a.meta_description as original_meta_description,
			a.author_id, a.featured_image, a.gallery, a.status, a.published_at, 
			a.scheduled_at, a.views, a.read_time, a.is_breaking, a.is_featured, 
			a.is_sticky, a.allow_comments, a.source, a.source_url, a.language as original_language,
			a.created_at, a.updated_at,
			COALESCE(t.title, a.title) as title,
			COALESCE(t.slug, a.slug) as slug,
			COALESCE(t.summary, a.summary) as summary,
			COALESCE(t.content, a.content) as content,
			a.meta_title as meta_title,
			COALESCE(t.meta_description, a.meta_description) as meta_description,
			CASE WHEN t.id IS NOT NULL THEN true ELSE false END as translation_exists,
			t.translation_source, t.quality_score as translation_quality
		FROM articles a
		LEFT JOIN article_translations t ON a.id = t.article_id 
			AND t.language = ? AND t.status = 'published' AND t.is_active = true
		WHERE a.status = ?
		ORDER BY a.published_at DESC, a.created_at DESC
		LIMIT ? OFFSET ?
	`, language, status, limit, offset).Rows()

	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Warning: Error closing rows: %v", closeErr)
		}
	}()

	for rows.Next() {
		var article models.LocalizedArticle
		var translationSource *string
		var translationQuality *float64

		var originalTitle, originalSlug, originalSummary, originalContent, originalMetaTitle, originalMetaDescription string

		err := rows.Scan(
			&article.ID, &originalTitle, &originalSlug, &originalSummary, &originalContent,
			&originalMetaTitle, &originalMetaDescription, &article.AuthorID, &article.FeaturedImage,
			&article.Gallery, &article.Status, &article.PublishedAt, &article.ScheduledAt,
			&article.Views, &article.ReadTime, &article.IsBreaking, &article.IsFeatured,
			&article.IsSticky, &article.AllowComments, &article.Source, &article.SourceURL,
			&article.OriginalLanguage, &article.CreatedAt, &article.UpdatedAt,
			&article.Title, &article.Slug, &article.Summary, &article.Content,
			&article.MetaTitle, &article.MetaDescription, &article.TranslationExists,
			&translationSource, &translationQuality,
		)
		if err != nil {
			return nil, 0, err
		}

		article.Language = language
		if translationSource != nil {
			article.TranslationType = *translationSource
		}
		if translationQuality != nil {
			article.TranslationQuality = translationQuality
		}

		localizedArticles = append(localizedArticles, article)
	}

	return localizedArticles, total, nil
}

// GetArticleWithTranslations retrieves an article with all its translations
func (r *ArticleTranslationRepository) GetArticleWithTranslations(articleID uint) (*models.ArticleWithTranslations, error) {
	var article models.Article

	// Get the base article with all necessary relations
	err := r.db.Preload("Author").Preload("Categories").Preload("Tags").
		Preload("ContentBlocks").
		First(&article, articleID).Error
	if err != nil {
		return nil, err
	}

	// Get all translations for this article
	translations, err := r.GetTranslationsByArticleID(articleID)
	if err != nil {
		return nil, err
	}

	// Get published translations
	publishedTranslations, err := r.GetPublishedTranslationsByArticleID(articleID)
	if err != nil {
		return nil, err
	}

	// Get translation stats
	stats, err := r.GetTranslationStatsForArticle(articleID)
	if err != nil {
		return nil, err
	}

	result := &models.ArticleWithTranslations{
		Article:               article,
		Translations:          translations,
		PublishedTranslations: publishedTranslations,
		Stats:                 *stats,
	}

	return result, nil
}

// SearchTranslatedArticles searches for articles in translated content
func (r *ArticleTranslationRepository) SearchTranslatedArticles(query, language string, offset, limit int) ([]models.LocalizedArticle, int64, error) {
	searchTerm := "%" + strings.ToLower(query) + "%"

	var total int64
	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM articles a
		LEFT JOIN article_translations t ON a.id = t.article_id 
			AND t.language = ? AND t.status = 'published' AND t.is_active = true
		WHERE a.status = 'published' AND (
			LOWER(COALESCE(t.title, a.title)) LIKE ? OR
			LOWER(COALESCE(t.content, a.content)) LIKE ? OR
			LOWER(COALESCE(t.summary, a.summary)) LIKE ?
		)
	`

	err := r.db.Raw(countQuery, language, searchTerm, searchTerm, searchTerm).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var localizedArticles []models.LocalizedArticle
	searchQuery := `
		SELECT 
			a.id, a.author_id, a.featured_image, a.gallery, a.status, a.published_at, 
			a.scheduled_at, a.views, a.read_time, a.is_breaking, a.is_featured, 
			a.is_sticky, a.allow_comments, a.source, a.source_url, a.language as original_language,
			a.created_at, a.updated_at,
			COALESCE(t.title, a.title) as title,
			COALESCE(t.slug, a.slug) as slug,
			COALESCE(t.summary, a.summary) as summary,
			COALESCE(t.content, a.content) as content,
			COALESCE(t.meta_title, a.meta_title) as meta_title,
			COALESCE(t.meta_description, a.meta_description) as meta_description,
			CASE WHEN t.id IS NOT NULL THEN true ELSE false END as translation_exists,
			t.translation_type, t.quality as translation_quality
		FROM articles a
		LEFT JOIN article_translations t ON a.id = t.article_id 
			AND t.language = ? AND t.status = 'published' AND t.is_active = true
		WHERE a.status = 'published' AND (
			LOWER(COALESCE(t.title, a.title)) LIKE ? OR
			LOWER(COALESCE(t.content, a.content)) LIKE ? OR
			LOWER(COALESCE(t.summary, a.summary)) LIKE ?
		)
		ORDER BY a.published_at DESC, a.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Raw(searchQuery, language, searchTerm, searchTerm, searchTerm, limit, offset).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Warning: Error closing search rows: %v", closeErr)
		}
	}()

	for rows.Next() {
		var article models.LocalizedArticle
		var translationType *string
		var translationQuality *float64

		err := rows.Scan(
			&article.ID, &article.AuthorID, &article.FeaturedImage, &article.Gallery,
			&article.Status, &article.PublishedAt, &article.ScheduledAt, &article.Views,
			&article.ReadTime, &article.IsBreaking, &article.IsFeatured, &article.IsSticky,
			&article.AllowComments, &article.Source, &article.SourceURL, &article.OriginalLanguage,
			&article.CreatedAt, &article.UpdatedAt, &article.Title, &article.Slug,
			&article.Summary, &article.Content, &article.MetaTitle, &article.MetaDescription,
			&article.TranslationExists, &translationType, &translationQuality,
		)
		if err != nil {
			return nil, 0, err
		}

		article.Language = language
		if translationType != nil {
			article.TranslationType = *translationType
		}
		if translationQuality != nil {
			article.TranslationQuality = translationQuality
		}

		localizedArticles = append(localizedArticles, article)
	}

	return localizedArticles, total, nil
}

// GetTranslationStats retrieves global translation statistics (no article ID parameter)
func (r *ArticleTranslationRepository) GetTranslationStats() (*models.GlobalTranslationStats, error) {
	var stats models.GlobalTranslationStats

	// Get total articles count
	var totalArticles int64
	r.db.Model(&models.Article{}).Where("status = ?", "published").Count(&totalArticles)
	stats.TotalArticles = int(totalArticles)

	// Get total translations count
	var totalTranslations int64
	r.db.Model(&models.ArticleTranslation{}).Where("is_active = ?", true).Count(&totalTranslations)
	stats.TotalTranslations = int(totalTranslations)

	// Get published translations count
	var publishedTranslations int64
	r.db.Model(&models.ArticleTranslation{}).
		Where("status = ? AND is_active = ?", "published", true).
		Count(&publishedTranslations)
	stats.PublishedTranslations = int(publishedTranslations)

	// Get pending translations count
	var pendingTranslations int64
	r.db.Model(&models.ArticleTranslation{}).
		Where("status IN ? AND is_active = ?", []string{"draft", "pending_review"}, true).
		Count(&pendingTranslations)
	stats.PendingTranslations = int(pendingTranslations)

	// Get language statistics
	type LanguageCount struct {
		Language string `json:"language"`
		Count    int64  `json:"count"`
	}

	var languageCounts []LanguageCount
	r.db.Model(&models.ArticleTranslation{}).
		Select("language, COUNT(*) as count").
		Where("status = ? AND is_active = ?", "published", true).
		Group("language").
		Order("count DESC").
		Scan(&languageCounts)

	stats.LanguageDistribution = make(map[string]int)
	for _, lc := range languageCounts {
		stats.LanguageDistribution[lc.Language] = int(lc.Count)
	}

	// Calculate coverage percentage
	if totalArticles > 0 {
		stats.CoveragePercentage = float64(publishedTranslations) / float64(totalArticles) * 100
	}

	return &stats, nil
}

// GetTranslationStatsForArticle retrieves translation statistics for a specific article
func (r *ArticleTranslationRepository) GetTranslationStatsForArticle(articleID uint) (*models.TranslationStats, error) {
	var stats models.TranslationStats
	stats.ArticleID = articleID

	// Get total translations count
	var total int64
	r.db.Model(&models.ArticleTranslation{}).
		Where("article_id = ? AND is_active = ?", articleID, true).
		Count(&total)
	stats.TotalTranslations = int(total)

	// Get published translations count
	var publishedCount int64
	r.db.Model(&models.ArticleTranslation{}).
		Where("article_id = ? AND status = ? AND is_active = ?", articleID, "published", true).
		Count(&publishedCount)
	stats.PublishedTranslations = int(publishedCount)

	// Get pending translations count
	var pendingCount int64
	r.db.Model(&models.ArticleTranslation{}).
		Where("article_id = ? AND status IN ? AND is_active = ?", articleID, []string{"draft", "pending_review"}, true).
		Count(&pendingCount)
	stats.PendingTranslations = int(pendingCount)

	// Get available languages
	var availableLanguages []string
	r.db.Model(&models.ArticleTranslation{}).
		Where("article_id = ? AND status = ? AND is_active = ?", articleID, "published", true).
		Pluck("language", &availableLanguages)
	stats.AvailableLanguages = availableLanguages

	// Define supported languages and find missing ones
	allLanguages := []string{"en", "tr", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"}
	missingLanguages := []string{}

	for _, lang := range allLanguages {
		found := false
		for _, available := range availableLanguages {
			if available == lang {
				found = true
				break
			}
		}
		if !found {
			missingLanguages = append(missingLanguages, lang)
		}
	}
	stats.MissingLanguages = missingLanguages

	return &stats, nil
}

// GenerateSlug creates a URL-friendly slug from title (exported for use in other packages)
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace Turkish characters
	replacements := map[string]string{
		"ı": "i", "ğ": "g", "ü": "u", "ş": "s", "ö": "o", "ç": "c",
		"İ": "i", "Ğ": "g", "Ü": "u", "Ş": "s", "Ö": "o", "Ç": "c",
		// Other common international characters
		"á": "a", "à": "a", "ä": "a", "â": "a", "ā": "a", "ă": "a", "ą": "a",
		"é": "e", "è": "e", "ë": "e", "ê": "e", "ē": "e", "ė": "e", "ę": "e",
		"í": "i", "ì": "i", "ï": "i", "î": "i", "ī": "i", "į": "i",
		"ó": "o", "ò": "o", "ô": "o", "ō": "o", "ő": "o", "ø": "o",
		"ú": "u", "ù": "u", "û": "u", "ū": "u", "ů": "u", "ű": "u",
		"ý": "y", "ÿ": "y", "ñ": "n", "ß": "ss",
	}

	for old, new := range replacements {
		slug = strings.ReplaceAll(slug, old, new)
	}

	// Replace spaces and non-alphanumeric characters with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any character that's not alphanumeric or hyphen
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()

	// Remove multiple consecutive hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Ensure slug is not empty
	if slug == "" {
		slug = "untitled"
	}

	return slug
}

// Global repository instance
var ArticleTranslationRepo *ArticleTranslationRepository

// InitializeArticleTranslationRepository initializes the global repository instance
func InitializeArticleTranslationRepository() {
	ArticleTranslationRepo = NewArticleTranslationRepository(database.DB)
}
