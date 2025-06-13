package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TranslationQueue struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EntityType   string    `gorm:"not null" json:"entity_type"`
	EntityID     uint      `gorm:"not null" json:"entity_id"`
	SourceLang   string    `gorm:"column:source_language;size:5;not null" json:"source_lang"`
	TargetLang   string    `gorm:"column:target_language;size:5;not null" json:"target_lang"`
	Status       string    `gorm:"default:'pending'" json:"status"`
	Priority     int       `gorm:"default:1" json:"priority"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (TranslationQueue) TableName() string {
	return "translation_queue"
}

// BeforeCreate validates the queue entry before creating it
func (tq *TranslationQueue) BeforeCreate(tx *gorm.DB) error {
	// Validate entity type
	validEntityTypes := []string{"article", "category", "tag", "menu", "notification"}
	isValidEntityType := false
	for _, validType := range validEntityTypes {
		if tq.EntityType == validType {
			isValidEntityType = true
			break
		}
	}
	if !isValidEntityType {
		return fmt.Errorf("invalid entity type: %s. Supported types: %v", tq.EntityType, validEntityTypes)
	}

	// Validate languages
	supportedLanguages := []string{"tr", "en", "es", "fr", "de", "ar"}
	isValidSourceLang := false
	isValidTargetLang := false

	for _, lang := range supportedLanguages {
		if tq.SourceLang == lang {
			isValidSourceLang = true
		}
		if tq.TargetLang == lang {
			isValidTargetLang = true
		}
	}

	if !isValidSourceLang {
		return fmt.Errorf("invalid source language: %s. Supported languages: %v", tq.SourceLang, supportedLanguages)
	}

	if !isValidTargetLang {
		return fmt.Errorf("invalid target language: %s. Supported languages: %v", tq.TargetLang, supportedLanguages)
	}

	// Validate that source and target languages are different
	if tq.SourceLang == tq.TargetLang {
		return fmt.Errorf("source and target language cannot be the same: %s", tq.SourceLang)
	}

	// Set default source language if empty
	if tq.SourceLang == "" {
		tq.SourceLang = "tr"
	}

	// Set default priority if not set
	if tq.Priority == 0 {
		tq.Priority = 1
	}

	return nil
}
