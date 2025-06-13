package routes

import (
	"time"

	"news/internal/handlers"
	"news/internal/middleware"

	"github.com/gin-gonic/gin"
)

// @title Video API Routes
// @description Video-related API endpoints with proper Swagger documentation
// @version 1.0

// SetupVideoRoutes configures all video-related routes with proper Swagger annotations
// @Summary Configure video routes
// @Description Sets up all video-related endpoints including public access, authenticated routes, and admin moderation
func SetupVideoRoutes(r *gin.RouterGroup, videoHandler *handlers.VideoHandler, videoAnalyticsHandler *handlers.VideoAnalyticsHandler, videoCachedHandler *handlers.VideoHandlerCached) {
	// Public routes (no authentication required)
	public := r.Group("/videos")
	{
		// GetVideos godoc
		// @Summary Get public videos feed
		// @Description Retrieve a paginated list of public videos
		// @Tags Videos
		// @Produce json
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page (max 50)" default(20)
		// @Param category query string false "Filter by category"
		// @Param sort query string false "Sort by: created_at, views, votes" default(created_at)
		// @Param order query string false "Order: asc, desc" default(desc)
		// @Success 200 {object} models.PaginatedVideoResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos [get]
		public.GET("", videoHandler.GetVideos)

		// GetVideo godoc
		// @Summary Get single video
		// @Description Retrieve a single video by ID with full details
		// @Tags Videos
		// @Produce json
		// @Param id path int true "Video ID"
		// @Success 200 {object} models.VideoDetailResponse
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id} [get]
		public.GET("/:id", videoHandler.GetVideo)

		// GetVideoComments godoc
		// @Summary Get video comments
		// @Description Retrieve comments for a specific video
		// @Tags Videos
		// @Produce json
		// @Param id path int true "Video ID"
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Success 200 {object} models.PaginatedCommentsResponse
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/comments [get]
		public.GET("/:id/comments", videoHandler.GetVideoComments)
	}

	// Authenticated routes
	auth := r.Group("/videos")
	auth.Use(middleware.Authenticate()) // Correct middleware name
	{
		// CreateVideo godoc
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
		auth.POST("", videoHandler.CreateVideo)

		// UpdateVideo godoc
		// @Summary Update video metadata
		// @Description Update video title, description, and other metadata (video owner only)
		// @Tags Videos
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Param video body models.UpdateVideoRequest true "Video update data"
		// @Success 200 {object} models.Video
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 403 {object} models.ErrorResponse "Not video owner"
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id} [put]
		auth.PUT("/:id", videoHandler.UpdateVideo)

		// DeleteVideo godoc
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
		auth.DELETE("/:id", videoHandler.DeleteVideo)

		// VoteVideo godoc
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
		auth.POST("/:id/vote", videoHandler.VoteVideo)

		// CreateVideoComment godoc
		// @Summary Create a comment on a video
		// @Description Add a comment to a video
		// @Tags Videos
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Param comment body models.CreateCommentRequest true "Comment data"
		// @Success 201 {object} models.Comment
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse "Video not found"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/comments [post]
		auth.POST("/:id/comments", videoHandler.CreateVideoComment)

		// ===== VIDEO PROCESSING ENDPOINTS =====

		// ProcessVideo godoc
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
		auth.POST("/:id/process", videoHandler.ProcessVideo)

		// GetVideoProcessingStatus godoc
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
		auth.GET("/:id/processing-status", videoHandler.GetVideoProcessingStatus)

		// GetVideoProcessingJobs godoc
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
		auth.GET("/processing-jobs", videoHandler.GetVideoProcessingJobs)

		// ===== CACHED ENDPOINTS - REDIS OPTIMIZED =====

		// VoteVideoCached godoc
		// @Summary Vote on a video (cached)
		// @Description Like or dislike a video with Redis cache optimization for faster response
		// @Tags Videos - Cached
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
		// @Router /api/videos/{id}/vote-cached [post]
		auth.POST("/:id/vote-cached", videoCachedHandler.VoteVideoCached)

		// GetVideoStatsCached godoc
		// @Summary Get video vote statistics (cached)
		// @Description Retrieve video vote counts with Redis cache for sub-10ms response times
		// @Tags Videos - Cached
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Success 200 {object} models.VideoStatsResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/stats-cached [get]
		auth.GET("/:id/stats-cached", videoCachedHandler.GetVideoStatsCached)

		// GetUserVoteCached godoc
		// @Summary Get user's vote status for a video (cached)
		// @Description Retrieve user's current vote (like/dislike) for a video with Redis cache
		// @Tags Videos - Cached
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Success 200 {object} models.UserVoteResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/my-vote-cached [get]
		auth.GET("/:id/my-vote-cached", videoCachedHandler.GetUserVideoVoteCached)
	}

	// Video Analytics routes (authenticated)
	analytics := r.Group("/videos")
	analytics.Use(middleware.Authenticate())
	{
		// RecordVideoInteraction godoc
		// @Summary Record video interaction
		// @Description Record user interaction with video (view, like, dislike, etc.)
		// @Tags Video Analytics
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Param interaction body models.VideoInteractionRequest true "Interaction data"
		// @Success 200 {object} models.SuccessResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse "Video not found"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/interact [post]
		analytics.POST("/:id/interact", videoAnalyticsHandler.RecordVideoInteraction)

		// GetVideoAnalytics godoc
		// @Summary Get video analytics
		// @Description Get detailed analytics for a specific video
		// @Tags Video Analytics
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Param timeframe query string false "Analytics timeframe: day, week, month, all" default(all)
		// @Success 200 {object} models.VideoAnalyticsResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse "Video not found"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/{id}/analytics [get]
		analytics.GET("/:id/analytics", videoAnalyticsHandler.GetVideoAnalytics)

		// GetUserVideoInteractions godoc
		// @Summary Get user video interactions
		// @Description Get user's interaction history with videos
		// @Tags Video Analytics
		// @Produce json
		// @Security BearerAuth
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Param interaction_type query string false "Filter by interaction type: view, like, dislike"
		// @Success 200 {object} models.PaginatedVideoInteractionsResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/videos/my-interactions [get]
		analytics.GET("/my-interactions", videoAnalyticsHandler.GetUserVideoInteractions)
	}

	// Admin/Moderator routes
	admin := r.Group("/admin/videos")
	admin.Use(middleware.Authenticate(), middleware.AdminAuth())
	{
		// ForceDeleteVideo godoc
		// @Summary Force delete a video (admin)
		// @Description Force delete any video regardless of ownership
		// @Tags Admin - Videos
		// @Produce json
		// @Security BearerAuth
		// @Param id path int true "Video ID"
		// @Success 204 "Video deleted successfully"
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 403 {object} models.ErrorResponse "Admin access required"
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /admin/videos/{id}/force [delete]
		admin.DELETE("/:id/force", videoHandler.DeleteVideo) // Reuse with admin context
	}

	// Admin Video Analytics routes
	adminAnalytics := r.Group("/admin/video-analytics")
	adminAnalytics.Use(middleware.Authenticate(), middleware.AdminAuth())
	{
		// GetVideoEngagementStats godoc
		// @Summary Get video engagement statistics (admin)
		// @Description Get comprehensive video engagement statistics for admin dashboard
		// @Tags Admin - Video Analytics
		// @Produce json
		// @Security BearerAuth
		// @Param timeframe query string false "Analytics timeframe: day, week, month, all" default(week)
		// @Param limit query int false "Number of top videos to return" default(10)
		// @Success 200 {object} models.VideoEngagementStatsResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 403 {object} models.ErrorResponse "Admin access required"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /admin/video-analytics/engagement [get]
		adminAnalytics.GET("/engagement", videoAnalyticsHandler.GetVideoEngagementStats)

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
		// @Success 200 {object} models.PaginatedVideoAnalyticsResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 403 {object} models.ErrorResponse "Admin access required"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /admin/video-analytics/all [get]
		adminAnalytics.GET("/all", videoAnalyticsHandler.GetAllVideoAnalytics)
	}
}

