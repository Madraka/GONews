package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
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

// SEOTranslation represents translations for SEO elements
type SEOTranslation struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	EntityID           uint           `gorm:"not null;index:idx_seo_translation_entity" json:"entity_id"`
	EntityType         string         `gorm:"size:50;not null;index:idx_seo_translation_entity" json:"entity_type"` // page, article
	Language           string         `gorm:"size:5;not null;index:idx_seo_translation_entity" json:"language"`
	Keywords           datatypes.JSON `gorm:"type:json" json:"keywords"` // Translated keywords array
	OGTitle            string         `gorm:"size:255" json:"og_title,omitempty"`
	OGDescription      string         `gorm:"type:text" json:"og_description,omitempty"`
	TwitterTitle       string         `gorm:"size:255" json:"twitter_title,omitempty"`
	TwitterDescription string         `gorm:"type:text" json:"twitter_description,omitempty"`
	Schema             string         `gorm:"type:text" json:"schema,omitempty"` // JSON-LD structured data
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for SEOTranslation
func (SEOTranslation) TableName() string {
	return "seo_translations"
}

// LocalizedSEOSettings represents SEO settings with translations
type LocalizedSEOSettings struct {
	Keywords           []string `json:"keywords,omitempty"`
	CanonicalURL       string   `json:"canonical_url,omitempty"`
	RobotsIndex        bool     `json:"robots_index"`
	RobotsFollow       bool     `json:"robots_follow"`
	OGTitle            string   `json:"og_title,omitempty"`
	OGDescription      string   `json:"og_description,omitempty"`
	OGImage            string   `json:"og_image,omitempty"`
	TwitterCard        string   `json:"twitter_card,omitempty"`
	TwitterTitle       string   `json:"twitter_title,omitempty"`
	TwitterDescription string   `json:"twitter_description,omitempty"`
	TwitterImage       string   `json:"twitter_image,omitempty"`
	Schema             string   `json:"schema,omitempty"`
}

// GetKeywords unmarshals and returns translated keywords
func (s *SEOTranslation) GetKeywords() []string {
	var keywords []string
	if len(s.Keywords) > 0 {
		if err := json.Unmarshal(s.Keywords, &keywords); err != nil {
			return []string{}
		}
	}
	return keywords
}

// SetKeywords marshals and sets keywords
func (s *SEOTranslation) SetKeywords(keywords []string) error {
	data, err := json.Marshal(keywords)
	if err != nil {
		return err
	}
	s.Keywords = data
	return nil
}

// TranslatedMetaData represents structured metadata for content blocks
type TranslatedMetaData struct {
	AltText     string `json:"alt_text,omitempty"`
	Caption     string `json:"caption,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Credits     string `json:"credits,omitempty"`
	Link        string `json:"link,omitempty"`
	LinkText    string `json:"link_text,omitempty"`
	ButtonText  string `json:"button_text,omitempty"`
}
