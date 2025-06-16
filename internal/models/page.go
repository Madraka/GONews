package models

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Page represents a content page with modern block-based architecture
type Page struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `gorm:"size:255;not null" json:"title"`
	Slug      string `gorm:"size:255;unique;not null" json:"slug"`
	MetaTitle string `gorm:"size:255" json:"meta_title"`
	MetaDesc  string `gorm:"size:255" json:"meta_description"`
	Template  string `gorm:"size:50;default:'default'" json:"template"`
	Layout    string `gorm:"size:50;default:'container'" json:"layout"`
	Status    string `gorm:"size:20;not null;default:'draft'" json:"status"`
	Language  string `gorm:"size:5;default:'tr'" json:"language"`

	// Hierarchy Support
	ParentID  *uint `gorm:"index" json:"parent_id"`
	SortOrder int   `gorm:"default:0" json:"sort_order"`

	// Content Management
	FeaturedImage string `gorm:"size:255" json:"featured_image"`
	ExcerptText   string `gorm:"type:text" json:"excerpt_text"`

	// Settings (JSON fields)
	SEOSettings  datatypes.JSON `gorm:"type:json" json:"seo_settings" swaggertype:"object"`
	PageSettings datatypes.JSON `gorm:"type:json" json:"page_settings" swaggertype:"object"`
	LayoutData   datatypes.JSON `gorm:"type:json" json:"layout_data" swaggertype:"object"`

	// Publishing
	AuthorID    uint       `gorm:"not null;index" json:"author_id"`
	PublishedAt *time.Time `json:"published_at"`
	ScheduledAt *time.Time `json:"scheduled_at"`

	// Analytics
	Views         int  `gorm:"default:0" json:"views"`
	IsHomepage    bool `gorm:"default:false" json:"is_homepage"`
	IsLandingPage bool `gorm:"default:false" json:"is_landing_page"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Author        User               `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Parent        *Page              `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children      []Page             `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	ContentBlocks []PageContentBlock `gorm:"foreignKey:PageID;orderBy:position ASC" json:"content_blocks,omitempty"`
	Translations  []PageTranslation  `gorm:"foreignKey:PageID" json:"translations,omitempty"`
}

// PageSEOSettings represents SEO-specific settings for a page
type PageSEOSettings struct {
	Keywords           []string `json:"keywords,omitempty"`
	CanonicalURL       string   `json:"canonical_url,omitempty"`
	RobotsIndex        bool     `json:"robots_index"`
	RobotsFollow       bool     `json:"robots_follow"`
	OGTitle            string   `json:"og_title,omitempty"`
	OGDescription      string   `json:"og_description,omitempty"`
	OGImage            string   `json:"og_image,omitempty"`
	TwitterCard        string   `json:"twitter_card,omitempty"`
	TwitterTitle       string   `json:"twitter_title,omitempty"`
	TwitterDescription string   `json:"twitter_description,omitempty"`
	TwitterImage       string   `json:"twitter_image,omitempty"`
	Schema             string   `json:"schema,omitempty"` // JSON-LD structured data
}

// PageSettings represents general page settings
type PageSettings struct {
	ShowTitle          bool                   `json:"show_title"`
	ShowAuthor         bool                   `json:"show_author"`
	ShowDate           bool                   `json:"show_date"`
	ShowComments       bool                   `json:"show_comments"`
	ShowSocialShare    bool                   `json:"show_social_share"`
	ShowBreadcrumbs    bool                   `json:"show_breadcrumbs"`
	EnableSearch       bool                   `json:"enable_search"`
	CustomCSS          string                 `json:"custom_css,omitempty"`
	CustomJS           string                 `json:"custom_js,omitempty"`
	LayoutSettings     map[string]interface{} `json:"layout_settings,omitempty"`
	ResponsiveSettings map[string]interface{} `json:"responsive_settings,omitempty"`
}

