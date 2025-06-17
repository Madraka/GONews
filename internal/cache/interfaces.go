package cache

import (
	"news/internal/models"
)

// TranslationCache defines translation-specific cache operations
type TranslationCache interface {
	// UI Translation methods
	GetUITranslation(language, key string) (string, error)
	
	// Content Translation methods
	GetArticleTranslation(articleID uint, language string) (*models.LocalizedArticle, error)
	GetCategoryTranslation(categoryID uint, language string) (*models.LocalizedCategory, error)
	GetTagTranslation(tagID uint, language string) (*models.LocalizedTag, error)
	GetPageTranslation(pageID uint, language string) (*models.LocalizedPage, error)
	GetMenuTranslation(menuID uint, language string) (*models.LocalizedMenu, error)
	
	// System Translation methods
	GetSEOTranslation(entityType string, entityID uint, language string) (*models.LocalizedSEOSettings, error)
	GetFormTranslation(formKey, fieldKey, language string) (*models.FormTranslation, error)
	GetErrorTranslation(errorCode, language string) (*models.ErrorMessageTranslation, error)
	
	// Cache management methods
	InvalidateCache(pattern string) error
	InvalidateEntityTranslations(entityType string, entityID uint) error
	InvalidateTranslationsByType(translationType string) error
	InvalidateLanguageTranslations(language string) error
	
	// Warmup methods
	WarmupCache(languages []string) error
	
	// Stats
	GetCacheStats() map[string]interface{}
}
