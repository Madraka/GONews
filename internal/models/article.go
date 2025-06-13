package models

import (
	"strings"
	"time"

	"news/internal/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Category represents article categories
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;unique;not null" json:"name"`
	Slug        string         `gorm:"size:100;unique;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"size:7" json:"color"` // Hex color code
	Icon        string         `gorm:"size:50" json:"icon"`
	ParentID    *uint          `gorm:"index" json:"parent_id"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Articles []Article  `gorm:"many2many:article_categories" json:"articles,omitempty"`
}

// Tag represents article tags
type Tag struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:50;unique;not null" json:"name"`
	Slug        string         `gorm:"size:50;unique;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"size:7" json:"color"`
	UsageCount  int            `gorm:"default:0;index" json:"usage_count"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Articles []Article `gorm:"many2many:article_tags" json:"articles,omitempty"`
}

// Article represents a news article with comprehensive features
type Article struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"size:255;not null" json:"title"`
	Slug          string         `gorm:"size:255;unique;not null" json:"slug"`
	Summary       string         `gorm:"type:text" json:"summary"`
	Content       string         `gorm:"type:text;not null" json:"content"`                     // Legacy content - will be migrated to blocks
	ContentType   string         `gorm:"size:20;default:'legacy';not null" json:"content_type"` // legacy, blocks, hybrid
	HasBlocks     bool           `gorm:"default:false" json:"has_blocks"`                       // True if article uses content blocks
	BlocksVersion int            `gorm:"default:1" json:"blocks_version"`                       // For versioning content blocks structure
	AuthorID      uint           `gorm:"not null;index" json:"author_id"`
	FeaturedImage string         `gorm:"size:255" json:"featured_image"`
	Gallery       datatypes.JSON `gorm:"type:json" json:"gallery" swaggertype:"array,string"` // JSON array of image URLs
	Status        string         `gorm:"size:20;not null;default:'draft';index" json:"status"`
	PublishedAt   *time.Time     `gorm:"index" json:"published_at"`
	ScheduledAt   *time.Time     `json:"scheduled_at"`
	Views         int            `gorm:"default:0;index" json:"views"`
	ReadTime      int            `gorm:"default:0" json:"read_time"` // in minutes
	IsBreaking    bool           `gorm:"default:false;index" json:"is_breaking"`
	IsFeatured    bool           `gorm:"default:false;index" json:"is_featured"`
	IsSticky      bool           `gorm:"default:false" json:"is_sticky"`
	AllowComments bool           `gorm:"default:true" json:"allow_comments"`
	MetaTitle     string         `gorm:"size:255" json:"meta_title"`
	MetaDesc      string         `gorm:"size:255;column:meta_description" json:"meta_description"`
	Source        string         `gorm:"size:255" json:"source"`
	SourceURL     string         `gorm:"size:255" json:"source_url"`
	Language      string         `gorm:"size:5;default:'tr';index" json:"language"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Author           User                     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Categories       []Category               `gorm:"many2many:article_categories" json:"categories,omitempty"`
	Tags             []Tag                    `gorm:"many2many:article_tags" json:"tags,omitempty"`
	Comments         []Comment                `gorm:"foreignKey:ArticleID" json:"comments,omitempty"`
	Votes            []Vote                   `gorm:"foreignKey:ArticleID" json:"votes,omitempty"`
	Bookmarks        []Bookmark               `gorm:"foreignKey:ArticleID" json:"bookmarks,omitempty"`
	Related          []RelatedArticle         `gorm:"foreignKey:ArticleID" json:"related,omitempty"`
	Translations     []ArticleTranslation     `gorm:"foreignKey:ArticleID" json:"translations,omitempty"`
	UserInteractions []UserArticleInteraction `gorm:"foreignKey:ArticleID" json:"user_interactions,omitempty"`
	ContentBlocks    []ArticleContentBlock    `gorm:"foreignKey:ArticleID;orderBy:position ASC" json:"content_blocks,omitempty"`
}

// ValidateStatus validates article status
func (a *Article) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"draft":     true,
		"published": true,
		"scheduled": true,
		"archived":  true,
		"trash":     true,
	}
	return allowedStatuses[a.Status]
}

// ValidateContentType validates article content type
func (a *Article) ValidateContentType() bool {
	allowedTypes := map[string]bool{
		"legacy": true,
		"blocks": true,
		"hybrid": true,
	}
	return allowedTypes[a.ContentType]
}

// IsUsingBlocks returns true if article uses content blocks system
func (a *Article) IsUsingBlocks() bool {
	return a.HasBlocks && a.ContentType != "legacy"
}

// GetContentForRendering returns content based on content type
func (a *Article) GetContentForRendering() interface{} {
	if a.IsUsingBlocks() && len(a.ContentBlocks) > 0 {
		return a.ContentBlocks
	}
	return a.Content
}

// MigrateToBlocks converts legacy content to content blocks
func (a *Article) MigrateToBlocks() []ArticleContentBlock {
	if a.IsUsingBlocks() {
		return a.ContentBlocks
	}

	var blocks []ArticleContentBlock

	// Create a single text block from legacy content
	if a.Content != "" {
		block := ArticleContentBlock{
			ArticleID: a.ID,
			BlockType: "text",
			Content:   a.Content,
			Position:  1,
			IsVisible: true,
		}

		// Set default settings as JSON
		defaultSettings := block.GetDefaultSettings()
		if settingsJSON, err := json.Marshal(defaultSettings); err == nil {
			block.Settings = datatypes.JSON(settingsJSON)
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// UpdateContentFromBlocks updates legacy content from blocks (for backward compatibility)
func (a *Article) UpdateContentFromBlocks() {
	if !a.IsUsingBlocks() || len(a.ContentBlocks) == 0 {
		return
	}

	var contentParts []string
	for _, block := range a.ContentBlocks {
		if !block.IsVisible {
			continue
		}

		switch block.BlockType {
		case "text", "paragraph":
			contentParts = append(contentParts, block.Content)
		case "heading":
			contentParts = append(contentParts, "# "+block.Content)
		case "quote":
			contentParts = append(contentParts, "> "+block.Content)
		case "image":
			if block.Content != "" {
				contentParts = append(contentParts, "[Image: "+block.Content+"]")
			}
		case "video":
			if block.Content != "" {
				contentParts = append(contentParts, "[Video: "+block.Content+"]")
			}
		default:
			if block.Content != "" {
				contentParts = append(contentParts, block.Content)
			}
		}
	}

	a.Content = strings.Join(contentParts, "\n\n")
}

// ArticleStats represents article statistics
type ArticleStats struct {
	ArticleID      uint `json:"article_id"`
	Views          int  `json:"views"`
	Likes          int  `json:"likes"`
	Dislikes       int  `json:"dislikes"`
	CommentsCount  int  `json:"comments_count"`
	BookmarksCount int  `json:"bookmarks_count"`
	SharesCount    int  `json:"shares_count"`
}

// RelatedArticle represents related articles
type RelatedArticle struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ArticleID uint   `gorm:"not null;index" json:"article_id"`
	RelatedID uint   `gorm:"not null;index" json:"related_id"`
	Score     int    `gorm:"default:0" json:"score"`             // Relevance score
	Type      string `gorm:"size:20;default:'auto'" json:"type"` // auto, manual

	// Relations
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Related Article `gorm:"foreignKey:RelatedID" json:"related,omitempty"`
}

// News is a type alias for Article to maintain backward compatibility
// with legacy handlers. New code should use Article directly.
type News = Article
