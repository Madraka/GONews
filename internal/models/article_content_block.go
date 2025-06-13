package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ArticleContentBlock represents individual content blocks within an article
type ArticleContentBlock struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ArticleID uint           `gorm:"not null;index" json:"article_id"`
	BlockType string         `gorm:"size:50;not null;index" json:"block_type"`       // text, image, video, gallery, quote, code, divider, etc.
	Content   string         `gorm:"type:text" json:"content"`                       // Main content for the block
	Settings  datatypes.JSON `gorm:"type:json" json:"settings" swaggertype:"object"` // JSON for block-specific settings
	Position  int            `gorm:"not null;index" json:"position"`                 // Order within the article
	IsVisible bool           `gorm:"default:true" json:"is_visible"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// ArticleContentBlockSettings represents different settings for various block types
type ArticleContentBlockSettings struct {
	// Text Block Settings
	TextAlign       string `json:"text_align,omitempty"`       // left, center, right, justify
	FontSize        string `json:"font_size,omitempty"`        // small, normal, large, x-large
	FontWeight      string `json:"font_weight,omitempty"`      // normal, bold
	TextColor       string `json:"text_color,omitempty"`       // hex color
	BackgroundColor string `json:"background_color,omitempty"` // hex color

	// Image Block Settings
	ImageURL     string `json:"image_url,omitempty"`
	AltText      string `json:"alt_text,omitempty"`
	Caption      string `json:"caption,omitempty"`
	Width        string `json:"width,omitempty"`         // percentage or pixels
	Height       string `json:"height,omitempty"`        // auto, or pixels
	Alignment    string `json:"alignment,omitempty"`     // left, center, right, full
	BorderRadius string `json:"border_radius,omitempty"` // for rounded corners

	// Video Block Settings
	VideoURL      string `json:"video_url,omitempty"`
	VideoProvider string `json:"video_provider,omitempty"` // youtube, vimeo, local, etc.
	VideoID       string `json:"video_id,omitempty"`       // for youtube/vimeo
	AutoPlay      bool   `json:"autoplay,omitempty"`
	ShowControls  bool   `json:"show_controls,omitempty"`
	Muted         bool   `json:"muted,omitempty"`

	// Gallery Block Settings
	Images       []GalleryImage `json:"images,omitempty"`
	Layout       string         `json:"layout,omitempty"`   // grid, slider, masonry
	Columns      int            `json:"columns,omitempty"`  // for grid layout
	GapSize      string         `json:"gap_size,omitempty"` // small, medium, large
	ShowCaptions bool           `json:"show_captions,omitempty"`

	// Quote Block Settings
	QuoteText   string `json:"quote_text,omitempty"`
	Author      string `json:"author,omitempty"`
	AuthorTitle string `json:"author_title,omitempty"`
	QuoteStyle  string `json:"quote_style,omitempty"` // simple, bordered, highlighted

	// Code Block Settings
	Language        string `json:"language,omitempty"` // programming language
	ShowLineNumbers bool   `json:"show_line_numbers,omitempty"`
	Theme           string `json:"theme,omitempty"` // dark, light

	// Divider Block Settings
	DividerStyle string `json:"divider_style,omitempty"` // line, dots, custom
	DividerColor string `json:"divider_color,omitempty"`
	DividerWidth string `json:"divider_width,omitempty"` // percentage

	// Embed Block Settings
	EmbedURL    string `json:"embed_url,omitempty"`
	EmbedType   string `json:"embed_type,omitempty"` // twitter, instagram, youtube, etc.
	EmbedWidth  string `json:"embed_width,omitempty"`
	EmbedHeight string `json:"embed_height,omitempty"`

	// List Block Settings
	ListType  string     `json:"list_type,omitempty"` // unordered, ordered, checklist
	ListItems []ListItem `json:"list_items,omitempty"`
	ListStyle string     `json:"list_style,omitempty"` // bullets, numbers, checkboxes

	// Table Block Settings
	TableData  [][]string `json:"table_data,omitempty"`
	HasHeader  bool       `json:"has_header,omitempty"`
	TableStyle string     `json:"table_style,omitempty"` // simple, striped, bordered
	Responsive bool       `json:"responsive,omitempty"`

	// Button Block Settings
	ButtonText   string `json:"button_text,omitempty"`
	ButtonURL    string `json:"button_url,omitempty"`
	ButtonStyle  string `json:"button_style,omitempty"` // primary, secondary, outline
	OpenInNewTab bool   `json:"open_in_new_tab,omitempty"`

	// Custom HTML Block Settings
	HTMLContent string `json:"html_content,omitempty"`
	IsSanitized bool   `json:"is_sanitized,omitempty"`

	// Chart Block Settings
	ChartType    string                 `json:"chart_type,omitempty"`  // line, bar, pie, doughnut, area, scatter
	DataSource   string                 `json:"data_source,omitempty"` // manual, api, csv
	ChartData    map[string]interface{} `json:"chart_data,omitempty"`
	ChartOptions map[string]interface{} `json:"chart_options,omitempty"`

	// Map Block Settings
	MapProvider     string      `json:"map_provider,omitempty"` // google, mapbox, openstreetmap
	Latitude        float64     `json:"latitude,omitempty"`
	Longitude       float64     `json:"longitude,omitempty"`
	ZoomLevel       int         `json:"zoom_level,omitempty"`
	MapType         string      `json:"map_type,omitempty"` // roadmap, satellite, hybrid, terrain
	Markers         []MapMarker `json:"markers,omitempty"`
	ShowMapControls bool        `json:"show_map_controls,omitempty"`

	// FAQ Block Settings
	FAQStyle      string    `json:"faq_style,omitempty"` // accordion, tabs, cards
	FAQItems      []FAQItem `json:"faq_items,omitempty"`
	SearchEnabled bool      `json:"search_enabled,omitempty"`
	Categories    []string  `json:"categories,omitempty"`

	// Newsletter Block Settings
	NewsletterTitle       string   `json:"newsletter_title,omitempty"`
	NewsletterDescription string   `json:"newsletter_description,omitempty"`
	FormStyle             string   `json:"form_style,omitempty"`      // inline, modal, sidebar
	RequiredFields        []string `json:"required_fields,omitempty"` // email, name, phone
	SuccessMessage        string   `json:"success_message,omitempty"`
	PrivacyNotice         bool     `json:"privacy_notice,omitempty"`
	GDPRCompliant         bool     `json:"gdpr_compliant,omitempty"`

	// Quiz/Poll Block Settings
	QuizType      string         `json:"quiz_type,omitempty"` // single, multiple, poll, survey
	QuizTitle     string         `json:"quiz_title,omitempty"`
	Questions     []QuizQuestion `json:"questions,omitempty"`
	ShowResults   bool           `json:"show_results,omitempty"`
	AllowRetake   bool           `json:"allow_retake,omitempty"`
	ResultSharing bool           `json:"result_sharing,omitempty"`

	// Comments Block Settings
	CommentSystem string `json:"comment_system,omitempty"` // internal, disqus, facebook
	Moderation    string `json:"moderation,omitempty"`     // auto, manual, none
	AllowReplies  bool   `json:"allow_replies,omitempty"`
	MaxDepth      int    `json:"max_depth,omitempty"`
	SortOrder     string `json:"sort_order,omitempty"` // newest, oldest, popular
	RequireLogin  bool   `json:"require_login,omitempty"`
	ShowCount     bool   `json:"show_count,omitempty"`

	// Rating Block Settings
	RatingType   string `json:"rating_type,omitempty"` // stars, thumbs, numeric
	MaxRating    int    `json:"max_rating,omitempty"`
	AllowReviews bool   `json:"allow_reviews,omitempty"`
	ShowAverage  bool   `json:"show_average,omitempty"`

	// Social Feed Block Settings
	Platform        string `json:"platform,omitempty"`  // twitter, instagram, linkedin, facebook
	FeedType        string `json:"feed_type,omitempty"` // hashtag, user, list
	FeedQuery       string `json:"feed_query,omitempty"`
	PostCount       int    `json:"post_count,omitempty"`
	ShowAvatars     bool   `json:"show_avatars,omitempty"`
	ShowTimestamps  bool   `json:"show_timestamps,omitempty"`
	AutoRefresh     bool   `json:"auto_refresh,omitempty"`
	RefreshInterval int    `json:"refresh_interval,omitempty"` // seconds

	// Hero Section Block Settings
	BackgroundType string      `json:"background_type,omitempty"` // image, video, gradient, color
	BackgroundURL  string      `json:"background_url,omitempty"`
	OverlayColor   string      `json:"overlay_color,omitempty"`
	HeroTitle      string      `json:"hero_title,omitempty"`
	HeroSubtitle   string      `json:"hero_subtitle,omitempty"`
	CTAButtons     []CTAButton `json:"cta_buttons,omitempty"`
	MinHeight      string      `json:"min_height,omitempty"`

	// Card Grid Block Settings
	GridColumns int        `json:"grid_columns,omitempty"`
	CardStyle   string     `json:"card_style,omitempty"` // minimal, shadow, bordered
	Cards       []GridCard `json:"cards,omitempty"`

	// Countdown Timer Block Settings
	TargetDate        string `json:"target_date,omitempty"` // ISO format
	Timezone          string `json:"timezone,omitempty"`
	CountdownFormat   string `json:"countdown_format,omitempty"`  // days, hours, minutes, seconds
	CountdownStyle    string `json:"countdown_style,omitempty"`   // digital, analog, minimal
	CompletionAction  string `json:"completion_action,omitempty"` // hide, show_message, redirect
	CompletionMessage string `json:"completion_message,omitempty"`

	// Search Block Settings
	SearchScope    string   `json:"search_scope,omitempty"` // site, articles, products
	Placeholder    string   `json:"placeholder,omitempty"`
	ShowFilters    bool     `json:"show_filters,omitempty"`
	Filters        []string `json:"filters,omitempty"`
	ResultsPerPage int      `json:"results_per_page,omitempty"`
	SearchAPI      string   `json:"search_api,omitempty"`

	// News Ticker Block Settings
	NewsSource        string `json:"news_source,omitempty"`   // internal, rss, api
	NewsCategory      string `json:"news_category,omitempty"` // breaking, sports, economy, tech
	ScrollSpeed       string `json:"scroll_speed,omitempty"`  // slow, medium, fast
	MaxItems          int    `json:"max_items,omitempty"`
	TickerAutoRefresh bool   `json:"ticker_auto_refresh,omitempty"`

	// Breaking News Banner Block Settings
	AlertLevel    string `json:"alert_level,omitempty"` // low, medium, high, critical
	BannerColor   string `json:"banner_color,omitempty"`
	Animation     string `json:"animation,omitempty"` // slide, fade, pulse
	AutoHide      bool   `json:"auto_hide,omitempty"`
	HideDelay     int    `json:"hide_delay,omitempty"` // milliseconds
	ShowTimestamp bool   `json:"show_timestamp,omitempty"`

	// Product Block Settings
	ProductID         string `json:"product_id,omitempty"`
	DisplayType       string `json:"display_type,omitempty"` // card, list, grid
	ShowPrice         bool   `json:"show_price,omitempty"`
	ShowRating        bool   `json:"show_rating,omitempty"`
	ShowStock         bool   `json:"show_stock,omitempty"`
	BuyButtonText     string `json:"buy_button_text,omitempty"`
	BuyButtonURL      string `json:"buy_button_url,omitempty"`
	AffiliateTracking bool   `json:"affiliate_tracking,omitempty"`
}

// GalleryImage represents an image in a gallery block
type GalleryImage struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text,omitempty"`
	Caption string `json:"caption,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
}

// ListItem represents an item in a list block
type ListItem struct {
	Text     string     `json:"text"`
	Checked  bool       `json:"checked,omitempty"`   // for checklist
	Indent   int        `json:"indent,omitempty"`    // for nested lists
	SubItems []ListItem `json:"sub_items,omitempty"` // for nested lists
}

// MapMarker represents a marker on a map block
type MapMarker struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Icon        string  `json:"icon,omitempty"`
}

// FAQItem represents a question-answer pair in FAQ block
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Category string `json:"category,omitempty"`
}

// QuizQuestion represents a question in quiz/poll block
type QuizQuestion struct {
	Question      string   `json:"question"`
	Type          string   `json:"type"` // single, multiple, text
	Options       []string `json:"options,omitempty"`
	CorrectAnswer int      `json:"correct_answer,omitempty"`
	Points        int      `json:"points,omitempty"`
}

// CTAButton represents a call-to-action button in hero blocks
type CTAButton struct {
	Text  string `json:"text"`
	URL   string `json:"url"`
	Style string `json:"style"` // primary, secondary, outline
}

// GridCard represents a card in card grid block
type GridCard struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image,omitempty"`
	Link    string `json:"link,omitempty"`
}

