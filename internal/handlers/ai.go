package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"news/internal/database"
	"news/internal/json"
	"news/internal/middleware"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("news-api/handlers/ai")

// GenerateHeadlines godoc
// @Summary Generate AI Headlines
// @Description Generate multiple headline suggestions for given content using AI
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.GenerateHeadlinesRequest true "Headlines generation request"
// @Success 200 {object} models.GenerateHeadlinesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/headlines [post]
func GenerateHeadlines(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.generate_headlines")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.GenerateHeadlinesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate request
	if request.Count == 0 {
		request.Count = 5
	}
	if request.Count > 10 {
		request.Count = 10
	}
	if request.Style == "" {
		request.Style = "news"
	}

	span.SetAttributes(
		attribute.Int("request.count", request.Count),
		attribute.String("request.style", request.Style),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	headlines, err := aiService.GenerateHeadlines(ctx, request.Content, request.Count, request.Style)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate headlines"})
		return
	}

	// Save suggestion to database
	suggestionData, _ := json.Marshal(map[string]interface{}{
		"style": request.Style,
		"count": request.Count,
	})

	suggestion := models.ContentSuggestion{
		Type:       "headline",
		Input:      request.Content,
		Suggestion: string(must(json.Marshal(headlines))),
		Context:    string(suggestionData),
		UserID:     userID.(uint),
		Confidence: 0.85, // Default confidence for headline generation
	}

	database.DB.Create(&suggestion)

	// Track usage
	trackAIUsage(userID.(uint), "generate_headlines", len(request.Content))

	response := models.GenerateHeadlinesResponse{
		Headlines: headlines,
		Count:     len(headlines),
	}

	span.SetAttributes(attribute.Int("response.count", len(headlines)))
	c.JSON(http.StatusOK, response)
}

// GenerateContent godoc
// @Summary Generate AI Content
// @Description Generate article content based on topic and parameters using AI
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.GenerateContentRequest true "Content generation request"
// @Success 200 {object} models.GenerateContentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/content [post]
func GenerateContent(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.generate_content")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.GenerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Set defaults
	if request.Style == "" {
		request.Style = "news"
	}
	if request.Length == "" {
		request.Length = "medium"
	}
	if request.Perspective == "" {
		request.Perspective = "objective"
	}

	span.SetAttributes(
		attribute.String("request.topic", request.Topic),
		attribute.String("request.style", request.Style),
		attribute.String("request.length", request.Length),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	content, err := aiService.GenerateContent(ctx, request.Topic, request.Style, request.Length, request.Keywords, request.Perspective)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate content"})
		return
	}

	// Generate summary and extract keywords
	summary, err := aiService.SummarizeContent(ctx, content, "short", "paragraph", "en")
	if err != nil {
		summary = content[:min(len(content), 200)] + "..." // Fallback summary
	}

	// Estimate read time (average 200 words per minute)
	wordCount := len(content) / 5 // Rough estimation
	readTime := max(1, wordCount/200)

	// Save suggestion
	suggestionData, _ := json.Marshal(map[string]interface{}{
		"style":       request.Style,
		"length":      request.Length,
		"keywords":    request.Keywords,
		"perspective": request.Perspective,
	})

	suggestion := models.ContentSuggestion{
		Type:       "content",
		Input:      request.Topic,
		Suggestion: content,
		Context:    string(suggestionData),
		UserID:     userID.(uint),
		Confidence: 0.80,
	}

	database.DB.Create(&suggestion)

	// Track usage
	trackAIUsage(userID.(uint), "generate_content", len(content))

	response := models.GenerateContentResponse{
		Content:  content,
		Summary:  summary,
		Keywords: request.Keywords,
		ReadTime: readTime,
	}

	span.SetAttributes(
		attribute.Int("response.word_count", wordCount),
		attribute.Int("response.read_time", readTime),
	)

	c.JSON(http.StatusOK, response)
}

