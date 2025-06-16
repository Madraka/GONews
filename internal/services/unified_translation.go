package services

import (
	"database/sql"
	"fmt"
	"log"

	"news/internal/config"
	"news/internal/database"
	"news/internal/json"
	"news/internal/models"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

// UnifiedTranslationService combines all translation functionality
type UnifiedTranslationService struct {
	// Core components
	config    *config.TranslationConfig
	bundle    *i18n.Bundle
	db        *gorm.DB
	aiService *AIService

	// Internal services
	uiTranslator      *UITranslator
	contentTranslator *ContentTranslator
	aiTranslator      *AITranslationService

	// Cache and performance
	cache map[string]interface{}
}

// UITranslator handles static UI translations from JSON files
type UITranslator struct {
	bundle *i18n.Bundle
	config *config.TranslationConfig
}

// ContentTranslator handles dynamic content translations from database
type ContentTranslator struct {
	db     *gorm.DB
	config *config.TranslationConfig
}

// NewUnifiedTranslationService creates a new unified translation service
func NewUnifiedTranslationService(db *sql.DB, localesPath string) (*UnifiedTranslationService, error) {
	gormDB := database.DB
	translationConfig := config.GetTranslationConfig()

	// Initialize UI translator (JSON-based)
	uiTranslator, err := NewUITranslator(localesPath, translationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize UI translator: %w", err)
	}

	// Initialize content translator (DB-based)
	contentTranslator := NewContentTranslator(gormDB, translationConfig)

	// Initialize AI translator
	aiService := GetAIService()
	var aiTranslator *AITranslationService
	if aiService != nil && translationConfig.EnableAITranslation {
		aiTranslator = NewAITranslationService(aiService)
	}

	service := &UnifiedTranslationService{
		config:            translationConfig,
		bundle:            uiTranslator.bundle,
		db:                gormDB,
		aiService:         aiService,
		uiTranslator:      uiTranslator,
		contentTranslator: contentTranslator,
		aiTranslator:      aiTranslator,
		cache:             make(map[string]interface{}),
	}

	return service, nil
}

// NewUITranslator creates a new UI translator
func NewUITranslator(localesPath string, config *config.TranslationConfig) (*UITranslator, error) {
	// Create bundle for default language
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load translation files for all supported languages
	for _, lang := range config.SupportedLanguages {
		filename := fmt.Sprintf("%s/%s.json", localesPath, lang)
		if _, err := bundle.LoadMessageFile(filename); err != nil {
			log.Printf("Warning: Failed to load UI translation file %s: %v", filename, err)
		}
	}

	return &UITranslator{
		bundle: bundle,
		config: config,
	}, nil
}

// NewContentTranslator creates a new content translator
func NewContentTranslator(db *gorm.DB, config *config.TranslationConfig) *ContentTranslator {
	return &ContentTranslator{
		db:     db,
		config: config,
	}
}

// GetBundle returns the i18n bundle for UI translations
func (uts *UnifiedTranslationService) GetBundle() *i18n.Bundle {
	return uts.bundle
}

// GetLocalizer creates a localizer for UI translations
func (uts *UnifiedTranslationService) GetLocalizer(language string) *i18n.Localizer {
	if !uts.config.ValidateLanguage(language) {
		language = uts.config.DefaultLanguage
	}
	return i18n.NewLocalizer(uts.bundle, language)
}

// TranslateUI translates UI elements using JSON-based translations
func (uts *UnifiedTranslationService) TranslateUI(language, messageID string, templateData map[string]interface{}) (string, error) {
	localizer := uts.GetLocalizer(language)

	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	message, err := localizer.Localize(config)
	if err != nil {
		// Fallback to default language
		if language != uts.config.DefaultLanguage {
			return uts.TranslateUI(uts.config.DefaultLanguage, messageID, templateData)
		}
		return messageID, err
	}

	return message, nil
}

// Localize translates a message for compatibility with pubsub.TranslationService interface
func (uts *UnifiedTranslationService) Localize(language, messageID string, templateData map[string]interface{}) (string, error) {
	return uts.TranslateUI(language, messageID, templateData)
}

// TranslateContent translates dynamic content using database translations
func (uts *UnifiedTranslationService) TranslateContent(entityType string, entityID uint, language string) (interface{}, error) {
	return uts.contentTranslator.TranslateEntity(entityType, entityID, language)
}

// TranslateEntity translates an entity to multiple target languages using AI
func (uts *UnifiedTranslationService) TranslateEntity(entityType string, entityID uint, targetLanguages []string) error {
	if uts.aiTranslator == nil {
		return fmt.Errorf("AI translation service not available")
	}

	switch entityType {
	case "article":
		return uts.aiTranslator.TranslateArticle(entityID, targetLanguages)
	case "category":
		return uts.aiTranslator.TranslateCategory(entityID, targetLanguages)
	case "tag":
		return uts.aiTranslator.TranslateTag(entityID, targetLanguages)
	case "menu":
		return uts.aiTranslator.TranslateMenu(entityID, targetLanguages)
	case "notification":
		return uts.aiTranslator.TranslateNotification(entityID, targetLanguages)
	case "page":
		return uts.aiTranslator.TranslatePage(entityID, targetLanguages)
	case "page-content-block":
		return uts.aiTranslator.TranslatePageContentBlock(entityID, targetLanguages)
	case "article-content-block":
		return uts.aiTranslator.TranslateArticleContentBlock(entityID, targetLanguages)
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// IsTranslationAvailable checks if a translation exists for the given entity and language
func (uts *UnifiedTranslationService) IsTranslationAvailable(entityType string, entityID uint, language string) bool {
	switch entityType {
	case "article":
		var translation models.ArticleTranslation
		err := uts.db.Where("article_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "category":
		var translation models.CategoryTranslation
		err := uts.db.Where("category_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "tag":
		var translation models.TagTranslation
		err := uts.db.Where("tag_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "menu":
		var translation models.MenuTranslation
		err := uts.db.Where("menu_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "notification":
		var translation models.NotificationTranslation
		err := uts.db.Where("notification_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "page":
		var translation models.PageTranslation
		err := uts.db.Where("page_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "page-content-block":
		var translation models.PageContentBlockTranslation
		err := uts.db.Where("block_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	case "article-content-block":
		var translation models.ArticleContentBlockTranslation
		err := uts.db.Where("block_id = ? AND language = ?", entityID, language).First(&translation).Error
		return err == nil
	}
	return false
}

// GetConfig returns the translation configuration
func (uts *UnifiedTranslationService) GetConfig() *config.TranslationConfig {
	return uts.config
}

// RequestAITranslation requests AI translation for content
func (uts *UnifiedTranslationService) RequestAITranslation(
	entityType string,
	entityID uint,
	sourceLanguage string,
	targetLanguages []string,
	priority int,
) (string, error) {
	if uts.aiTranslator == nil {
		return "", fmt.Errorf("AI translation service not available")
	}

	// Set default source language if not provided
	if sourceLanguage == "" {
		sourceLanguage = uts.config.DefaultLanguage
	}

	// For now, we'll create a simple job ID and delegate to the AI translator
	// In a real implementation, this would return a proper job queue ID
	var err error
	switch entityType {
	case "article":
		err = uts.aiTranslator.TranslateArticle(entityID, targetLanguages)
	case "category":
		err = uts.aiTranslator.TranslateCategory(entityID, targetLanguages)
	case "tag":
		err = uts.aiTranslator.TranslateTag(entityID, targetLanguages)
	case "menu":
		err = uts.aiTranslator.TranslateMenu(entityID, targetLanguages)
	case "notification":
		err = uts.aiTranslator.TranslateNotification(entityID, targetLanguages)
	case "page":
		err = uts.aiTranslator.TranslatePage(entityID, targetLanguages)
	case "page-content-block":
		err = uts.aiTranslator.TranslatePageContentBlock(entityID, targetLanguages)
	case "article-content-block":
		err = uts.aiTranslator.TranslateArticleContentBlock(entityID, targetLanguages)
	default:
		return "", fmt.Errorf("unsupported entity type: %s", entityType)
	}

	if err != nil {
		return "", fmt.Errorf("failed to queue AI translation: %w", err)
	}

	// Generate a job ID (in production, this would come from the queue system)
	jobID := fmt.Sprintf("%s-%d-%d", entityType, entityID, len(targetLanguages))
	return jobID, nil
}

// GetTranslationJobStatus gets the status of a translation job
func (uts *UnifiedTranslationService) GetTranslationJobStatus(jobID string) (*models.TranslationJobResponse, error) {
	if uts.aiTranslator == nil {
		return nil, fmt.Errorf("AI translation service not available")
	}

	// For now, return a basic status
	// In a real implementation, this would query the actual job queue
	response := &models.TranslationJobResponse{
		JobID:   jobID,
		Status:  "processing",
		Message: "Translation job is being processed",
	}

	return response, nil
}

// ContentTranslator methods

// TranslateEntity translates a single entity to the specified language
func (ct *ContentTranslator) TranslateEntity(entityType string, entityID uint, language string) (interface{}, error) {
	switch entityType {
	case "article":
		return ct.translateArticle(entityID, language)
	case "category":
		return ct.translateCategory(entityID, language)
	case "tag":
		return ct.translateTag(entityID, language)
	case "menu":
		return ct.translateMenu(entityID, language)
	case "page":
		return ct.translatePage(entityID, language)
	case "page-content-block":
		return ct.translatePageContentBlock(entityID, language)
	case "article-content-block":
		return ct.translateArticleContentBlock(entityID, language)
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

func (ct *ContentTranslator) translateArticle(articleID uint, language string) (*models.LocalizedArticle, error) {
	// Get original article with related data to avoid N+1 queries
	var article models.Article
	if err := ct.db.Preload("Author").Preload("Categories").Preload("Tags").First(&article, articleID).Error; err != nil {
		return nil, fmt.Errorf("article not found: %w", err)
	}

	// Try to get translation
	var translation models.ArticleTranslation
	if err := ct.db.Where("article_id = ? AND language = ?", articleID, language).First(&translation).Error; err == nil {
		// Return translated version
		localized := &models.LocalizedArticle{
			ID:              article.ID,
			Title:           translation.Title,
			Slug:            translation.Slug,
			Summary:         translation.Summary,
			Content:         translation.Content,
			MetaTitle:       translation.MetaTitle,
			MetaDescription: article.MetaDesc,
			FeaturedImage:   article.FeaturedImage,
			Status:          article.Status,
			PublishedAt:     article.PublishedAt,
			Language:        language,
			CreatedAt:       article.CreatedAt,
			UpdatedAt:       article.UpdatedAt,
		}
		return localized, nil
	}

	// Fallback to original
	localized := &models.LocalizedArticle{
		ID:              article.ID,
		Title:           article.Title,
		Slug:            article.Slug,
		Summary:         article.Summary,
		Content:         article.Content,
		MetaTitle:       article.MetaTitle,
		MetaDescription: article.MetaDesc,
		FeaturedImage:   article.FeaturedImage,
		Status:          article.Status,
		PublishedAt:     article.PublishedAt,
		Language:        ct.config.DefaultLanguage,
		CreatedAt:       article.CreatedAt,
		UpdatedAt:       article.UpdatedAt,
	}
	return localized, nil
}

func (ct *ContentTranslator) translateCategory(categoryID uint, language string) (*models.LocalizedCategory, error) {
	// Get original category with related data to avoid N+1 queries
	var category models.Category
	if err := ct.db.Preload("Parent").Preload("Children").First(&category, categoryID).Error; err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Try to get translation
	var translation models.CategoryTranslation
	if err := ct.db.Where("category_id = ? AND language = ?", categoryID, language).First(&translation).Error; err == nil {
		// Return translated version
		localized := &models.LocalizedCategory{
			ID:          category.ID,
			Name:        translation.Name,
			Slug:        translation.Slug,
			Description: translation.Description,
			MetaTitle:   translation.MetaTitle,
			MetaDesc:    translation.MetaDesc,
			Color:       category.Color,
			Language:    language,
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		}
		return localized, nil
	}

	// Fallback to original
	localized := &models.LocalizedCategory{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Color:       category.Color,
		Language:    ct.config.DefaultLanguage,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
	return localized, nil
}

func (ct *ContentTranslator) translateTag(tagID uint, language string) (*models.LocalizedTag, error) {
	// Get original tag
	var tag models.Tag
	if err := ct.db.First(&tag, tagID).Error; err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}

	// Try to get translation
	var translation models.TagTranslation
	if err := ct.db.Where("tag_id = ? AND language = ?", tagID, language).First(&translation).Error; err == nil {
		// Return translated version
		localized := &models.LocalizedTag{
			ID:          tag.ID,
			Name:        translation.Name,
			Slug:        translation.Slug,
			Description: translation.Description,
			Color:       tag.Color,
			UsageCount:  tag.UsageCount,
			Language:    language,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		}
		return localized, nil
	}

	// Fallback to original
	localized := &models.LocalizedTag{
		ID:          tag.ID,
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		Color:       tag.Color,
		UsageCount:  tag.UsageCount,
		Language:    ct.config.DefaultLanguage,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
	return localized, nil
}

func (ct *ContentTranslator) translateMenu(menuID uint, language string) (*models.LocalizedMenu, error) {
	// Get original menu
	var menu models.Menu
	if err := ct.db.Preload("Items").First(&menu, menuID).Error; err != nil {
		return nil, fmt.Errorf("menu not found: %w", err)
	}

	// Try to get translation
	var translation models.MenuTranslation
	if err := ct.db.Where("menu_id = ? AND language = ?", menuID, language).First(&translation).Error; err == nil {
		// Return translated version
		localized := &models.LocalizedMenu{
			ID:          menu.ID,
			Name:        translation.Name,
			Slug:        menu.Slug,
			Location:    menu.Location,
			Description: translation.Description,
			Language:    language,
			CreatedAt:   menu.CreatedAt,
			UpdatedAt:   menu.UpdatedAt,
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
				Language:  ct.config.DefaultLanguage,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}

			if err := ct.db.Where("menu_item_id = ? AND language = ?", item.ID, language).
				First(&itemTranslation).Error; err == nil {
				localizedItem.Title = itemTranslation.Title
				localizedItem.URL = itemTranslation.URL
				localizedItem.Language = language
			}

			localized.Items = append(localized.Items, localizedItem)
		}

		return localized, nil
	}

	// Fallback to original
	localized := &models.LocalizedMenu{
		ID:        menu.ID,
		Name:      menu.Name,
		Slug:      menu.Slug,
		Location:  menu.Location,
		Language:  ct.config.DefaultLanguage,
		CreatedAt: menu.CreatedAt,
		UpdatedAt: menu.UpdatedAt,
	}

	for _, item := range menu.Items {
		localizedItem := models.LocalizedMenuItem{
			ID:        item.ID,
			MenuID:    item.MenuID,
			ParentID:  item.ParentID,
			Title:     item.Title,
			URL:       item.URL,
			Icon:      item.Icon,
			Target:    item.Target,
			SortOrder: item.SortOrder,
			Language:  ct.config.DefaultLanguage,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		localized.Items = append(localized.Items, localizedItem)
	}

	return localized, nil
}

func (ct *ContentTranslator) translatePage(pageID uint, language string) (*models.LocalizedPage, error) {
	// Get original page with related data to avoid N+1 queries
	var page models.Page
	if err := ct.db.Preload("Author").Preload("Parent").Preload("Children").Preload("ContentBlocks").First(&page, pageID).Error; err != nil {
		return nil, fmt.Errorf("page not found: %w", err)
	}

	// Try to get translation
	var translation models.PageTranslation
	if err := ct.db.Where("page_id = ? AND language = ?", pageID, language).First(&translation).Error; err == nil {
		// Return translated version
		localized := &models.LocalizedPage{
			ID:              page.ID,
			Title:           translation.Title,
			Slug:            translation.Slug,
			MetaTitle:       translation.MetaTitle,
			MetaDescription: translation.MetaDesc,
			ExcerptText:     translation.ExcerptText,
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
		return localized, nil
	}

	// Fallback to original
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
		Language:        ct.config.DefaultLanguage,
		IsHomepage:      page.IsHomepage,
		IsLandingPage:   page.IsLandingPage,
		Views:           page.Views,
		PublishedAt:     page.PublishedAt,
		CreatedAt:       page.CreatedAt,
		UpdatedAt:       page.UpdatedAt,
	}
	return localized, nil
}

// TranslatePageContentBlock translates page content blocks
func (ct *ContentTranslator) translatePageContentBlock(blockID uint, language string) (interface{}, error) {
	// Get original block with page relation
	var block models.PageContentBlock
	if err := ct.db.Preload("Page").First(&block, blockID).Error; err != nil {
		return nil, fmt.Errorf("page content block not found: %w", err)
	}

	// Try to get translation
	var translation models.PageContentBlockTranslation
	if err := ct.db.Where("block_id = ? AND language = ?", blockID, language).First(&translation).Error; err == nil {
		// Return translated version
		return map[string]interface{}{
			"id":         block.ID,
			"page_id":    block.PageID,
			"block_type": block.BlockType,
			"content":    translation.Content,
			"settings":   block.Settings,
			"position":   block.Position,
			"language":   language,
			"created_at": block.CreatedAt,
			"updated_at": block.UpdatedAt,
		}, nil
	}

	// Fallback to original
	return map[string]interface{}{
		"id":         block.ID,
		"page_id":    block.PageID,
		"block_type": block.BlockType,
		"content":    block.Content,
		"settings":   block.Settings,
		"position":   block.Position,
		"language":   ct.config.DefaultLanguage,
		"created_at": block.CreatedAt,
		"updated_at": block.UpdatedAt,
	}, nil
}

// TranslateArticleContentBlock translates article content blocks
func (ct *ContentTranslator) translateArticleContentBlock(blockID uint, language string) (interface{}, error) {
	// Get original block with article relation
	var block models.ArticleContentBlock
	if err := ct.db.Preload("Article").First(&block, blockID).Error; err != nil {
		return nil, fmt.Errorf("article content block not found: %w", err)
	}

	// Try to get translation
	var translation models.ArticleContentBlockTranslation
	if err := ct.db.Where("block_id = ? AND language = ?", blockID, language).First(&translation).Error; err == nil {
		// Return translated version
		return map[string]interface{}{
			"id":         block.ID,
			"article_id": block.ArticleID,
			"block_type": block.BlockType,
			"content":    translation.Content,
			"settings":   block.Settings,
			"position":   block.Position,
			"language":   language,
			"created_at": block.CreatedAt,
			"updated_at": block.UpdatedAt,
		}, nil
	}

	// Fallback to original
	return map[string]interface{}{
		"id":         block.ID,
		"article_id": block.ArticleID,
		"block_type": block.BlockType,
		"content":    block.Content,
		"settings":   block.Settings,
		"position":   block.Position,
		"language":   ct.config.DefaultLanguage,
		"created_at": block.CreatedAt,
		"updated_at": block.UpdatedAt,
	}, nil
}

// Global service instance
var globalUnifiedTranslationService *UnifiedTranslationService

// InitializeUnifiedTranslationService initializes the global unified translation service
func InitializeUnifiedTranslationService(db *sql.DB, localesPath string) error {
	service, err := NewUnifiedTranslationService(db, localesPath)
	if err != nil {
		return err
	}
	globalUnifiedTranslationService = service
	return nil
}

// GetUnifiedTranslationService returns the global unified translation service
func GetUnifiedTranslationService() *UnifiedTranslationService {
	return globalUnifiedTranslationService
}

// SetGlobalUnifiedTranslationService sets the global unified translation service
func SetGlobalUnifiedTranslationService(service *UnifiedTranslationService) {
	globalUnifiedTranslationService = service
}
