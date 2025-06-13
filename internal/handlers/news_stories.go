package handlers

import (
	"net/http"
	"time"

	"news/internal/database"
	"news/internal/models"
	"news/internal/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// NewsStoriesHandler handles Instagram/Facebook-style news stories operations
type NewsStoriesHandler struct {
	DB *gorm.DB
}

// NewNewsStoriesHandler creates a new news stories handler
func NewNewsStoriesHandler() *NewsStoriesHandler {
	return &NewsStoriesHandler{
		DB: database.DB,
	}
}

// @Summary Get active news stories
// @Description Retrieves all currently active news stories for display
// @Tags News Stories
// @Produce json
// @Success 200 {array} models.NewsStory
// @Failure 500 {object} models.ErrorResponse
// @Router /api/news-stories [get]
func (h *NewsStoriesHandler) GetActiveStories(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.GetActiveStories")
	defer span.End()

	var stories []models.NewsStory
	now := time.Now()

	// Get active stories that haven't expired
	// A story is active if:
	// 1. Current time is after StartTime
	// 2. Current time is before EndTime (EndTime = StartTime + Duration)
	result := h.DB.Where("start_time <= ? AND start_time + (duration || ' seconds')::interval >= ?",
		now, now).
		Order("sort_order ASC, start_time DESC").
		Preload("Article", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Title, Slug")
		}).
		Find(&stories)

	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch news stories",
		})
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(stories)))
	c.JSON(http.StatusOK, stories)
}

// @Summary Get story by ID
// @Description Retrieves a specific news story by ID
// @Tags News Stories
// @Produce json
// @Param id path int true "Story ID"
// @Success 200 {object} models.NewsStory
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/news-stories/{id} [get]
func (h *NewsStoriesHandler) GetStoryByID(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.GetStoryByID")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("story.id", id))

	var story models.NewsStory
	if err := h.DB.First(&story, id).Error; err != nil {
		span.RecordError(err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "News story not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch news story",
			})
		}
		return
	}

	// Increment view count
	h.DB.Model(&story).Update("view_count", gorm.Expr("view_count + ?", 1))

	// Record user view if authenticated
	if userID, exists := c.Get("user_id"); exists {
		storyView := models.StoryView{
			StoryID:  story.ID,
			UserID:   userID.(uint),
			ViewedAt: time.Now(),
		}
		h.DB.Create(&storyView)
	}

	c.JSON(http.StatusOK, story)
}

// @Summary Create news story
// @Description Creates a new news story
// @Tags News Stories
// @Accept json
// @Produce json
// @Param story body models.NewsStory true "News story info"
// @Success 201 {object} models.NewsStory
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/news-stories [post]
func (h *NewsStoriesHandler) CreateStory(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.CreateStory")
	defer span.End()

	var story models.NewsStory
	if err := c.ShouldBindJSON(&story); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if story.Headline == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Headline is required",
		})
		return
	}

	if story.ImageURL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Image URL is required",
		})
		return
	}

	// Set default values if not provided
	if story.Duration == 0 {
		story.Duration = 5 // 5 seconds default duration
	}

	if story.BackgroundColor == "" {
		story.BackgroundColor = "#000000"
	}

	if story.TextColor == "" {
		story.TextColor = "#FFFFFF"
	}

	// Get the creator user ID from the context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "User ID not found in request context",
		})
		return
	}
	story.CreateUserID = userID.(uint)

	// If start time not set, use current time
	if story.StartTime.IsZero() {
		story.StartTime = time.Now()
	}

	// Create the story
	if err := h.DB.Create(&story).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create news story: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, story)
}

// @Summary Update news story
// @Description Updates an existing news story
// @Tags News Stories
// @Accept json
// @Produce json
// @Param id path int true "Story ID"
// @Param story body models.NewsStory true "Updated story info"
// @Success 200 {object} models.NewsStory
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/news-stories/{id} [put]
func (h *NewsStoriesHandler) UpdateStory(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.UpdateStory")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("story.id", id))

	var story models.NewsStory
	if err := h.DB.First(&story, id).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "News story not found",
		})
		return
	}

	var updatedStory models.NewsStory
	if err := c.ShouldBindJSON(&updatedStory); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Update fields if provided
	if updatedStory.Headline != "" {
		story.Headline = updatedStory.Headline
	}

	if updatedStory.ImageURL != "" {
		story.ImageURL = updatedStory.ImageURL
	}

	if updatedStory.BackgroundColor != "" {
		story.BackgroundColor = updatedStory.BackgroundColor
	}

	if updatedStory.TextColor != "" {
		story.TextColor = updatedStory.TextColor
	}

	if updatedStory.Duration > 0 {
		story.Duration = updatedStory.Duration
	}

	if !updatedStory.StartTime.IsZero() {
		story.StartTime = updatedStory.StartTime
	}

	if updatedStory.ExternalURL != "" {
		story.ExternalURL = updatedStory.ExternalURL
	}

	if updatedStory.SortOrder > 0 {
		story.SortOrder = updatedStory.SortOrder
	}

	// Update related article if provided
	if updatedStory.ArticleID != nil {
		story.ArticleID = updatedStory.ArticleID
	}

	// Save the updated story
	if err := h.DB.Save(&story).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update news story: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, story)
}

// @Summary Delete news story
// @Description Deletes a news story
// @Tags News Stories
// @Produce json
// @Param id path int true "Story ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/news-stories/{id} [delete]
func (h *NewsStoriesHandler) DeleteStory(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.DeleteStory")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("story.id", id))

	var story models.NewsStory
	if err := h.DB.First(&story, id).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "News story not found",
		})
		return
	}

	if err := h.DB.Delete(&story).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete news story: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "News story deleted successfully",
	})
}

// @Summary Get user stories that haven't been viewed
// @Description Retrieves news stories that the authenticated user hasn't seen yet
// @Tags News Stories
// @Produce json
// @Success 200 {array} models.NewsStory
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/news-stories/unviewed [get]
func (h *NewsStoriesHandler) GetUnviewedStories(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "NewsStoriesHandler.GetUnviewedStories")
	defer span.End()

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return
	}

	var stories []models.NewsStory
	now := time.Now()

	// Get currently active stories that the user hasn't viewed
	result := h.DB.Where("start_time <= ? AND start_time + (duration || ' seconds')::interval >= ?",
		now, now).
		// Using a subquery to get unviewed stories
		Where("NOT EXISTS (SELECT 1 FROM story_views WHERE story_views.story_id = news_stories.id AND story_views.user_id = ?)",
			userID).
		Order("sort_order ASC, start_time DESC").
		Find(&stories)

	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch unviewed stories",
		})
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(stories)))
	c.JSON(http.StatusOK, stories)
}