// ValidateBlockType validates if the block type is supported
func (acb *ArticleContentBlock) ValidateBlockType() bool {
	allowedTypes := map[string]bool{
		"text":      true,
		"heading":   true,
		"paragraph": true,
		"image":     true,
		"video":     true,
		"gallery":   true,
		"quote":     true,
		"code":      true,
		"divider":   true,
		"spacer":    true,
		"embed":     true,
		"list":      true,
		"table":     true,
		"button":    true,
		"html":      true,
		"columns":   true,
		"accordion": true,
		"tabs":      true,
		"alert":     true,
		"callout":   true,
		// New priority block types
		"chart":         true,
		"map":           true,
		"faq":           true,
		"newsletter":    true,
		"quiz":          true,
		"poll":          true,
		"comments":      true,
		"rating":        true,
		"social_feed":   true,
		"hero":          true,
		"card_grid":     true,
		"countdown":     true,
		"search":        true,
		"news_ticker":   true,
		"breaking_news": true,
		"product":       true,
	}
	return allowedTypes[acb.BlockType]
}

// GetDefaultSettings returns default settings for a block type
func (acb *ArticleContentBlock) GetDefaultSettings() ArticleContentBlockSettings {
	switch acb.BlockType {
	case "text", "paragraph":
		return ArticleContentBlockSettings{
			TextAlign:  "left",
			FontSize:   "normal",
			FontWeight: "normal",
			TextColor:  "#000000",
		}
	case "heading":
		return ArticleContentBlockSettings{
			TextAlign:  "left",
			FontSize:   "large",
			FontWeight: "bold",
			TextColor:  "#000000",
		}
	case "image":
		return ArticleContentBlockSettings{
			Alignment:    "center",
			Width:        "100%",
			Height:       "auto",
			BorderRadius: "0px",
		}
	case "video":
		return ArticleContentBlockSettings{
			Width:        "100%",
			AutoPlay:     false,
			ShowControls: true,
			Muted:        false,
		}
	case "gallery":
		return ArticleContentBlockSettings{
			Layout:       "grid",
			Columns:      3,
			GapSize:      "medium",
			ShowCaptions: true,
		}
	case "quote":
		return ArticleContentBlockSettings{
			QuoteStyle: "simple",
			TextAlign:  "center",
		}
	case "code":
		return ArticleContentBlockSettings{
			Language:        "javascript",
			ShowLineNumbers: true,
			Theme:           "dark",
		}
	case "divider":
		return ArticleContentBlockSettings{
			DividerStyle: "line",
			DividerColor: "#cccccc",
			DividerWidth: "100%",
		}
	case "list":
		return ArticleContentBlockSettings{
			ListType:  "unordered",
			ListStyle: "bullets",
		}
	case "table":
		return ArticleContentBlockSettings{
			HasHeader:  true,
			TableStyle: "simple",
			Responsive: true,
		}
	case "button":
		return ArticleContentBlockSettings{
			ButtonStyle:  "primary",
			OpenInNewTab: false,
		}
	case "chart":
		return ArticleContentBlockSettings{
			ChartType: "line",
			ChartOptions: map[string]interface{}{
				"responsive":      true,
				"legend_position": "top",
				"show_grid":       true,
				"animation":       true,
			},
		}
	case "map":
		return ArticleContentBlockSettings{
			MapProvider:     "openstreetmap",
			ZoomLevel:       10,
			MapType:         "roadmap",
			ShowMapControls: true,
			Height:          "400px",
		}
	case "faq":
		return ArticleContentBlockSettings{
			FAQStyle:      "accordion",
			SearchEnabled: false,
		}
	case "newsletter":
		return ArticleContentBlockSettings{
			FormStyle:      "inline",
			RequiredFields: []string{"email"},
			PrivacyNotice:  true,
			GDPRCompliant:  true,
		}
	case "quiz", "poll":
		return ArticleContentBlockSettings{
			QuizType:      "single",
			ShowResults:   true,
			AllowRetake:   true,
			ResultSharing: false,
		}
	case "comments":
		return ArticleContentBlockSettings{
			CommentSystem: "internal",
			Moderation:    "manual",
			AllowReplies:  true,
			MaxDepth:      3,
			SortOrder:     "newest",
			RequireLogin:  true,
			ShowCount:     true,
		}
	case "rating":
		return ArticleContentBlockSettings{
			RatingType:   "stars",
			MaxRating:    5,
			AllowReviews: true,
			ShowAverage:  true,
			RequireLogin: true,
		}
	case "social_feed":
		return ArticleContentBlockSettings{
			Platform:        "twitter",
			FeedType:        "hashtag",
			PostCount:       5,
			ShowAvatars:     true,
			ShowTimestamps:  true,
			AutoRefresh:     false,
			RefreshInterval: 300,
		}
	case "hero":
		return ArticleContentBlockSettings{
			BackgroundType: "image",
			TextAlign:      "center",
			MinHeight:      "500px",
		}
	case "card_grid":
		return ArticleContentBlockSettings{
			GridColumns: 3,
			GapSize:     "medium",
			CardStyle:   "shadow",
		}
	case "countdown":
		return ArticleContentBlockSettings{
			CountdownFormat:  "days",
			CountdownStyle:   "digital",
			CompletionAction: "hide",
			Timezone:         "Europe/Istanbul",
		}
	case "search":
		return ArticleContentBlockSettings{
			SearchScope:    "articles",
			ShowFilters:    true,
			ResultsPerPage: 10,
			Placeholder:    "Arama yapın...",
		}
	case "news_ticker":
		return ArticleContentBlockSettings{
			NewsSource:        "internal",
			ScrollSpeed:       "medium",
			MaxItems:          10,
			TickerAutoRefresh: true,
		}
	case "breaking_news":
		return ArticleContentBlockSettings{
			AlertLevel:    "medium",
			BannerColor:   "#ff0000",
			TextColor:     "#ffffff",
			Animation:     "slide",
			AutoHide:      true,
			HideDelay:     10000,
			ShowTimestamp: true,
		}
	case "product":
		return ArticleContentBlockSettings{
			DisplayType:       "card",
			ShowPrice:         true,
			ShowRating:        true,
			ShowStock:         false,
			BuyButtonText:     "Satın Al",
			AffiliateTracking: false,
		}
	default:
		return ArticleContentBlockSettings{}
	}
}

// TableName specifies the table name for ArticleContentBlock
func (ArticleContentBlock) TableName() string {
	return "article_content_blocks"
}
