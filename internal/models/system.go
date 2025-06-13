package models

import (
	"time"

	"gorm.io/gorm"
)

// Newsletter represents newsletter campaigns
type Newsletter struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Subject     string         `gorm:"size:255;not null" json:"subject"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Status      string         `gorm:"size:20;not null;default:'draft'" json:"status"`
	SentAt      *time.Time     `json:"sent_at"`
	ScheduledAt *time.Time     `json:"scheduled_at"`
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// Notification represents system notifications
type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Type      string         `gorm:"size:50;not null" json:"type"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Data      string         `gorm:"type:json" json:"data"` // Additional data as JSON
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	ReadAt    *time.Time     `json:"read_at"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ValidateNotificationType validates notification type
func (n *Notification) ValidateNotificationType() bool {
	allowedTypes := map[string]bool{
		"article_published": true,
		"comment_reply":     true,
		"comment_like":      true,
		"article_like":      true,
		"new_follower":      true,
		"mention":           true,
		"newsletter":        true,
		"system":            true,
	}
	return allowedTypes[n.Type]
}

// Menu represents website navigation menu
type Menu struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Slug      string         `gorm:"size:100;unique;not null" json:"slug"`
	Location  string         `gorm:"size:50;not null" json:"location"` // header, footer, sidebar
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Items []MenuItem `gorm:"foreignKey:MenuID" json:"items,omitempty"`
}

// MenuItem represents menu items
type MenuItem struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	MenuID     uint           `gorm:"not null;index" json:"menu_id"`
	ParentID   *uint          `gorm:"index" json:"parent_id"`
	Title      string         `gorm:"size:100;not null" json:"title"`
	URL        string         `gorm:"size:255" json:"url"`
	CategoryID *uint          `gorm:"index" json:"category_id"`
	Icon       string         `gorm:"size:50" json:"icon"`
	Target     string         `gorm:"size:20;default:'_self'" json:"target"`
	SortOrder  int            `gorm:"default:0" json:"sort_order"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Menu     Menu       `gorm:"foreignKey:MenuID" json:"menu,omitempty"`
	Parent   *MenuItem  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []MenuItem `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Category *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// Setting represents system settings
type Setting struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Key         string         `gorm:"size:100;unique;not null" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`
	Type        string         `gorm:"size:20;not null;default:'string'" json:"type"`
	Description string         `gorm:"type:text" json:"description"`
	Group       string         `gorm:"size:50;column:group" json:"group"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ValidateSettingType validates setting type
func (s *Setting) ValidateSettingType() bool {
	allowedTypes := map[string]bool{
		"string":  true,
		"integer": true,
		"boolean": true,
		"json":    true,
		"text":    true,
	}
	return allowedTypes[s.Type]
}

// Media represents uploaded media files
type Media struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	FileName     string         `gorm:"size:255;not null" json:"file_name"`
	OriginalName string         `gorm:"size:255;not null" json:"original_name"`
	MimeType     string         `gorm:"size:100;not null" json:"mime_type"`
	Size         int64          `gorm:"not null" json:"size"`
	Path         string         `gorm:"size:500;not null" json:"path"`
	URL          string         `gorm:"size:500;not null" json:"url"`
	AltText      string         `gorm:"size:255" json:"alt_text"`
	Caption      string         `gorm:"type:text" json:"caption"`
	UploadedBy   uint           `gorm:"not null" json:"uploaded_by"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Uploader User `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}
