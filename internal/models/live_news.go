package models

import (
	"time"
)

// LiveNewsStream represents a live news feed for real-time coverage
type LiveNewsStream struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Title         string     `gorm:"column:title;size:255;not null" json:"title"`
	Description   string     `gorm:"column:description;type:text" json:"description"`
	Status        string     `gorm:"column:status;size:20;not null;default:'scheduled'" json:"status"` // scheduled, live, ended
	StartTime     *time.Time `gorm:"column:start_time" json:"start_time"`
	EndTime       *time.Time `gorm:"column:end_time" json:"end_time"`
	IsHighlighted bool       `gorm:"column:is_highlighted;default:false" json:"is_highlighted"`
	ViewerCount   int        `gorm:"column:viewer_count;default:0" json:"viewer_count"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// These fields are used in the handler but not in the DB schema
	CoverImageURL string `gorm:"-" json:"cover_image_url,omitempty"`
	CategoryID    *uint  `gorm:"-" json:"category_id,omitempty"`

	// Relations
	Updates []LiveNewsUpdate `gorm:"foreignKey:StreamID" json:"updates,omitempty"`
}

// LiveNewsUpdate represents individual updates in a live news stream
type LiveNewsUpdate struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	StreamID   uint       `gorm:"column:stream_id;not null;index" json:"stream_id"`
	Title      string     `gorm:"column:title;size:255;not null" json:"title"`
	Content    string     `gorm:"column:content;type:text;not null" json:"content"`
	UpdateType string     `gorm:"column:update_type;size:50;default:'update'" json:"update_type"`
	Importance string     `gorm:"column:importance;size:20;default:'normal'" json:"importance"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// Relations
	Stream LiveNewsStream `gorm:"foreignKey:StreamID" json:"stream,omitempty"`
}

// ValidateStatus validates live stream status
func (l *LiveNewsStream) ValidateStatus() bool {
	allowedStatuses := map[string]bool{
		"draft": true,
		"live":  true,
		"ended": true,
	}
	return allowedStatuses[l.Status]
}

// IsActive checks if the live stream is currently active
func (l *LiveNewsStream) IsActive() bool {
	now := time.Now()
	isLive := l.Status == "live"
	inTimeRange := true

	if l.StartTime != nil {
		inTimeRange = inTimeRange && now.After(*l.StartTime)
	}

	if l.EndTime != nil {
		inTimeRange = inTimeRange && now.Before(*l.EndTime)
	}

	return isLive && inTimeRange
}
