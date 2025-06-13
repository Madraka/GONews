package models

import (
	"time"

	"gorm.io/gorm"
)

// NewsStory represents Instagram/Facebook-style story format for news
type NewsStory struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Headline        string         `gorm:"size:255;not null" json:"headline"`
	ImageURL        string         `gorm:"size:255;not null" json:"image_url"`
	BackgroundColor string         `gorm:"size:7;default:'#000000'" json:"background_color"`
	TextColor       string         `gorm:"size:7;default:'#FFFFFF'" json:"text_color"`
	Duration        int            `gorm:"default:5" json:"duration"` // Duration in seconds
	ArticleID       *uint          `gorm:"index" json:"article_id"`
	ExternalURL     string         `gorm:"size:255" json:"external_url"`
	SortOrder       int            `gorm:"default:0" json:"sort_order"`
	StartTime       time.Time      `gorm:"not null" json:"start_time"`
	ViewCount       int            `gorm:"default:0" json:"view_count"`
	CreateUserID    uint           `gorm:"not null" json:"create_user_id"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// TableName specifies the table name for NewsStory
func (NewsStory) TableName() string {
	return "news_stories"
}

// StoryGroup represents a collection of related stories
type StoryGroup struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `gorm:"size:100;not null" json:"title"`
	Description   string         `gorm:"size:255" json:"description"`
	CoverImageURL string         `gorm:"size:255" json:"cover_image_url"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Stories []NewsStory `gorm:"many2many:story_group_items" json:"stories,omitempty"`
}

// StoryGroupItem represents the mapping between story groups and stories
type StoryGroupItem struct {
	StoryGroupID uint      `gorm:"primaryKey;autoIncrement:false" json:"story_group_id"`
	StoryID      uint      `gorm:"primaryKey;autoIncrement:false" json:"story_id"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	StoryGroup StoryGroup `gorm:"foreignKey:StoryGroupID" json:"story_group,omitempty"`
	Story      NewsStory  `gorm:"foreignKey:StoryID" json:"story,omitempty"`
}

// StoryView tracks which users have viewed which stories
type StoryView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	StoryID   uint      `gorm:"not null;index" json:"story_id"`
	ViewedAt  time.Time `gorm:"not null" json:"viewed_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User  User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Story NewsStory `gorm:"foreignKey:StoryID" json:"story,omitempty"`
}

// TableName specifies the table name for StoryView
func (StoryView) TableName() string {
	return "story_views"
}
