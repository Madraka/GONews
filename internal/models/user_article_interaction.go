package models

import (
	"time"

	"gorm.io/gorm"
)

// UserArticleInteraction represents user interactions with articles
// This model tracks reading history, views, bookmarks, and other user activities
type UserArticleInteraction struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	UserID          uint           `gorm:"not null;index" json:"user_id"`
	ArticleID       uint           `gorm:"not null;index" json:"article_id"`
	InteractionType string         `gorm:"size:20;not null;index" json:"interaction_type"` // view, bookmark, upvote, downvote, share, comment
	Duration        *int           `json:"duration"`                                       // Reading duration in seconds (for view interactions)
	CompletionRate  *float64       `gorm:"type:decimal(3,2)" json:"completion_rate"`       // Percentage of article read (0.0 - 1.0)
	Platform        string         `gorm:"size:20" json:"platform"`                        // web, mobile, app
	UserAgent       string         `gorm:"size:255" json:"user_agent"`                     // Browser/device info
	IPAddress       string         `gorm:"size:45" json:"ip_address"`                      // IPv4/IPv6 address
	ReferrerURL     string         `gorm:"size:500" json:"referrer_url"`                   // How user found the article
	CreatedAt       time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// TableName overrides the table name used by UserArticleInteraction to `user_article_interactions`
func (UserArticleInteraction) TableName() string {
	return "user_article_interactions"
}

// ValidateInteractionType validates the interaction type
func (uai *UserArticleInteraction) ValidateInteractionType() bool {
	validTypes := []string{"view", "bookmark", "upvote", "downvote", "share", "comment", "like", "dislike"}
	for _, validType := range validTypes {
		if uai.InteractionType == validType {
			return true
		}
	}
	return false
}

// IsReadingInteraction checks if this is a reading-related interaction
func (uai *UserArticleInteraction) IsReadingInteraction() bool {
	readingTypes := []string{"view", "bookmark"}
	for _, readingType := range readingTypes {
		if uai.InteractionType == readingType {
			return true
		}
	}
	return false
}

// IsEngagementInteraction checks if this is an engagement-related interaction
func (uai *UserArticleInteraction) IsEngagementInteraction() bool {
	engagementTypes := []string{"upvote", "downvote", "like", "dislike", "share", "comment"}
	for _, engagementType := range engagementTypes {
		if uai.InteractionType == engagementType {
			return true
		}
	}
	return false
}

// ReadingHistory represents a simplified view for reading history responses
type ReadingHistory struct {
	ArticleID       uint      `json:"article_id"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug"`
	Summary         string    `json:"summary"`
	AuthorName      string    `json:"author_name"`
	FeaturedImage   string    `json:"featured_image"`
	PublishedAt     time.Time `json:"published_at"`
	LastReadAt      time.Time `json:"last_read_at"`
	ReadingProgress float64   `json:"reading_progress"` // 0.0 - 1.0
	ReadCount       int       `json:"read_count"`
	Categories      []string  `json:"categories"`
	Tags            []string  `json:"tags"`
}

// UserReadingStats represents user reading statistics
type UserReadingStats struct {
	UserID              uint       `json:"user_id"`
	TotalArticlesRead   int        `json:"total_articles_read"`
	TotalReadingTime    int        `json:"total_reading_time"` // in seconds
	AverageReadingTime  int        `json:"average_reading_time"`
	PreferredCategories []string   `json:"preferred_categories"`
	ReadingStreak       int        `json:"reading_streak"` // consecutive days
	LastReadDate        *time.Time `json:"last_read_date"`
	ArticlesThisWeek    int        `json:"articles_this_week"`
	ArticlesThisMonth   int        `json:"articles_this_month"`
	CompletionRate      float64    `json:"completion_rate"` // Average completion rate
}
