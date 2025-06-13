package handlers

import (
	"net/http"
	"news/internal/database"
	"news/internal/models"
	"news/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRecommendedArticles godoc
// @Summary Get recommended articles
// @Description Retrieve a list of recommended articles for the authenticated user or based on general popularity if not authenticated.
// @Tags Articles
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of recommendations to return" default(10)
// @Success 200 {array} models.Article
// @Failure 401 {object} models.ErrorResponse "If authentication is required and fails"
// @Failure 500 {object} models.ErrorResponse "If an internal error occurs"
// @Router /api/articles/recommendations [get]
func GetRecommendedArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 { // Cap the limit to prevent excessive load
		limit = 50
	}

	recommendationService := services.NewRecommendationService(database.DB)

	// Check if user is authenticated
	userID, exists := c.Get("userID")
	var articles []models.Article

	if exists && userID != nil {
		// Get personalized recommendations for authenticated user
		if uid, ok := userID.(uint); ok {
			articles, err = recommendationService.GetPersonalizedRecommendations(uid, limit)
		} else {
			// Fallback to popular recommendations if userID type assertion fails
			articles, err = recommendationService.GetPopularRecommendations(limit)
		}
	} else {
		// Get popular recommendations for unauthenticated users
		articles, err = recommendationService.GetPopularRecommendations(limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch recommended articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

// GetSimilarArticles godoc
// @Summary Get similar articles
// @Description Retrieve a list of articles similar to a given article.
// @Tags Articles
// @Produce json
// @Param article_id path int true "ID of the article to find similar articles for"
// @Param limit query int false "Number of similar articles to return" default(5)
// @Success 200 {array} models.Article
// @Failure 400 {object} models.ErrorResponse "If article_id is invalid"
// @Failure 404 {object} models.ErrorResponse "If the source article is not found"
// @Failure 500 {object} models.ErrorResponse "If an internal error occurs"
// @Router /api/articles/{article_id}/similar [get]
func GetSimilarArticles(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID format"})
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	if limit > 20 { // Cap the limit
		limit = 20
	}

	recommendationService := services.NewRecommendationService(database.DB)
	articles, err := recommendationService.GetSimilarArticles(uint(articleID), limit)

	if err != nil {
		if err.Error() == "source article not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Source article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch similar articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

// GetTrendingArticles godoc
// @Summary Get trending articles
// @Description Retrieve a list of trending articles based on recent interactions
// @Tags Articles
// @Produce json
// @Param limit query int false "Number of trending articles to return" default(10)
// @Success 200 {array} models.Article
// @Failure 500 {object} models.ErrorResponse "If an internal error occurs"
// @Router /api/articles/trending [get]
func GetTrendingArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	recommendationService := services.NewRecommendationService(database.DB)
	articles, err := recommendationService.GetPopularRecommendations(limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch trending articles"})
		return
	}

	c.JSON(http.StatusOK, articles)
}

// GetReadingHistory godoc
// @Summary Get user's reading history
// @Description Retrieve articles that the authenticated user has previously viewed or bookmarked
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of articles to return" default(20)
// @Success 200 {array} models.Article
// @Failure 401 {object} models.ErrorResponse "If authentication fails"
// @Failure 500 {object} models.ErrorResponse "If an internal error occurs"
// @Router /api/user/reading-history [get]
func GetReadingHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	recommendationService := services.NewRecommendationService(database.DB)
	articles, err := recommendationService.GetUserReadingHistory(uid, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch reading history"})
		return
	}

	c.JSON(http.StatusOK, articles)
}