// ImproveContent godoc
// @Summary Improve Content with AI
// @Description Get AI suggestions to improve existing content
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.ImproveContentRequest true "Content improvement request"
// @Success 200 {object} models.ImproveContentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/improve [post]
func ImproveContent(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.improve_content")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.ImproveContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Set defaults
	if len(request.Goals) == 0 {
		request.Goals = []string{"clarity", "engagement"}
	}
	if request.TargetLevel == "" {
		request.TargetLevel = "intermediate"
	}

	span.SetAttributes(
		attribute.StringSlice("request.goals", request.Goals),
		attribute.String("request.target_level", request.TargetLevel),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	improved, suggestions, qualityScore, err := aiService.ImproveContent(ctx, request.Content, request.Goals, request.TargetLevel)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to improve content"})
		return
	}

	// Save suggestion
	suggestionData, _ := json.Marshal(map[string]interface{}{
		"goals":        request.Goals,
		"target_level": request.TargetLevel,
		"suggestions":  suggestions,
	})

	suggestion := models.ContentSuggestion{
		Type:       "improvement",
		Input:      request.Content,
		Suggestion: improved,
		Context:    string(suggestionData),
		UserID:     userID.(uint),
		Confidence: qualityScore,
	}

	database.DB.Create(&suggestion)

	// Track usage
	trackAIUsage(userID.(uint), "improve_content", len(request.Content))

	response := models.ImproveContentResponse{
		ImprovedContent: improved,
		Suggestions:     suggestions,
		QualityScore:    qualityScore,
	}

	span.SetAttributes(attribute.Float64("response.quality_score", qualityScore))
	c.JSON(http.StatusOK, response)
}

// ModerateContent godoc
// @Summary Moderate Content with AI
// @Description Use AI to moderate content for inappropriate material
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.ModerateContentRequest true "Content moderation request"
// @Success 200 {object} models.ModerateContentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/moderate [post]
func ModerateContent(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.moderate_content")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.ModerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	if request.ContentType == "" {
		request.ContentType = "general"
	}

	span.SetAttributes(
		attribute.String("request.content_type", request.ContentType),
		attribute.Bool("request.strict", request.Strict),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	isApproved, confidence, reason, categories, severity, err := aiService.ModerateComment(ctx, request.Content, request.Strict)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to moderate content"})
		return
	}

	// Save moderation result
	categoriesJSON, _ := json.Marshal(categories)
	moderation := models.ModerationResult{
		ContentType: request.ContentType,
		ContentID:   0, // Will be set when associated with specific content
		Content:     request.Content,
		IsApproved:  isApproved,
		Confidence:  confidence,
		Reason:      reason,
		Categories:  string(categoriesJSON),
		Severity:    severity,
	}

	database.DB.Create(&moderation)

	// Track usage
	trackAIUsage(userID.(uint), "moderate_content", len(request.Content))

	response := models.ModerateContentResponse{
		IsApproved: isApproved,
		Confidence: confidence,
		Reason:     reason,
		Categories: categories,
		Severity:   severity,
	}

	span.SetAttributes(
		attribute.Bool("response.is_approved", isApproved),
		attribute.Float64("response.confidence", confidence),
		attribute.String("response.severity", severity),
	)

	c.JSON(http.StatusOK, response)
}

// SummarizeContent godoc
// @Summary Summarize Content with AI
// @Description Generate a summary of given content using AI
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.SummarizeContentRequest true "Content summarization request"
// @Success 200 {object} models.SummarizeContentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/summarize [post]
func SummarizeContent(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.summarize_content")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.SummarizeContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Set defaults
	if request.Length == "" {
		request.Length = "medium"
	}
	if request.Style == "" {
		request.Style = "paragraph"
	}
	if request.Language == "" {
		request.Language = "en"
	}

	span.SetAttributes(
		attribute.String("request.length", request.Length),
		attribute.String("request.style", request.Style),
		attribute.String("request.language", request.Language),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	summary, err := aiService.SummarizeContent(ctx, request.Content, request.Length, request.Style, request.Language)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to summarize content"})
		return
	}

	// Extract key points (simplified)
	keyPoints := extractKeyPoints(summary)

	// Calculate word counts and reduction
	originalWords := len(request.Content) / 5 // Rough estimation
	summaryWords := len(summary) / 5
	reduction := float64(originalWords-summaryWords) / float64(originalWords) * 100

	// Save suggestion
	suggestionData, _ := json.Marshal(map[string]interface{}{
		"length":   request.Length,
		"style":    request.Style,
		"language": request.Language,
	})

	suggestion := models.ContentSuggestion{
		Type:       "summary",
		Input:      request.Content,
		Suggestion: summary,
		Context:    string(suggestionData),
		UserID:     userID.(uint),
		Confidence: 0.85,
	}

	database.DB.Create(&suggestion)

	// Track usage
	trackAIUsage(userID.(uint), "summarize_content", len(request.Content))

	response := models.SummarizeContentResponse{
		Summary:   summary,
		KeyPoints: keyPoints,
		WordCount: summaryWords,
		Reduction: reduction,
	}

	span.SetAttributes(
		attribute.Int("response.word_count", summaryWords),
		attribute.Float64("response.reduction", reduction),
	)

	c.JSON(http.StatusOK, response)
}

