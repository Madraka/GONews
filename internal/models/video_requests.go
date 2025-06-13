package models

// Video-related request/response structures

// VoteRequest represents a vote request for videos or comments
type VoteRequest struct {
	Type string `json:"type" binding:"required" validate:"oneof=like dislike"` // like or dislike
}

// VoteResponse represents the response after voting
type VoteResponse struct {
	Likes    int64 `json:"likes"`
	Dislikes int64 `json:"dislikes"`
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required,min=5,max=1000"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

// UpdateCommentRequest represents a request to update a comment
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=5,max=1000"`
}

// VideoUploadRequest represents the form data for video upload
type VideoUploadRequest struct {
	Title       string `form:"title" binding:"required,max=200"`
	Description string `form:"description,omitempty" binding:"max=1000"`
	CategoryID  *uint  `form:"category_id,omitempty"`
	Tags        string `form:"tags,omitempty"`
	IsPublic    bool   `form:"is_public,omitempty"`
}

// VideoUpdateRequest represents video metadata update request
type VideoUpdateRequest struct {
	Title       string `json:"title,omitempty" binding:"max=200"`
	Description string `json:"description,omitempty" binding:"max=1000"`
	CategoryID  *uint  `json:"category_id,omitempty"`
	Tags        string `json:"tags,omitempty"`
	IsPublic    *bool  `json:"is_public,omitempty"`
	IsFeatured  *bool  `json:"is_featured,omitempty"`
}

// VideoFilter represents query parameters for filtering videos
type VideoFilter struct {
	Page       int    `form:"page,default=1" binding:"min=1"`
	Limit      int    `form:"limit,default=20" binding:"min=1,max=50"`
	CategoryID *uint  `form:"category_id,omitempty"`
	UserID     *uint  `form:"user_id,omitempty"`
	Status     string `form:"status,omitempty" binding:"omitempty,oneof=draft pending published rejected archived"`
	Sort       string `form:"sort,default=created_at" binding:"omitempty,oneof=created_at views votes likes title"`
	Order      string `form:"order,default=desc" binding:"omitempty,oneof=asc desc"`
	Search     string `form:"search,omitempty" binding:"max=100"`
	Featured   *bool  `form:"featured,omitempty"`
	Public     *bool  `form:"public,omitempty"`
}

// VideoStatsResponse represents video statistics
type VideoStatsResponse struct {
	VideoID      uint  `json:"video_id"`
	Views        int64 `json:"views"`
	Likes        int64 `json:"likes"`
	Dislikes     int64 `json:"dislikes"`
	Comments     int64 `json:"comments"`
	Shares       int64 `json:"shares"`
	WatchTime    int64 `json:"watch_time"`     // Total watch time in seconds
	AvgWatchTime int64 `json:"avg_watch_time"` // Average watch time per view
}

// VideoViewRequest represents a video view event
type VideoViewRequest struct {
	Duration     int     `json:"duration,omitempty"`      // How long watched in seconds
	WatchPercent float64 `json:"watch_percent,omitempty"` // Percentage watched
}

// CommentFilter represents query parameters for filtering comments
type CommentFilter struct {
	Page   int    `form:"page,default=1" binding:"min=1"`
	Limit  int    `form:"limit,default=20" binding:"min=1,max=50"`
	Sort   string `form:"sort,default=newest" binding:"omitempty,oneof=newest oldest likes"`
	Status string `form:"status,default=active" binding:"omitempty,oneof=active hidden deleted flagged"`
}

// VideoProcessingOptions represents options for video processing
type VideoProcessingOptions struct {
	Quality        string   `json:"quality,omitempty" binding:"omitempty,oneof=720p 1080p 1440p 4k"`     // Target quality
	GenerateThumbs bool     `json:"generate_thumbs,omitempty"`                                           // Generate thumbnail images
	Formats        []string `json:"formats,omitempty" binding:"omitempty,dive,oneof=mp4 webm hls"`       // Output formats
	Priority       string   `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent"` // Processing priority
	Webhooks       []string `json:"webhooks,omitempty" binding:"omitempty,dive,url"`                     // Callback URLs
	AIAnalysis     bool     `json:"ai_analysis,omitempty"`                                               // Enable AI content analysis
	Moderation     bool     `json:"moderation,omitempty"`                                                // Enable content moderation
	AutoPublish    bool     `json:"auto_publish,omitempty"`                                              // Auto-publish after processing
}

// Note: ErrorResponse and PaginatedResponse are already defined in other model files