// PageContentBlock represents individual content blocks within a page
type PageContentBlock struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	PageID      uint           `gorm:"not null;index" json:"page_id"`
	ContainerID *uint          `gorm:"index" json:"container_id"` // For nested containers
	BlockType   string         `gorm:"size:50;not null;index" json:"block_type"`
	Content     string         `gorm:"type:text" json:"content"`
	Settings    datatypes.JSON `gorm:"type:json" json:"settings" swaggertype:"object"`
	Styles      datatypes.JSON `gorm:"type:json" json:"styles" swaggertype:"object"` // Custom CSS styles
	Position    int            `gorm:"not null;index" json:"position"`
	IsVisible   bool           `gorm:"default:true" json:"is_visible"`

	// Container Properties
	IsContainer   bool           `gorm:"default:false" json:"is_container"`
	ContainerType string         `gorm:"size:30" json:"container_type"` // section, row, column, card
	GridSettings  datatypes.JSON `gorm:"type:json" json:"grid_settings" swaggertype:"object"`

	// Responsive Design
	ResponsiveData datatypes.JSON `gorm:"type:json" json:"responsive_data" swaggertype:"object"`

	// AI & Analytics
	AIGenerated     bool           `gorm:"default:false" json:"ai_generated"`
	PerformanceData datatypes.JSON `gorm:"type:json" json:"performance_data" swaggertype:"object"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Page         Page                          `gorm:"foreignKey:PageID" json:"page,omitempty"`
	Container    *PageContentBlock             `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	ChildBlocks  []PageContentBlock            `gorm:"foreignKey:ContainerID" json:"child_blocks,omitempty"`
	Translations []PageContentBlockTranslation `gorm:"foreignKey:BlockID" json:"translations,omitempty"`
}

// PageTemplate represents reusable page templates
type PageTemplate struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Description    string         `gorm:"type:text" json:"description"`
	Category       string         `gorm:"size:50" json:"category"`
	Thumbnail      string         `gorm:"size:255" json:"thumbnail"`
	PreviewImage   string         `gorm:"size:255" json:"preview_image"`
	BlockStructure datatypes.JSON `gorm:"type:json;not null" json:"block_structure" swaggertype:"object"`
	DefaultStyles  datatypes.JSON `gorm:"type:json" json:"default_styles" swaggertype:"object"`
	IsPublic       bool           `gorm:"default:false" json:"is_public"`
	IsPremium      bool           `gorm:"default:false" json:"is_premium"`
	UsageCount     int            `gorm:"default:0" json:"usage_count"`
	Rating         float64        `gorm:"default:0" json:"rating"`
	Tags           datatypes.JSON `gorm:"type:json" json:"tags" swaggertype:"array,string"` // JSON array of tags

	// Author
	CreatorID uint `gorm:"not null;index" json:"creator_id"`

	// Timestamps
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Creator User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// ContainerSettings represents settings for container blocks
type ContainerSettings struct {
	// Layout Settings
	DisplayType    string `json:"display_type,omitempty"`    // flex, grid, block
	FlexDirection  string `json:"flex_direction,omitempty"`  // row, column
	JustifyContent string `json:"justify_content,omitempty"` // flex-start, center, space-between, etc.
	AlignItems     string `json:"align_items,omitempty"`     // stretch, center, flex-start, etc.
	FlexWrap       string `json:"flex_wrap,omitempty"`       // nowrap, wrap
	Gap            string `json:"gap,omitempty"`             // spacing between items

	// Grid Settings (for grid containers)
	GridColumns  int    `json:"grid_columns,omitempty"`
	GridRows     int    `json:"grid_rows,omitempty"`
	GridTemplate string `json:"grid_template,omitempty"`
	GridGap      string `json:"grid_gap,omitempty"`

	// Sizing
	Width     string `json:"width,omitempty"`  // auto, 100%, px, %
	Height    string `json:"height,omitempty"` // auto, vh, px, %
	MaxWidth  string `json:"max_width,omitempty"`
	MinHeight string `json:"min_height,omitempty"`

	// Spacing
	Padding map[string]string `json:"padding,omitempty"` // top, right, bottom, left
	Margin  map[string]string `json:"margin,omitempty"`  // top, right, bottom, left

	// Visual Styling
	BackgroundType  string                 `json:"background_type,omitempty"` // color, image, gradient, video
	BackgroundColor string                 `json:"background_color,omitempty"`
	BackgroundImage string                 `json:"background_image,omitempty"`
	BackgroundVideo string                 `json:"background_video,omitempty"`
	Gradient        map[string]interface{} `json:"gradient,omitempty"`

	// Border & Effects
	BorderRadius string `json:"border_radius,omitempty"`
	BorderWidth  string `json:"border_width,omitempty"`
	BorderColor  string `json:"border_color,omitempty"`
	BorderStyle  string `json:"border_style,omitempty"`
	BoxShadow    string `json:"box_shadow,omitempty"`

	// Responsive Behavior
	ResponsiveRules map[string]interface{} `json:"responsive_rules,omitempty"`

	// Animation & Interaction
	AnimationIn    string                 `json:"animation_in,omitempty"` // fadeIn, slideIn, etc.
	AnimationDelay string                 `json:"animation_delay,omitempty"`
	HoverEffects   map[string]interface{} `json:"hover_effects,omitempty"`

	// Advanced
	CSSClasses []string `json:"css_classes,omitempty"`
	CustomCSS  string   `json:"custom_css,omitempty"`
	ZIndex     int      `json:"z_index,omitempty"`
}

// Model validation and utility methods

