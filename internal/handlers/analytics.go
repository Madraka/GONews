package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// RecordInteractionRequest represents the request payload for recording interactions
type RecordInteractionRequest struct {
	InteractionType string   `json:"interaction_type" binding:"required"` // view, bookmark, upvote, downvote, share, comment
	Duration        *int     `json:"duration,omitempty"`                  // For view interactions
	CompletionRate  *float64 `json:"completion_rate,omitempty"`           // For view interactions (0.0-1.0)
	Platform        string   `json:"platform,omitempty"`                  // web, mobile, app
	ReferrerURL     string   `json:"referrer_url,omitempty"`              // How user found the article
}

// RecordArticleInteraction godoc
// @Summary Record user interaction with article
// @Description Record a user's interaction with an article (view, bookmark, vote, etc.)
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Param interaction body RecordInteractionRequest true "Interaction data"
// @Success 201 {object} models.UserArticleInteraction
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/articles/{id}/interactions [post]
func RecordArticleInteraction(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var req RecordInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// Verify article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Create interaction record
	interaction := models.UserArticleInteraction{
		UserID:          userID.(uint),
		ArticleID:       uint(articleID),
		InteractionType: req.InteractionType,
		Duration:        req.Duration,
		CompletionRate:  req.CompletionRate,
		Platform:        req.Platform,
		UserAgent:       c.GetHeader("User-Agent"),
		IPAddress:       c.ClientIP(),
		ReferrerURL:     req.ReferrerURL,
	}

	// Validate interaction type
	if !interaction.ValidateInteractionType() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid interaction type"})
		return
	}

	// For certain interaction types, check for duplicates or update existing
	if req.InteractionType == "view" {
		// For views, we might want to update existing view with latest duration/completion
		var existingInteraction models.UserArticleInteraction
		err := database.DB.Where("user_id = ? AND article_id = ? AND interaction_type = ?",
			userID, articleID, "view").First(&existingInteraction).Error

		if err == nil {
			// Update existing view
			if req.Duration != nil {
				existingInteraction.Duration = req.Duration
			}
			if req.CompletionRate != nil {
				existingInteraction.CompletionRate = req.CompletionRate
			}
			existingInteraction.UpdatedAt = time.Now()
			database.DB.Save(&existingInteraction)
			c.JSON(http.StatusOK, existingInteraction)
			return
		}
	}

	// Create new interaction
	if err := database.DB.Create(&interaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to record interaction"})
		return
	}

	c.JSON(http.StatusCreated, interaction)
}

// GetUserInteractions godoc
// @Summary Get user's article interactions
// @Description Get paginated list of user's interactions with articles
// @Tags Analytics
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Filter by interaction type"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/user/interactions [get]
func GetUserInteractions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	interactionType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := database.DB.Where("user_id = ?", userID)
	if interactionType != "" {
		query = query.Where("interaction_type = ?", interactionType)
	}

	// Get total count
	var total int64
	query.Model(&models.UserArticleInteraction{}).Count(&total)

	// Get interactions with article details
	var interactions []models.UserArticleInteraction
	if err := query.Preload("Article").Preload("Article.Author").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&interactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch interactions"})
		return
	}

	response := map[string]interface{}{
		"page":         page,
		"limit":        limit,
		"total":        total,
		"interactions": interactions,
	}

	c.JSON(http.StatusOK, response)
}

// GetArticleAnalytics godoc
// @Summary Get article analytics
// @Description Get analytics data for a specific article
// @Tags Analytics
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/articles/{id}/analytics [get]
func GetArticleAnalytics(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	// Verify article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Get interaction statistics
	var stats struct {
		TotalViews     int64   `json:"total_views"`
		UniqueViews    int64   `json:"unique_views"`
		TotalBookmarks int64   `json:"total_bookmarks"`
		TotalUpvotes   int64   `json:"total_upvotes"`
		TotalDownvotes int64   `json:"total_downvotes"`
		TotalShares    int64   `json:"total_shares"`
		TotalComments  int64   `json:"total_comments"`
		AvgReadTime    float64 `json:"avg_read_time"`
		AvgCompletion  float64 `json:"avg_completion_rate"`
	}

	// Count different interaction types
	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ?", articleID, "view").
		Count(&stats.TotalViews)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ?", articleID, "view").
		Distinct("user_id").
		Count(&stats.UniqueViews)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ?", articleID, "bookmark").
		Count(&stats.TotalBookmarks)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type IN ?", articleID, []string{"upvote", "like"}).
		Count(&stats.TotalUpvotes)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type IN ?", articleID, []string{"downvote", "dislike"}).
		Count(&stats.TotalDownvotes)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ?", articleID, "share").
		Count(&stats.TotalShares)

	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ?", articleID, "comment").
		Count(&stats.TotalComments)

	// Calculate averages
	var avgDuration float64
	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ? AND duration IS NOT NULL", articleID, "view").
		Select("AVG(duration)").
		Scan(&avgDuration)
	stats.AvgReadTime = avgDuration

	var avgCompletion float64
	database.DB.Model(&models.UserArticleInteraction{}).
		Where("article_id = ? AND interaction_type = ? AND completion_rate IS NOT NULL", articleID, "view").
		Select("AVG(completion_rate)").
		Scan(&avgCompletion)
	stats.AvgCompletion = avgCompletion

	response := map[string]interface{}{
		"article_id":   articleID,
		"title":        article.Title,
		"author":       article.Author,
		"stats":        stats,
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
