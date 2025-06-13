package services

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"

	"news/internal/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// TranslationService handles all translation-related operations
type TranslationService struct {
	bundle *i18n.Bundle
	db     *sql.DB
}

// NewTranslationService creates a new translation service
func NewTranslationService(db *sql.DB, localesPath string) (*TranslationService, error) {
	// Create bundle for default language
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load translation files
	supportedLanguages := []string{"en", "tr", "es"}
	for _, lang := range supportedLanguages {
		filename := filepath.Join(localesPath, lang+".json")
		if _, err := bundle.LoadMessageFile(filename); err != nil {
			log.Printf("Warning: Failed to load translation file %s: %v", filename, err)
		}
	}

	return &TranslationService{
		bundle: bundle,
		db:     db,
	}, nil
}

// GetBundle returns the i18n bundle
func (ts *TranslationService) GetBundle() *i18n.Bundle {
	return ts.bundle
}

// GetLocalizer creates a localizer for the specified language
func (ts *TranslationService) GetLocalizer(language string) *i18n.Localizer {
	return i18n.NewLocalizer(ts.bundle, language)
}

// Localize translates a message
func (ts *TranslationService) Localize(language, messageID string, templateData map[string]interface{}) (string, error) {
	localizer := ts.GetLocalizer(language)

	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	return localizer.Localize(config)
}

// ArticleTranslation represents a translation record in the database
type ArticleTranslation struct {
	ID                int64    `json:"id" db:"id"`
	ArticleID         int64    `json:"article_id" db:"article_id"`
	Language          string   `json:"language" db:"language"`
	Title             string   `json:"title" db:"title"`
	Content           string   `json:"content" db:"content"`
	Summary           *string  `json:"summary" db:"summary"`
	MetaDescription   *string  `json:"meta_description" db:"meta_description"`
	Slug              string   `json:"slug" db:"slug"`
	TranslationStatus string   `json:"translation_status" db:"translation_status"`
	TranslationSource string   `json:"translation_source" db:"translation_source"`
	QualityScore      *float64 `json:"quality_score" db:"quality_score"`
	TranslatorID      *int64   `json:"translator_id" db:"translator_id"`
	ReviewerID        *int64   `json:"reviewer_id" db:"reviewer_id"`
	CreatedAt         string   `json:"created_at" db:"created_at"`
	UpdatedAt         string   `json:"updated_at" db:"updated_at"`
}

// CreateArticleTranslation creates a new article translation
func (ts *TranslationService) CreateArticleTranslation(ctx context.Context, translation *ArticleTranslation) error {
	query := `
		INSERT INTO article_translations (
			article_id, language, title, content, summary, meta_description, slug,
			translation_status, translation_source, quality_score, translator_id, reviewer_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id, created_at, updated_at`

	err := ts.db.QueryRowContext(ctx, query,
		translation.ArticleID,
		translation.Language,
		translation.Title,
		translation.Content,
		translation.Summary,
		translation.MetaDescription,
		translation.Slug,
		translation.TranslationStatus,
		translation.TranslationSource,
		translation.QualityScore,
		translation.TranslatorID,
		translation.ReviewerID,
	).Scan(&translation.ID, &translation.CreatedAt, &translation.UpdatedAt)

	return err
}

