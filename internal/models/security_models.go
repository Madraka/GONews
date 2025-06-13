package models

import (
	"time"

	"gorm.io/gorm"
)

// UserSession represents an active user session
type UserSession struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	TokenID   string         `gorm:"size:255;not null" json:"token_id"`
	IP        string         `gorm:"size:50" json:"ip"`
	UserAgent string         `gorm:"size:255" json:"user_agent"`
	Device    string         `gorm:"size:100" json:"device"`
	Location  string         `gorm:"size:100" json:"location,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt int64          `json:"expires_at"`
	Active    bool           `gorm:"default:true" json:"active"`
	RevokedAt *time.Time     `json:"revoked_at,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// LoginAttempt tracks login attempts (successful and failed)
type LoginAttempt struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        *uint     `json:"user_id,omitempty"` // Nullable because we might not have a valid user
	Username      string    `gorm:"size:50;not null" json:"username"`
	IP            string    `gorm:"size:50" json:"ip"`
	UserAgent     string    `gorm:"size:255" json:"user_agent"`
	Location      string    `gorm:"size:100" json:"location,omitempty"`
	Timestamp     time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Success       bool      `gorm:"default:false" json:"success"`
	FailureReason string    `gorm:"size:255" json:"failure_reason,omitempty"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null" json:"user_id"`
	EventType   string    `gorm:"size:50;not null" json:"event_type"`
	Description string    `gorm:"size:255" json:"description"`
	IP          string    `gorm:"size:50" json:"ip"`
	UserAgent   string    `gorm:"size:255" json:"user_agent"`
	Metadata    string    `gorm:"type:text" json:"metadata,omitempty"` // JSON data
	Timestamp   time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Severity    string    `gorm:"size:20;default:'info'" json:"severity"` // info, warning, critical
}

// UserTOTP represents a user's TOTP configuration for 2FA
type UserTOTP struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;uniqueIndex" json:"user_id"`
	Secret      string     `gorm:"size:255;not null" json:"-"` // TOTP secret
	BackupCodes string     `gorm:"type:text" json:"-"`         // JSON-encoded backup codes
	Enabled     bool       `gorm:"default:false" json:"enabled"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}