// CategorizeContent godoc
// @Summary Categorize Content with AI
// @Description Automatically categorize and tag content using AI
// @Tags AI
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CategorizeContentRequest true "Content categorization request"
// @Success 200 {object} models.CategorizeContentResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/categorize [post]
func CategorizeContent(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.categorize_content")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.CategorizeContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	span.SetAttributes(
		attribute.StringSlice("request.options", request.Options),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Get AI service
	aiService := services.GetAIService()
	categories, tags, err := aiService.CategorizeContent(ctx, request.Content, request.Options)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to categorize content"})
		return
	}

	// Generate tags
	generatedTags, err := aiService.GenerateTags(ctx, request.Content, 10)
	if err == nil {
		// Merge AI generated tags with categorization tags
		for _, tag := range generatedTags {
			tags = append(tags, models.TagSuggestion{
				Name:       tag,
				Confidence: 0.75,
				Relevance:  "medium",
			})
		}
	}

	// Save suggestion
	suggestionData, _ := json.Marshal(map[string]interface{}{
		"categories": categories,
		"tags":       tags,
		"options":    request.Options,
	})

	suggestion := models.ContentSuggestion{
		Type:       "categorization",
		Input:      request.Content,
		Suggestion: string(suggestionData),
		Context:    string(suggestionData),
		UserID:     userID.(uint),
		Confidence: 0.80,
	}

	database.DB.Create(&suggestion)

	// Track usage
	trackAIUsage(userID.(uint), "categorize_content", len(request.Content))

	response := models.CategorizeContentResponse{
		Categories: categories,
		Tags:       tags,
	}

	span.SetAttributes(
		attribute.Int("response.categories_count", len(categories)),
		attribute.Int("response.tags_count", len(tags)),
	)

	c.JSON(http.StatusOK, response)
}

// GetAISuggestions godoc
// @Summary Get AI Suggestions History
// @Description Retrieve user's AI content suggestions history
// @Tags AI
// @Produce json
// @Security Bearer
// @Param type query string false "Suggestion type filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/suggestions [get]
func GetAISuggestions(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ai.get_suggestions")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	suggestionType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Where("user_id = ?", userID)
	if suggestionType != "" {
		query = query.Where("type = ?", suggestionType)
	}

	var total int64
	query.Model(&models.ContentSuggestion{}).Count(&total)

	var suggestions []models.ContentSuggestion
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&suggestions).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch suggestions"})
		return
	}

	span.SetAttributes(
		attribute.Int("response.count", len(suggestions)),
		attribute.Int64("response.total", total),
	)

	response := models.PaginatedResponse{
		Data:       suggestions,
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: (int(total) + limit - 1) / limit,
		HasNext:    page < (int(total)+limit-1)/limit,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// GetAIUsageStats godoc
// @Summary Get AI Usage Statistics
// @Description Retrieve user's AI service usage statistics
// @Tags AI
// @Produce json
// @Security Bearer
// @Param days query int false "Number of days to include" default(30)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ai/usage-stats [get]
func GetAIUsageStats(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ai.get_usage_stats")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var stats []models.AIUsageStats
	if err := database.DB.Where("user_id = ? AND date >= ?", userID, startDate).
		Order("date DESC").Find(&stats).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch usage stats"})
		return
	}

	// Aggregate statistics
	summary := make(map[string]interface{})
	dailyStats := make([]map[string]interface{}, 0)
	serviceStats := make(map[string]map[string]int)

	totalRequests := 0
	totalTokens := 0

	for _, stat := range stats {
		totalRequests += stat.RequestCount
		totalTokens += stat.TokensUsed

		if serviceStats[stat.ServiceType] == nil {
			serviceStats[stat.ServiceType] = make(map[string]int)
		}
		serviceStats[stat.ServiceType]["requests"] += stat.RequestCount
		serviceStats[stat.ServiceType]["tokens"] += stat.TokensUsed

		dailyStats = append(dailyStats, map[string]interface{}{
			"date":          stat.Date.Format("2006-01-02"),
			"service_type":  stat.ServiceType,
			"request_count": stat.RequestCount,
			"tokens_used":   stat.TokensUsed,
		})
	}

	summary["total_requests"] = totalRequests
	summary["total_tokens"] = totalTokens
	summary["days_included"] = days
	summary["service_breakdown"] = serviceStats
	summary["daily_stats"] = dailyStats

	span.SetAttributes(
		attribute.Int("response.total_requests", totalRequests),
		attribute.Int("response.total_tokens", totalTokens),
	)

	c.JSON(http.StatusOK, summary)
}