// GetArticleTranslation retrieves a translation by article ID and language
func (ts *TranslationService) GetArticleTranslation(ctx context.Context, articleID int64, language string) (*ArticleTranslation, error) {
	query := `
		SELECT id, article_id, language, title, content, summary, meta_description, slug,
			   translation_status, translation_source, quality_score, translator_id, reviewer_id,
			   created_at, updated_at
		FROM article_translations
		WHERE article_id = $1 AND language = $2`

	translation := &ArticleTranslation{}
	err := ts.db.QueryRowContext(ctx, query, articleID, language).Scan(
		&translation.ID,
		&translation.ArticleID,
		&translation.Language,
		&translation.Title,
		&translation.Content,
		&translation.Summary,
		&translation.MetaDescription,
		&translation.Slug,
		&translation.TranslationStatus,
		&translation.TranslationSource,
		&translation.QualityScore,
		&translation.TranslatorID,
		&translation.ReviewerID,
		&translation.CreatedAt,
		&translation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return translation, err
}

// GetArticleTranslations retrieves all translations for an article
func (ts *TranslationService) GetArticleTranslations(ctx context.Context, articleID int64) ([]*ArticleTranslation, error) {
	query := `
		SELECT id, article_id, language, title, content, summary, meta_description, slug,
			   translation_status, translation_source, quality_score, translator_id, reviewer_id,
			   created_at, updated_at
		FROM article_translations
		WHERE article_id = $1
		ORDER BY language`

	rows, err := ts.db.QueryContext(ctx, query, articleID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Warning: Error closing translation rows: %v", closeErr)
		}
	}()

	var translations []*ArticleTranslation
	for rows.Next() {
		translation := &ArticleTranslation{}
		err := rows.Scan(
			&translation.ID,
			&translation.ArticleID,
			&translation.Language,
			&translation.Title,
			&translation.Content,
			&translation.Summary,
			&translation.MetaDescription,
			&translation.Slug,
			&translation.TranslationStatus,
			&translation.TranslationSource,
			&translation.QualityScore,
			&translation.TranslatorID,
			&translation.ReviewerID,
			&translation.CreatedAt,
			&translation.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		translations = append(translations, translation)
	}

	return translations, rows.Err()
}

// UpdateArticleTranslation updates an existing article translation
func (ts *TranslationService) UpdateArticleTranslation(ctx context.Context, translation *ArticleTranslation) error {
	query := `
		UPDATE article_translations SET
			title = $3, content = $4, summary = $5, meta_description = $6, slug = $7,
			translation_status = $8, translation_source = $9, quality_score = $10,
			translator_id = $11, reviewer_id = $12, updated_at = CURRENT_TIMESTAMP
		WHERE article_id = $1 AND language = $2`

	_, err := ts.db.ExecContext(ctx, query,
		translation.ArticleID,
		translation.Language,
		translation.Title,
		translation.Content,
		translation.Summary,
		translation.MetaDescription,
		translation.Slug,
		translation.TranslationStatus,
		translation.TranslationSource,
		translation.QualityScore,
		translation.TranslatorID,
		translation.ReviewerID,
	)

	return err
}

// DeleteArticleTranslation deletes a translation
func (ts *TranslationService) DeleteArticleTranslation(ctx context.Context, articleID int64, language string) error {
	query := `DELETE FROM article_translations WHERE article_id = $1 AND language = $2`
	_, err := ts.db.ExecContext(ctx, query, articleID, language)
	return err
}

// GetMissingTranslations returns articles that don't have translations for the specified language
func (ts *TranslationService) GetMissingTranslations(ctx context.Context, language string, limit, offset int) ([]int64, error) {
	query := `
		SELECT a.id 
		FROM articles a
		LEFT JOIN article_translations at ON a.id = at.article_id AND at.language = $1
		WHERE at.id IS NULL AND a.status = 'published'
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := ts.db.QueryContext(ctx, query, language, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: Failed to close rows: %v", err)
		}
	}()

	var articleIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		articleIDs = append(articleIDs, id)
	}

	return articleIDs, rows.Err()
}

// GetTranslationStats returns translation statistics
func (ts *TranslationService) GetTranslationStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			language,
			COUNT(*) as total_translations,
			COUNT(CASE WHEN translation_status = 'approved' THEN 1 END) as approved_translations,
			COUNT(CASE WHEN translation_status = 'machine_translated' THEN 1 END) as machine_translations,
			COUNT(CASE WHEN translation_status = 'human_reviewed' THEN 1 END) as human_reviewed,
			AVG(CASE WHEN quality_score IS NOT NULL THEN quality_score END) as avg_quality_score
		FROM article_translations
		GROUP BY language`

	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Warning: Failed to close stats rows: %v", err)
		}
	}()

	stats := make(map[string]interface{})
	languageStats := make(map[string]map[string]interface{})

	for rows.Next() {
		var language string
		var total, approved, machine, humanReviewed int
		var avgQuality *float64

		err := rows.Scan(&language, &total, &approved, &machine, &humanReviewed, &avgQuality)
		if err != nil {
			return nil, err
		}

		languageStats[language] = map[string]interface{}{
			"total_translations":    total,
			"approved_translations": approved,
			"machine_translations":  machine,
			"human_reviewed":        humanReviewed,
			"completion_rate":       float64(approved) / float64(total) * 100,
			"avg_quality_score":     avgQuality,
		}
	}

	// Get total articles count
	var totalArticles int
	err = ts.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM articles WHERE status = 'published'").Scan(&totalArticles)
	if err != nil {
		return nil, err
	}

	stats["languages"] = languageStats
	stats["total_articles"] = totalArticles

	return stats, rows.Err()
}
