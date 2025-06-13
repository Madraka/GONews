# Video Module Implementation Guide

## Integration with Existing Go Backend

### Phase 1: Foundation Setup (Week 1-2)

#### 1.1 Database Migration
```bash
# Run the video module migration
psql -d your_database -f migrations/001_create_video_tables.sql

# Verify tables were created
psql -d your_database -c "\dt video*"
```

#### 1.2 Update Main Application Structure
```go
// main.go - Add video module initialization
func main() {
    // ... existing setup ...
    
    // Initialize video services
    storageService := storage.NewS3Storage(config.AWS)
    aiService := services.NewAIService(config.OpenAI)
    
    // Setup video processing queue
    processingQueue := services.NewVideoProcessingQueue(
        services.NewVideoProcessingService(db, storageService, aiService),
    )
    
    // Start background workers
    go processingQueue.ProcessJobs(context.Background())
    
    // Initialize handlers
    videoHandler := handlers.NewVideoHandler(db, storageService, aiService, processingQueue)
    
    // Setup routes
    routes.SetupVideoRoutes(router.Group("/api/v1"), videoHandler)
    
    // ... rest of setup ...
}
```

#### 1.3 Environment Configuration
```bash
# Add to .env file
VIDEO_UPLOAD_MAX_SIZE=500MB
VIDEO_PROCESSING_CONCURRENCY=4
VIDEO_STORAGE_BUCKET=your-video-bucket
CDN_BASE_URL=https://your-cdn.cloudfront.net
FFMPEG_PATH=/usr/bin/ffmpeg
TEMP_DIR=/tmp/video-processing
```

### Phase 2: Core Video CRUD (Week 3-4)

#### 2.1 Extend Existing User Model
```go
// internal/models/user.go - Add video-related fields
type User struct {
    // ... existing fields ...
    
    // Video-specific settings
    VideoUploadQuota    int64 `json:"video_upload_quota" gorm:"default:5368709120"` // 5GB default
    VideoUploadUsed     int64 `json:"video_upload_used" gorm:"default:0"`
    CanUploadVideos     bool  `json:"can_upload_videos" gorm:"default:true"`
    VideoQualityPrefs   string `json:"video_quality_prefs" gorm:"size:50;default:'auto'"` // auto, high, medium, low
    
    // Relations
    Videos []Video `json:"videos,omitempty" gorm:"foreignKey:UserID"`
}
```

#### 2.2 Update Existing Permission System
```go
// internal/middleware/permissions.go - Add video permissions
func checkVideoPermissions(userID uint, videoID uint, action string) bool {
    switch action {
    case "view":
        return canViewVideo(userID, videoID)
    case "edit":
        return isVideoOwner(userID, videoID) || hasRole(userID, "admin")
    case "delete":
        return isVideoOwner(userID, videoID) || hasRole(userID, "moderator")
    case "moderate":
        return hasRole(userID, "moderator") || hasRole(userID, "admin")
    default:
        return false
    }
}
```

#### 2.3 Integration with Existing Storage System
```go
// Extend existing storage service to handle videos
func (s *S3Storage) UploadVideo(file multipart.File, filename string) (string, error) {
    // Use existing upload logic but with video-specific paths and settings
    objectKey := fmt.Sprintf("videos/%s/%s", 
        time.Now().Format("2006/01/02"), 
        filename,
    )
    
    return s.UploadFile(file, objectKey)
}
```

### Phase 3: Comments and Reactions Integration (Week 5)

#### 3.1 Extend Existing Comment System
The video comment system reuses your existing comment patterns. Update your comment handlers to support video comments:

```go
// internal/handlers/comments.go - Add video comment support
func (h *CommentHandler) CreateComment(c *gin.Context) {
    var req struct {
        Content    string `json:"content" binding:"required"`
        ArticleID  *uint  `json:"article_id"`  // Existing
        VideoID    *uint  `json:"video_id"`    // New
        ParentID   *uint  `json:"parent_id"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Validate that either article_id or video_id is provided
    if req.ArticleID == nil && req.VideoID == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Either article_id or video_id is required"})
        return
    }
    
    // Create appropriate comment type
    if req.VideoID != nil {
        h.createVideoComment(c, req)
    } else {
        h.createArticleComment(c, req) // Existing logic
    }
}
```

#### 3.2 Extend Existing Vote System
```go
// internal/handlers/interactions.go - Add video voting
func (h *InteractionHandler) CreateVote(c *gin.Context) {
    var req struct {
        Type      string `json:"type" binding:"required"` // like, dislike
        ArticleID *uint  `json:"article_id"`
        VideoID   *uint  `json:"video_id"`
        CommentID *uint  `json:"comment_id"`
    }
    
    // Handle video votes using existing vote logic patterns
    if req.VideoID != nil {
        h.createVideoVote(c, req)
    } else if req.ArticleID != nil {
        h.createArticleVote(c, req) // Existing
    }
}
```

### Phase 4: AI Integration (Week 6-7)

#### 4.1 Extend Existing AI Service
```go
// internal/services/ai_service.go - Add video analysis capabilities
func (ai *AIService) AnalyzeVideo(videoPath string) (*VideoAnalysisResult, error) {
    // Reuse existing AI moderation patterns
    frames, err := ai.extractFrames(videoPath)
    if err != nil {
        return nil, err
    }
    
    // Use existing content moderation for each frame
    var violations []string
    var confidence float64
    
    for _, frame := range frames {
        result, err := ai.ModerateContent(frame) // Existing method
        if err != nil {
            continue
        }
        
        if !result.Safe {
            violations = append(violations, result.Reason)
        }
        confidence += result.Confidence
    }
    
    return &VideoAnalysisResult{
        IsAppropriate:    len(violations) == 0,
        ContentWarnings:  violations,
        Confidence:      confidence / float64(len(frames)),
    }, nil
}
```

#### 4.2 Queue Integration with Existing Job System
```go
// Extend existing translation queue pattern for video processing
type VideoProcessingTask struct {
    models.AgentTask // Reuse existing task structure
    VideoID          uint   `json:"video_id"`
    ProcessingType   string `json:"processing_type"` // thumbnail, transcode, ai_analysis
}

func (s *VideoProcessingService) QueueVideoJob(videoID uint, jobType string) error {
    task := VideoProcessingTask{
        AgentTask: models.AgentTask{
            Type:     "video_processing",
            Status:   "pending",
            Priority: 1,
        },
        VideoID:        videoID,
        ProcessingType: jobType,
    }
    
    return s.db.Create(&task).Error
}
```

### Phase 5: Feed Integration (Week 8)

#### 5.1 Extend Existing Feed System
```go
// internal/handlers/feed.go - Add video content to existing feeds
func (h *FeedHandler) GetUserFeed(c *gin.Context) {
    userID := getUserIDFromContext(c)
    
    // Get articles (existing)
    articles, err := h.getArticleFeed(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get articles"})
        return
    }
    
    // Get videos (new)
    videos, err := h.getVideoFeed(userID)
    if err != nil {
        log.Printf("Failed to get videos for feed: %v", err)
        videos = []models.Video{} // Don't fail the entire feed
    }
    
    // Merge and sort by engagement/recency
    feed := h.mergeFeedContent(articles, videos)
    
    c.JSON(http.StatusOK, gin.H{
        "feed": feed,
        "pagination": h.getPaginationInfo(c),
    })
}
```

#### 5.2 Update Existing Follow System
```go
// The existing follow system automatically works with videos since
// videos have user_id foreign key. Just update feed queries:

