package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system with comprehensive profile
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Username    string         `gorm:"size:50;unique;not null" json:"username"`
	Email       string         `gorm:"size:100;unique;not null" json:"email"`
	Password    string         `gorm:"size:255;not null" json:"-"`
	FirstName   string         `gorm:"size:50" json:"first_name"`
	LastName    string         `gorm:"size:50" json:"last_name"`
	Avatar      string         `gorm:"size:255" json:"avatar"`
	Bio         string         `gorm:"type:text" json:"bio"`
	Website     string         `gorm:"size:255" json:"website"`
	Location    string         `gorm:"size:100" json:"location"`
	Role        string         `gorm:"size:20;not null;default:'user';index" json:"role"`
	Status      string         `gorm:"size:20;not null;default:'active';index" json:"status"`
	IsVerified  bool           `gorm:"default:false;index" json:"is_verified"`
	LastLoginAt *time.Time     `gorm:"index" json:"last_login_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Articles            []Article                `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
	Comments            []Comment                `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Votes               []Vote                   `gorm:"foreignKey:UserID" json:"votes,omitempty"`
	Bookmarks           []Bookmark               `gorm:"foreignKey:UserID" json:"bookmarks,omitempty"`
	Followers           []Follow                 `gorm:"foreignKey:FollowingID" json:"followers,omitempty"`
	Following           []Follow                 `gorm:"foreignKey:FollowerID" json:"following,omitempty"`
	Subscriptions       []Subscription           `gorm:"foreignKey:UserID" json:"subscriptions,omitempty"`
	ArticleInteractions []UserArticleInteraction `gorm:"foreignKey:UserID" json:"article_interactions,omitempty"`
}

// ValidateRole validates if the role is allowed
func (u *User) ValidateRole() bool {
	allowedRoles := map[string]bool{
		"admin":     true,
		"editor":    true,
		"author":    true,
		"moderator": true,
		"user":      true,
	}
	return allowedRoles[u.Role]
}

// ValidateStatus validates if the status is allowed
func (u *User) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"active":    true,
		"inactive":  true,
		"suspended": true,
		"banned":    true,
	}
	return allowedStatuses[u.Status]
}

// UserProfile represents a public user profile
type UserProfile struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	Website   string    `json:"website"`
	Location  string    `json:"location"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID         uint `json:"user_id"`
	ArticlesCount  int  `json:"articles_count"`
	CommentsCount  int  `json:"comments_count"`
	FollowersCount int  `json:"followers_count"`
	FollowingCount int  `json:"following_count"`
	BookmarksCount int  `json:"bookmarks_count"`
	TotalViews     int  `json:"total_views"`
	TotalLikes     int  `json:"total_likes"`
}

// UpdateUserProfileRequest defines the structure for updating a user profile.
// Pointers are used to distinguish between zero values and fields not provided.
type UpdateUserProfileRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"` // New password
	Avatar    *string `json:"avatar,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Website   *string `json:"website,omitempty"`
	Location  *string `json:"location,omitempty"`
}

// UserProfileResponse defines the structure for a user's public profile.
type UserProfileResponse struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name,omitempty"`
	LastName       string    `json:"last_name,omitempty"`
	Avatar         string    `json:"avatar,omitempty"`
	Bio            string    `json:"bio,omitempty"`
	Website        string    `json:"website,omitempty"`
	Location       string    `json:"location,omitempty"`
	Role           string    `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	RecentArticles []Article `json:"recent_articles,omitempty"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
}

// PaginatedArticlesResponse defines the structure for paginated articles list.
type PaginatedArticlesResponse struct {
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Total    int64     `json:"total"`
	Articles []Article `json:"articles"`
}

// PaginatedNotificationsResponse defines the structure for paginated notifications list.
type PaginatedNotificationsResponse struct {
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
	Total         int64          `json:"total"`
	Notifications []Notification `json:"notifications"`
}
