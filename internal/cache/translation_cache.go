package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// TranslationCacheImpl handles caching of translation data
type TranslationCacheImpl struct {
	redis      *redis.Client
	db         *gorm.DB
	localCache sync.Map // In-memory cache for frequently accessed translations
}

// CacheEntry represents a cached translation entry
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	Language  string      `json:"language"`
	EntityID  uint        `json:"entity_id,omitempty"`
}

// NewTranslationCache creates a new translation cache
func NewTranslationCache(redisClient *redis.Client) TranslationCache {
	return &TranslationCacheImpl{
		redis: redisClient,
		db:    database.DB,
	}
}

// Cache keys
const (
	CacheKeyUITranslation       = "ui_translation:%s:%s"       // language:key
	CacheKeyArticleTranslation  = "article_translation:%d:%s"  // id:language
	CacheKeyCategoryTranslation = "category_translation:%d:%s" // id:language
	CacheKeyTagTranslation      = "tag_translation:%d:%s"      // id:language
	CacheKeyPageTranslation     = "page_translation:%d:%s"     // id:language
	CacheKeyMenuTranslation     = "menu_translation:%d:%s"     // id:language
	CacheKeySEOTranslation      = "seo_translation:%s:%d:%s"   // type:id:language
	CacheKeyCommentTranslation  = "comment_translation:%d:%s"  // id:language
	CacheKeyFormTranslation     = "form_translation:%s:%s:%s"  // form:field:language
	CacheKeyErrorTranslation    = "error_translation:%s:%s"    // code:language
	CacheKeyEmailTranslation    = "email_translation:%s:%s"    // template:language
)

// Cache expiration times
var (
	DefaultCacheExpiration       = 1 * time.Hour
	UITranslationExpiration      = 24 * time.Hour // UI translations change less frequently
	ContentTranslationExpiration = 2 * time.Hour
	SystemTranslationExpiration  = 12 * time.Hour
)

// GetUITranslation retrieves UI translation from cache or database
func (tc *TranslationCacheImpl) GetUITranslation(language, key string) (string, error) {
	cacheKey := fmt.Sprintf(CacheKeyUITranslation, language, key)

	// Try Redis first
	if tc.redis != nil {
		val, err := tc.redis.Get(context.Background(), cacheKey).Result()
		if err == nil {
			var entry CacheEntry
			if json.Unmarshal([]byte(val), &entry) == nil && time.Now().Before(entry.ExpiresAt) {
				if str, ok := entry.Value.(string); ok {
					return str, nil
				}
			}
		}
	}

	// Try local cache
	if val, ok := tc.localCache.Load(cacheKey); ok {
		if entry, ok := val.(CacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
			if str, ok := entry.Value.(string); ok {
				return str, nil
			}
		}
	}

	// Get from database
	var translation models.Translation
	if err := tc.db.Where("key = ? AND language = ? AND is_active = ?", key, language, true).
		First(&translation).Error; err != nil {
		return "", err
	}

	// Cache the result
	tc.cacheValue(cacheKey, translation.Value, UITranslationExpiration)

	return translation.Value, nil
}

// GetArticleTranslation retrieves article translation from cache or database
func (tc *TranslationCacheImpl) GetArticleTranslation(articleID uint, language string) (*models.LocalizedArticle, error) {
	cacheKey := fmt.Sprintf(CacheKeyArticleTranslation, articleID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if article, ok := cached.(*models.LocalizedArticle); ok {
			return article, nil
		}
	}

	// Get from database
	var article models.Article
	if err := tc.db.Preload("Author").Preload("Category").Preload("Tags").
		First(&article, articleID).Error; err != nil {
		return nil, err
	}

	var translation models.ArticleTranslation
	localized := &models.LocalizedArticle{
		ID:              article.ID,
		Title:           article.Title,
		Slug:            article.Slug,
		Content:         article.Content,
		Summary:         article.Summary,
		MetaTitle:       article.MetaTitle,
		MetaDescription: article.MetaDesc,
		Language:        language,
		Status:          article.Status,
		PublishedAt:     article.PublishedAt,
		CreatedAt:       article.CreatedAt,
		UpdatedAt:       article.UpdatedAt,
	}

	// If not English, get translated version
	if language != "en" {
		if err := tc.db.Where("article_id = ? AND language = ? AND is_active = ?",
			articleID, language, true).First(&translation).Error; err == nil {
			localized.Title = translation.Title
			localized.Slug = translation.Slug
			localized.Content = translation.Content
			localized.Summary = translation.Summary
			localized.MetaTitle = translation.MetaTitle
			localized.MetaDescription = translation.MetaDescription
		}
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, ContentTranslationExpiration)

	return localized, nil
}

