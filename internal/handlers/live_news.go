package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"
	"news/internal/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

// LiveNewsHandler handles live news stream operations
type LiveNewsHandler struct {
	DB *gorm.DB
}

// NewLiveNewsHandler creates a new live news handler
func NewLiveNewsHandler() *LiveNewsHandler {
	return &LiveNewsHandler{
		DB: database.DB,
	}
}

// @Summary Get active live news streams
// @Description Retrieves all currently active live news streams
// @Tags Live News
// @Produce json
// @Success 200 {array} models.LiveNewsStream
// @Failure 500 {object} models.ErrorResponse
// @Router /api/live-news [get]
func (h *LiveNewsHandler) GetActiveLiveStreams(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.GetActiveLiveStreams")
	defer span.End()

	var streams []models.LiveNewsStream

	// Get active streams (status=live)
	result := h.DB.Where("status = ?", "live").
		Order("is_highlighted DESC, start_time DESC").
		Preload("Updates", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(5) // Get the 5 most recent updates
		}).
		Find(&streams)

	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch live news streams",
		})
		return
	}

	span.SetAttributes(attribute.Int("result.count", len(streams)))
	c.JSON(http.StatusOK, streams)
}

// @Summary Get live stream by ID
// @Description Retrieves a specific live news stream by ID
// @Tags Live News
// @Produce json
// @Param id path int true "Stream ID"
// @Success 200 {object} models.LiveNewsStream
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/live-news/{id} [get]
func (h *LiveNewsHandler) GetLiveStreamByID(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.GetLiveStreamByID")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("stream.id", id))

	var stream models.LiveNewsStream
	if err := h.DB.Preload("Updates", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	}).First(&stream, id).Error; err != nil {
		span.RecordError(err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Live news stream not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: "Failed to fetch live news stream",
			})
		}
		return
	}

	// Increment viewer count
	h.DB.Model(&stream).Update("viewer_count", gorm.Expr("viewer_count + ?", 1))

	c.JSON(http.StatusOK, stream)
}

// @Summary Create live news stream
// @Description Creates a new live news stream
// @Tags Live News
// @Accept json
// @Produce json
// @Param stream body models.LiveNewsStream true "Live news stream info"
// @Success 201 {object} models.LiveNewsStream
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/live-news [post]
func (h *LiveNewsHandler) CreateLiveStream(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.CreateLiveStream")
	defer span.End()

	var stream models.LiveNewsStream
	if err := c.ShouldBindJSON(&stream); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if stream.Title == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Title is required",
		})
		return
	}

	// User ID check no longer needed as CreateUserID field was removed from the model
	// userID, exists := c.Get("userID")
	// if !exists {
	// 	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
	// 		Error: "User ID not found in request context",
	// 	})
	// 	return
	// }
	// stream.CreateUserID = userID.(uint)

	// If status not provided, set default to draft
	if stream.Status == "" {
		stream.Status = "draft"
	}

	// Create the stream
	if err := h.DB.Create(&stream).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create live news stream: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, stream)
}

// @Summary Update live news stream
// @Description Updates an existing live news stream
// @Tags Live News
// @Accept json
// @Produce json
// @Param id path int true "Stream ID"
// @Param stream body models.LiveNewsStream true "Updated stream info"
// @Success 200 {object} models.LiveNewsStream
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/live-news/{id} [put]
func (h *LiveNewsHandler) UpdateLiveStream(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.UpdateLiveStream")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("stream.id", id))

	var stream models.LiveNewsStream
	if err := h.DB.First(&stream, id).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Live news stream not found",
		})
		return
	}

	var updatedStream models.LiveNewsStream
	if err := c.ShouldBindJSON(&updatedStream); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Update fields if provided
	if updatedStream.Title != "" {
		stream.Title = updatedStream.Title
	}

	if updatedStream.Description != "" {
		stream.Description = updatedStream.Description
	}

	// Cover image URL is not stored in DB schema, skip this update
	// if updatedStream.CoverImageURL != "" {
	//	stream.CoverImageURL = updatedStream.CoverImageURL
	// }

	// Update status if provided
	if updatedStream.Status != "" {
		// Validate status transitions
		if updatedStream.Status == "live" && stream.Status == "draft" {
			// If going from draft to live, set start time to now if not already set
			if stream.StartTime == nil {
				now := time.Now()
				stream.StartTime = &now
			}
		}

		if updatedStream.Status == "ended" && stream.Status == "live" {
			// If ending a live stream, set end time to now if not already set
			if stream.EndTime == nil {
				now := time.Now()
				stream.EndTime = &now
			}
		}

		stream.Status = updatedStream.Status
	}

	// Update timestamps if provided
	if updatedStream.StartTime != nil {
		stream.StartTime = updatedStream.StartTime
	}

	if updatedStream.EndTime != nil {
		stream.EndTime = updatedStream.EndTime
	}

	// Update related category if provided (commented out as CategoryID field doesn't exist in DB)
	// if updatedStream.CategoryID != nil {
	// 	stream.CategoryID = updatedStream.CategoryID
	// }

	// Update highlight status if provided
	stream.IsHighlighted = updatedStream.IsHighlighted

	// Save the updated stream
	if err := h.DB.Save(&stream).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to update live news stream: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stream)
}

