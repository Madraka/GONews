package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"news/internal/database"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoHandler struct {
	db *gorm.DB
}

func NewVideoHandler() *VideoHandler {
	return &VideoHandler{
		db: database.DB,
	}
}

// CreateVideo handles video upload and metadata creation
// @Summary Upload a new video
// @Description Upload and create a new video with metadata
// @Tags Videos
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param video formData file true "Video file to upload"
// @Param title formData string true "Video title"
// @Param description formData string false "Video description"
// @Param category_id formData int false "Category ID"
// @Param tags formData string false "Comma-separated tags"
// @Param is_public formData boolean false "Is video public" default(true)
// @Success 201 {object} models.Video
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse "File too large"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos [post]
func (h *VideoHandler) CreateVideo(c *gin.Context) {
	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(100 << 20) // 100MB limit
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to parse form data"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Video file is required"})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Failed to close video file: %v", err)
		}
	}()

	// Validate file type and size
	if !isVideoFileValid(header.Filename) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video format"})
		return
	}

	// For now, store video URL as placeholder (would integrate with storage service later)
	videoURL := "/uploads/videos/" + header.Filename

	// Create video record
	video := models.Video{
		Title:       c.PostForm("title"),
		Description: c.PostForm("description"),
		VideoURL:    videoURL,
		UserID:      userID,
		Status:      "pending",
		FileSize:    header.Size,
	}

	// Parse optional fields
	if categoryID, _ := strconv.Atoi(c.PostForm("category_id")); categoryID > 0 {
		catID := uint(categoryID)
		video.CategoryID = &catID
	}

	video.Tags = c.PostForm("tags")
	video.IsPublic = c.PostForm("is_public") != "false"

	if err := h.db.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create video"})
		return
	}

	c.JSON(http.StatusCreated, video)
}

// GetVideo retrieves a single video with details
// @Summary Get single video
// @Description Retrieve a single video by ID with full details
// @Tags Videos
// @Produce json
// @Param id path int true "Video ID"
// @Success 200 {object} models.Video
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id} [get]
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id := c.Param("id")

	var video models.Video
	if err := h.db.Preload("User").Preload("Category").First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		}
		return
	}

	// Record view if video is found
	userID := getVideoUserIDFromContext(c)
	if userID > 0 {
		h.recordVideoView(video.ID, userID, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(http.StatusOK, video)
}

// GetVideos retrieves a paginated list of videos
// @Summary Get videos feed
// @Description Retrieve a paginated list of videos with filtering options
// @Tags Videos
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 50)" default(20)
// @Param category query string false "Filter by category"
// @Param sort query string false "Sort by: created_at, views, votes" default(created_at)
// @Param order query string false "Order: asc, desc" default(desc)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos [get]
func (h *VideoHandler) GetVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := h.db.Preload("User").Preload("Category").Where("status = ? AND is_public = ?", "published", true)

	// Apply filters
	if category != "" {
		query = query.Joins("JOIN categories ON categories.id = videos.category_id").
			Where("categories.name = ?", category)
	}

	// Apply sorting
	orderBy := sort + " " + order
	if sort == "votes" {
		orderBy = "(like_count - dislike_count) " + order
	}

	var videos []models.Video
	var total int64

	// Get total count
	if err := query.Model(&models.Video{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	// Get videos with pagination
	if err := query.Order(orderBy).Offset(offset).Limit(limit).Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       videos,
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateVideo updates video metadata
// @Summary Update video metadata
// @Description Update video title, description, and other metadata (video owner only)
// @Tags Videos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param video body models.Video true "Video update data"
// @Success 200 {object} models.Video
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse "Not video owner"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id} [put]
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	id := c.Param("id")
	userID := getVideoUserIDFromContext(c)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	var video models.Video
	if err := h.db.First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		}
		return
	}

	// Check ownership or admin
	if video.UserID != userID && !isVideoAdmin(c) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to update this video"})
		return
	}

	var updateData models.Video
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update allowed fields
	if updateData.Title != "" {
		video.Title = updateData.Title
	}
	if updateData.Description != "" {
		video.Description = updateData.Description
	}
	if updateData.Tags != "" {
		video.Tags = updateData.Tags
	}
	video.IsPublic = updateData.IsPublic

	if err := h.db.Save(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update video"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// DeleteVideo deletes a video
// @Summary Delete a video
// @Description Delete a video (video owner only)
// @Tags Videos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Success 204 "Video deleted successfully"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse "Not video owner"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id} [delete]
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id := c.Param("id")
	userID := getVideoUserIDFromContext(c)

	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	var video models.Video
	if err := h.db.First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		}
		return
	}

	// Check ownership or admin
	if video.UserID != userID && !isVideoAdmin(c) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to delete this video"})
		return
	}

	if err := h.db.Delete(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete video"})
		return
	}

	c.Status(http.StatusNoContent)
}

