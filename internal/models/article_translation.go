package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ArticleTranslation represents translated content for articles
type ArticleTranslation struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ArticleID       uint           `gorm:"not null;index" json:"article_id"`
	Language        string         `gorm:"size:5;not null;index" json:"language"` // ISO 639-1 code (en, tr, es, etc.)
	Title           string         `gorm:"size:255;not null" json:"title"`
	Slug            string         `gorm:"size:255;not null" json:"slug"`
	Summary         string         `gorm:"type:text" json:"summary"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	MetaTitle       string         `gorm:"column:meta_title;size:255" json:"meta_title"`
	MetaDescription string         `gorm:"column:meta_description;size:255" json:"meta_description"`
	Status          string         `gorm:"column:translation_status;size:20;not null;default:'draft'" json:"status"`   // draft, published, pending_review
	TranslatedBy    *uint          `gorm:"column:translator_id;index" json:"translated_by"`                            // User who created/edited the translation
	TranslationType string         `gorm:"column:translation_source;size:20;default:'manual'" json:"translation_type"` // manual, ai_generated, auto_translated
	Quality         *float64       `gorm:"column:quality_score;type:decimal(3,2)" json:"quality"`                      // Quality score 0-1 for AI translations
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Article    Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Translator *User   `gorm:"foreignKey:TranslatedBy" json:"translator,omitempty"`
}

// TableName specifies the table name for ArticleTranslation
func (ArticleTranslation) TableName() string {
	return "article_translations"
}

// ValidateStatus validates translation status
func (at *ArticleTranslation) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"draft":              true,
		"machine_translated": true,
		"human_reviewed":     true,
		"approved":           true,
	}
	return allowedStatuses[at.Status]
}

// ValidateLanguage validates language code
func (at *ArticleTranslation) ValidateLanguage() bool {
	allowedLanguages := map[string]bool{
		"en": true, // English
		"tr": true, // Turkish
		"es": true, // Spanish
		"fr": true, // French
		"de": true, // German
		"ar": true, // Arabic
		"zh": true, // Chinese
		"ru": true, // Russian
		"ja": true, // Japanese
		"ko": true, // Korean
	}
	return allowedLanguages[at.Language]
}

// ValidateTranslationType validates translation type
func (at *ArticleTranslation) ValidateTranslationType() bool {
	allowedTypes := map[string]bool{
		"manual": true,
		"openai": true,
		"deepl":  true,
		"google": true,
	}
	return allowedTypes[at.TranslationType]
}

// BeforeCreate GORM hook to validate before creating
func (at *ArticleTranslation) BeforeCreate(tx *gorm.DB) error {
	if !at.ValidateStatus() {
		return gorm.ErrInvalidData
	}
	if !at.ValidateLanguage() {
		return gorm.ErrInvalidData
	}
	if !at.ValidateTranslationType() {
		return gorm.ErrInvalidData
	}
	return nil
}

// BeforeUpdate GORM hook to validate before updating
func (at *ArticleTranslation) BeforeUpdate(tx *gorm.DB) error {
	if !at.ValidateStatus() {
		return gorm.ErrInvalidData
	}
	if !at.ValidateLanguage() {
		return gorm.ErrInvalidData
	}
	if !at.ValidateTranslationType() {
		return gorm.ErrInvalidData
	}
	return nil
}

// TranslationRequest represents a request for creating/updating translations
type TranslationRequest struct {
	Language        string   `json:"language" binding:"required"`
	Title           string   `json:"title" binding:"required"`
	Summary         string   `json:"summary"`
	Content         string   `json:"content" binding:"required"`
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	Status          string   `json:"status"`
	TranslationType string   `json:"translation_type"`
	Quality         *float64 `json:"quality"`
}

// TranslationResponse represents a translation response with additional metadata
type TranslationResponse struct {
	ArticleTranslation
	ArticleTitle string `json:"article_title"`
	AuthorName   string `json:"author_name"`
}

// ArticleWithTranslations represents an article with all its translations and stats
type ArticleWithTranslations struct {
	Article               Article              `json:"article"`
	Translations          []ArticleTranslation `json:"translations"`
	PublishedTranslations []ArticleTranslation `json:"published_translations"`
	Stats                 TranslationStats     `json:"stats"`
}

// LocalizedArticle represents an article with localized content
type LocalizedArticle struct {
	// Base article fields
	ID            uint           `json:"id"`
	AuthorID      uint           `json:"author_id"`
	FeaturedImage string         `json:"featured_image"`
	Gallery       datatypes.JSON `json:"gallery" swaggertype:"array,string"`
	Status        string         `json:"status"`
	PublishedAt   *time.Time     `json:"published_at"`
	ScheduledAt   *time.Time     `json:"scheduled_at"`
	Views         int            `json:"views"`
	ReadTime      int            `json:"read_time"`
	IsBreaking    bool           `json:"is_breaking"`
	IsFeatured    bool           `json:"is_featured"`
	IsSticky      bool           `json:"is_sticky"`
	AllowComments bool           `json:"allow_comments"`
	Source        string         `json:"source"`
	SourceURL     string         `json:"source_url"`
	Language      string         `json:"language"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	// Localized content fields
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Summary         string `json:"summary"`
	Content         string `json:"content"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`

	// Additional metadata
	TranslationExists  bool     `json:"translation_exists"`
	OriginalLanguage   string   `json:"original_language"`
	TranslationType    string   `json:"translation_type,omitempty"`
	TranslationQuality *float64 `json:"translation_quality,omitempty"`

	// Relations
	Author     User       `json:"author,omitempty"`
	Categories []Category `json:"categories,omitempty"`
	Tags       []Tag      `json:"tags,omitempty"`
}