// ValidateStatus validates page status
func (p *Page) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"draft":     true,
		"published": true,
		"scheduled": true,
		"private":   true,
		"archived":  true,
	}
	return allowedStatuses[p.Status]
}

// ValidateTemplate validates page template
func (p *Page) ValidateTemplate() bool {
	allowedTemplates := map[string]bool{
		"default":   true,
		"landing":   true,
		"blog":      true,
		"portfolio": true,
		"ecommerce": true,
		"contact":   true,
		"about":     true,
		"custom":    true,
	}
	return allowedTemplates[p.Template]
}

// IsUsingBlocks returns true if page uses content blocks system
func (p *Page) IsUsingBlocks() bool {
	return len(p.ContentBlocks) > 0
}

// GetSEOSettings unmarshals and returns SEO settings
func (p *Page) GetSEOSettings() PageSEOSettings {
	var settings PageSEOSettings
	if len(p.SEOSettings) > 0 {
		if err := json.Unmarshal(p.SEOSettings, &settings); err != nil {
			// Log error and return default settings
			// You might want to log this error properly in your application
			return PageSEOSettings{}
		}
	}
	return settings
}

// GetPageSettings unmarshals and returns page settings
func (p *Page) GetPageSettings() PageSettings {
	var settings PageSettings
	if len(p.PageSettings) > 0 {
		if err := json.Unmarshal(p.PageSettings, &settings); err != nil {
			// Log error and return default settings
			return PageSettings{}
		}
	}
	return settings
}

// GenerateExcerpt creates an excerpt from content blocks
func (p *Page) GenerateExcerpt(maxLength int) string {
	if p.ExcerptText != "" {
		return p.ExcerptText
	}

	var textParts []string
	for _, block := range p.ContentBlocks {
		if !block.IsVisible || block.IsContainer {
			continue
		}

		if block.BlockType == "text" || block.BlockType == "paragraph" {
			textParts = append(textParts, block.Content)
		}
	}

	fullText := strings.Join(textParts, " ")
	if len(fullText) <= maxLength {
		return fullText
	}

	// Truncate at word boundary
	truncated := fullText[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// ValidateBlockType validates if the block type is supported for pages
func (pcb *PageContentBlock) ValidateBlockType() bool {
	// Include all article block types plus page-specific ones
	allowedTypes := map[string]bool{
		// Basic content blocks
		"text": true, "heading": true, "paragraph": true, "image": true, "video": true,
		"gallery": true, "quote": true, "code": true, "divider": true, "spacer": true,
		"embed": true, "list": true, "table": true, "button": true, "html": true,

		// Layout containers
		"container": true, "section": true, "row": true, "column": true, "card": true,
		"tabs": true, "accordion": true, "modal": true,

		// Advanced blocks
		"hero": true, "cta": true, "testimonial": true, "pricing": true, "team": true,
		"features": true, "stats": true, "timeline": true, "portfolio": true,

		// Interactive blocks
		"form": true, "contact_form": true, "newsletter": true, "poll": true, "quiz": true,
		"comments": true, "rating": true, "search": true,

		// Media blocks
		"carousel": true, "slider": true, "lightbox": true, "video_playlist": true,

		// Dynamic content
		"article_list": true, "category_list": true, "tag_cloud": true, "author_bio": true,
		"related_content": true, "popular_content": true, "recent_content": true,

		// Social & Integration
		"social_feed": true, "social_share": true, "instagram_feed": true, "twitter_feed": true,

		// E-commerce (future)
		"product": true, "product_grid": true, "cart": true, "checkout": true,

		// Analytics & SEO
		"analytics": true, "breadcrumbs": true, "sitemap": true,
	}
	return allowedTypes[pcb.BlockType]
}

// IsContainerBlock returns true if this block can contain other blocks
func (pcb *PageContentBlock) IsContainerBlock() bool {
	containerTypes := map[string]bool{
		"container": true, "section": true, "row": true, "column": true,
		"card": true, "tabs": true, "accordion": true, "modal": true,
		"hero": true, "carousel": true, "slider": true,
	}
	return containerTypes[pcb.BlockType] || pcb.IsContainer
}

// GetContainerSettings unmarshals and returns container settings
func (pcb *PageContentBlock) GetContainerSettings() ContainerSettings {
	var settings ContainerSettings
	if len(pcb.Settings) > 0 {
		if err := json.Unmarshal(pcb.Settings, &settings); err != nil {
			// Log error and return default settings
			return ContainerSettings{}
		}
	}
	return settings
}

// Table names
func (Page) TableName() string {
	return "pages"
}

func (PageContentBlock) TableName() string {
	return "page_content_blocks"
}

func (PageTemplate) TableName() string {
	return "page_templates"
}