// GetCategoryTranslation retrieves category translation from cache or database
func (tc *TranslationCacheImpl) GetCategoryTranslation(categoryID uint, language string) (*models.LocalizedCategory, error) {
	cacheKey := fmt.Sprintf(CacheKeyCategoryTranslation, categoryID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if category, ok := cached.(*models.LocalizedCategory); ok {
			return category, nil
		}
	}

	// Get from database
	var category models.Category
	if err := tc.db.First(&category, categoryID).Error; err != nil {
		return nil, err
	}

	var translation models.CategoryTranslation
	localized := &models.LocalizedCategory{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Color:       category.Color,
		Language:    language,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}

	// If not the default language, get translated version
	if language != "tr" {
		if err := tc.db.Where("category_id = ? AND language = ? AND is_active = ?",
			categoryID, language, true).First(&translation).Error; err == nil {
			localized.Name = translation.Name
			localized.Slug = translation.Slug
			localized.Description = translation.Description
			localized.MetaTitle = translation.MetaTitle
			localized.MetaDesc = translation.MetaDesc
		}
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, ContentTranslationExpiration)

	return localized, nil
}

// GetTagTranslation retrieves tag translation from cache or database
func (tc *TranslationCacheImpl) GetTagTranslation(tagID uint, language string) (*models.LocalizedTag, error) {
	cacheKey := fmt.Sprintf(CacheKeyTagTranslation, tagID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if tag, ok := cached.(*models.LocalizedTag); ok {
			return tag, nil
		}
	}

	// Get from database
	var tag models.Tag
	if err := tc.db.First(&tag, tagID).Error; err != nil {
		return nil, err
	}

	var translation models.TagTranslation
	localized := &models.LocalizedTag{
		ID:          tag.ID,
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		UsageCount:  tag.UsageCount,
		Language:    language,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}

	// If not the default language, get translated version
	if language != "tr" {
		if err := tc.db.Where("tag_id = ? AND language = ? AND is_active = ?",
			tagID, language, true).First(&translation).Error; err == nil {
			localized.Name = translation.Name
			localized.Slug = translation.Slug
			localized.Description = translation.Description
		}
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, ContentTranslationExpiration)

	return localized, nil
}

// GetPageTranslation retrieves page translation from cache or database
func (tc *TranslationCacheImpl) GetPageTranslation(pageID uint, language string) (*models.LocalizedPage, error) {
	cacheKey := fmt.Sprintf(CacheKeyPageTranslation, pageID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if page, ok := cached.(*models.LocalizedPage); ok {
			return page, nil
		}
	}

	// Get from database
	var page models.Page
	if err := tc.db.First(&page, pageID).Error; err != nil {
		return nil, err
	}

	var translation models.PageTranslation
	localized := &models.LocalizedPage{
		ID:              page.ID,
		Title:           page.Title,
		Slug:            page.Slug,
		MetaTitle:       page.MetaTitle,
		MetaDescription: page.MetaDesc,
		ExcerptText:     page.ExcerptText,
		Template:        page.Template,
		Layout:          page.Layout,
		Status:          page.Status,
		FeaturedImage:   page.FeaturedImage,
		Language:        language,
		IsHomepage:      page.IsHomepage,
		IsLandingPage:   page.IsLandingPage,
		Views:           page.Views,
		PublishedAt:     page.PublishedAt,
		CreatedAt:       page.CreatedAt,
		UpdatedAt:       page.UpdatedAt,
	}

	// If not the default language, get translated version
	if language != "tr" {
		if err := tc.db.Where("page_id = ? AND language = ? AND is_active = ?",
			pageID, language, true).First(&translation).Error; err == nil {
			localized.Title = translation.Title
			localized.Slug = translation.Slug
			localized.MetaTitle = translation.MetaTitle
			localized.MetaDescription = translation.MetaDesc
			localized.ExcerptText = translation.ExcerptText
		}
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, ContentTranslationExpiration)

	return localized, nil
}

