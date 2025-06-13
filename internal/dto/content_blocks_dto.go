package dto

import (
	"news/internal/models"
)

// Content Block Base DTOs
type UpdateContentBlockRequest struct {
	Content   *string `json:"content,omitempty"`
	Settings  *string `json:"settings,omitempty"`
	Position  *int    `json:"position,omitempty"`
	IsVisible *bool   `json:"is_visible,omitempty"`
}

type ReorderBlocksRequest struct {
	BlockPositions map[uint]int `json:"block_positions" binding:"required"`
}

type UpdateArticleBlocksRequest struct {
	Blocks []models.ArticleContentBlock `json:"blocks" binding:"required"`
}

// Embed Functionality DTOs
type DetectEmbedsRequest struct {
	Content string `json:"content" binding:"required"`
}

// Embed suggestion type for DTO usage
type EmbedSuggestion struct {
	URL         string                 `json:"url"`
	EmbedType   string                 `json:"embed_type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"settings"`
	Preview     string                 `json:"preview"`
}

type DetectEmbedsResponse struct {
	Suggestions []EmbedSuggestion `json:"suggestions"`
	Count       int               `json:"count"`
}

type CreateEmbedRequest struct {
	URL   string `json:"url" binding:"required"`
	Title string `json:"title,omitempty"`
}

type AnalyzeURLRequest struct {
	URL string `json:"url" binding:"required"`
}

type AnalyzeURLResponse struct {
	IsEmbeddable bool                   `json:"is_embeddable"`
	URL          string                 `json:"url"`
	EmbedType    string                 `json:"embed_type,omitempty"`
	Title        string                 `json:"title,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
	Preview      string                 `json:"preview,omitempty"`
	Message      string                 `json:"message"`
}

// Advanced Block Creation DTOs
type CreateChartRequest struct {
	ChartData map[string]interface{} `json:"chart_data" binding:"required"`
	Position  int                    `json:"position"`
}

type CreateMapRequest struct {
	Latitude  float64            `json:"latitude" binding:"required"`
	Longitude float64            `json:"longitude" binding:"required"`
	Markers   []models.MapMarker `json:"markers,omitempty"`
	Position  int                `json:"position"`
}

type CreateFAQRequest struct {
	FAQItems []models.FAQItem `json:"faq_items" binding:"required"`
	Position int              `json:"position"`
}

type CreateNewsletterRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
}

type CreateQuizRequest struct {
	QuizType  string                `json:"quiz_type" binding:"required"` // quiz, poll, survey
	Title     string                `json:"title" binding:"required"`
	Questions []models.QuizQuestion `json:"questions" binding:"required"`
	Position  int                   `json:"position"`
}

type CreateCountdownRequest struct {
	TargetDate string `json:"target_date" binding:"required"` // RFC3339 format
	Title      string `json:"title"`
	Position   int    `json:"position"`
}

type CreateNewsTickerRequest struct {
	NewsSource string `json:"news_source"`        // internal, rss, api
	Category   string `json:"category,omitempty"` // breaking, sports, economy, tech
	Position   int    `json:"position"`
}

type CreateBreakingNewsRequest struct {
	Content    string `json:"content" binding:"required"`
	AlertLevel string `json:"alert_level,omitempty"` // low, medium, high, critical
	Position   int    `json:"position"`
}

// Social Feed DTOs
type CreateSocialFeedRequest struct {
	Platform        string `json:"platform" binding:"required"`   // twitter, instagram, linkedin, facebook
	FeedType        string `json:"feed_type" binding:"required"`  // hashtag, user, list
	FeedQuery       string `json:"feed_query" binding:"required"` // #hashtag, @username, etc.
	PostCount       int    `json:"post_count,omitempty"`          // default: 5
	ShowAvatars     bool   `json:"show_avatars,omitempty"`        // default: true
	ShowTimestamps  bool   `json:"show_timestamps,omitempty"`     // default: true
	AutoRefresh     bool   `json:"auto_refresh,omitempty"`        // default: false
	RefreshInterval int    `json:"refresh_interval,omitempty"`    // seconds, default: 300
	Position        int    `json:"position"`
}

// Hero Section DTOs
type CreateHeroRequest struct {
	BackgroundType string       `json:"background_type"` // image, video, gradient, color
	BackgroundURL  string       `json:"background_url"`
	OverlayColor   string       `json:"overlay_color"` // rgba(0,0,0,0.5)
	Title          string       `json:"title" binding:"required"`
	Subtitle       string       `json:"subtitle"`
	CTAButtons     []HeroButton `json:"cta_buttons,omitempty"`
	TextAlign      string       `json:"text_align"` // center, left, right
	MinHeight      string       `json:"min_height"` // 500px
	Position       int          `json:"position"`
}

type HeroButton struct {
	Text  string `json:"text" binding:"required"`
	URL   string `json:"url" binding:"required"`
	Style string `json:"style"` // primary, secondary, outline
}

// Card Grid DTOs
type CreateCardGridRequest struct {
	Columns   int        `json:"columns"`    // default: 3
	GapSize   string     `json:"gap_size"`   // small, medium, large
	CardStyle string     `json:"card_style"` // minimal, shadow, bordered
	Cards     []GridCard `json:"cards" binding:"required"`
	Position  int        `json:"position"`
}

type GridCard struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content"`
	Image   string `json:"image,omitempty"`
	Link    string `json:"link,omitempty"`
}

// Search Block DTOs
type CreateSearchRequest struct {
	SearchScope    string   `json:"search_scope"`      // site, articles, products
	Placeholder    string   `json:"placeholder"`       // "Arama yapın..."
	ShowFilters    bool     `json:"show_filters"`      // default: true
	Filters        []string `json:"filters,omitempty"` // ["kategori", "tarih", "yazar"]
	ResultsPerPage int      `json:"results_per_page"`  // default: 10
	SearchAPI      string   `json:"search_api"`        // "/api/search"
	Position       int      `json:"position"`
}

// Comments Block DTOs
type CreateCommentsRequest struct {
	CommentSystem string `json:"comment_system"` // internal, disqus, facebook
	Moderation    string `json:"moderation"`     // auto, manual, none
	AllowReplies  bool   `json:"allow_replies"`  // default: true
	MaxDepth      int    `json:"max_depth"`      // default: 3
	SortOrder     string `json:"sort_order"`     // newest, oldest, popular
	RequireLogin  bool   `json:"require_login"`  // default: true
	ShowCount     bool   `json:"show_count"`     // default: true
	Position      int    `json:"position"`
}

// Rating Block DTOs
type CreateRatingRequest struct {
	RatingType   string `json:"rating_type"`   // stars, thumbs, numeric
	MaxRating    int    `json:"max_rating"`    // default: 5
	AllowReviews bool   `json:"allow_reviews"` // default: true
	ShowAverage  bool   `json:"show_average"`  // default: true
	RequireLogin bool   `json:"require_login"` // default: true
	Position     int    `json:"position"`
}

// Product Block DTOs
type CreateProductRequest struct {
	ProductID         string `json:"product_id" binding:"required"`
	DisplayType       string `json:"display_type"`    // card, list, grid
	ShowPrice         bool   `json:"show_price"`      // default: true
	ShowRating        bool   `json:"show_rating"`     // default: true
	ShowStock         bool   `json:"show_stock"`      // default: false
	BuyButtonText     string `json:"buy_button_text"` // "Satın Al"
	BuyButtonURL      string `json:"buy_button_url"`
	AffiliateTracking bool   `json:"affiliate_tracking"` // default: false
	Position          int    `json:"position"`
}
