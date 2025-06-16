package routes

import (
	"news/internal/auth"
	"news/internal/cache"
	"news/internal/database"
	"news/internal/handlers"
	"news/internal/middleware"
	"news/internal/tracing"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterRoutes(r *gin.Engine) {
	// Initialize API keys
	middleware.InitAPIKeys()

	// Initialize semantic search rate limiter
	redisClient := cache.GetRedisClient()
	var searchLimiter *middleware.SemanticSearchLimiter
	if redisClient != nil {
		searchLimiter = middleware.NewSemanticSearchLimiter(middleware.DefaultSearchLimitConfig(), redisClient.GetClient())
	} else {
		searchLimiter = middleware.NewSemanticSearchLimiter(middleware.DefaultSearchLimitConfig(), nil)
	}

	// Add enhanced OpenTelemetry middleware for distributed tracing
	r.Use(tracing.TracingMiddleware("news-api"))

	// Generate request ID for each request
	r.Use(func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})

	// Create and apply logger middleware
	logger := middleware.NewLogger()
	r.Use(middleware.LoggingMiddleware(logger))

	// Apply error handling middleware
	r.Use(middleware.ErrorHandlingMiddleware())

	// Apply CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length", "RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset", "X-Request-ID", "X-RateLimit-AI-Remaining", "X-RateLimit-AI-Used", "X-RateLimit-Local-Used", "X-RateLimit-User-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply gzip compression with error handling
	r.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Handle gzip panics gracefully (e.g., "flate: closed writer")
				if err, ok := r.(error); ok && err.Error() == "flate: closed writer" {
					// Client disconnected during compression - this is normal
					c.Abort()
					return
				}
				// Re-panic for other errors
				panic(r)
			}
		}()
		gzip.Gzip(gzip.DefaultCompression)(c)
	})

	// Favicon route to prevent 404 errors (both GET and HEAD)
	favicon := func(c *gin.Context) {
		c.Status(204) // No Content - prevents browser from making repeated requests
	}
	r.GET("/favicon.ico", favicon)
	r.HEAD("/favicon.ico", favicon)

	// Chrome DevTools route to prevent 404 errors
	r.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c *gin.Context) {
		c.Status(204) // No Content - Chrome DevTools integration not supported
	})

	// Debug routes for JSON Sonic performance testing (development only)
	debug := r.Group("/debug")
	{
		debug.GET("/json-engine", handlers.GetJSONEngineStatus)
		debug.GET("/json-performance", handlers.TestJSONPerformance)

		// HTTP/2 debug endpoints
		http2Group := debug.Group("/http2")
		{
			http2Group.GET("/status", handlers.GetHTTP2Status)
			http2Group.GET("/push-test", handlers.TestHTTP2Push)
		}
	}

	// Health check endpoint - no rate limiting, enhanced with optimized cache monitoring
	r.GET("/health", func(c *gin.Context) {
		// Check database connection
		dbHealthy := true
		if database.DB == nil {
			dbHealthy = false
		} else {
			sqlDB, err := database.DB.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbHealthy = false
			}
		}

		// Check Redis connection (legacy check)
		redisHealthy := true
		client := cache.GetRedisClient()
		if client == nil {
			redisHealthy = false
		} else {
			if err := client.Ping(); err != nil {
				redisHealthy = false
			}
		}

		// Check Standard Unified Cache health (L1 + L2)
		unifiedCache := cache.GetUnifiedCache()
		cacheHealth := unifiedCache.Health()
		unifiedCacheHealthy := cacheHealth["l1_healthy"] && cacheHealth["l2_healthy"]

		// Enhanced health check - try to get optimized cache health if available
		var optimizedCacheHealth map[string]interface{}
		var overallOptimizedHealthy bool

		// Try to get optimized cache manager health status
		if optimizedCache := cache.GetOptimizedUnifiedCache(); optimizedCache != nil {
			healthStatus := optimizedCache.GetHealthStatus()

			optimizedCacheHealth = map[string]interface{}{
				"overall_healthy":   healthStatus.OverallHealthy,
				"l1_healthy":        healthStatus.L1Healthy,
				"l2_healthy":        healthStatus.L2Healthy,
				"l1_hit_rate":       healthStatus.L1HitRate,
				"l2_hit_rate":       healthStatus.L2HitRate,
				"overall_hit_rate":  healthStatus.OverallHitRate,
				"avg_latency_l1":    healthStatus.AvgLatencyL1.Milliseconds(),
				"avg_latency_l2":    healthStatus.AvgLatencyL2.Milliseconds(),
				"request_count_l1":  healthStatus.RequestCountL1,
				"request_count_l2":  healthStatus.RequestCountL2,
				"singleflight_hits": healthStatus.SingleflightHits,
				"last_health_check": healthStatus.LastHealthCheck,
				"redis_connection": map[string]interface{}{
					"healthy":           healthStatus.RedisHealth.IsHealthy,
					"last_latency":      healthStatus.RedisHealth.LastLatency.Milliseconds(),
					"avg_latency":       healthStatus.RedisHealth.AvgLatency.Milliseconds(),
					"consecutive_fails": healthStatus.RedisHealth.ConsecutiveFails,
					"connection_time":   healthStatus.RedisHealth.ConnectionTime,
					"request_count":     healthStatus.RedisHealth.RequestCount,
				},
			}

			overallOptimizedHealthy = healthStatus.OverallHealthy
		} else {
			// Fallback to standard cache health if optimized not available
			optimizedCacheHealth = map[string]interface{}{
				"available": false,
				"message":   "Optimized cache not initialized, using standard cache",
			}
			overallOptimizedHealthy = unifiedCacheHealthy
		}

		// Determine overall system health
		overallHealthy := dbHealthy && redisHealthy && overallOptimizedHealthy
		healthStatus := "healthy"
		if !overallHealthy {
			if dbHealthy && (redisHealthy || overallOptimizedHealthy) {
				healthStatus = "degraded"
			} else {
				healthStatus = "unhealthy"
			}
		}

		response := gin.H{
			"status":        healthStatus,
			"time":          time.Now().String(),
			"db_healthy":    dbHealthy,
			"cache_healthy": redisHealthy,
			"unified_cache": map[string]interface{}{
				"healthy":    unifiedCacheHealthy,
				"l1_healthy": cacheHealth["l1_healthy"],
				"l2_healthy": cacheHealth["l2_healthy"],
			},
			"optimized_cache": optimizedCacheHealth,
			"version":         "1.0.0",
			"performance": map[string]interface{}{
				"cache_optimization_active": optimizedCacheHealth["available"] != false,
				"connection_pool_active":    overallOptimizedHealthy,
				"singleflight_active":       optimizedCacheHealth["singleflight_hits"] != nil,
			},
		}

		// Set appropriate HTTP status code based on health
		statusCode := 200
		if healthStatus == "degraded" {
			statusCode = 207 // Multi-Status - some services degraded
		} else if healthStatus == "unhealthy" {
			statusCode = 503 // Service Unavailable
		}

		c.JSON(statusCode, response)
	})

	// Version endpoint - no auth required
	r.GET("/version", handlers.GetVersion)

	// Global rate limiter (optimized for high-concurrency load testing)
	r.Use(middleware.RateLimit(50000, 100000, false)) // 50000 reqs/min, burst 100000

	// Auth routes (optimized rate limiting for high-concurrency load testing)
	authRoutes := r.Group("/api/auth")
	authRoutes.Use(middleware.RateLimit(50, 100, true)) // 50 reqs/sec per path, burst 100
	{
		// Secure authentication handlers
		authRoutes.POST("/register", handlers.RegisterWithSecurity)
		authRoutes.POST("/login", handlers.LoginWithSecurity)
		authRoutes.POST("/logout", middleware.Authenticate(), handlers.LogoutWithSecurity)
		authRoutes.POST("/refresh", handlers.RefreshToken)

		// User Profile Management (authenticated users)
		authRoutes.PUT("/profile", middleware.Authenticate(), handlers.UpdateProfile)
		authRoutes.GET("/notifications", middleware.Authenticate(), handlers.GetUserNotifications)
		authRoutes.PATCH("/notifications/:notification_id/read", middleware.Authenticate(), handlers.MarkNotificationRead)
		authRoutes.PATCH("/notifications/read-all", middleware.Authenticate(), handlers.MarkAllNotificationsRead)
	}

	// Security and 2FA routes (authenticated users only)
	security := r.Group("/")
	security.Use(middleware.Authenticate(), middleware.RateLimit(3, 6, true)) // 3 reqs/sec, burst of 6
	{
		// Initialize handlers
		tokenManager := auth.NewTokenManager(
			[]byte(middleware.GetJWTSecret()),
			24*time.Hour,
			7*24*time.Hour,
			cache.GetRedisClient(),
		)
		securityHandler := handlers.NewSecurityAuditHandler(tokenManager)
		twoFactorHandler := handlers.NewTwoFactorHandler()

		// Two-Factor Authentication
		security.POST("/2fa/setup", twoFactorHandler.Setup2FA)     // Setup 2FA
		security.POST("/2fa/enable", twoFactorHandler.Enable2FA)   // Enable 2FA
		security.POST("/2fa/disable", twoFactorHandler.Disable2FA) // Disable 2FA
		security.POST("/2fa/verify", twoFactorHandler.Verify2FA)   // Verify 2FA code
		security.GET("/2fa/status", twoFactorHandler.Get2FAStatus) // Get 2FA status

		// Security Audit
		security.GET("/security/sessions", securityHandler.GetUserSessions)              // Get active sessions
		security.DELETE("/security/sessions/:session_id", securityHandler.RevokeSession) // Revoke specific session
		security.DELETE("/security/sessions", securityHandler.RevokeAllSessions)         // Revoke all other sessions
		security.GET("/security/login-history", securityHandler.GetLoginHistory)         // Get login history
		security.GET("/security/events", securityHandler.GetSecurityEvents)              // Get security events
	}

	// Public API routes
	api := r.Group("/api")
	api.Use(middleware.RateLimit(50, 100, true)) // Increased public API rate limits for high-concurrency testing
	{
		// Public articles endpoints (optimized with raw JSON cache)
		// @Summary Get articles with pagination
		// @Description Retrieve a list of articles with pagination using cached JSON
		// @Tags Articles
		// @Accept json
		// @Produce json
		// @Param page query int false "Page number (default: 1)"
		// @Param limit query int false "Number of items per page (default: 10, max: 50)"
		// @Param category query string false "Filter by category"
		// @Success 200 {object} models.PaginatedResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/articles [get]
		api.GET("/articles", handlers.GetArticles)

		// Article specific routes (order matters: specific before wildcards)
		api.GET("/articles/recommendations", handlers.GetRecommendedArticles) // Get recommended articles
		api.GET("/articles/trending", handlers.GetTrendingArticles)           // Get trending articles
		api.GET("/articles/:id/similar", handlers.GetSimilarArticles)         // Get similar articles by ID

		// Single article route (cached JSON optimized)
		// @Summary Get a single article by ID
		// @Description Retrieve a single article by its ID using cached JSON for optimal performance
		// @Tags Articles
		// @Produce json
		// @Param id path int true "Article ID"
		// @Success 200 {object} models.Article
		// @Failure 404 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/articles/{id} [get]
		api.GET("/articles/:id", handlers.GetArticleById)

		// Get article with content blocks for editing
		api.GET("/articles/:id/with-blocks", handlers.GetArticleWithBlocks)

		// News content handlers (Breaking News, Stories, Live Streams)
		breakingNewsHandler := handlers.NewBreakingNewsHandler()
		newsStoriesHandler := handlers.NewNewsStoriesHandler()
		liveNewsHandler := handlers.NewLiveNewsHandler()

		// Breaking News endpoints (Public)
		api.GET("/breaking-news", breakingNewsHandler.GetActiveBreakingNews)

		// News Stories endpoints (Public)
		api.GET("/news-stories", newsStoriesHandler.GetActiveStories)
		api.GET("/news-stories/:id", newsStoriesHandler.GetStoryByID)

		// Live News Stream endpoints (Public)
		api.GET("/live-news", liveNewsHandler.GetActiveLiveStreams)
		api.GET("/live-news/:id", liveNewsHandler.GetLiveStreamByID)
		api.GET("/live-news/:id/updates", liveNewsHandler.GetLiveUpdates)

		// Video endpoints (Public and Authenticated) - using external route setup
		videoHandler := handlers.NewVideoHandler()
		videoAnalyticsHandler := handlers.NewVideoAnalyticsHandler()
		videoCachedHandler := handlers.NewVideoHandlerCached()
		SetupVideoRoutes(api, videoHandler, videoAnalyticsHandler, videoCachedHandler)

		// Categories & Tags (Public)
		api.GET("/categories", handlers.GetCategories)
		api.GET("/categories/:slug", handlers.GetCategoryBySlug)
		api.GET("/tags", handlers.GetTags)
		api.GET("/tags/:slug", handlers.GetTagBySlug)

		// User Profiles (Public)
		api.GET("/users/:username/profile", handlers.GetUserProfile)   // Get user's public profile
		api.GET("/users/:username/articles", handlers.GetUserArticles) // Get user's published articles

		// Menus (Public)
		api.GET("/menus", handlers.GetMenus)
		api.GET("/menus/:slug", handlers.GetMenuBySlug)
		api.GET("/menu-items", handlers.GetMenuItems)
		api.GET("/menu-items/:id", handlers.GetMenuItem)

		// Settings (Public - only public settings)
		api.GET("/settings", handlers.GetSettings)
		api.GET("/settings/:key", handlers.GetSettingByKey)
		api.GET("/settings/groups", handlers.GetSettingGroups)

		// Media (Public)
		api.GET("/media", handlers.GetMedia)
		api.GET("/media/:id", handlers.GetMediaByID)

		// Comments (Public view, authenticated to create/interact)
		api.GET("/articles/:id/comments", handlers.GetComments) // Get comments for an article

		// Article Translations (Public read with language support)
		translationHandler := handlers.NewArticleTranslationHandlers()
		api.GET("/articles/localized", translationHandler.GetLocalizedArticles)           // Get localized articles
		api.GET("/articles/:id/localized", translationHandler.GetLocalizedArticle)        // Get localized article
		api.GET("/articles/:id/translations", translationHandler.GetArticleTranslations)  // Get all translations for article
		api.GET("/articles/search/localized", translationHandler.SearchLocalizedArticles) // Search localized articles
		api.GET("/translations/stats", translationHandler.GetTranslationStats)            // Get translation statistics

		// System Translations (Public read with language support)
		api.GET("/content/:entity_type/:entity_id", handlers.GetLocalizedContent) // Get localized content (categories, tags, menus, notifications)

		// Unified Translation API (New comprehensive translation system)
		unifiedTranslationHandler := handlers.NewUnifiedTranslationHandler()
		api.GET("/translations/languages", unifiedTranslationHandler.GetSupportedLanguages)                            // Get supported languages
		api.POST("/translations/ui/:language", unifiedTranslationHandler.TranslateUI)                                  // Translate UI message
		api.GET("/translations/content/:entity_type/:entity_id/:language", unifiedTranslationHandler.TranslateContent) // Translate dynamic content
		api.POST("/translations/ai", middleware.Authenticate(), unifiedTranslationHandler.RequestAITranslation)        // Request AI translation
		api.GET("/translations/status/:job_id", unifiedTranslationHandler.GetTranslationStatus)                        // Get translation job status

		// Article Translations (Authenticated CRUD)
		api.POST("/articles/:id/translations", middleware.Authenticate(), translationHandler.CreateArticleTranslation)             // Create translation
		api.PUT("/articles/:id/translations/:language", middleware.Authenticate(), translationHandler.UpdateArticleTranslation)    // Update translation
		api.DELETE("/articles/:id/translations/:language", middleware.Authenticate(), translationHandler.DeleteArticleTranslation) // Delete translation

		// Cache Monitoring Endpoints (Public read-only) - currently implemented
		api.GET("/cache/stats", handlers.GetCacheStats)         // Cache performance statistics
		api.GET("/cache/health", handlers.GetCacheHealth)       // Cache health check
		api.GET("/cache/analytics", handlers.GetCacheAnalytics) // Advanced cache performance analytics
		api.POST("/cache/preload", handlers.PreloadCache)       // Manual cache warming trigger

		// Content Blocks (Public read, authenticated for creation/modification)
		api.GET("/articles/:id/blocks", handlers.GetContentBlocks)                                         // Get content blocks for an article (public)
		api.POST("/articles/:id/blocks", middleware.Authenticate(), handlers.CreateContentBlock)           // Create content block (authenticated)
		api.PUT("/blocks/:block_id", middleware.Authenticate(), handlers.UpdateContentBlock)               // Update content block (authenticated)
		api.DELETE("/blocks/:block_id", middleware.Authenticate(), handlers.DeleteContentBlock)            // Delete content block (authenticated)
		api.POST("/articles/:id/blocks/reorder", middleware.Authenticate(), handlers.ReorderContentBlocks) // Reorder content blocks (authenticated)

		// Pages (Public read endpoints)
		api.GET("/pages", handlers.GetPages)                   // Get all published pages with pagination
		api.GET("/pages/:id", handlers.GetPageByID)            // Get page by ID
		api.GET("/pages/slug/:slug", handlers.GetPageBySlug)   // Get page by slug
		api.GET("/pages/hierarchy", handlers.GetPageHierarchy) // Get page hierarchy
		api.GET("/pages/:id/blocks", handlers.GetPageBlocks)   // Get content blocks for a page

		// Content Block Utilities (Public and authenticated endpoints)
		api.POST("/content-blocks/detect-embeds", handlers.DetectEmbeds)                          // Detect embeds from URLs (public)
		api.POST("/content-blocks/analyze-url", handlers.AnalyzeURL)                              // Analyze URL for content extraction (public)
		api.POST("/content-blocks/embed", middleware.Authenticate(), handlers.CreateEmbedFromURL) // Create embed block from URL (authenticated)

		// Article Content Migration
		api.POST("/articles/:id/migrate-to-blocks", middleware.Authenticate(), handlers.MigrateArticleToBlocks) // Migrate article content to blocks
		api.PUT("/articles/:id/blocks", middleware.Authenticate(), handlers.UpdateArticleBlocks)                // Update all article blocks

		// Advanced Content Blocks (Authenticated endpoints for specialized block types)
		api.POST("/content-blocks/chart", middleware.Authenticate(), handlers.CreateChartBlock)                // Create chart block
		api.POST("/content-blocks/map", middleware.Authenticate(), handlers.CreateMapBlock)                    // Create map block
		api.POST("/content-blocks/faq", middleware.Authenticate(), handlers.CreateFAQBlock)                    // Create FAQ block
		api.POST("/content-blocks/newsletter", middleware.Authenticate(), handlers.CreateNewsletterBlock)      // Create newsletter signup block
		api.POST("/content-blocks/quiz", middleware.Authenticate(), handlers.CreateQuizBlock)                  // Create quiz block
		api.POST("/content-blocks/countdown", middleware.Authenticate(), handlers.CreateCountdownBlock)        // Create countdown timer block
		api.POST("/content-blocks/news-ticker", middleware.Authenticate(), handlers.CreateNewsTickerBlock)     // Create news ticker block
		api.POST("/content-blocks/breaking-news", middleware.Authenticate(), handlers.CreateBreakingNewsBlock) // Create breaking news banner

		// ...existing routes...
	}

	// v1 API routes with advanced features
	v1 := r.Group("/api/v1")
	v1.Use(middleware.Authenticate(), middleware.RateLimit(5, 10, true)) // Auth required for v1 endpoints
	{
		// Semantic Search endpoint (authenticated users with higher AI limits)
		v1.GET("/search", middleware.SemanticSearchRateLimit(searchLimiter), handlers.SemanticSearch)
		// Search limit status endpoint for authenticated users
		v1.GET("/search/limits", handlers.GetSearchLimitStatus)
	}

	// Public semantic search API (limited AI usage)
	publicSearch := r.Group("/api/search")
	publicSearch.Use(middleware.RateLimit(5, 10, true)) // Basic rate limiting for public endpoints
	{
		// Public semantic search endpoint (limited AI usage, falls back to local search)
		publicSearch.GET("/semantic", middleware.SemanticSearchRateLimit(searchLimiter), handlers.SemanticSearch)
		// Public search limit status endpoint
		publicSearch.GET("/limits", handlers.GetSearchLimitStatus)
	}

	// Authenticated User Interactions (Interactions like votes, bookmarks, follows)
	interactions := r.Group("/api")
	interactions.Use(middleware.Authenticate(), middleware.RateLimit(5, 10, true))
	{
		// Media Upload (Authenticated users)
		interactions.POST("/media/upload", handlers.UploadMedia)
		interactions.PUT("/media/:id", handlers.UpdateMedia)
		interactions.DELETE("/media/:id", handlers.DeleteMedia)

		// Comments (Create, Update, Delete, Vote)
		// @Summary Create a comment
		// @Description Add a new comment to an article
		// @Tags Comments
		// @Accept json
		// @Produce json
		// @Param id path string true "Article ID"
		// @Param comment body models.Comment true "Comment body"
		// @Success 201 {object} models.Comment
		// @Failure 400 {object} models.ErrorResponse
		// @Failure 401 {object} models.ErrorResponse
		// @Failure 500 {object} models.ErrorResponse
		// @Router /api/articles/{id}/comments [post]
		interactions.POST("/articles/:id/comments", handlers.CreateComment)
		interactions.PUT("/comments/:id", handlers.UpdateComment)
		interactions.DELETE("/comments/:id", handlers.DeleteComment)
		interactions.POST("/comments/:id/vote", handlers.VoteComment) // upvote/downvote a comment

		// Article Interactions
		interactions.POST("/articles/:id/bookmark", handlers.BookmarkArticle)
		interactions.POST("/articles/:id/vote", handlers.VoteArticle) // upvote/downvote an article
		interactions.GET("/bookmarks", handlers.GetUserBookmarksPage) // Get user's bookmarked articles

		// User Following System
		interactions.POST("/follow/:user_id", handlers.FollowUser)
		interactions.GET("/followers/:user_id", handlers.GetUserFollowers)
		interactions.GET("/following/:user_id", handlers.GetUserFollowing)

		// User Reading History (Authenticated)
		interactions.GET("/user/reading-history", handlers.GetReadingHistory) // Get user's reading history

		// News Stories - User specific interactions
		newsStoriesHandler := handlers.NewNewsStoriesHandler()
		interactions.GET("/news-stories/unviewed", newsStoriesHandler.GetUnviewedStories)

		// Analytics - User Article Interactions
		interactions.POST("/articles/:id/interactions", handlers.RecordArticleInteraction) // Record user interaction
		interactions.GET("/user/interactions", handlers.GetUserInteractions)               // Get user's interactions
		interactions.GET("/articles/:id/analytics", handlers.GetArticleAnalytics)          // Get article analytics
	}

	// Admin routes with JWT auth
	admin := r.Group("/admin")
	admin.Use(middleware.Authenticate(), middleware.AdminAuth(), middleware.RateLimit(5, 10, true))
	{
		admin.POST("/articles", handlers.CreateArticle)
		admin.PUT("/articles/:id", handlers.UpdateArticle)
		admin.DELETE("/articles/:id", handlers.DeleteArticle)

		// Admin Category Management
		admin.POST("/categories", handlers.CreateCategory)
		admin.PUT("/categories/:id", handlers.UpdateCategory)
		admin.DELETE("/categories/:id", handlers.DeleteCategory)

		// Admin Tag Management
		admin.POST("/tags", handlers.CreateTag)
		admin.PUT("/tags/:id", handlers.UpdateTag)
		admin.DELETE("/tags/:id", handlers.DeleteTag)

		// Page Management
		admin.POST("/pages", handlers.CreatePage)                  // Create page
		admin.PUT("/pages/:id", handlers.UpdatePage)               // Update page
		admin.DELETE("/pages/:id", handlers.DeletePage)            // Delete page
		admin.POST("/pages/:id/publish", handlers.PublishPage)     // Publish page
		admin.POST("/pages/:id/unpublish", handlers.UnpublishPage) // Unpublish page
		admin.POST("/pages/:id/duplicate", handlers.DuplicatePage) // Duplicate page

		// Page Content Block Management
		admin.POST("/pages/:id/blocks", handlers.CreatePageBlock)             // Create content block for page
		admin.GET("/page-blocks/:id", handlers.GetPageBlock)                  // Get content block
		admin.PUT("/page-blocks/:id", handlers.UpdatePageBlock)               // Update content block
		admin.DELETE("/page-blocks/:id", handlers.DeletePageBlock)            // Delete content block
		admin.POST("/page-blocks/:id/duplicate", handlers.DuplicatePageBlock) // Duplicate content block
		admin.POST("/page-blocks/validate", handlers.ValidatePageBlock)       // Validate content block

		// Content Management handlers initialization
		breakingNewsHandler := handlers.NewBreakingNewsHandler()
		newsStoriesHandler := handlers.NewNewsStoriesHandler()
		liveNewsHandler := handlers.NewLiveNewsHandler()

		// Breaking News Management
		admin.POST("/breaking-news", breakingNewsHandler.CreateBreakingNews)
		admin.PUT("/breaking-news/:id", breakingNewsHandler.UpdateBreakingNews)
		admin.DELETE("/breaking-news/:id", breakingNewsHandler.DeleteBreakingNews)

		// News Stories Management
		admin.POST("/news-stories", newsStoriesHandler.CreateStory)
		admin.PUT("/news-stories/:id", newsStoriesHandler.UpdateStory)
		admin.DELETE("/news-stories/:id", newsStoriesHandler.DeleteStory)

		// Live News Stream Management
		admin.POST("/live-news", liveNewsHandler.CreateLiveStream)
		admin.PUT("/live-news/:id", liveNewsHandler.UpdateLiveStream)
		admin.DELETE("/live-news/:id", liveNewsHandler.DeleteLiveStream)
		admin.POST("/live-news/:id/updates", liveNewsHandler.AddLiveUpdate)

		// Newsletter Management
		admin.GET("/newsletters", handlers.GetNewsletters)
		admin.GET("/newsletters/:id", handlers.GetNewsletter)
		admin.POST("/newsletters", handlers.CreateNewsletter)
		admin.PUT("/newsletters/:id", handlers.UpdateNewsletter)
		admin.DELETE("/newsletters/:id", handlers.DeleteNewsletter)
		admin.POST("/newsletters/:id/send", handlers.SendNewsletter)

		// Menu Management
		admin.POST("/menus", handlers.CreateMenu)
		admin.PUT("/menus/:id", handlers.UpdateMenu)
		admin.DELETE("/menus/:id", handlers.DeleteMenu)

		// Menu Item Management
		admin.POST("/menu-items", handlers.CreateMenuItem)
		admin.PUT("/menu-items/:id", handlers.UpdateMenuItem)
		admin.DELETE("/menu-items/:id", handlers.DeleteMenuItem)
		admin.PUT("/menu-items/reorder", handlers.ReorderMenuItems)

		// Settings Management
		admin.POST("/settings", handlers.CreateSetting)
		admin.PUT("/settings/:id", handlers.UpdateSetting)
		admin.PUT("/settings/key/:key", handlers.UpdateSettingByKey)
		admin.DELETE("/settings/:id", handlers.DeleteSetting)
		admin.PUT("/settings/bulk", handlers.BulkUpdateSettings)

		// Media Management
		admin.GET("/media/stats", handlers.GetMediaStats)

		// Translation Management
		admin.GET("/translations/progress", handlers.GetTranslationProgress)          // Get translation progress
		admin.POST("/translations/bulk", handlers.BulkTranslateContent)               // Bulk translate content
		admin.GET("/translations/queue", handlers.GetTranslationQueue)                // Get translation queue
		admin.POST("/translations/process", handlers.ProcessTranslationQueue)         // Process translation queue
		admin.POST("/translations/:entity_type/:entity_id", handlers.TranslateEntity) // Translate specific entity

		// Test endpoint for debugging translation queue (will be removed in production)
		admin.GET("/translations/test", handlers.TestTranslationSystem)

		// Unified Analytics Management (Cross-platform analytics)
		unifiedAnalyticsHandler := handlers.NewUnifiedAnalyticsHandler()
		admin.GET("/analytics/dashboard", unifiedAnalyticsHandler.GetUnifiedDashboard)           // Unified dashboard
		admin.GET("/analytics/content-comparison", unifiedAnalyticsHandler.GetContentComparison) // Articles vs Videos comparison
		admin.GET("/analytics/user-engagement", unifiedAnalyticsHandler.GetUserEngagementReport) // User engagement across platforms

		// Cache Management (Admin operations)
		admin.GET("/cache/stats", handlers.GetCacheStats)         // Cache statistics
		admin.GET("/cache/health", handlers.GetCacheHealth)       // Cache health check
		admin.GET("/cache/analytics", handlers.GetCacheAnalytics) // Advanced cache analytics
		admin.POST("/cache/preload", handlers.PreloadCache)       // Preload popular content
		admin.DELETE("/cache/clear", handlers.ClearCache)         // Clear cache (admin only)
		admin.POST("/cache/warm", handlers.WarmCache)             // Warm cache (admin only)
	}

	// Editor routes with JWT auth
	editor := r.Group("/editor")
	editor.Use(middleware.Authenticate(), middleware.EditorOnly())
	{
		editor.PUT("/articles/:id", handlers.UpdateArticle)
	}

	// Author routes with JWT auth
	author := r.Group("/author")
	author.Use(middleware.Authenticate(), middleware.AuthorOnly(), middleware.RateLimit(5, 10, true))
	{
		author.POST("/articles", handlers.CreateArticle)
		author.PUT("/articles/:id", handlers.UpdateArticle) // Authors can only edit their own articles
	}

	// API tier-specific routes (require API key authentication)
	apiKeyRoutes := r.Group("/api")
	apiKeyRoutes.Use(middleware.APIKeyAuth()) // Apply API key auth only to these routes
	{
		// Analytics endpoints (available to Pro and Enterprise tiers)
		apiKeyRoutes.GET("/analytics", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Analytics endpoint - requires Pro or Enterprise API key",
				"data": map[string]interface{}{
					"views":       12500,
					"users":       1200,
					"click_rate":  "4.5%",
					"top_article": "How to implement API key tiers",
				},
			})
		})

		// Export endpoints (available to Enterprise tier only)
		apiKeyRoutes.GET("/export", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Export endpoint - requires Enterprise API key",
				"data": map[string]interface{}{
					"export_url": "https://api.example.com/exports/latest.csv",
					"format":     "CSV",
					"generated":  time.Now().Format(time.RFC3339),
					"expires_in": "24 hours",
				},
			})
		})

		// Bulk operations (available to Enterprise tier only)
		apiKeyRoutes.POST("/bulk", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Bulk operations endpoint - requires Enterprise API key",
				"job_id":  "bulk-job-" + strconv.FormatInt(time.Now().Unix(), 10),
				"status":  "queued",
			})
		})
	}

	// AI routes with JWT auth (authenticated users only)
	ai := r.Group("/api/ai")
	ai.Use(middleware.Authenticate(), middleware.RateLimit(3, 6, true)) // 3 reqs/sec, burst of 6 for AI endpoints
	{
		ai.POST("/headlines", handlers.GenerateHeadlines)
		ai.POST("/content", handlers.GenerateContent)
		ai.POST("/improve", handlers.ImproveContent)
		ai.POST("/moderate", handlers.ModerateContent)
		ai.POST("/summarize", handlers.SummarizeContent)
		ai.POST("/categorize", handlers.CategorizeContent)
		ai.GET("/suggestions", handlers.GetAISuggestions)
		ai.GET("/usage-stats", handlers.GetAIUsageStats)
	}

	// Agent API routes for n8n integration
	agent := r.Group("/api/agent")
	agent.Use(middleware.Authenticate(), middleware.RateLimit(5, 10, true)) // 5 reqs/sec, burst of 10
	{
		agent.POST("/tasks", handlers.CreateAgentTask)
		agent.GET("/tasks", handlers.GetAgentTasks)
		agent.GET("/tasks/:id", handlers.GetAgentTask)
		agent.PUT("/tasks/:id", handlers.UpdateAgentTask)
		agent.DELETE("/tasks/:id", handlers.DeleteAgentTask)
		agent.POST("/tasks/:id/process", handlers.ProcessAgentTask)
	}

	// WebSocket routes for real-time notifications
	ws := r.Group("/ws")
	ws.Use(middleware.RateLimit(5, 10, true)) // 5 reqs/sec, burst of 10 for WebSocket connections
	{
		// WebSocket connection for authenticated users (handles auth manually due to query param)
		ws.GET("/notifications", handlers.HandleWebSocketNotifications)

		// WebSocket management endpoints
		ws.GET("/stats", middleware.Authenticate(), handlers.GetNotificationStats)
		ws.GET("/user/:user_id/status", middleware.Authenticate(), handlers.GetUserConnectionStatus)
		ws.POST("/test", middleware.Authenticate(), handlers.SendTestNotification)
	}
}