// GetMenuTranslation retrieves menu translation from cache or database
func (tc *TranslationCacheImpl) GetMenuTranslation(menuID uint, language string) (*models.LocalizedMenu, error) {
	cacheKey := fmt.Sprintf(CacheKeyMenuTranslation, menuID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if menu, ok := cached.(*models.LocalizedMenu); ok {
			return menu, nil
		}
	}

	// Get from database
	var menu models.Menu
	if err := tc.db.Preload("Items").First(&menu, menuID).Error; err != nil {
		return nil, err
	}

	var translation models.MenuTranslation
	localized := &models.LocalizedMenu{
		ID:        menu.ID,
		Name:      menu.Name,
		Slug:      menu.Slug,
		Location:  menu.Location,
		Language:  language,
		CreatedAt: menu.CreatedAt,
		UpdatedAt: menu.UpdatedAt,
	}

	// If not the default language, get translated version
	if language != "tr" {
		if err := tc.db.Where("menu_id = ? AND language = ? AND is_active = ?",
			menuID, language, true).First(&translation).Error; err == nil {
			localized.Name = translation.Name
			if translation.Description != nil {
				localized.Description = translation.Description
			}
		}
	}

	// Get localized menu items
	for _, item := range menu.Items {
		var itemTranslation models.MenuItemTranslation
		localizedItem := models.LocalizedMenuItem{
			ID:        item.ID,
			MenuID:    item.MenuID,
			ParentID:  item.ParentID,
			Title:     item.Title,
			URL:       item.URL,
			Icon:      item.Icon,
			Target:    item.Target,
			SortOrder: item.SortOrder,
			Language:  language,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}

		if language != "tr" {
			if err := tc.db.Where("menu_item_id = ? AND language = ? AND is_active = ?",
				item.ID, language, true).First(&itemTranslation).Error; err == nil {
				localizedItem.Title = itemTranslation.Title
				localizedItem.URL = itemTranslation.URL
			}
		}

		localized.Items = append(localized.Items, localizedItem)
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, ContentTranslationExpiration)

	return localized, nil
}

// GetSEOTranslation retrieves SEO translation from cache or database
func (tc *TranslationCacheImpl) GetSEOTranslation(entityType string, entityID uint, language string) (*models.LocalizedSEOSettings, error) {
	cacheKey := fmt.Sprintf(CacheKeySEOTranslation, entityType, entityID, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if seo, ok := cached.(*models.LocalizedSEOSettings); ok {
			return seo, nil
		}
	}

	// Get original SEO settings
	var originalSettings models.PageSEOSettings
	if entityType == "page" {
		var page models.Page
		if err := tc.db.First(&page, entityID).Error; err != nil {
			return nil, err
		}
		originalSettings = page.GetSEOSettings()
	}

	localized := &models.LocalizedSEOSettings{
		Keywords:           originalSettings.Keywords,
		CanonicalURL:       originalSettings.CanonicalURL,
		RobotsIndex:        originalSettings.RobotsIndex,
		RobotsFollow:       originalSettings.RobotsFollow,
		OGTitle:            originalSettings.OGTitle,
		OGDescription:      originalSettings.OGDescription,
		OGImage:            originalSettings.OGImage,
		TwitterCard:        originalSettings.TwitterCard,
		TwitterTitle:       originalSettings.TwitterTitle,
		TwitterDescription: originalSettings.TwitterDescription,
		TwitterImage:       originalSettings.TwitterImage,
		Schema:             originalSettings.Schema,
	}

	// If not English, get translated version
	if language != "en" {
		var seoTranslation models.SEOTranslation
		if err := tc.db.Where("entity_id = ? AND entity_type = ? AND language = ?",
			entityID, entityType, language).First(&seoTranslation).Error; err == nil {

			if keywords := seoTranslation.GetKeywords(); len(keywords) > 0 {
				localized.Keywords = keywords
			}
			if seoTranslation.OGTitle != "" {
				localized.OGTitle = seoTranslation.OGTitle
			}
			if seoTranslation.OGDescription != "" {
				localized.OGDescription = seoTranslation.OGDescription
			}
			if seoTranslation.TwitterTitle != "" {
				localized.TwitterTitle = seoTranslation.TwitterTitle
			}
			if seoTranslation.TwitterDescription != "" {
				localized.TwitterDescription = seoTranslation.TwitterDescription
			}
		}
	}

	// Cache the result
	tc.cacheValue(cacheKey, localized, SystemTranslationExpiration)

	return localized, nil
}

// GetFormTranslation retrieves form field translation from cache or database
func (tc *TranslationCacheImpl) GetFormTranslation(formKey, fieldKey, language string) (*models.FormTranslation, error) {
	cacheKey := fmt.Sprintf(CacheKeyFormTranslation, formKey, fieldKey, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if form, ok := cached.(*models.FormTranslation); ok {
			return form, nil
		}
	}

	// Get from database
	var translation models.FormTranslation
	if err := tc.db.Where("form_key = ? AND field_key = ? AND language = ? AND is_active = ?",
		formKey, fieldKey, language, true).First(&translation).Error; err != nil {
		return nil, err
	}

	// Cache the result
	tc.cacheValue(cacheKey, &translation, SystemTranslationExpiration)

	return &translation, nil
}

