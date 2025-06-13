package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"
	"news/internal/pubsub"
	"news/internal/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// BreakingNewsHandler handles breaking news banner operations
type BreakingNewsHandler struct {
	DB *gorm.DB
}

// NewBreakingNewsHandler creates a new breaking news handler
func NewBreakingNewsHandler() *BreakingNewsHandler {
	return &BreakingNewsHandler{
		DB: database.DB,
	}
}

// @Summary Get active breaking news banners
// @Description Retrieves all currently active breaking news banners
// @Tags Breaking News
// @Produce json
// @Success 200 {array} models.BreakingNewsBanner
// @Failure 500 {object} models.ErrorResponse
// @Router /api/breaking-news [get]
func (h *BreakingNewsHandler) GetActiveBreakingNews(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "BreakingNewsHandler.GetActiveBreakingNews")
	defer span.End()

	var banners []models.BreakingNewsBanner
	now := time.Now()

	result := h.DB.Where("is_active = ? AND start_time <= ? AND (end_time IS NULL OR end_time >= ?)",
		true, now, now).
		Order("priority DESC").
		Preload("Article", func(db *gorm.DB) *gorm.DB {
			return db.Select("ID, Title, Slug")
		}).
		Find(&banners)

	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch breaking news banners",
		})
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(banners)))
	c.JSON(http.StatusOK, banners)
}

// @Summary Create breaking news banner
// @Description Creates a new breaking news banner
// @Tags Breaking News
// @Accept json
// @Produce json
// @Param breakingNews body models.BreakingNewsBanner true "Breaking news banner info"
// @Success 201 {object} models.BreakingNewsBanner
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/breaking-news [post]
func (h *BreakingNewsHandler) CreateBreakingNews(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "BreakingNewsHandler.CreateBreakingNews")
	defer span.End()

	var banner models.BreakingNewsBanner
	if err := c.ShouldBindJSON(&banner); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid breaking news data: " + err.Error(),
		})
		return
	}

	// Set the create user ID from the authenticated user - commented out since field doesn't exist in DB
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	span.RecordError(fmt.Errorf("user ID not found in context"))
	// 	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
	// 		Error: "User information not available",
	// 	})
	// 	return
	// }
	// banner.CreateUserID = userID.(uint)

	// If startTime is not set, default to now
	if banner.StartTime.IsZero() {
		banner.StartTime = time.Now()
	}

	result := h.DB.Create(&banner)
	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create breaking news banner: " + result.Error.Error(),
		})
		return
	}

	// Send breaking news notification for the banner
	if banner.IsActive {
		// Create a notification message for the breaking news banner
		notification := pubsub.NotificationMessage{
			Type: "breaking_news_banner",
			Data: map[string]interface{}{
				"id":               banner.ID,
				"title":            banner.Title,
				"content":          banner.Content,
				"priority":         banner.Priority,
				"style":            banner.Style,
				"background_color": banner.BackgroundColor,
				"text_color":       banner.TextColor,
				"start_time":       banner.StartTime,
				"end_time":         banner.EndTime,
			},
		}

		if err := pubsub.PublishNotification(pubsub.ChannelBreakingNews, notification); err != nil {
			// Log error but don't fail the request
			span.RecordError(err)
			// Use Printf instead of log to match the existing pattern in the codebase
			// log.Printf("Failed to send breaking news banner notification: %v", err)
		}
	}

	span.SetAttributes(attribute.Int("banner.id", int(banner.ID)))
	c.JSON(http.StatusCreated, banner)
}

// @Summary Update breaking news banner
// @Description Updates an existing breaking news banner
// @Tags Breaking News
// @Accept json
// @Produce json
// @Param id path int true "Breaking news banner ID"
// @Param breakingNews body models.BreakingNewsBanner true "Breaking news banner updated info"
// @Success 200 {object} models.BreakingNewsBanner
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/breaking-news/{id} [put]
func (h *BreakingNewsHandler) UpdateBreakingNews(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "BreakingNewsHandler.UpdateBreakingNews")
	defer span.End()

	// Get the banner ID from the URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid banner ID",
		})
		return
	}

	span.SetAttributes(attribute.Int("banner.id", int(id)))

	// Find the existing banner
	var existingBanner models.BreakingNewsBanner
	if result := h.DB.First(&existingBanner, id); result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Breaking news banner not found",
		})
		return
	}

	// Bind the updated data
	var updatedBanner models.BreakingNewsBanner
	if err := c.ShouldBindJSON(&updatedBanner); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid breaking news data: " + err.Error(),
		})
		return
	}

	// Update only allowed fields
	existingBanner.Title = updatedBanner.Title
	existingBanner.Content = updatedBanner.Content
	existingBanner.ArticleID = updatedBanner.ArticleID
	existingBanner.Priority = updatedBanner.Priority
	existingBanner.Style = updatedBanner.Style
	existingBanner.TextColor = updatedBanner.TextColor
	existingBanner.BackgroundColor = updatedBanner.BackgroundColor
	existingBanner.StartTime = updatedBanner.StartTime
	existingBanner.EndTime = updatedBanner.EndTime
	existingBanner.IsActive = updatedBanner.IsActive

	// Save the updated banner
	if result := h.DB.Save(&existingBanner); result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update breaking news banner: " + result.Error.Error(),
		})
		return
	}

	// Send breaking news notification for the updated banner if it's active
	if existingBanner.IsActive {
		notification := pubsub.NotificationMessage{
			Type: "breaking_news_banner_updated",
			Data: map[string]interface{}{
				"id":               existingBanner.ID,
				"title":            existingBanner.Title,
				"content":          existingBanner.Content,
				"priority":         existingBanner.Priority,
				"style":            existingBanner.Style,
				"background_color": existingBanner.BackgroundColor,
				"text_color":       existingBanner.TextColor,
				"start_time":       existingBanner.StartTime,
				"end_time":         existingBanner.EndTime,
			},
		}

		if err := pubsub.PublishNotification(pubsub.ChannelBreakingNews, notification); err != nil {
			// Log error but don't fail the request
			span.RecordError(err)
		}
	}

	c.JSON(http.StatusOK, existingBanner)
}

// @Summary Delete breaking news banner
// @Description Deletes a breaking news banner
// @Tags Breaking News
// @Produce json
// @Param id path int true "Breaking news banner ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/breaking-news/{id} [delete]
func (h *BreakingNewsHandler) DeleteBreakingNews(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "BreakingNewsHandler.DeleteBreakingNews")
	defer span.End()

	// Get the banner ID from the URL
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid banner ID",
		})
		return
	}

	span.SetAttributes(attribute.Int("banner.id", int(id)))

	// Delete the banner
	result := h.DB.Delete(&models.BreakingNewsBanner{}, id)
	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete breaking news banner: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Breaking news banner not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Breaking news banner deleted successfully",
	})
}