func (h *FeedHandler) getFollowingFeed(userID uint) ([]interface{}, error) {
    // Articles from followed users (existing)
    var articles []models.Article
    h.db.Joins("JOIN follows ON articles.user_id = follows.followed_id").
        Where("follows.follower_id = ? AND follows.status = ?", userID, "active").
        Find(&articles)
    
    // Videos from followed users (new)
    var videos []models.Video
    h.db.Joins("JOIN follows ON videos.user_id = follows.followed_id").
        Where("follows.follower_id = ? AND follows.status = ?", userID, "active").
        Find(&videos)
    
    return mergeFeedContent(articles, videos), nil
}
```

### Phase 6: Performance Optimization (Week 9-10)

#### 6.1 Database Optimization
```sql
-- Add indexes that work with existing patterns
CREATE INDEX CONCURRENTLY idx_videos_user_feed ON videos(user_id, status, published_at DESC) 
WHERE is_public = true AND deleted_at IS NULL;

-- Optimize existing queries to include videos
CREATE INDEX CONCURRENTLY idx_follows_with_content ON follows(follower_id, followed_id, status)
WHERE status = 'active';
```

#### 6.2 Caching Integration
```go
// Extend existing caching patterns
func (h *VideoHandler) GetTrendingVideos(c *gin.Context) {
    cacheKey := "trending:videos:24h"
    
    // Try existing cache first
    if cached, err := h.cache.Get(cacheKey); err == nil {
        c.JSON(http.StatusOK, cached)
        return
    }
    
    // Query database
    videos, err := h.getTrendingVideosFromDB()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending videos"})
        return
    }
    
    // Cache result using existing cache service
    h.cache.Set(cacheKey, videos, 1*time.Hour)
    
    c.JSON(http.StatusOK, videos)
}
```

### Phase 7: Testing and Deployment (Week 11-12)

#### 7.1 Integration Tests
```go
// tests/integration/video_test.go
func TestVideoWorkflow(t *testing.T) {
    // Test complete video lifecycle
    // 1. Upload video
    // 2. Verify processing jobs are queued
    // 3. Test video appears in feeds
    // 4. Test comments and votes work
    // 5. Test moderation workflow
}
```

#### 7.2 Migration Strategy
```bash
# 1. Deploy database migrations
kubectl apply -f k8s/video-migration-job.yaml

# 2. Deploy new application version with video support
kubectl set image deployment/news-api api=news-api:v2.0.0

# 3. Start video processing workers
kubectl apply -f k8s/video-processing-workers.yaml

# 4. Update frontend to support video upload
# (Frontend deployment)
```

## Integration Checklist

### ✅ Database Integration
- [ ] Run video table migrations
- [ ] Update existing models to reference videos
- [ ] Test foreign key constraints
- [ ] Verify indexes are created

### ✅ API Integration  
- [ ] Add video routes to existing router
- [ ] Test video CRUD operations
- [ ] Verify authentication/authorization works
- [ ] Test error handling

### ✅ Storage Integration
- [ ] Configure S3 bucket for videos
- [ ] Test file upload functionality
- [ ] Verify CDN integration
- [ ] Test file cleanup processes

### ✅ AI/Processing Integration
- [ ] Configure FFmpeg on servers
- [ ] Test video processing pipeline
- [ ] Verify AI moderation works
- [ ] Test queue processing

### ✅ Feed Integration
- [ ] Videos appear in user feeds
- [ ] Following feed includes videos
- [ ] Trending/recommended feeds work
- [ ] Pagination works correctly

### ✅ Performance Testing
- [ ] Load test video uploads
- [ ] Test concurrent processing
- [ ] Verify caching works
- [ ] Monitor database performance

### ✅ Security Testing
- [ ] Test file upload validation
- [ ] Verify content moderation
- [ ] Test permission systems
- [ ] Audit logging works

## Monitoring and Maintenance

### Key Metrics to Monitor
- Video upload success rate
- Processing queue depth
- Storage usage growth
- CDN bandwidth costs
- Database query performance
- User engagement rates

### Automated Alerts
- Video processing failures
- Storage quota approaching limits
- High error rates
- Database connection pool exhaustion
- CDN cache hit rate drops

This implementation guide provides a structured approach to integrating the video module with your existing Go backend while reusing as much existing infrastructure as possible.
