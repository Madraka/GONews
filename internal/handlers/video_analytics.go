package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VideoAnalyticsHandler handles video analytics operations
type VideoAnalyticsHandler struct {
	db *gorm.DB
}

// NewVideoAnalyticsHandler creates a new video analytics handler
func NewVideoAnalyticsHandler() *VideoAnalyticsHandler {
	return &VideoAnalyticsHandler{
		db: database.DB,
	}
}

// VideoInteractionRequest represents the request payload for recording video interactions
type VideoInteractionRequest struct {
	InteractionType string   `json:"interaction_type" binding:"required"` // view, like, dislike, share, comment
	WatchPercent    *float64 `json:"watch_percent,omitempty"`             // For view interactions (0.0-1.0)
	Duration        *int     `json:"duration,omitempty"`                  // How long the video was watched (seconds)
	Platform        string   `json:"platform,omitempty"`                  // web, mobile, app
	Quality         string   `json:"quality,omitempty"`                   // 720p, 1080p, 4K, auto
	ReferrerURL     string   `json:"referrer_url,omitempty"`              // How user found the video
}

// RecordVideoInteraction godoc
// @Summary Record user interaction with video
// @Description Record a user's interaction with a video (view, like, dislike, share, etc.)
// @Tags Video Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param interaction body VideoInteractionRequest true "Interaction data"
// @Success 201 {object} models.VideoView
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/videos/{id}/interact [post]
func (h *VideoAnalyticsHandler) RecordVideoInteraction(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	userID, exists := c.Get("user_id")
	var userIDPtr *uint
	if exists {
		uid := userID.(uint)
		userIDPtr = &uid
	}

	var req VideoInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	// Verify video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	// Handle different interaction types
	switch req.InteractionType {
	case "view":
		// Record video view
		view := models.VideoView{
			VideoID:   uint(videoID),
			UserID:    userIDPtr,
			UserAgent: c.GetHeader("User-Agent"),
			IPAddress: c.ClientIP(),
		}

		// Set watch percent and duration from request
		if req.WatchPercent != nil {
			view.WatchPercent = *req.WatchPercent
		}
		if req.Duration != nil {
			view.Duration = *req.Duration
		}

		// For views, update existing view or create new one
		if userIDPtr != nil {
			var existingView models.VideoView
			err := h.db.Where("user_id = ? AND video_id = ?", *userIDPtr, videoID).First(&existingView).Error
			if err == nil {
				// Update existing view
				if req.WatchPercent != nil && *req.WatchPercent > existingView.WatchPercent {
					existingView.WatchPercent = *req.WatchPercent
				}
				if req.Duration != nil && *req.Duration > existingView.Duration {
					existingView.Duration = *req.Duration
				}
				existingView.UpdatedAt = time.Now()
				h.db.Save(&existingView)
				c.JSON(http.StatusOK, existingView)
				return
			}
		}

		// Create new view
		if err := h.db.Create(&view).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to record view"})
			return
		}
		c.JSON(http.StatusCreated, view)

	case "like", "dislike":
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required for voting"})
			return
		}

		// Handle video vote
		var existingVote models.VideoVote
		err := h.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&existingVote).Error

		if err == nil {
			// User already voted
			if existingVote.Type == req.InteractionType {
				// Same vote - remove it
				h.db.Delete(&existingVote)
			} else {
				// Different vote - update it
				existingVote.Type = req.InteractionType
				h.db.Save(&existingVote)
			}
		} else {
			// New vote
			vote := models.VideoVote{
				UserID:  userID.(uint),
				VideoID: uint(videoID),
				Type:    req.InteractionType,
			}
			h.db.Create(&vote)
		}

		// Get updated vote counts
		var likes, dislikes int64
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "like").Count(&likes)
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "dislike").Count(&dislikes)

		response := map[string]interface{}{
			"likes":    likes,
			"dislikes": dislikes,
		}
		c.JSON(http.StatusOK, response)

	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid interaction type"})
	}
}