// SemanticSearch godoc
// @Summary Semantic Search
// @Description Perform semantic search using AI-generated embeddings for better relevance. Rate limited to control costs.
// @Tags AI
// @Accept json
// @Produce json
// @Param query query string true "Search query"
// @Param lang query string false "Language filter"
// @Param region query string false "Region filter"
// @Param limit query int false "Number of results (max 50)" default(10)
// @Success 200 {object} models.SemanticSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 429 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/search [get]
func SemanticSearch(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.semantic_search")
	defer span.End()

	startTime := time.Now()

	// Parse query parameters
	query := c.Query("query")

	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Query parameter is required"})
		return
	}

	request := models.SemanticSearchRequest{
		Query:  query,
		Lang:   c.Query("lang"),
		Region: c.Query("region"),
		Limit:  10,
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 50 {
			request.Limit = limit
		}
	}

	span.SetAttributes(
		attribute.String("search.query", request.Query),
		attribute.String("search.lang", request.Lang),
		attribute.String("search.region", request.Region),
		attribute.Int("search.limit", request.Limit),
	)

	// Check rate limiting and AI availability from middleware (if exists)
	useAI, rateLimitExists := c.Get("use_ai_search")

	var response *models.SemanticSearchResponse

	if !rateLimitExists {
		// No rate limiting middleware - use traditional AI search (backward compatibility)
		span.AddEvent("no_rate_limiting_using_traditional_ai_search")
		response = performTraditionalAISearch(ctx, &request, startTime, span)
	} else if useAI.(bool) {
		// Rate limiting allows AI search
		span.AddEvent("rate_limit_allows_ai_search")
		response = performAISearchWithTracking(ctx, &request, startTime, span, c)
	} else {
		// Rate limited - use local search only
		span.AddEvent("rate_limited_using_local_search")
		response = performLocalSearchWithTracking(ctx, &request, startTime, span, c)
	}

	// Update metadata
	response.Meta.ProcessingTime = time.Since(startTime).String()

	span.SetAttributes(
		attribute.Int("results.count", len(response.Results)),
		attribute.Int("results.total", response.Total),
		attribute.String("search.method", response.Method),
	)

	c.JSON(http.StatusOK, response)
}

// performFallbackSearch performs traditional text-based search when vector search fails
func performFallbackSearch(ctx context.Context, request *models.SemanticSearchRequest, startTime time.Time) *models.SemanticSearchResponse {
	ctx, span := tracer.Start(ctx, "ai.fallback_search")
	defer span.End()

	esService := services.GetElasticSearchService()

	// Perform fallback search
	response, err := esService.FallbackSearch(ctx, request)
	if err != nil {
		span.RecordError(err)
		// Return empty results if fallback also fails
		return &models.SemanticSearchResponse{
			Query:   request.Query,
			Results: []models.SemanticSearchResult{},
			Total:   0,
			Method:  "fallback_failed",
			Meta: models.SemanticSearchMeta{
				ProcessingTime: time.Since(startTime).String(),
				QueryEmbedding: false,
				IndexUsed:      "none",
			},
		}
	}

	// Update metadata for fallback
	response.Meta.ProcessingTime = time.Since(startTime).String()
	response.Meta.QueryEmbedding = false
	response.Method = "fallback"

	span.SetAttributes(
		attribute.Int("fallback.results.count", len(response.Results)),
		attribute.Int("fallback.results.total", response.Total),
	)

	return response
}