// GetErrorTranslation retrieves error message translation from cache or database
func (tc *TranslationCacheImpl) GetErrorTranslation(errorCode, language string) (*models.ErrorMessageTranslation, error) {
	cacheKey := fmt.Sprintf(CacheKeyErrorTranslation, errorCode, language)

	// Try cache first
	if cached, err := tc.getCachedValue(cacheKey); err == nil {
		if error, ok := cached.(*models.ErrorMessageTranslation); ok {
			return error, nil
		}
	}

	// Get from database
	var translation models.ErrorMessageTranslation
	if err := tc.db.Where("error_code = ? AND language = ? AND is_active = ?",
		errorCode, language, true).First(&translation).Error; err != nil {
		return nil, err
	}

	// Cache the result
	tc.cacheValue(cacheKey, &translation, SystemTranslationExpiration)

	return &translation, nil
}

// cacheValue stores a value in both Redis and local cache
func (tc *TranslationCacheImpl) cacheValue(key string, value interface{}, duration time.Duration) {
	entry := CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(duration),
	}

	// Store in local cache
	tc.localCache.Store(key, entry)

	// Store in Redis if available
	if tc.redis != nil {
		data, err := json.Marshal(entry)
		if err == nil {
			tc.redis.Set(context.Background(), key, data, duration)
		}
	}
}

// getCachedValue retrieves a value from cache
func (tc *TranslationCacheImpl) getCachedValue(key string) (interface{}, error) {
	// Try Redis first
	if tc.redis != nil {
		val, err := tc.redis.Get(context.Background(), key).Result()
		if err == nil {
			var entry CacheEntry
			if json.Unmarshal([]byte(val), &entry) == nil && time.Now().Before(entry.ExpiresAt) {
				return entry.Value, nil
			}
		}
	}

	// Try local cache
	if val, ok := tc.localCache.Load(key); ok {
		if entry, ok := val.(CacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
			return entry.Value, nil
		}
	}

	return nil, fmt.Errorf("not found in cache")
}

// InvalidateCache removes cached translations
func (tc *TranslationCacheImpl) InvalidateCache(pattern string) error {
	// Clear local cache entries matching pattern
	tc.localCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.Contains(keyStr, pattern) {
				tc.localCache.Delete(key)
			}
		}
		return true
	})

	// Clear Redis cache if available
	if tc.redis != nil {
		keys, err := tc.redis.Keys(context.Background(), pattern+"*").Result()
		if err == nil && len(keys) > 0 {
			tc.redis.Del(context.Background(), keys...)
		}
	}

	return nil
}

// InvalidateEntityTranslations invalidates all translations for a specific entity
func (tc *TranslationCacheImpl) InvalidateEntityTranslations(entityType string, entityID uint) error {
	patterns := []string{
		fmt.Sprintf("%s_translation:%d", entityType, entityID),
		fmt.Sprintf("seo_translation:%s:%d", entityType, entityID),
	}

	for _, pattern := range patterns {
		if err := tc.InvalidateCache(pattern); err != nil {
			log.Printf("Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}

	return nil
}

// InvalidateTranslationsByType invalidates all cached translations for a specific type
func (tc *TranslationCacheImpl) InvalidateTranslationsByType(translationType string) error {
	var pattern string

	switch translationType {
	case "article":
		pattern = "article_translation:"
	case "category":
		pattern = "category_translation:"
	case "tag":
		pattern = "tag_translation:"
	case "page":
		pattern = "page_translation:"
	case "menu":
		pattern = "menu_translation:"
	case "ui":
		pattern = "ui_translation:"
	case "seo":
		pattern = "seo_translation:"
	case "form":
		pattern = "form_translation:"
	case "error":
		pattern = "error_translation:"
	default:
		return fmt.Errorf("unknown translation type: %s", translationType)
	}

	return tc.InvalidateCache(pattern)
}

// InvalidateLanguageTranslations invalidates all cached translations for a specific language
func (tc *TranslationCacheImpl) InvalidateLanguageTranslations(language string) error {
	// Clear local cache entries for the language
	tc.localCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.Contains(keyStr, ":"+language) {
				tc.localCache.Delete(key)
			}
		}
		return true
	})

	// Clear Redis cache if available
	if tc.redis != nil {
		keys, err := tc.redis.Keys(context.Background(), "*:"+language).Result()
		if err == nil && len(keys) > 0 {
			tc.redis.Del(context.Background(), keys...)
		}
	}

	return nil
}