// GetVideoAnalytics godoc
// @Summary Get video analytics
// @Description Get analytics data for a specific video
// @Tags Video Analytics
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/videos/{id}/analytics [get]
func (h *VideoAnalyticsHandler) GetVideoAnalytics(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	// Verify video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	// Get video analytics
	var stats struct {
		TotalViews      int64   `json:"total_views"`
		UniqueViews     int64   `json:"unique_views"`
		TotalLikes      int64   `json:"total_likes"`
		TotalDislikes   int64   `json:"total_dislikes"`
		TotalComments   int64   `json:"total_comments"`
		AvgWatchPercent float64 `json:"avg_watch_percent"`
		AvgWatchTime    float64 `json:"avg_watch_time"`
		ViewRetention   float64 `json:"view_retention"`
	}

	// Count views
	h.db.Model(&models.VideoView{}).
		Where("video_id = ?", videoID).
		Count(&stats.TotalViews)

	// Count unique views (authenticated users only)
	h.db.Model(&models.VideoView{}).
		Where("video_id = ? AND user_id IS NOT NULL", videoID).
		Distinct("user_id").
		Count(&stats.UniqueViews)

	// Count votes
	h.db.Model(&models.VideoVote{}).
		Where("video_id = ? AND type = ?", videoID, "like").
		Count(&stats.TotalLikes)

	h.db.Model(&models.VideoVote{}).
		Where("video_id = ? AND type = ?", videoID, "dislike").
		Count(&stats.TotalDislikes)

	// Count comments
	h.db.Model(&models.VideoComment{}).
		Where("video_id = ?", videoID).
		Count(&stats.TotalComments)

	// Calculate averages
	var avgWatchPercent float64
	h.db.Model(&models.VideoView{}).
		Where("video_id = ? AND watch_percent IS NOT NULL", videoID).
		Select("AVG(watch_percent)").
		Scan(&avgWatchPercent)
	stats.AvgWatchPercent = avgWatchPercent

	var avgWatchTime float64
	h.db.Model(&models.VideoView{}).
		Where("video_id = ? AND duration IS NOT NULL", videoID).
		Select("AVG(duration)").
		Scan(&avgWatchTime)
	stats.AvgWatchTime = avgWatchTime

	// Calculate retention rate (users who watched more than 50%)
	var retentionViews int64
	h.db.Model(&models.VideoView{}).
		Where("video_id = ? AND watch_percent > ?", videoID, 0.5).
		Count(&retentionViews)

	if stats.TotalViews > 0 {
		stats.ViewRetention = float64(retentionViews) / float64(stats.TotalViews)
	}

	response := map[string]interface{}{
		"video_id":     videoID,
		"title":        video.Title,
		"duration":     video.Duration,
		"stats":        stats,
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserVideoInteractions godoc
// @Summary Get user's video interactions
// @Description Get paginated list of user's interactions with videos
// @Tags Video Analytics
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Filter by interaction type"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Router /api/videos/my-interactions [get]
func (h *VideoAnalyticsHandler) GetUserVideoInteractions(c *gin.Context) {
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

	response := map[string]interface{}{
		"page":  page,
		"limit": limit,
	}

	switch interactionType {
	case "view", "":
		// Get video views
		var views []models.VideoView
		var total int64

		query := h.db.Where("user_id = ?", userID)
		query.Model(&models.VideoView{}).Count(&total)

		if err := query.Preload("Video").Preload("Video.User").
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&views).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch video views"})
			return
		}

		response["total"] = total
		response["views"] = views

	case "like", "dislike":
		// Get video votes
		var votes []models.VideoVote
		var total int64

		query := h.db.Where("user_id = ?", userID)
		if interactionType != "" {
			query = query.Where("type = ?", interactionType)
		}
		query.Model(&models.VideoVote{}).Count(&total)

		if err := query.Preload("Video").Preload("Video.User").
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&votes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch video votes"})
			return
		}

		response["total"] = total
		response["votes"] = votes

	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid interaction type"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoEngagementStats godoc
// @Summary Get video engagement statistics
// @Description Get engagement metrics for video content (admin only)
// @Tags Video Analytics
// @Produce json
// @Security BearerAuth
// @Param timeframe query string false "Timeframe: day, week, month" default(week)
// @Param video_id query int false "Filter by specific video ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/video-analytics/engagement [get]
func (h *VideoAnalyticsHandler) GetVideoEngagementStats(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "week")
	videoIDStr := c.Query("video_id")

	var startDate time.Time
	switch timeframe {
	case "day":
		startDate = time.Now().AddDate(0, 0, -1)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	default:
		startDate = time.Now().AddDate(0, 0, -7)
	}

	stats := map[string]interface{}{
		"timeframe":    timeframe,
		"start_date":   startDate,
		"end_date":     time.Now(),
		"generated_at": time.Now(),
	}

	query := h.db.Where("created_at >= ?", startDate)
	if videoIDStr != "" {
		videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
			return
		}
		query = query.Where("video_id = ?", videoID)
		stats["video_id"] = videoID
	}

	// Total views in timeframe
	var totalViews int64
	query.Model(&models.VideoView{}).Count(&totalViews)
	stats["total_views"] = totalViews

	// Total votes in timeframe
	var totalVotes int64
	query.Model(&models.VideoVote{}).Count(&totalVotes)
	stats["total_votes"] = totalVotes

	// Top performing videos
	type VideoPerformance struct {
		VideoID    uint    `json:"video_id"`
		Title      string  `json:"title"`
		Views      int64   `json:"views"`
		Likes      int64   `json:"likes"`
		Dislikes   int64   `json:"dislikes"`
		Comments   int64   `json:"comments"`
		Engagement float64 `json:"engagement_rate"`
	}

	var topVideos []VideoPerformance
	h.db.Raw(`
		SELECT 
			v.id as video_id,
			v.title,
			COALESCE(views.count, 0) as views,
			COALESCE(likes.count, 0) as likes,
			COALESCE(dislikes.count, 0) as dislikes,
			COALESCE(comments.count, 0) as comments,
			CASE 
				WHEN COALESCE(views.count, 0) > 0 
				THEN (COALESCE(likes.count, 0) + COALESCE(comments.count, 0)) * 100.0 / views.count
				ELSE 0 
			END as engagement_rate
		FROM videos v
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_views 
			WHERE created_at >= ? 
			GROUP BY video_id
		) views ON v.id = views.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_votes 
			WHERE type = 'like' AND created_at >= ? 
			GROUP BY video_id
		) likes ON v.id = likes.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_votes 
			WHERE type = 'dislike' AND created_at >= ? 
			GROUP BY video_id
		) dislikes ON v.id = dislikes.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_comments 
			WHERE created_at >= ? 
			GROUP BY video_id
		) comments ON v.id = comments.video_id
		WHERE v.is_public = true
		ORDER BY engagement_rate DESC, views DESC
		LIMIT 10
	`, startDate, startDate, startDate, startDate).Scan(&topVideos)

	stats["top_videos"] = topVideos

	c.JSON(http.StatusOK, stats)
}