// VoteVideo handles video voting (like/dislike)
// @Summary Vote on a video
// @Description Like or dislike a video
// @Tags Videos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param vote body models.VoteRequest true "Vote data (type: like/dislike)"
// @Success 200 {object} models.VoteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/vote [post]
func (h *VideoHandler) VoteVideo(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	var request models.VoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	if request.Type != "like" && request.Type != "dislike" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Vote type must be 'like' or 'dislike'"})
		return
	}

	// Check if video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	// Check if user already voted on this video
	var existingVote models.VideoVote
	err = h.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&existingVote).Error

	if err == nil {
		// User already voted
		if existingVote.Type == request.Type {
			// Same vote - remove it
			h.db.Delete(&existingVote)
		} else {
			// Different vote - update it
			existingVote.Type = request.Type
			h.db.Save(&existingVote)
		}
	} else {
		// New vote
		vote := models.VideoVote{
			UserID:  userID,
			VideoID: uint(videoID),
			Type:    request.Type,
		}
		h.db.Create(&vote)
	}

	// Get updated vote counts from the correct VideoVote table
	var likes, dislikes int64
	h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "like").Count(&likes)
	h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "dislike").Count(&dislikes)

	// Update the cached counts in the Video model for better performance
	h.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
		"like_count":    likes,
		"dislike_count": dislikes,
	})

	response := models.VoteResponse{
		Likes:    likes,
		Dislikes: dislikes,
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoComments retrieves comments for a video
// @Summary Get video comments
// @Description Retrieve comments for a specific video
// @Tags Videos
// @Produce json
// @Param id path int true "Video ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/comments [get]
func (h *VideoHandler) GetVideoComments(c *gin.Context) {
	videoID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Check if video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	// Use the proper VideoComment model
	var comments []models.VideoComment
	var total int64

	query := h.db.Preload("User").Where("video_id = ?", videoID)

	// Get total count
	if err := query.Model(&models.VideoComment{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	// Get comments with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       comments,
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// CreateVideoComment creates a comment on a video
// @Summary Create a comment on a video
// @Description Add a comment to a video
// @Tags Videos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param comment body models.CreateCommentRequest true "Comment data"
// @Success 201 {object} models.VideoComment
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "Video not found"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/comments [post]
func (h *VideoHandler) CreateVideoComment(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	// Check if video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	var commentData models.CreateCommentRequest
	if err := c.ShouldBindJSON(&commentData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Create video comment using the proper VideoComment model
	comment := models.VideoComment{
		Content: commentData.Content,
		UserID:  userID,
		VideoID: uint(videoID),
		Status:  "active",
	}

	if commentData.ParentID != nil {
		comment.ParentID = commentData.ParentID
	}

	if err := h.db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create comment"})
		return
	}

	// Load user relation
	h.db.Preload("User").First(&comment, comment.ID)

	c.JSON(http.StatusCreated, comment)
}

// ===== VIDEO PROCESSING HANDLERS =====

// ProcessVideo handles manual video processing requests
// @Summary Trigger video processing
// @Description Manually trigger video processing (transcoding, thumbnails, AI analysis)
// @Tags Video Processing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param options body models.VideoProcessingOptions false "Processing options"
// @Success 202 {object} models.ProcessingJobResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "Video not found"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/process [post]
func (h *VideoHandler) ProcessVideo(c *gin.Context) {
	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	// Verify video exists and user owns it
	var video models.Video
	if err := h.db.Where("id = ? AND user_id = ?", videoID, userID).First(&video).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		}
		return
	}

	// Get video processing service
	videoProcessingService := services.GetGlobalVideoProcessingService()
	if videoProcessingService == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "Video processing service not available"})
		return
	}

	// Get video processing queue
	videoProcessingQueue := services.GetGlobalVideoProcessingQueue()
	if videoProcessingQueue == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Error: "Video processing queue not available"})
		return
	}

	// Parse processing options if provided
	var options struct {
		GenerateThumbnails bool `json:"generate_thumbnails"`
		TranscodeVideo     bool `json:"transcode_video"`
		AnalyzeContent     bool `json:"analyze_content"`
		GenerateSubtitles  bool `json:"generate_subtitles"`
	}
	// Set defaults
	options.GenerateThumbnails = true
	options.TranscodeVideo = true
	options.AnalyzeContent = true
	options.GenerateSubtitles = false

	// Try to bind JSON, but don't fail if no body provided
	if err := c.ShouldBindJSON(&options); err != nil {
		// Log error but continue with default options
		// This is optional JSON binding
	}

	// Add processing job to queue
	job := services.ProcessingJob{
		Type:     "video_processing",
		VideoID:  video.ID,
		Priority: 1,
	}

	if err := videoProcessingQueue.AddJob(job); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to queue processing job"})
		return
	}

	response := models.ProcessingJobResponse{
		Message: "Video processing job queued successfully",
		VideoID: video.ID,
		Status:  "queued",
	}

	c.JSON(http.StatusAccepted, response)
}

