package models

import (
	"time"

	"gorm.io/gorm"
)

// CategoryTranslation represents translated content for categories
type CategoryTranslation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CategoryID  uint           `gorm:"not null;index" json:"category_id"`
	Language    string         `gorm:"size:5;not null;index" json:"language"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Slug        string         `gorm:"size:100;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	MetaTitle   string         `gorm:"size:255" json:"meta_title"`
	MetaDesc    string         `gorm:"size:255" json:"meta_description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName specifies the table name for CategoryTranslation
func (CategoryTranslation) TableName() string {
	return "category_translations"
}

// TagTranslation represents translated content for tags
type TagTranslation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TagID       uint           `gorm:"not null;index" json:"tag_id"`
	Language    string         `gorm:"size:5;not null;index" json:"language"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Slug        string         `gorm:"size:50;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tag Tag `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// TableName specifies the table name for TagTranslation
func (TagTranslation) TableName() string {
	return "tag_translations"
}

// MenuTranslation represents translated content for menus
type MenuTranslation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	MenuID      uint           `gorm:"not null;index" json:"menu_id"`
	Language    string         `gorm:"size:5;not null;index" json:"language"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description *string        `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Menu Menu `gorm:"foreignKey:MenuID" json:"menu,omitempty"`
}

// TableName specifies the table name for MenuTranslation
func (MenuTranslation) TableName() string {
	return "menu_translations"
}

// MenuItemTranslation represents translated content for menu items
type MenuItemTranslation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	MenuItemID uint           `gorm:"not null;index" json:"menu_item_id"`
	Language   string         `gorm:"size:5;not null;index" json:"language"`
	Title      string         `gorm:"size:100;not null" json:"title"`
	URL        string         `gorm:"size:255" json:"url"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	MenuItem MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}

// TableName specifies the table name for MenuItemTranslation
func (MenuItemTranslation) TableName() string {
	return "menu_item_translations"
}

// NotificationTranslation represents translated content for notifications
type NotificationTranslation struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	NotificationID uint           `gorm:"not null;index" json:"notification_id"`
	Language       string         `gorm:"size:5;not null;index" json:"language"`
	Title          string         `gorm:"size:255;not null" json:"title"`
	Message        string         `gorm:"type:text;not null" json:"message"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Notification Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
}

// TableName specifies the table name for NotificationTranslation
func (NotificationTranslation) TableName() string {
	return "notification_translations"
}

// SettingTranslation represents translated content for system settings
type SettingTranslation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	SettingKey  string         `gorm:"size:100;not null;index" json:"setting_key"`
	Language    string         `gorm:"size:5;not null;index" json:"language"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SettingTranslation
func (SettingTranslation) TableName() string {
	return "setting_translations"
}

// NOTE: TranslationQueue model moved to translation_queue.go to avoid duplication

// LocalizedCategory represents a category with localized content
type LocalizedCategory struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	MetaTitle   string    `json:"meta_title"`
	MetaDesc    string    `json:"meta_description"`
	Color       string    `json:"color"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LocalizedTag represents a tag with localized content
type LocalizedTag struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	UsageCount  int       `json:"usage_count"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LocalizedMenu represents a menu with localized content
type LocalizedMenu struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Location    string              `json:"location"`
	Description *string             `json:"description"`
	Language    string              `json:"language"`
	Items       []LocalizedMenuItem `json:"items,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// LocalizedMenuItem represents a menu item with localized content
type LocalizedMenuItem struct {
	ID        uint                `json:"id"`
	MenuID    uint                `json:"menu_id"`
	ParentID  *uint               `json:"parent_id"`
	Title     string              `json:"title"`
	URL       string              `json:"url"`
	Icon      string              `json:"icon"`
	Target    string              `json:"target"`
	SortOrder int                 `json:"sort_order"`
	Language  string              `json:"language"`
	Children  []LocalizedMenuItem `json:"children,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// LocalizedNotification represents a notification with localized content
type LocalizedNotification struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Data      string     `json:"data"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
}

// TranslationProgress represents translation progress for entities
type TranslationProgress struct {
	EntityType         string   `json:"entity_type"`
	TotalEntities      int      `json:"total_entities"`
	TranslatedCount    int      `json:"translated_count"`
	PendingCount       int      `json:"pending_count"`
	CompletionRate     float64  `json:"completion_rate"`
	AvailableLanguages []string `json:"available_languages"`
}

// BulkTranslationRequest represents bulk translation request
type BulkTranslationRequest struct {
	EntityType       string   `json:"entity_type" binding:"required"`
	EntityIDs        []uint   `json:"entity_ids"`
	TargetLanguages  []string `json:"target_languages" binding:"required"`
	SourceLanguage   string   `json:"source_language" binding:"omitempty"`
	Priority         int      `json:"priority"`
	ForceRetranslate bool     `json:"force_retranslate"`
}

// BulkTranslationResponse represents bulk translation response
type BulkTranslationResponse struct {
	QueuedJobs    int    `json:"queued_jobs"`
	SkippedJobs   int    `json:"skipped_jobs"`
	FailedJobs    int    `json:"failed_jobs"`
	EstimatedTime string `json:"estimated_time"`
	JobIDs        []uint `json:"job_ids"`
}

// PageTranslation represents translated content for pages
type PageTranslation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PageID      uint           `gorm:"not null;index" json:"page_id"`
	Language    string         `gorm:"size:5;not null;index" json:"language"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Slug        string         `gorm:"size:255;not null;index" json:"slug"`
	MetaTitle   string         `gorm:"size:255" json:"meta_title"`
	MetaDesc    string         `gorm:"size:255" json:"meta_description"`
	ExcerptText string         `gorm:"type:text" json:"excerpt_text"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Page Page `gorm:"foreignKey:PageID" json:"page,omitempty"`
}

// TableName specifies the table name for PageTranslation
func (PageTranslation) TableName() string {
	return "page_translations"
}

// LocalizedPage represents a page with localized content
type LocalizedPage struct {
	ID              uint       `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	MetaTitle       string     `json:"meta_title"`
	MetaDescription string     `json:"meta_description"`
	ExcerptText     string     `json:"excerpt_text"`
	Template        string     `json:"template"`
	Layout          string     `json:"layout"`
	Status          string     `json:"status"`
	FeaturedImage   string     `json:"featured_image"`
	Language        string     `json:"language"`
	IsHomepage      bool       `json:"is_homepage"`
	IsLandingPage   bool       `json:"is_landing_page"`
	Views           int        `json:"views"`
	PublishedAt     *time.Time `json:"published_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// PageContentBlockTranslation represents translated content for page content blocks
type PageContentBlockTranslation struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BlockID   uint           `gorm:"not null;index" json:"block_id"`
	Language  string         `gorm:"size:5;not null;index" json:"language"`
	Content   string         `gorm:"type:text" json:"content"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Block PageContentBlock `gorm:"foreignKey:BlockID" json:"block,omitempty"`
}

// TableName specifies the table name for PageContentBlockTranslation
func (PageContentBlockTranslation) TableName() string {
	return "page_content_block_translations"
}

// ArticleContentBlockTranslation represents translated content for article content blocks
type ArticleContentBlockTranslation struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BlockID   uint           `gorm:"not null;index" json:"block_id"`
	Language  string         `gorm:"size:5;not null;index" json:"language"`
	Content   string         `gorm:"type:text" json:"content"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Block ArticleContentBlock `gorm:"foreignKey:BlockID" json:"block,omitempty"`
}

// TableName specifies the table name for ArticleContentBlockTranslation
func (ArticleContentBlockTranslation) TableName() string {
	return "article_content_block_translations"
}
