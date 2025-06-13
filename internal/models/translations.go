package models

import (
	"time"

	"gorm.io/gorm"
)

// Translation represents system-wide UI translations
type Translation struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Key       string         `gorm:"size:100;not null;index:idx_translation_key_lang,unique" json:"key"`
	Language  string         `gorm:"size:5;not null;index:idx_translation_key_lang,unique" json:"language"`
	Value     string         `gorm:"type:text;not null" json:"value"`
	Category  string         `gorm:"size:50;not null;index" json:"category"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Translation
func (Translation) TableName() string {
	return "translations"
}

// LocalizedTranslation represents a translation with additional context
type LocalizedTranslation struct {
	ID       uint   `json:"id"`
	Key      string `json:"key"`
	Language string `json:"language"`
	Value    string `json:"value"`
	Category string `json:"category"`
}
