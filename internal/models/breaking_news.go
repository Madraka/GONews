package models

import (
	"time"

	"gorm.io/gorm"
)

// BreakingNewsBanner represents a breaking news banner that displays at the top of the site
type BreakingNewsBanner struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Title           string     `gorm:"column:title;size:255;not null" json:"title"`
	Content         string     `gorm:"column:content;type:text" json:"content"`
	ArticleID       *uint      `gorm:"column:article_id;index" json:"article_id"`
	Priority        int        `gorm:"column:priority;not null;default:1" json:"priority"` // Higher number = higher priority
	Style           string     `gorm:"column:style;size:50;default:'urgent'" json:"style"`
	TextColor       string     `gorm:"column:text_color;size:7;default:'#FFFFFF'" json:"text_color"`
	BackgroundColor string     `gorm:"column:background_color;size:7;default:'#DC2626'" json:"background_color"`
	StartTime       time.Time  `gorm:"column:start_time;not null;default:CURRENT_TIMESTAMP" json:"start_time"`
	EndTime         *time.Time `gorm:"column:end_time" json:"end_time"`
	IsActive        bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// Relations
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// IsVisible determines if the breaking news banner should be visible
func (b *BreakingNewsBanner) IsVisible() bool {
	now := time.Now()
	isInTimeRange := now.After(b.StartTime) && (b.EndTime == nil || now.Before(*b.EndTime))
	return b.IsActive && isInTimeRange
}

// BreakingNewsService represents a service for managing breaking news banners
type BreakingNewsService struct {
	DB *gorm.DB
}

// GetActiveBreakingNews returns all currently active breaking news banners
func (s *BreakingNewsService) GetActiveBreakingNews() ([]BreakingNewsBanner, error) {
	var banners []BreakingNewsBanner
	now := time.Now()

	err := s.DB.Where("is_active = ? AND start_time <= ? AND (end_time IS NULL OR end_time >= ?)",
		true, now, now).
		Order("priority DESC").
		Find(&banners).Error

	return banners, err
}