// GetVideoProcessingStatus returns the processing status of a video
// @Summary Get video processing status
// @Description Get the current processing status of a video
// @Tags Video Processing
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Success 200 {object} models.VideoProcessingStatusResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse "Video not found"
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/processing-status [get]
func (h *VideoHandler) GetVideoProcessingStatus(c *gin.Context) {
	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	// Verify video exists and user owns it
	var video models.Video
	if err := h.db.Where("id = ? AND user_id = ?", videoID, userID).First(&video).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Database error"})
		}
		return
	}

	// Get processing jobs for this video
	var processingJobs []models.VideoProcessingJob
	if err := h.db.Where("video_id = ?", videoID).Order("created_at DESC").Find(&processingJobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch processing status"})
		return
	}

	// Determine overall status
	overallStatus := "not_processed"
	latestProgress := 0
	var latestError string
	var lastProcessedAt *time.Time

	if len(processingJobs) > 0 {
		// Check if any jobs are currently processing
		for _, job := range processingJobs {
			if job.Status == "processing" {
				overallStatus = "processing"
				if job.Progress > latestProgress {
					latestProgress = job.Progress
				}
			} else if job.Status == "failed" && overallStatus != "processing" {
				overallStatus = "failed"
				if job.ErrorMsg != "" {
					latestError = job.ErrorMsg
				}
			} else if job.Status == "completed" && overallStatus != "processing" && overallStatus != "failed" {
				overallStatus = "completed"
			}
			if job.CompletedAt != nil && (lastProcessedAt == nil || job.CompletedAt.After(*lastProcessedAt)) {
				lastProcessedAt = job.CompletedAt
			}
		}
	}

	// Convert processing jobs to response format
	jobResponses := make([]models.VideoProcessingJobResponse, len(processingJobs))
	for i, job := range processingJobs {
		var startedAt, completedAt *int64
		if job.StartedAt != nil {
			timestamp := job.StartedAt.Unix()
			startedAt = &timestamp
		}
		if job.CompletedAt != nil {
			timestamp := job.CompletedAt.Unix()
			completedAt = &timestamp
		}

		jobResponses[i] = models.VideoProcessingJobResponse{
			ID:            job.ID,
			VideoID:       job.VideoID,
			VideoTitle:    video.Title, // Use the video title we already have
			JobType:       job.JobType,
			Status:        job.Status,
			Progress:      job.Progress,
			ErrorMsg:      job.ErrorMsg,
			StartedAt:     startedAt,
			CompletedAt:   completedAt,
			CreatedAt:     job.CreatedAt.Unix(),
			LastUpdatedAt: job.UpdatedAt.Unix(),
		}
	}

	// Convert lastProcessedAt to Unix timestamp
	var lastProcessedAtUnix *int64
	if lastProcessedAt != nil {
		timestamp := lastProcessedAt.Unix()
		lastProcessedAtUnix = &timestamp
	}

	// Return structured video processing status
	response := models.VideoProcessingStatusResponse{
		VideoID:            video.ID,
		ProcessingStatus:   overallStatus,
		ThumbnailURL:       video.ThumbnailURL,
		ProcessingProgress: latestProgress,
		ProcessingError:    latestError,
		LastProcessedAt:    lastProcessedAtUnix,
		ProcessingJobs:     jobResponses,
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoProcessingJobs returns processing jobs for user's videos
// @Summary Get user's video processing jobs
// @Description Get all processing jobs for the authenticated user's videos
// @Tags Video Processing
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status: pending, processing, completed, failed"
// @Success 200 {object} models.PaginatedProcessingJobsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/processing-jobs [get]
func (h *VideoHandler) GetVideoProcessingJobs(c *gin.Context) {
	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	statusFilter := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query for processing jobs with video ownership validation
	query := h.db.Table("video_processing_jobs").
		Select("video_processing_jobs.*, videos.title as video_title").
		Joins("JOIN videos ON videos.id = video_processing_jobs.video_id").
		Where("videos.user_id = ?", userID)

	// Apply status filter if provided
	if statusFilter != "" {
		query = query.Where("video_processing_jobs.status = ?", statusFilter)
	}

	var total int64

	// Get total count
	countQuery := h.db.Table("video_processing_jobs").
		Joins("JOIN videos ON videos.id = video_processing_jobs.video_id").
		Where("videos.user_id = ?", userID)
	if statusFilter != "" {
		countQuery = countQuery.Where("video_processing_jobs.status = ?", statusFilter)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to count processing jobs"})
		return
	}

	// Create a struct to hold the joined data
	type JobWithVideoTitle struct {
		models.VideoProcessingJob
		VideoTitle string `json:"video_title"`
	}

	var jobsWithTitles []JobWithVideoTitle

	// Get processing jobs with pagination
	if err := query.Offset(offset).Limit(limit).Order("video_processing_jobs.updated_at DESC").Find(&jobsWithTitles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch processing jobs"})
		return
	}

	// Transform to API response format
	jobs := make([]models.VideoProcessingJobResponse, len(jobsWithTitles))
	for i, job := range jobsWithTitles {
		var startedAt, completedAt *int64
		if job.StartedAt != nil {
			timestamp := job.StartedAt.Unix()
			startedAt = &timestamp
		}
		if job.CompletedAt != nil {
			timestamp := job.CompletedAt.Unix()
			completedAt = &timestamp
		}

		jobs[i] = models.VideoProcessingJobResponse{
			ID:            job.ID,
			VideoID:       job.VideoID,
			VideoTitle:    job.VideoTitle,
			JobType:       job.JobType,
			Status:        job.Status,
			Progress:      job.Progress,
			ErrorMsg:      job.ErrorMsg,
			StartedAt:     startedAt,
			CompletedAt:   completedAt,
			CreatedAt:     job.CreatedAt.Unix(),
			LastUpdatedAt: job.UpdatedAt.Unix(),
		}
	}

	response := models.PaginatedProcessingJobsResponse{
		Jobs: jobs,
		Pagination: models.PaginationInfo{
			CurrentPage: page,
			PerPage:     limit,
			Total:       total,
			TotalPages:  (total + int64(limit) - 1) / int64(limit),
		},
	}

	c.JSON(http.StatusOK, response)
}

// ===== END VIDEO PROCESSING HANDLERS =====

// Helper functions

func getVideoUserIDFromContext(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

func isVideoAdmin(c *gin.Context) bool {
	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr == "admin"
		}
	}
	return false
}

func isVideoFileValid(filename string) bool {
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	allowedExts := []string{".mp4", ".webm", ".mov", ".avi", ".mkv"}
	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (h *VideoHandler) recordVideoView(videoID, userID uint, ip, userAgent string) {
	// Use the proper VideoView model
	view := models.VideoView{
		VideoID:      videoID,
		UserID:       &userID,
		IPAddress:    ip,
		UserAgent:    userAgent,
		Duration:     0,   // Will be updated when user interaction is recorded
		WatchPercent: 0.0, // Will be updated when user interaction is recorded
	}

	// Only create view if user ID is provided (not anonymous)
	if userID > 0 {
		h.db.Create(&view)
	}

	// Update view count
	h.db.Model(&models.Video{}).Where("id = ?", videoID).UpdateColumn("view_count", gorm.Expr("view_count + 1"))
}
