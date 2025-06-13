package services

import (
	"encoding/json"
	"fmt"
	"time"

	"news/internal/dto"
	"news/internal/models"

	"gorm.io/datatypes"
)

// AdvancedBlockService handles creation and management of advanced content blocks
type AdvancedBlockService struct{}

// NewAdvancedBlockService creates a new advanced block service
func NewAdvancedBlockService() *AdvancedBlockService {
	return &AdvancedBlockService{}
}

// CreateChartBlock creates a chart block with data validation
func (abs *AdvancedBlockService) CreateChartBlock(articleID uint, chartData map[string]interface{}, position int) (*models.ArticleContentBlock, error) {
	// Validate chart data structure
	if chartData == nil {
		return nil, fmt.Errorf("chart data is required")
	}

	chartType, ok := chartData["chart_type"].(string)
	if !ok || chartType == "" {
		chartType = "line" // default
	}

	// Validate chart type
	allowedChartTypes := map[string]bool{
		"line": true, "bar": true, "pie": true, "doughnut": true, "area": true, "scatter": true,
	}
	if !allowedChartTypes[chartType] {
		return nil, fmt.Errorf("invalid chart type: %s", chartType)
	}

	settings := models.ArticleContentBlockSettings{
		ChartType: chartType,
		ChartData: chartData,
		ChartOptions: map[string]interface{}{
			"responsive":      true,
			"legend_position": "top",
			"show_grid":       true,
			"animation":       true,
		},
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize chart settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "chart",
		Content:   "Chart",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateMapBlock creates a map block with coordinates and markers
func (abs *AdvancedBlockService) CreateMapBlock(articleID uint, lat, lng float64, markers []models.MapMarker, position int) (*models.ArticleContentBlock, error) {
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("invalid latitude: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("invalid longitude: %f", lng)
	}

	settings := models.ArticleContentBlockSettings{
		MapProvider:     "openstreetmap",
		Latitude:        lat,
		Longitude:       lng,
		ZoomLevel:       10,
		MapType:         "roadmap",
		Markers:         markers,
		ShowMapControls: true,
		Height:          "400px",
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize map settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "map",
		Content:   "Interactive Map",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateFAQBlock creates an FAQ block with questions and answers
func (abs *AdvancedBlockService) CreateFAQBlock(articleID uint, faqItems []models.FAQItem, position int) (*models.ArticleContentBlock, error) {
	if len(faqItems) == 0 {
		return nil, fmt.Errorf("at least one FAQ item is required")
	}

	// Validate FAQ items
	for i, item := range faqItems {
		if item.Question == "" || item.Answer == "" {
			return nil, fmt.Errorf("FAQ item %d: question and answer are required", i+1)
		}
	}

	settings := models.ArticleContentBlockSettings{
		FAQStyle:      "accordion",
		FAQItems:      faqItems,
		SearchEnabled: false,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize FAQ settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "faq",
		Content:   "Frequently Asked Questions",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateNewsletterBlock creates a newsletter signup block
func (abs *AdvancedBlockService) CreateNewsletterBlock(articleID uint, title, description string, position int) (*models.ArticleContentBlock, error) {
	if title == "" {
		title = "Newsletter'a Abone Ol"
	}
	if description == "" {
		description = "En son haberleri kaçırma!"
	}

	settings := models.ArticleContentBlockSettings{
		NewsletterTitle:       title,
		NewsletterDescription: description,
		FormStyle:             "inline",
		RequiredFields:        []string{"email"},
		SuccessMessage:        "Başarıyla abone oldunuz!",
		PrivacyNotice:         true,
		GDPRCompliant:         true,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize newsletter settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "newsletter",
		Content:   title,
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateQuizBlock creates a quiz or poll block
func (abs *AdvancedBlockService) CreateQuizBlock(articleID uint, quizType, title string, questions []models.QuizQuestion, position int) (*models.ArticleContentBlock, error) {
	if quizType != "quiz" && quizType != "poll" && quizType != "survey" {
		return nil, fmt.Errorf("invalid quiz type: %s", quizType)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("at least one question is required")
	}

	// Validate questions
	for i, q := range questions {
		if q.Question == "" {
			return nil, fmt.Errorf("question %d: question text is required", i+1)
		}
		if len(q.Options) < 2 && q.Type != "text" {
			return nil, fmt.Errorf("question %d: at least 2 options required for multiple choice", i+1)
		}
	}

	settings := models.ArticleContentBlockSettings{
		QuizType:      quizType,
		QuizTitle:     title,
		Questions:     questions,
		ShowResults:   true,
		AllowRetake:   true,
		ResultSharing: false,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize quiz settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: quizType,
		Content:   title,
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateCountdownBlock creates a countdown timer block
func (abs *AdvancedBlockService) CreateCountdownBlock(articleID uint, targetDate time.Time, title string, position int) (*models.ArticleContentBlock, error) {
	if targetDate.Before(time.Now()) {
		return nil, fmt.Errorf("target date must be in the future")
	}

	if title == "" {
		title = "Countdown Timer"
	}

	settings := models.ArticleContentBlockSettings{
		TargetDate:        targetDate.Format(time.RFC3339),
		CountdownFormat:   "days",
		CountdownStyle:    "digital",
		CompletionAction:  "hide",
		Timezone:          "Europe/Istanbul",
		CompletionMessage: "Süre doldu!",
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize countdown settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "countdown",
		Content:   title,
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateNewsTickerBlock creates a news ticker block
func (abs *AdvancedBlockService) CreateNewsTickerBlock(articleID uint, newsSource, category string, position int) (*models.ArticleContentBlock, error) {
	if newsSource == "" {
		newsSource = "internal"
	}
	if category == "" {
		category = "breaking"
	}

	// Validate news source
	allowedSources := map[string]bool{
		"internal": true, "rss": true, "api": true,
	}
	if !allowedSources[newsSource] {
		return nil, fmt.Errorf("invalid news source: %s", newsSource)
	}

	settings := models.ArticleContentBlockSettings{
		NewsSource:        newsSource,
		NewsCategory:      category,
		ScrollSpeed:       "medium",
		MaxItems:          10,
		TickerAutoRefresh: true,
		RefreshInterval:   60,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize news ticker settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "news_ticker",
		Content:   "Breaking News",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateBreakingNewsBanner creates a breaking news banner block
func (abs *AdvancedBlockService) CreateBreakingNewsBanner(articleID uint, content, alertLevel string, position int) (*models.ArticleContentBlock, error) {
	if content == "" {
		return nil, fmt.Errorf("breaking news content is required")
	}

	if alertLevel == "" {
		alertLevel = "medium"
	}

	// Validate alert level
	allowedLevels := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	if !allowedLevels[alertLevel] {
		return nil, fmt.Errorf("invalid alert level: %s", alertLevel)
	}

	settings := models.ArticleContentBlockSettings{
		AlertLevel:    alertLevel,
		BannerColor:   "#ff0000",
		TextColor:     "#ffffff",
		Animation:     "slide",
		AutoHide:      true,
		HideDelay:     10000,
		ShowTimestamp: true,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize breaking news settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "breaking_news",
		Content:   content,
		Settings:  datatypes.JSON(settingsJSON),
		Position:  position,
		IsVisible: true,
	}

	return block, nil
}

// CreateSocialFeedBlock creates a social media feed block
func (abs *AdvancedBlockService) CreateSocialFeedBlock(articleID uint, request dto.CreateSocialFeedRequest) (*models.ArticleContentBlock, error) {
	// Validate platform
	allowedPlatforms := map[string]bool{
		"twitter": true, "instagram": true, "linkedin": true, "facebook": true,
	}
	if !allowedPlatforms[request.Platform] {
		return nil, fmt.Errorf("invalid platform: %s", request.Platform)
	}

	// Validate feed type
	allowedFeedTypes := map[string]bool{
		"hashtag": true, "user": true, "list": true,
	}
	if !allowedFeedTypes[request.FeedType] {
		return nil, fmt.Errorf("invalid feed type: %s", request.FeedType)
	}

	if request.FeedQuery == "" {
		return nil, fmt.Errorf("feed query is required")
	}

	// Set defaults
	if request.PostCount == 0 {
		request.PostCount = 5
	}
	if request.RefreshInterval == 0 {
		request.RefreshInterval = 300
	}

	settings := models.ArticleContentBlockSettings{
		Platform:        request.Platform,
		FeedType:        request.FeedType,
		FeedQuery:       request.FeedQuery,
		PostCount:       request.PostCount,
		ShowAvatars:     request.ShowAvatars,
		ShowTimestamps:  request.ShowTimestamps,
		AutoRefresh:     request.AutoRefresh,
		RefreshInterval: request.RefreshInterval,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize social feed settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "social_feed",
		Content:   fmt.Sprintf("%s %s Feed", request.Platform, request.FeedType),
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateHeroBlock creates a hero section block
func (abs *AdvancedBlockService) CreateHeroBlock(articleID uint, request dto.CreateHeroRequest) (*models.ArticleContentBlock, error) {
	if request.Title == "" {
		return nil, fmt.Errorf("hero title is required")
	}

	// Set defaults
	if request.BackgroundType == "" {
		request.BackgroundType = "image"
	}
	if request.TextAlign == "" {
		request.TextAlign = "center"
	}
	if request.MinHeight == "" {
		request.MinHeight = "500px"
	}

	// Convert DTO buttons to model buttons
	var ctaButtons []models.CTAButton
	for _, btn := range request.CTAButtons {
		ctaButtons = append(ctaButtons, models.CTAButton{
			Text:  btn.Text,
			URL:   btn.URL,
			Style: btn.Style,
		})
	}

	settings := models.ArticleContentBlockSettings{
		BackgroundType: request.BackgroundType,
		BackgroundURL:  request.BackgroundURL,
		OverlayColor:   request.OverlayColor,
		HeroTitle:      request.Title,
		HeroSubtitle:   request.Subtitle,
		CTAButtons:     ctaButtons,
		TextAlign:      request.TextAlign,
		MinHeight:      request.MinHeight,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize hero settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "hero",
		Content:   request.Title,
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateCardGridBlock creates a card grid block
func (abs *AdvancedBlockService) CreateCardGridBlock(articleID uint, request dto.CreateCardGridRequest) (*models.ArticleContentBlock, error) {
	if len(request.Cards) == 0 {
		return nil, fmt.Errorf("at least one card is required")
	}

	// Set defaults
	if request.Columns == 0 {
		request.Columns = 3
	}
	if request.GapSize == "" {
		request.GapSize = "medium"
	}
	if request.CardStyle == "" {
		request.CardStyle = "shadow"
	}

	// Convert DTO cards to model cards
	var cards []models.GridCard
	for _, card := range request.Cards {
		cards = append(cards, models.GridCard{
			Title:   card.Title,
			Content: card.Content,
			Image:   card.Image,
			Link:    card.Link,
		})
	}

	settings := models.ArticleContentBlockSettings{
		GridColumns: request.Columns,
		GapSize:     request.GapSize,
		CardStyle:   request.CardStyle,
		Cards:       cards,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize card grid settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "card_grid",
		Content:   "Card Grid",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateSearchBlock creates a search interface block
func (abs *AdvancedBlockService) CreateSearchBlock(articleID uint, request dto.CreateSearchRequest) (*models.ArticleContentBlock, error) {
	// Set defaults
	if request.SearchScope == "" {
		request.SearchScope = "articles"
	}
	if request.Placeholder == "" {
		request.Placeholder = "Arama yapın..."
	}
	if request.ResultsPerPage == 0 {
		request.ResultsPerPage = 10
	}
	if request.SearchAPI == "" {
		request.SearchAPI = "/api/search"
	}

	settings := models.ArticleContentBlockSettings{
		SearchScope:    request.SearchScope,
		Placeholder:    request.Placeholder,
		ShowFilters:    request.ShowFilters,
		Filters:        request.Filters,
		ResultsPerPage: request.ResultsPerPage,
		SearchAPI:      request.SearchAPI,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize search settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "search",
		Content:   "Search",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateCommentsBlock creates a comments section block
func (abs *AdvancedBlockService) CreateCommentsBlock(articleID uint, request dto.CreateCommentsRequest) (*models.ArticleContentBlock, error) {
	// Set defaults
	if request.CommentSystem == "" {
		request.CommentSystem = "internal"
	}
	if request.Moderation == "" {
		request.Moderation = "manual"
	}
	if request.MaxDepth == 0 {
		request.MaxDepth = 3
	}
	if request.SortOrder == "" {
		request.SortOrder = "newest"
	}

	settings := models.ArticleContentBlockSettings{
		CommentSystem: request.CommentSystem,
		Moderation:    request.Moderation,
		AllowReplies:  request.AllowReplies,
		MaxDepth:      request.MaxDepth,
		SortOrder:     request.SortOrder,
		RequireLogin:  request.RequireLogin,
		ShowCount:     request.ShowCount,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize comments settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "comments",
		Content:   "Comments",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateRatingBlock creates a rating/review block
func (abs *AdvancedBlockService) CreateRatingBlock(articleID uint, request dto.CreateRatingRequest) (*models.ArticleContentBlock, error) {
	// Set defaults
	if request.RatingType == "" {
		request.RatingType = "stars"
	}
	if request.MaxRating == 0 {
		request.MaxRating = 5
	}

	// Validate rating type
	allowedRatingTypes := map[string]bool{
		"stars": true, "thumbs": true, "numeric": true,
	}
	if !allowedRatingTypes[request.RatingType] {
		return nil, fmt.Errorf("invalid rating type: %s", request.RatingType)
	}

	settings := models.ArticleContentBlockSettings{
		RatingType:   request.RatingType,
		MaxRating:    request.MaxRating,
		AllowReviews: request.AllowReviews,
		ShowAverage:  request.ShowAverage,
		RequireLogin: request.RequireLogin,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize rating settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "rating",
		Content:   "Rating & Reviews",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// CreateProductBlock creates a product showcase block
func (abs *AdvancedBlockService) CreateProductBlock(articleID uint, request dto.CreateProductRequest) (*models.ArticleContentBlock, error) {
	if request.ProductID == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	// Set defaults
	if request.DisplayType == "" {
		request.DisplayType = "card"
	}
	if request.BuyButtonText == "" {
		request.BuyButtonText = "Satın Al"
	}

	settings := models.ArticleContentBlockSettings{
		ProductID:         request.ProductID,
		DisplayType:       request.DisplayType,
		ShowPrice:         request.ShowPrice,
		ShowRating:        request.ShowRating,
		ShowStock:         request.ShowStock,
		BuyButtonText:     request.BuyButtonText,
		BuyButtonURL:      request.BuyButtonURL,
		AffiliateTracking: request.AffiliateTracking,
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize product settings: %w", err)
	}

	block := &models.ArticleContentBlock{
		ArticleID: articleID,
		BlockType: "product",
		Content:   "Product Showcase",
		Settings:  datatypes.JSON(settingsJSON),
		Position:  request.Position,
		IsVisible: true,
	}

	return block, nil
}

// ValidateAdvancedBlockData validates data for advanced block types
func (abs *AdvancedBlockService) ValidateAdvancedBlockData(blockType string, data map[string]interface{}) error {
	switch blockType {
	case "chart":
		if data["chart_data"] == nil {
			return fmt.Errorf("chart_data is required for chart blocks")
		}
	case "map":
		lat, hasLat := data["latitude"]
		lng, hasLng := data["longitude"]
		if !hasLat || !hasLng {
			return fmt.Errorf("latitude and longitude are required for map blocks")
		}
		if latFloat, ok := lat.(float64); !ok || latFloat < -90 || latFloat > 90 {
			return fmt.Errorf("invalid latitude value")
		}
		if lngFloat, ok := lng.(float64); !ok || lngFloat < -180 || lngFloat > 180 {
			return fmt.Errorf("invalid longitude value")
		}
	case "faq":
		if data["faq_items"] == nil {
			return fmt.Errorf("faq_items is required for FAQ blocks")
		}
	case "countdown":
		if data["target_date"] == nil {
			return fmt.Errorf("target_date is required for countdown blocks")
		}
	}

	return nil
}

// Global instance
var advancedBlockService *AdvancedBlockService

// GetAdvancedBlockService returns the global advanced block service instance
func GetAdvancedBlockService() *AdvancedBlockService {
	if advancedBlockService == nil {
		advancedBlockService = NewAdvancedBlockService()
	}
	return advancedBlockService
}
