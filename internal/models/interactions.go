package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment represents user comments on articles
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ArticleID uint           `gorm:"not null;index" json:"article_id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	ParentID  *uint          `gorm:"index" json:"parent_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Status    string         `gorm:"size:20;not null;default:'approved';index" json:"status"`
	IsEdited  bool           `gorm:"default:false" json:"is_edited"`
	CreatedAt time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Article Article   `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
	Votes   []Vote    `gorm:"foreignKey:CommentID" json:"votes,omitempty"`
}

// ValidateStatus validates comment status
func (c *Comment) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
		"spam":     true,
	}
	return allowedStatuses[c.Status]
}

// Vote represents user votes (likes/dislikes) on articles and comments
type Vote struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ArticleID *uint     `gorm:"index" json:"article_id"`
	CommentID *uint     `gorm:"index" json:"comment_id"`
	Type      string    `gorm:"size:10;not null;index" json:"type"` // like, dislike
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`

	// Relations
	User    User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	Comment *Comment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
}

// ValidateType validates vote type
func (v *Vote) ValidateType() bool {
	allowedTypes := map[string]bool{
		"like":    true,
		"dislike": true,
	}
	return allowedTypes[v.Type]
}

// Bookmark represents user bookmarks
type Bookmark struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ArticleID uint      `gorm:"not null;index" json:"article_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// Follow represents user following relationships
type Follow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FollowerID  uint      `gorm:"not null;index" json:"follower_id"`
	FollowingID uint      `gorm:"not null;index" json:"following_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Follower  User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// Subscription represents newsletter and notification subscriptions
type Subscription struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     *uint          `gorm:"index" json:"user_id"`
	Email      string         `gorm:"size:100;not null" json:"email"`
	Type       string         `gorm:"size:20;not null" json:"type"` // newsletter, notifications, category
	CategoryID *uint          `gorm:"index" json:"category_id"`
	TagID      *uint          `gorm:"index" json:"tag_id"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	Token      string         `gorm:"size:255;unique" json:"token"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tag      *Tag      `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// ValidateSubscriptionType validates subscription type
func (s *Subscription) ValidateSubscriptionType() bool {
	allowedTypes := map[string]bool{
		"newsletter":    true,
		"notifications": true,
		"category":      true,
		"tag":           true,
		"author":        true,
	}
	return allowedTypes[s.Type]
}
