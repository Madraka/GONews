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

// UnifiedAnalyticsHandler handles cross-platform analytics operations
type UnifiedAnalyticsHandler struct {
	db *gorm.DB
}

// NewUnifiedAnalyticsHandler creates a new unified analytics handler
func NewUnifiedAnalyticsHandler() *UnifiedAnalyticsHandler {
	return &UnifiedAnalyticsHandler{
		db: database.DB,
	}
}

// GetUnifiedDashboard godoc
// @Summary Get unified analytics dashboard
// @Description Get comprehensive analytics across articles and videos
// @Tags Unified Analytics
// @Produce json
// @Security BearerAuth
// @Param timeframe query string false "Timeframe: day, week, month, all" default(week)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/analytics/dashboard [get]
func (h *UnifiedAnalyticsHandler) GetUnifiedDashboard(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "week")

	var startDate time.Time
	switch timeframe {
	case "day":
		startDate = time.Now().AddDate(0, 0, -1)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "all":
		startDate = time.Time{} // Beginning of time
	default:
		startDate = time.Now().AddDate(0, 0, -7)
	}

	dashboard := map[string]interface{}{
		"timeframe":    timeframe,
		"start_date":   startDate,
		"end_date":     time.Now(),
		"generated_at": time.Now(),
	}

	// Content Overview
	contentOverview := map[string]interface{}{}

	// Total content counts
	var totalArticles, totalVideos int64
	h.db.Model(&models.Article{}).Count(&totalArticles)
	h.db.Model(&models.Video{}).Count(&totalVideos)

	contentOverview["total_articles"] = totalArticles
	contentOverview["total_videos"] = totalVideos
	contentOverview["total_content"] = totalArticles + totalVideos

	// Content published in timeframe
	var articlesPublished, videosPublished int64
	if timeframe != "all" {
		h.db.Model(&models.Article{}).Where("created_at >= ?", startDate).Count(&articlesPublished)
		h.db.Model(&models.Video{}).Where("created_at >= ?", startDate).Count(&videosPublished)
	} else {
		articlesPublished = totalArticles
		videosPublished = totalVideos
	}

	contentOverview["articles_published"] = articlesPublished
	contentOverview["videos_published"] = videosPublished
	contentOverview["content_published"] = articlesPublished + videosPublished

	dashboard["content_overview"] = contentOverview

	// Engagement Overview
	engagementOverview := map[string]interface{}{}

	// Article interactions
	var articleViews, articleLikes, articleComments int64
	articleQuery := h.db.Model(&models.UserArticleInteraction{}).Where("interaction_type = ?", "view")
	articleLikesQuery := h.db.Model(&models.UserArticleInteraction{}).Where("interaction_type = ?", "like")
	articleCommentsQuery := h.db.Model(&models.Comment{})

	if timeframe != "all" {
		articleQuery = articleQuery.Where("created_at >= ?", startDate)
		articleLikesQuery = articleLikesQuery.Where("created_at >= ?", startDate)
		articleCommentsQuery = articleCommentsQuery.Where("created_at >= ?", startDate)
	}

	articleQuery.Count(&articleViews)
	articleLikesQuery.Count(&articleLikes)
	articleCommentsQuery.Count(&articleComments)

	// Video interactions
	var videoViews, videoLikes, videoComments int64
	videoViewQuery := h.db.Model(&models.VideoView{})
	videoLikesQuery := h.db.Model(&models.VideoVote{}).Where("type = ?", "like")
	videoCommentsQuery := h.db.Model(&models.VideoComment{})

	if timeframe != "all" {
		videoViewQuery = videoViewQuery.Where("created_at >= ?", startDate)
		videoLikesQuery = videoLikesQuery.Where("created_at >= ?", startDate)
		videoCommentsQuery = videoCommentsQuery.Where("created_at >= ?", startDate)
	}

	videoViewQuery.Count(&videoViews)
	videoLikesQuery.Count(&videoLikes)
	videoCommentsQuery.Count(&videoComments)

	engagementOverview["article_views"] = articleViews
	engagementOverview["article_likes"] = articleLikes
	engagementOverview["article_comments"] = articleComments
	engagementOverview["video_views"] = videoViews
	engagementOverview["video_likes"] = videoLikes
	engagementOverview["video_comments"] = videoComments
	engagementOverview["total_views"] = articleViews + videoViews
	engagementOverview["total_likes"] = articleLikes + videoLikes
	engagementOverview["total_comments"] = articleComments + videoComments

	dashboard["engagement_overview"] = engagementOverview

	// Top Performing Content
	topContent := map[string]interface{}{}

	// Top articles
	type ContentPerformance struct {
		ID         uint      `json:"id"`
		Title      string    `json:"title"`
		Type       string    `json:"type"` // article or video
		Views      int64     `json:"views"`
		Likes      int64     `json:"likes"`
		Comments   int64     `json:"comments"`
		Engagement float64   `json:"engagement_rate"`
		AuthorName string    `json:"author_name"`
		CreatedAt  time.Time `json:"created_at"`
	}

	var topArticles []ContentPerformance
	articleAnalyticsQuery := `
		SELECT 
			a.id,
			a.title,
			'article' as type,
			COALESCE(views.count, 0) as views,
			COALESCE(likes.count, 0) as likes,
			COALESCE(comments.count, 0) as comments,
			CASE 
				WHEN COALESCE(views.count, 0) > 0 
				THEN (COALESCE(likes.count, 0) + COALESCE(comments.count, 0)) * 100.0 / views.count
				ELSE 0 
			END as engagement_rate,
			u.username as author_name,
			a.created_at
		FROM articles a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN (
			SELECT article_id, COUNT(*) as count 
			FROM user_article_interactions 
			WHERE interaction_type = 'view'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY article_id
		) views ON a.id = views.article_id
		LEFT JOIN (
			SELECT article_id, COUNT(*) as count 
			FROM user_article_interactions 
			WHERE interaction_type = 'like'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY article_id
		) likes ON a.id = likes.article_id
		LEFT JOIN (
			SELECT article_id, COUNT(*) as count 
			FROM comments` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY article_id
		) comments ON a.id = comments.article_id
		WHERE a.status = 'published'
		ORDER BY engagement_rate DESC, views DESC
		LIMIT 5
	`

	if timeframe != "all" {
		h.db.Raw(articleAnalyticsQuery, startDate, startDate, startDate).Scan(&topArticles)
	} else {
		h.db.Raw(articleAnalyticsQuery).Scan(&topArticles)
	}

	// Top videos
	var topVideos []ContentPerformance
	videoAnalyticsQuery := `
		SELECT 
			v.id,
			v.title,
			'video' as type,
			COALESCE(views.count, 0) as views,
			COALESCE(likes.count, 0) as likes,
			COALESCE(comments.count, 0) as comments,
			CASE 
				WHEN COALESCE(views.count, 0) > 0 
				THEN (COALESCE(likes.count, 0) + COALESCE(comments.count, 0)) * 100.0 / views.count
				ELSE 0 
			END as engagement_rate,
			u.username as author_name,
			v.created_at
		FROM videos v
		LEFT JOIN users u ON v.user_id = u.id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_views` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY video_id
		) views ON v.id = views.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_votes 
			WHERE type = 'like'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY video_id
		) likes ON v.id = likes.video_id
		LEFT JOIN (
			SELECT video_id, COUNT(*) as count 
			FROM video_comments` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
			GROUP BY video_id
		) comments ON v.id = comments.video_id
		WHERE v.is_public = true
		ORDER BY engagement_rate DESC, views DESC
		LIMIT 5
	`

	if timeframe != "all" {
		h.db.Raw(videoAnalyticsQuery, startDate, startDate, startDate).Scan(&topVideos)
	} else {
		h.db.Raw(videoAnalyticsQuery).Scan(&topVideos)
	}

	topContent["articles"] = topArticles
	topContent["videos"] = topVideos
	dashboard["top_content"] = topContent

	// User Activity Overview
	userActivity := map[string]interface{}{}

	var totalUsers, activeUsers int64
	h.db.Model(&models.User{}).Count(&totalUsers)

	// Active users (users who created content or interacted in timeframe)
	activeUsersQuery := `
		SELECT COUNT(DISTINCT user_id) 
		FROM (
			SELECT user_id FROM articles WHERE created_at >= ?
			UNION
			SELECT user_id FROM videos WHERE created_at >= ?
			UNION 
			SELECT user_id FROM user_article_interactions WHERE created_at >= ?
			UNION
			SELECT user_id FROM video_views WHERE created_at >= ? AND user_id IS NOT NULL
		) active_users
	`

	if timeframe != "all" {
		h.db.Raw(activeUsersQuery, startDate, startDate, startDate, startDate).Scan(&activeUsers)
	} else {
		activeUsers = totalUsers
	}

	userActivity["total_users"] = totalUsers
	userActivity["active_users"] = activeUsers
	if totalUsers > 0 {
		userActivity["activity_rate"] = float64(activeUsers) / float64(totalUsers) * 100
	} else {
		userActivity["activity_rate"] = 0
	}

	dashboard["user_activity"] = userActivity

	// Growth Trends (if not all time)
	if timeframe != "all" {
		trends := map[string]interface{}{}

		// Calculate growth compared to previous period
		var prevStartDate time.Time
		switch timeframe {
		case "day":
			prevStartDate = startDate.AddDate(0, 0, -1)
		case "week":
			prevStartDate = startDate.AddDate(0, 0, -7)
		case "month":
			prevStartDate = startDate.AddDate(0, -1, 0)
		}

		// Previous period metrics
		var prevArticleViews, prevVideoViews, prevContentPublished int64
		h.db.Model(&models.UserArticleInteraction{}).Where("interaction_type = ? AND created_at >= ? AND created_at < ?", "view", prevStartDate, startDate).Count(&prevArticleViews)
		h.db.Model(&models.VideoView{}).Where("created_at >= ? AND created_at < ?", prevStartDate, startDate).Count(&prevVideoViews)
		h.db.Raw("SELECT COUNT(*) FROM (SELECT 1 FROM articles WHERE created_at >= ? AND created_at < ? UNION ALL SELECT 1 FROM videos WHERE created_at >= ? AND created_at < ?) counts", prevStartDate, startDate, prevStartDate, startDate).Scan(&prevContentPublished)

		currentViews := articleViews + videoViews
		prevViews := prevArticleViews + prevVideoViews
		currentContentPublished := articlesPublished + videosPublished

		// Calculate growth rates
		if prevViews > 0 {
			trends["views_growth"] = float64(currentViews-prevViews) / float64(prevViews) * 100
		} else {
			trends["views_growth"] = 0
		}

		if prevContentPublished > 0 {
			trends["content_growth"] = float64(currentContentPublished-prevContentPublished) / float64(prevContentPublished) * 100
		} else {
			trends["content_growth"] = 0
		}

		dashboard["growth_trends"] = trends
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetContentComparison godoc
// @Summary Compare articles vs videos performance
// @Description Get comparative analytics between articles and videos
// @Tags Unified Analytics
// @Produce json
// @Security BearerAuth
// @Param timeframe query string false "Timeframe: day, week, month, all" default(month)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/analytics/comparison [get]
func (h *UnifiedAnalyticsHandler) GetContentComparison(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "month")

	var startDate time.Time
	switch timeframe {
	case "day":
		startDate = time.Now().AddDate(0, 0, -1)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "all":
		startDate = time.Time{}
	default:
		startDate = time.Now().AddDate(0, -1, 0)
	}

	comparison := map[string]interface{}{
		"timeframe":    timeframe,
		"start_date":   startDate,
		"end_date":     time.Now(),
		"generated_at": time.Now(),
	}

	// Article metrics
	articleMetrics := map[string]interface{}{}
	var articleCount, articleViews, articleLikes, articleComments int64
	var avgArticleEngagement float64

	articleCountQuery := h.db.Model(&models.Article{}).Where("status = ?", "published")
	articleViewQuery := h.db.Model(&models.UserArticleInteraction{}).Where("interaction_type = ?", "view")
	articleLikeQuery := h.db.Model(&models.UserArticleInteraction{}).Where("interaction_type = ?", "like")
	articleCommentQuery := h.db.Model(&models.Comment{})

	if timeframe != "all" {
		articleCountQuery = articleCountQuery.Where("created_at >= ?", startDate)
		articleViewQuery = articleViewQuery.Where("created_at >= ?", startDate)
		articleLikeQuery = articleLikeQuery.Where("created_at >= ?", startDate)
		articleCommentQuery = articleCommentQuery.Where("created_at >= ?", startDate)
	}

	articleCountQuery.Count(&articleCount)
	articleViewQuery.Count(&articleViews)
	articleLikeQuery.Count(&articleLikes)
	articleCommentQuery.Count(&articleComments)

	if articleViews > 0 {
		avgArticleEngagement = float64(articleLikes+articleComments) / float64(articleViews) * 100
	}

	articleMetrics["count"] = articleCount
	articleMetrics["views"] = articleViews
	articleMetrics["likes"] = articleLikes
	articleMetrics["comments"] = articleComments
	articleMetrics["avg_engagement"] = avgArticleEngagement
	if articleCount > 0 {
		articleMetrics["avg_views_per_article"] = float64(articleViews) / float64(articleCount)
	} else {
		articleMetrics["avg_views_per_article"] = 0
	}

	// Video metrics
	videoMetrics := map[string]interface{}{}
	var videoCount, videoViews, videoLikes, videoComments int64
	var avgVideoEngagement float64

	videoCountQuery := h.db.Model(&models.Video{}).Where("is_public = ?", true)
	videoViewQuery := h.db.Model(&models.VideoView{})
	videoLikeQuery := h.db.Model(&models.VideoVote{}).Where("type = ?", "like")
	videoCommentQuery := h.db.Model(&models.VideoComment{})

	if timeframe != "all" {
		videoCountQuery = videoCountQuery.Where("created_at >= ?", startDate)
		videoViewQuery = videoViewQuery.Where("created_at >= ?", startDate)
		videoLikeQuery = videoLikeQuery.Where("created_at >= ?", startDate)
		videoCommentQuery = videoCommentQuery.Where("created_at >= ?", startDate)
	}

	videoCountQuery.Count(&videoCount)
	videoViewQuery.Count(&videoViews)
	videoLikeQuery.Count(&videoLikes)
	videoCommentQuery.Count(&videoComments)

	if videoViews > 0 {
		avgVideoEngagement = float64(videoLikes+videoComments) / float64(videoViews) * 100
	}

	videoMetrics["count"] = videoCount
	videoMetrics["views"] = videoViews
	videoMetrics["likes"] = videoLikes
	videoMetrics["comments"] = videoComments
	videoMetrics["avg_engagement"] = avgVideoEngagement
	if videoCount > 0 {
		videoMetrics["avg_views_per_video"] = float64(videoViews) / float64(videoCount)
	} else {
		videoMetrics["avg_views_per_video"] = 0
	}

	comparison["articles"] = articleMetrics
	comparison["videos"] = videoMetrics

	// Overall comparison
	overallMetrics := map[string]interface{}{
		"total_content":  articleCount + videoCount,
		"total_views":    articleViews + videoViews,
		"total_likes":    articleLikes + videoLikes,
		"total_comments": articleComments + videoComments,
		"content_ratio": map[string]interface{}{
			"articles_percentage": func() float64 {
				if articleCount+videoCount > 0 {
					return float64(articleCount) / float64(articleCount+videoCount) * 100
				}
				return 0
			}(),
			"videos_percentage": func() float64 {
				if articleCount+videoCount > 0 {
					return float64(videoCount) / float64(articleCount+videoCount) * 100
				}
				return 0
			}(),
		},
		"engagement_winner": func() string {
			if avgArticleEngagement > avgVideoEngagement {
				return "articles"
			} else if avgVideoEngagement > avgArticleEngagement {
				return "videos"
			}
			return "tie"
		}(),
	}

	comparison["overall"] = overallMetrics

	c.JSON(http.StatusOK, comparison)
}

// GetUserEngagementReport godoc
// @Summary Get user engagement report
// @Description Get detailed user engagement metrics across content types
// @Tags Unified Analytics
// @Produce json
// @Security BearerAuth
// @Param timeframe query string false "Timeframe: day, week, month, all" default(month)
// @Param user_id query int false "Filter by specific user ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/analytics/user-engagement [get]
func (h *UnifiedAnalyticsHandler) GetUserEngagementReport(c *gin.Context) {
	timeframe := c.DefaultQuery("timeframe", "month")
	userIDStr := c.Query("user_id")

	var startDate time.Time
	switch timeframe {
	case "day":
		startDate = time.Now().AddDate(0, 0, -1)
	case "week":
		startDate = time.Now().AddDate(0, 0, -7)
	case "month":
		startDate = time.Now().AddDate(0, -1, 0)
	case "all":
		startDate = time.Time{}
	default:
		startDate = time.Now().AddDate(0, -1, 0)
	}

	report := map[string]interface{}{
		"timeframe":    timeframe,
		"start_date":   startDate,
		"end_date":     time.Now(),
		"generated_at": time.Now(),
	}

	// Top content creators
	type CreatorStats struct {
		UserID        uint    `json:"user_id"`
		Username      string  `json:"username"`
		ArticlesCount int64   `json:"articles_count"`
		VideosCount   int64   `json:"videos_count"`
		TotalViews    int64   `json:"total_views"`
		TotalLikes    int64   `json:"total_likes"`
		TotalComments int64   `json:"total_comments"`
		AvgEngagement float64 `json:"avg_engagement"`
	}

	var topCreators []CreatorStats
	creatorQuery := `
		SELECT 
			u.id as user_id,
			u.username,
			COALESCE(content_stats.articles_count, 0) as articles_count,
			COALESCE(content_stats.videos_count, 0) as videos_count,
			COALESCE(engagement_stats.total_views, 0) as total_views,
			COALESCE(engagement_stats.total_likes, 0) as total_likes,
			COALESCE(engagement_stats.total_comments, 0) as total_comments,
			CASE 
				WHEN COALESCE(engagement_stats.total_views, 0) > 0 
				THEN (COALESCE(engagement_stats.total_likes, 0) + COALESCE(engagement_stats.total_comments, 0)) * 100.0 / engagement_stats.total_views
				ELSE 0 
			END as avg_engagement
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				SUM(CASE WHEN content_type = 'article' THEN 1 ELSE 0 END) as articles_count,
				SUM(CASE WHEN content_type = 'video' THEN 1 ELSE 0 END) as videos_count
			FROM (
				SELECT user_id, 'article' as content_type FROM articles WHERE status = 'published'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
				UNION ALL
				SELECT user_id, 'video' as content_type FROM videos WHERE is_public = true` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
			) content
			GROUP BY user_id
		) content_stats ON u.id = content_stats.user_id
		LEFT JOIN (
			SELECT 
				user_id,
				SUM(views) as total_views,
				SUM(likes) as total_likes,
				SUM(comments) as total_comments
			FROM (
				SELECT 
					a.user_id,
					COALESCE(article_views.count, 0) as views,
					COALESCE(article_likes.count, 0) as likes,
					COALESCE(article_comments.count, 0) as comments
				FROM articles a
				LEFT JOIN (
					SELECT article_id, COUNT(*) as count 
					FROM user_article_interactions 
					WHERE interaction_type = 'view'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY article_id
				) article_views ON a.id = article_views.article_id
				LEFT JOIN (
					SELECT article_id, COUNT(*) as count 
					FROM user_article_interactions 
					WHERE interaction_type = 'like'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY article_id
				) article_likes ON a.id = article_likes.article_id
				LEFT JOIN (
					SELECT article_id, COUNT(*) as count 
					FROM comments` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY article_id
				) article_comments ON a.id = article_comments.article_id
				WHERE a.status = 'published'
				UNION ALL
				SELECT 
					v.user_id,
					COALESCE(video_views.count, 0) as views,
					COALESCE(video_likes.count, 0) as likes,
					COALESCE(video_comments.count, 0) as comments
				FROM videos v
				LEFT JOIN (
					SELECT video_id, COUNT(*) as count 
					FROM video_views` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY video_id
				) video_views ON v.id = video_views.video_id
				LEFT JOIN (
					SELECT video_id, COUNT(*) as count 
					FROM video_votes 
					WHERE type = 'like'` + func() string {
		if timeframe != "all" {
			return " AND created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY video_id
				) video_likes ON v.id = video_likes.video_id
				LEFT JOIN (
					SELECT video_id, COUNT(*) as count 
					FROM video_comments` + func() string {
		if timeframe != "all" {
			return " WHERE created_at >= ?"
		}
		return ""
	}() + `
					GROUP BY video_id
				) video_comments ON v.id = video_comments.video_id
				WHERE v.is_public = true
			) all_engagement
			GROUP BY user_id
		) engagement_stats ON u.id = engagement_stats.user_id
		WHERE (content_stats.articles_count > 0 OR content_stats.videos_count > 0)` + func() string {
		if userIDStr != "" {
			return " AND u.id = ?"
		}
		return ""
	}() + `
		ORDER BY avg_engagement DESC, total_views DESC
		LIMIT 20
	`

	var args []interface{}
	if timeframe != "all" {
		args = append(args, startDate, startDate, startDate, startDate, startDate, startDate, startDate, startDate)
	}
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
			return
		}
		args = append(args, userID)
		report["user_id"] = userID
	}

	h.db.Raw(creatorQuery, args...).Scan(&topCreators)
	report["top_creators"] = topCreators

	// User activity patterns
	activityPatterns := map[string]interface{}{}

	// Most active content type
	var articleCreators, videoCreators int64
	creatorCountQuery := h.db.Model(&models.User{}).Distinct("id")

	articleCreatorQuery := creatorCountQuery.Joins("JOIN articles ON users.id = articles.user_id").Where("articles.status = ?", "published")
	videoCreatorQuery := creatorCountQuery.Joins("JOIN videos ON users.id = videos.user_id").Where("videos.is_public = ?", true)

	if timeframe != "all" {
		articleCreatorQuery = articleCreatorQuery.Where("articles.created_at >= ?", startDate)
		videoCreatorQuery = videoCreatorQuery.Where("videos.created_at >= ?", startDate)
	}

	articleCreatorQuery.Count(&articleCreators)
	videoCreatorQuery.Count(&videoCreators)

	activityPatterns["article_creators"] = articleCreators
	activityPatterns["video_creators"] = videoCreators
	activityPatterns["preferred_content_type"] = func() string {
		if articleCreators > videoCreators {
			return "articles"
		} else if videoCreators > articleCreators {
			return "videos"
		}
		return "balanced"
	}()

	report["activity_patterns"] = activityPatterns

	c.JSON(http.StatusOK, report)
}