// @Summary Delete live news stream
// @Description Deletes a live news stream
// @Tags Live News
// @Produce json
// @Param id path int true "Stream ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/live-news/{id} [delete]
func (h *LiveNewsHandler) DeleteLiveStream(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.DeleteLiveStream")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("stream.id", id))

	var stream models.LiveNewsStream
	if err := h.DB.First(&stream, id).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Live news stream not found",
		})
		return
	}

	if err := h.DB.Delete(&stream).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete live news stream: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Live news stream deleted successfully",
	})
}

// @Summary Add update to live news stream
// @Description Adds a new update to an existing live news stream
// @Tags Live News
// @Accept json
// @Produce json
// @Param id path int true "Stream ID"
// @Param update body models.LiveNewsUpdate true "Live news update info"
// @Success 201 {object} models.LiveNewsUpdate
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /admin/live-news/{id}/updates [post]
func (h *LiveNewsHandler) AddLiveUpdate(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.AddLiveUpdate")
	defer span.End()

	streamID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid stream ID",
		})
		return
	}

	// Verify the stream exists
	var stream models.LiveNewsStream
	if err := h.DB.First(&stream, streamID).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Live news stream not found",
		})
		return
	}

	// Check if stream is in live status
	if stream.Status != "live" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Updates can only be added to streams with 'live' status",
		})
		return
	}

	var update models.LiveNewsUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request payload: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if update.Content == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Content is required",
		})
		return
	}

	// If title is not provided, use a default based on the content
	if update.Title == "" {
		if len(update.Content) > 50 {
			update.Title = update.Content[:47] + "..."
		} else {
			update.Title = update.Content
		}
	}

	update.StreamID = uint(streamID)

	// If update type not provided, set default
	if update.UpdateType == "" {
		update.UpdateType = "update"
	}

	// If importance not provided, set default
	if update.Importance == "" {
		update.Importance = "normal"
	}

	// Create the update
	if err := h.DB.Create(&update).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create live update: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, update)
}

// @Summary Get updates for a live news stream
// @Description Gets all updates for a specific live news stream with pagination
// @Tags Live News
// @Produce json
// @Param id path int true "Stream ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/live-news/{id}/updates [get]
func (h *LiveNewsHandler) GetLiveUpdates(c *gin.Context) {
	_, span := tracing.StartSpanWithAttributes(c.Request.Context(), "LiveNewsHandler.GetLiveUpdates")
	defer span.End()

	streamID := c.Param("id")
	span.SetAttributes(attribute.String("stream.id", streamID))

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Verify the stream exists
	var stream models.LiveNewsStream
	if err := h.DB.First(&stream, streamID).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Live news stream not found",
		})
		return
	}

	var updates []models.LiveNewsUpdate
	var total int64

	// Get total count
	h.DB.Model(&models.LiveNewsUpdate{}).
		Where("stream_id = ?", streamID).
		Count(&total)

	// Get paginated updates
	if err := h.DB.Where("stream_id = ?", streamID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&updates).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to fetch live updates",
		})
		return
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit)) // ceiling division

	span.SetAttributes(
		attribute.Int("response.total_items", int(total)),
		attribute.Int("response.total_pages", totalPages),
		attribute.Int("response.returned_items", len(updates)),
	)

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       updates,
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}