// TranslationStats represents translation statistics for an article
type TranslationStats struct {
	ArticleID             uint     `json:"article_id"`
	TotalTranslations     int      `json:"total_translations"`
	PublishedTranslations int      `json:"published_translations"`
	PendingTranslations   int      `json:"pending_translations"`
	AvailableLanguages    []string `json:"available_languages"`
	MissingLanguages      []string `json:"missing_languages"`
}

// AITranslationJob represents a job for AI translation
type AITranslationJob struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ArticleID   uint      `gorm:"not null;index" json:"article_id"`
	TargetLang  string    `gorm:"size:5;not null" json:"target_language"`
	SourceLang  string    `gorm:"size:5;not null" json:"source_language"`
	Status      string    `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	Provider    string    `gorm:"size:20;not null" json:"provider"`                 // openai, deepl, google
	JobData     string    `gorm:"type:json" json:"job_data"`                        // Provider-specific job data
	Result      string    `gorm:"type:json" json:"result"`                          // Translation result
	ErrorMsg    string    `gorm:"type:text" json:"error_message"`
	RequestedBy uint      `gorm:"not null" json:"requested_by"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Article   Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Requester User    `gorm:"foreignKey:RequestedBy" json:"requester,omitempty"`
}

// TableName specifies the table name for AITranslationJob
func (AITranslationJob) TableName() string {
	return "ai_translation_jobs"
}

// PaginatedLocalizedArticlesResponse represents a paginated response for localized articles
type PaginatedLocalizedArticlesResponse struct {
	Page     int                `json:"page"`
	Limit    int                `json:"limit"`
	Total    int64              `json:"total"`
	Language string             `json:"language"`
	Query    string             `json:"query,omitempty"`
	Articles []LocalizedArticle `json:"articles"`
}

// TranslationStatusUpdate represents a request to update translation status
type TranslationStatusUpdate struct {
	Status   string   `json:"status" binding:"required"`
	Language string   `json:"language" binding:"required"`
	Quality  *float64 `json:"quality,omitempty"`
}

// GlobalTranslationStats represents global translation statistics
type GlobalTranslationStats struct {
	TotalArticles         int            `json:"total_articles"`
	TotalTranslations     int            `json:"total_translations"`
	PublishedTranslations int            `json:"published_translations"`
	PendingTranslations   int            `json:"pending_translations"`
	LanguageDistribution  map[string]int `json:"language_distribution"`
	CoveragePercentage    float64        `json:"coverage_percentage"`
}