// Helper functions

func trackAIUsage(userID uint, serviceType string, tokensUsed int) {
	today := time.Now().Truncate(24 * time.Hour)

	var usage models.AIUsageStats
	result := database.DB.Where("user_id = ? AND service_type = ? AND date = ?", userID, serviceType, today).First(&usage)

	if result.Error != nil {
		// Create new record
		usage = models.AIUsageStats{
			UserID:       userID,
			ServiceType:  serviceType,
			RequestCount: 1,
			TokensUsed:   tokensUsed,
			Date:         today,
		}
		database.DB.Create(&usage)
	} else {
		// Update existing record
		usage.RequestCount++
		usage.TokensUsed += tokensUsed
		database.DB.Save(&usage)
	}
}

func extractKeyPoints(text string) []string {
	// Simple implementation - split by sentences and take first few
	// In production, this could use more sophisticated NLP
	sentences := strings.Split(text, ".")
	keyPoints := make([]string, 0)

	for i, sentence := range sentences {
		if i >= 5 { // Limit to 5 key points
			break
		}
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 10 { // Filter out very short sentences
			keyPoints = append(keyPoints, sentence)
		}
	}

	return keyPoints
}

func must(data []byte, err error) []byte {
	if err != nil {
		return []byte{}
	}
	return data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper functions for rate-limited search

// performTraditionalAISearch performs traditional AI search without rate limiting
func performTraditionalAISearch(ctx context.Context, request *models.SemanticSearchRequest, startTime time.Time, span interface{}) *models.SemanticSearchResponse {
	// Get AI service for embedding generation
	aiService := services.GetAIService()

	// Generate embedding for the search query
	embedding, err := aiService.GenerateEmbedding(ctx, request.Query)
	if err != nil {
		// Fallback to non-semantic search if embedding fails
		return performFallbackSearch(ctx, request, startTime)
	}

	// Set the embedding in the request
	request.Embedding = embedding

	// Get ElasticSearch service
	esService := services.GetElasticSearchService()

	// Perform vector search
	response, err := esService.VectorSearch(ctx, request)
	if err != nil {
		// Fallback to text search if vector search fails
		return performFallbackSearch(ctx, request, startTime)
	}

	response.Meta.QueryEmbedding = true
	response.Method = "vector"

	return response
}

// performAISearchWithTracking performs AI search and records the request in rate limiter
func performAISearchWithTracking(ctx context.Context, request *models.SemanticSearchRequest, startTime time.Time, span interface{}, c *gin.Context) *models.SemanticSearchResponse {
	response := performTraditionalAISearch(ctx, request, startTime, span)

	// Record AI request in rate limiter
	if limiterInterface, exists := c.Get("search_limiter"); exists {
		userIDInterface, _ := c.Get("search_user_id")
		ip, _ := c.Get("search_ip")

		// Type assertion to middleware.SemanticSearchLimiter
		if limiter, ok := limiterInterface.(*middleware.SemanticSearchLimiter); ok {
			userID, _ := userIDInterface.(*uint)
			limiter.RecordAIRequest(userID, ip.(string))
		}
	}

	return response
}

// performLocalSearchWithTracking performs local search and records the request
func performLocalSearchWithTracking(ctx context.Context, request *models.SemanticSearchRequest, startTime time.Time, span interface{}, c *gin.Context) *models.SemanticSearchResponse {
	response := performFallbackSearch(ctx, request, startTime)

	// Add rate limit reason to response
	if rateLimitReason, exists := c.Get("rate_limit_reason"); exists {
		if reason, ok := rateLimitReason.(string); ok && reason != "" {
			response.Meta.RateLimitReason = reason
		}
	}

	// Record local request in rate limiter
	if limiterInterface, exists := c.Get("search_limiter"); exists {
		userIDInterface, _ := c.Get("search_user_id")
		ip, _ := c.Get("search_ip")

		// Type assertion to middleware.SemanticSearchLimiter
		if limiter, ok := limiterInterface.(*middleware.SemanticSearchLimiter); ok {
			userID, _ := userIDInterface.(*uint)
			limiter.RecordLocalRequest(userID, ip.(string))
		}
	}

	return response
}

// ...existing helper functions...