// GetAllVideoAnalytics godoc
// @Summary Get analytics for all videos (admin)
// @Description Get comprehensive analytics across all videos in the system
// @Tags Admin - Video Analytics
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort by: views, engagement, created_at" default(views)
// @Param order query string false "Order: asc, desc" default(desc)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/video-analytics/all [get]
func (h *VideoAnalyticsHandler) GetAllVideoAnalytics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort", "views")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Validate sort parameters
	validSorts := map[string]string{
		"views":      "views",
		"engagement": "engagement_rate",
		"created_at": "v.created_at",
	}
	sortColumn, valid := validSorts[sortBy]
	if !valid {
		sortColumn = "views"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	type VideoAnalyticsSummary struct {
		VideoID        uint      `json:"video_id"`
		Title          string    `json:"title"`
		UserID         uint      `json:"user_id"`
		Username       string    `json:"username"`
		CreatedAt      time.Time `json:"created_at"`
		Views          int64     `json:"views"`
		UniqueViews    int64     `json:"unique_views"`
		Likes          int64     `json:"likes"`
		Dislikes       int64     `json:"dislikes"`
		Comments       int64     `json:"comments"`
		EngagementRate float64   `json:"engagement_rate"`
		AvgWatchTime   float64   `json:"avg_watch_time"`
		ViewRetention  float64   `json:"view_retention"`
	}

	var videos []VideoAnalyticsSummary
	var total int64

	// Count total videos
	h.db.Model(&models.Video{}).Count(&total)

	// Get comprehensive video analytics
	query := `
		SELECT 
			v.id as video_id,
			v.title,
			v.user_id,
			u.username,
			v.created_at,
			COALESCE(views.total_views, 0) as views,
			COALESCE(views.unique_views, 0) as unique_views,
			COALESCE(likes.count, 0) as likes,
			COALESCE(dislikes.count, 0) as dislikes,
			COALESCE(comments.count, 0) as comments,
			CASE 
				WHEN COALESCE(views.total_views, 0) > 0 
				THEN (COALESCE(likes.count, 0) + COALESCE(comments.count, 0)) * 100.0 / views.total_views
				ELSE 0 
			END as engagement_rate,
			COALESCE(views.avg_watch_time, 0) as avg_watch_time,
			COALESCE(views.retention_rate, 0) as view_retention
		FROM videos v
		LEFT JOIN users u ON v.user_id = u.id
		LEFT JOIN (
			SELECT 
				video_id, 
				COUNT(*) as total_views,
				COUNT(DISTINCT user_id) as unique_views,
				AVG(COALESCE(duration, 0)) as avg_watch_time,
				COUNT(CASE WHEN watch_percent > 0.5 THEN 1 END) * 100.0 / COUNT(*) as retention_rate
			FROM video_views 
			GROUP BY video_id
		) views ON v.id = views.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_votes 
			WHERE type = 'like' 
			GROUP BY video_id
		) likes ON v.id = likes.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_votes 
			WHERE type = 'dislike' 
			GROUP BY video_id
		) dislikes ON v.id = dislikes.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_comments 
			GROUP BY video_id
		) comments ON v.id = comments.video_id
		ORDER BY ` + sortColumn + ` ` + order + `
		LIMIT ? OFFSET ?
	`

	if err := h.db.Raw(query, limit, offset).Scan(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch video analytics"})
		return
	}

	response := map[string]interface{}{
		"page":         page,
		"limit":        limit,
		"total":        total,
		"sort":         sortBy,
		"order":        order,
		"videos":       videos,
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