// WarmupCache preloads frequently accessed translations
func (tc *TranslationCacheImpl) WarmupCache(languages []string) error {
	log.Println("Starting translation cache warmup...")

	// Warmup UI translations
	go tc.warmupUITranslations(languages)

	// Warmup recent articles
	go tc.warmupRecentContent(languages)

	// Warmup form and error translations
	go tc.warmupSystemTranslations(languages)

	log.Println("Translation cache warmup initiated")
	return nil
}

// warmupUITranslations preloads UI translations
func (tc *TranslationCacheImpl) warmupUITranslations(languages []string) {
	for _, lang := range languages {
		var translations []models.Translation
		tc.db.Where("language = ? AND is_active = ?", lang, true).
			Limit(100).Find(&translations)

		for _, t := range translations {
			cacheKey := fmt.Sprintf(CacheKeyUITranslation, lang, t.Key)
			tc.cacheValue(cacheKey, t.Value, UITranslationExpiration)
		}
	}
}

// warmupRecentContent preloads recent article translations
func (tc *TranslationCacheImpl) warmupRecentContent(languages []string) {
	var recentArticles []models.Article
	tc.db.Where("status = ? AND published_at IS NOT NULL", "published").
		Order("published_at DESC").Limit(20).Find(&recentArticles)

	for _, article := range recentArticles {
		for _, lang := range languages {
			if _, err := tc.GetArticleTranslation(article.ID, lang); err != nil {
				log.Printf("Failed to warmup article %d translation for %s: %v", article.ID, lang, err)
			}
		}
	}

	// Warmup popular categories
	var categories []models.Category
	tc.db.Where("is_active = ?", true).Limit(10).Find(&categories)
	for _, category := range categories {
		for _, lang := range languages {
			if _, err := tc.GetCategoryTranslation(category.ID, lang); err != nil {
				log.Printf("Failed to warmup category %d translation for %s: %v", category.ID, lang, err)
			}
		}
	}

	// Warmup popular tags
	var tags []models.Tag
	tc.db.Order("usage_count DESC").Limit(15).Find(&tags)
	for _, tag := range tags {
		for _, lang := range languages {
			if _, err := tc.GetTagTranslation(tag.ID, lang); err != nil {
				log.Printf("Failed to warmup tag %d translation for %s: %v", tag.ID, lang, err)
			}
		}
	}

	// Warmup active menus
	var menus []models.Menu
	tc.db.Where("is_active = ?", true).Find(&menus)
	for _, menu := range menus {
		for _, lang := range languages {
			if _, err := tc.GetMenuTranslation(menu.ID, lang); err != nil {
				log.Printf("Failed to warmup menu %d translation for %s: %v", menu.ID, lang, err)
			}
		}
	}

	// Warmup important pages
	var pages []models.Page
	tc.db.Where("status = ? AND (is_homepage = ? OR is_landing_page = ?)", "published", true, true).
		Limit(10).Find(&pages)
	for _, page := range pages {
		for _, lang := range languages {
			if _, err := tc.GetPageTranslation(page.ID, lang); err != nil {
				log.Printf("Failed to warmup page %d translation for %s: %v", page.ID, lang, err)
			}
		}
	}
}

// warmupSystemTranslations preloads system translations
func (tc *TranslationCacheImpl) warmupSystemTranslations(languages []string) {
	for _, lang := range languages {
		// Warmup form translations
		var formTranslations []models.FormTranslation
		tc.db.Where("language = ? AND is_active = ?", lang, true).Find(&formTranslations)

		for _, ft := range formTranslations {
			cacheKey := fmt.Sprintf(CacheKeyFormTranslation, ft.FormKey, ft.FieldKey, lang)
			tc.cacheValue(cacheKey, &ft, SystemTranslationExpiration)
		}

		// Warmup error translations
		var errorTranslations []models.ErrorMessageTranslation
		tc.db.Where("language = ? AND is_active = ?", lang, true).Find(&errorTranslations)

		for _, et := range errorTranslations {
			cacheKey := fmt.Sprintf(CacheKeyErrorTranslation, et.ErrorCode, lang)
			tc.cacheValue(cacheKey, &et, SystemTranslationExpiration)
		}
	}
}

// GetCacheStats returns cache statistics
func (tc *TranslationCacheImpl) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Count local cache entries
	localCount := 0
	tc.localCache.Range(func(key, value interface{}) bool {
		localCount++
		return true
	})

	stats["local_cache_entries"] = localCount

	// Redis stats if available
	if tc.redis != nil {
		if info, err := tc.redis.Info(context.Background(), "memory").Result(); err == nil {
			stats["redis_info"] = info
		}
	}

	return stats
}