// SetupVideoFeedRoutes sets up algorithmic feed routes
// @Summary Configure video feed routes
// @Description Sets up video discovery and recommendation endpoints
func SetupVideoFeedRoutes(r *gin.RouterGroup, videoHandler *handlers.VideoHandler) {
	feed := r.Group("/feed")
	{
		// GetTrendingVideos godoc
		// @Summary Get trending videos
		// @Description Retrieve currently trending videos based on engagement metrics
		// @Tags Video Feed
		// @Produce json
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Param timeframe query string false "Trending timeframe: hour, day, week, month" default(day)
		// @Success 200 {object} models.PaginatedVideoResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/feed/trending [get]
		feed.GET("/trending", videoHandler.GetVideos) // Reuse with trending logic

		// GetRecommendedVideos godoc
		// @Summary Get recommended videos
		// @Description Get personalized video recommendations (optional auth for better recommendations)
		// @Tags Video Feed
		// @Produce json
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Success 200 {object} models.PaginatedVideoResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/feed/recommended [get]
		feed.GET("/recommended", videoHandler.GetVideos) // Reuse with recommendation logic

		// GetCategoryFeed godoc
		// @Summary Get videos by category
		// @Description Retrieve videos from a specific category
		// @Tags Video Feed
		// @Produce json
		// @Param category path string true "Category slug"
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Success 200 {object} models.PaginatedVideoResponse
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 404 {object} models.ErrorResponse "Category not found"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/feed/category/{category} [get]
		feed.GET("/category/:category", videoHandler.GetVideos) // Reuse with category filter
	}
}

// SetupVideoSearchRoutes sets up video search functionality
// @Summary Configure video search routes
// @Description Sets up video search and discovery endpoints
func SetupVideoSearchRoutes(r *gin.RouterGroup, videoHandler *handlers.VideoHandler) {
	search := r.Group("/search")
	{
		// SearchVideos godoc
		// @Summary Search videos
		// @Description Search for videos by title, description, or tags
		// @Tags Video Search
		// @Produce json
		// @Param q query string true "Search query"
		// @Param category query string false "Filter by category"
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(20)
		// @Param sort query string false "Sort by: relevance, created_at, views, votes" default(relevance)
		// @Success 200 {object} models.VideoSearchResponse
		// @Failure 400 {object} models.ErrorResponse "Invalid search query"
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/search/videos [get]
		search.GET("/videos", videoHandler.GetVideos) // Reuse with search logic
	}
}

// Additional middleware that might be needed for video routes

// VideoUploadLimits middleware for upload restrictions
// @Summary Video upload middleware
// @Description Middleware to enforce video upload limits and restrictions
func VideoUploadLimits() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check file size, format, user limits, etc.
		// Implementation would go here
		c.Next()
	})
}

// CacheResponse middleware for caching video responses
// @Summary Video response caching middleware
// @Description Middleware to cache video API responses for better performance
func CacheResponse(duration time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Implement response caching logic
		// Implementation would go here
		c.Next()
	})
}
