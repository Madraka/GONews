package models

import (
	"time"

	"gorm.io/gorm"
)

// Video represents a short-form video content
type Video struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"size:200;not null"`
	Description string `json:"description" gorm:"type:text"`

	// File information
	VideoURL     string `json:"video_url" gorm:"size:500;not null"`
	ThumbnailURL string `json:"thumbnail_url" gorm:"size:500"`
	Duration     int    `json:"duration"` // Duration in seconds
	FileSize     int64  `json:"file_size"`
	Resolution   string `json:"resolution" gorm:"size:20"` // e.g., "1080x1920"

	// Content metadata
	CategoryID *uint    `json:"category_id" gorm:"index"`
	Category   Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Tags       string   `json:"tags" gorm:"type:text"` // JSON array of tags

	// User and ownership
	UserID      uint `json:"user_id" gorm:"not null;index"`
	User        User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsGenerated bool `json:"is_generated" gorm:"default:false"` // AI-generated content

	// Status and moderation (reusing existing patterns)
	Status     string `json:"status" gorm:"size:20;default:'pending';index;check:status IN ('draft','pending','published','rejected','archived')"`
	IsPublic   bool   `json:"is_public" gorm:"default:true;index"`
	IsFeatured bool   `json:"is_featured" gorm:"default:false;index"`

	// AI and content analysis
	AIGenerated    bool    `json:"ai_generated" gorm:"default:false;index"`
	AIConfidence   float64 `json:"ai_confidence" gorm:"default:0"`
	ContentWarning string  `json:"content_warning" gorm:"size:100"`

	// Engagement metrics
	ViewCount    int64 `json:"view_count" gorm:"default:0;index"`
	LikeCount    int64 `json:"like_count" gorm:"default:0;index"`
	DislikeCount int64 `json:"dislike_count" gorm:"default:0"`
	CommentCount int64 `json:"comment_count" gorm:"default:0"`
	ShareCount   int64 `json:"share_count" gorm:"default:0"`

	// Timestamps
	PublishedAt *time.Time     `json:"published_at" gorm:"index"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// Relations
	Comments []VideoComment `json:"comments,omitempty" gorm:"foreignKey:VideoID"`
	Votes    []VideoVote    `json:"votes,omitempty" gorm:"foreignKey:VideoID"`
	Views    []VideoView    `json:"views,omitempty" gorm:"foreignKey:VideoID"`
}

// VideoComment extends the existing comment pattern for videos
type VideoComment struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	VideoID uint   `json:"video_id" gorm:"not null;index"`
	Video   Video  `json:"video,omitempty" gorm:"foreignKey:VideoID"`
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	User    User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Content string `json:"content" gorm:"type:text;not null"`

	// Nested comments (reusing existing pattern)
	ParentID *uint          `json:"parent_id" gorm:"index"`
	Parent   *VideoComment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies  []VideoComment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`

	// Moderation (reusing existing patterns)
	Status   string     `json:"status" gorm:"size:20;default:'active';check:status IN ('active','hidden','deleted','flagged')"`
	IsEdited bool       `json:"is_edited" gorm:"default:false"`
	EditedAt *time.Time `json:"edited_at"`

	// AI moderation
	AIModerated   bool    `json:"ai_moderated" gorm:"default:false"`
	AIConfidence  float64 `json:"ai_confidence" gorm:"default:0"`
	ToxicityScore float64 `json:"toxicity_score" gorm:"default:0"`

	// Engagement
	LikeCount    int64 `json:"like_count" gorm:"default:0"`
	DislikeCount int64 `json:"dislike_count" gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// Relations
	Votes []VideoCommentVote `json:"votes,omitempty" gorm:"foreignKey:CommentID"`
}

// VideoVote represents likes/dislikes on videos (extending existing Vote pattern)
type VideoVote struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	VideoID uint   `json:"video_id" gorm:"not null;index"`
	Video   Video  `json:"video,omitempty" gorm:"foreignKey:VideoID"`
	UserID  uint   `json:"user_id" gorm:"not null;index"`
	User    User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Type    string `json:"type" gorm:"size:10;not null;check:type IN ('like','dislike')"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// Unique constraint to prevent duplicate votes
	// Will be added in migration
}

// VideoCommentVote represents likes/dislikes on video comments
type VideoCommentVote struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	CommentID uint         `json:"comment_id" gorm:"not null;index"`
	Comment   VideoComment `json:"comment,omitempty" gorm:"foreignKey:CommentID"`
	UserID    uint         `json:"user_id" gorm:"not null;index"`
	User      User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Type      string       `json:"type" gorm:"size:10;not null;check:type IN ('like','dislike')"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`
}

// VideoView tracks video views for analytics and recommendations
type VideoView struct {
	ID      uint  `json:"id" gorm:"primaryKey"`
	VideoID uint  `json:"video_id" gorm:"not null;index"`
	Video   Video `json:"video,omitempty" gorm:"foreignKey:VideoID"`
	UserID  *uint `json:"user_id" gorm:"index"` // Nullable for anonymous views
	User    *User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// View metadata
	IPAddress    string  `json:"ip_address" gorm:"size:45"` // IPv6 support
	UserAgent    string  `json:"user_agent" gorm:"size:500"`
	Duration     int     `json:"duration"`      // How long the video was watched (seconds)
	WatchPercent float64 `json:"watch_percent"` // Percentage of video watched

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VideoPlaylist for organizing videos
type VideoPlaylist struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"type:text"`
	UserID      uint   `json:"user_id" gorm:"not null;index"`
	User        User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	IsPublic    bool   `json:"is_public" gorm:"default:true"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index" swaggerignore:"true"`

	// Relations
	Items []VideoPlaylistItem `json:"items,omitempty" gorm:"foreignKey:PlaylistID"`
}

// VideoPlaylistItem represents videos in a playlist
type VideoPlaylistItem struct {
	ID         uint          `json:"id" gorm:"primaryKey"`
	PlaylistID uint          `json:"playlist_id" gorm:"not null;index"`
	Playlist   VideoPlaylist `json:"playlist,omitempty" gorm:"foreignKey:PlaylistID"`
	VideoID    uint          `json:"video_id" gorm:"not null;index"`
	Video      Video         `json:"video,omitempty" gorm:"foreignKey:VideoID"`
	Order      int           `json:"order" gorm:"default:0"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VideoProcessingJob tracks video processing tasks
type VideoProcessingJob struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	VideoID  uint   `json:"video_id" gorm:"not null;index"`
	Video    Video  `json:"video,omitempty" gorm:"foreignKey:VideoID"`
	JobType  string `json:"job_type" gorm:"size:50;not null"` // transcoding, thumbnail, tts, etc.
	Status   string `json:"status" gorm:"size:20;default:'pending'"`
	Progress int    `json:"progress" gorm:"default:0"` // 0-100
	ErrorMsg string `json:"error_msg" gorm:"type:text"`

	// Job parameters (JSON)
	Parameters string `json:"parameters" gorm:"type:text"`
	Result     string `json:"result" gorm:"type:text"`

	// Processing metadata
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Table names for GORM
func (Video) TableName() string {
	return "videos"
}

func (VideoComment) TableName() string {
	return "video_comments"
}

func (VideoVote) TableName() string {
	return "video_votes"
}

func (VideoCommentVote) TableName() string {
	return "video_comment_votes"
}

func (VideoView) TableName() string {
	return "video_views"
}

func (VideoPlaylist) TableName() string {
	return "video_playlists"
}

func (VideoPlaylistItem) TableName() string {
	return "video_playlist_items"
}

func (VideoProcessingJob) TableName() string {
	return "video_processing_jobs"
}
