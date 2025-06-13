package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"news/internal/middleware"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
)

// Updated SemanticSearch with rate limiting and fallback
func SemanticSearchWithRateLimit(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "ai.semantic_search_rate_limited")
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

	// Check rate limiting and AI availability from middleware
	useAI, exists := c.Get("use_ai_search")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Rate limiting middleware not configured"})
		return
	}

	rateLimitReason, _ := c.Get("rate_limit_reason")
	limiterInterface, _ := c.Get("search_limiter")
	userIDInterface, _ := c.Get("search_user_id")
	ip, _ := c.Get("search_ip")

	var response *models.SemanticSearchResponse

	// Determine search method based on rate limiting
	if useAI.(bool) {
		// Use AI-powered semantic search
		span.AddEvent("using_ai_semantic_search")
		response = performAISearch(ctx, &request, startTime, span)

		// Record successful AI request
		if limiter, ok := limiterInterface.(*middleware.SemanticSearchLimiter); ok {
			userID, _ := userIDInterface.(*uint)
			limiter.RecordAIRequest(userID, ip.(string))
		}

		span.SetAttributes(
			attribute.Bool("search.ai_used", true),
			attribute.String("search.method", response.Method),
		)
	} else {
		// Use local search only (rate limited)
		span.AddEvent("using_local_search_due_to_rate_limit")
		response = performFallbackSearch(ctx, &request, startTime)

		// Add rate limit reason to response
		if reason, ok := rateLimitReason.(string); ok && reason != "" {
			response.Meta.RateLimitReason = reason
		}

		// Record local request
		if limiter, ok := limiterInterface.(*middleware.SemanticSearchLimiter); ok {
			userID, _ := userIDInterface.(*uint)
			limiter.RecordLocalRequest(userID, ip.(string))
		}

		span.SetAttributes(
			attribute.Bool("search.ai_used", false),
			attribute.String("search.rate_limit_reason", rateLimitReason.(string)),
		)
	}

	// Update metadata
	response.Meta.ProcessingTime = time.Since(startTime).String()

	span.SetAttributes(
		attribute.Int("results.count", len(response.Results)),
		attribute.Int("results.total", response.Total),
		attribute.String("search.final_method", response.Method),
	)

	c.JSON(http.StatusOK, response)
}

// performAISearch performs AI-powered semantic search
func performAISearch(ctx context.Context, request *models.SemanticSearchRequest, startTime time.Time, span interface{}) *models.SemanticSearchResponse {
	// Get AI service for embedding generation
	aiService := services.GetAIService()

	// Generate embedding for the search query
	embedding, err := aiService.GenerateEmbedding(ctx, request.Query)
	if err != nil {
		// Fallback to local search if embedding fails
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

	response.Method = "vector"
	response.Meta.QueryEmbedding = true

	return response
}

// GetSearchLimitStatus godoc
// @Summary Get Search Rate Limit Status
// @Description Get current search rate limit status for the user/IP
// @Tags AI
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/search/limits [get]
func GetSearchLimitStatus(c *gin.Context) {
	// Extract user info
	userIDInterface, userExists := c.Get("userID")
	var userID *uint
	if userExists {
		if uid, ok := userIDInterface.(uint); ok {
			userID = &uid
		}
	}

	// Get IP
	ip := c.ClientIP()

	// Create limiter to check status
	limiter := middleware.NewSemanticSearchLimiter(middleware.DefaultSearchLimitConfig(), nil)
	status := limiter.GetLimitStatus(userID, ip)

	c.JSON(http.StatusOK, status)
}
