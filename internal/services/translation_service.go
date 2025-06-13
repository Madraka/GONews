package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"news/internal/database"
	"news/internal/models"

	"gorm.io/gorm"
)

// AITranslationService handles automatic translation of content using AI
type AITranslationService struct {
	db        *gorm.DB
	aiService *AIService
	languages []string
}

// NewAITranslationService creates a new AI translation service
func NewAITranslationService(aiService *AIService) *AITranslationService {
	return &AITranslationService{
		db:        database.DB,
		aiService: aiService,
		languages: []string{"en", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"},
	}
}

// TranslationJob represents a translation task
// Using models.TranslationQueue instead of separate TranslationJob struct

// TranslateCategory translates a category to target languages
func (ts *AITranslationService) TranslateCategory(categoryID uint, targetLanguages []string) error {
	var category models.Category
	if err := ts.db.First(&category, categoryID).Error; err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	for _, targetLang := range targetLanguages {
		// Check if translation already exists
		var existing models.CategoryTranslation
		if err := ts.db.Where("category_id = ? AND language = ?", categoryID, targetLang).
			First(&existing).Error; err == nil {
			continue // Translation already exists
		}

		// Translate using AI
		translatedName, err := ts.aiService.TranslateText(context.Background(), category.Name, "tr", targetLang)
		if err != nil {
			log.Printf("Failed to translate category name: %v", err)
			continue
		}

		translatedDesc := ""
		if category.Description != "" {
			translatedDesc, err = ts.aiService.TranslateText(context.Background(), category.Description, "tr", targetLang)
			if err != nil {
				log.Printf("Failed to translate category description: %v", err)
			}
		}

		// Create translation record
		translation := models.CategoryTranslation{
			CategoryID:  categoryID,
			Language:    targetLang,
			Name:        translatedName,
			Description: translatedDesc,
			Slug:        generateSlug(translatedName),
		}

		if err := ts.db.Create(&translation).Error; err != nil {
			log.Printf("Failed to save category translation: %v", err)
		}
	}

	return nil
}

// TranslateTag translates a tag to target languages
func (ts *AITranslationService) TranslateTag(tagID uint, targetLanguages []string) error {
	var tag models.Tag
	if err := ts.db.First(&tag, tagID).Error; err != nil {
		return fmt.Errorf("tag not found: %w", err)
	}

	for _, targetLang := range targetLanguages {
		// Check if translation already exists
		var existing models.TagTranslation
		if err := ts.db.Where("tag_id = ? AND language = ?", tagID, targetLang).
			First(&existing).Error; err == nil {
			continue
		}

		// Translate using AI
		translatedName, err := ts.aiService.TranslateText(context.Background(), tag.Name, "tr", targetLang)
		if err != nil {
			log.Printf("Failed to translate tag: %v", err)
			continue
		}

		// Create translation record
		translation := models.TagTranslation{
			TagID:    tagID,
			Language: targetLang,
			Name:     translatedName,
			Slug:     generateSlug(translatedName),
		}

		if err := ts.db.Create(&translation).Error; err != nil {
			log.Printf("Failed to save tag translation: %v", err)
		}
	}

	return nil
}

// TranslateMenu translates menu and its items
func (ts *AITranslationService) TranslateMenu(menuID uint, targetLanguages []string) error {
	var menu models.Menu
	if err := ts.db.Preload("Items").First(&menu, menuID).Error; err != nil {
		return fmt.Errorf("menu not found: %w", err)
	}

	for _, targetLang := range targetLanguages {
		// Translate menu
		var existingMenu models.MenuTranslation
		err := ts.db.Where("menu_id = ? AND language = ?", menuID, targetLang).
			First(&existingMenu).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}

		if err == gorm.ErrRecordNotFound {
			translatedName, err := ts.aiService.TranslateText(context.Background(), menu.Name, "tr", targetLang)
			if err != nil {
				log.Printf("Failed to translate menu name: %v", err)
				continue
			}

			menuTranslation := models.MenuTranslation{
				MenuID:   menuID,
				Language: targetLang,
				Name:     translatedName,
			}

			ts.db.Create(&menuTranslation)
		}

		// Translate menu items
		for _, item := range menu.Items {
			var existingItem models.MenuItemTranslation
			if err := ts.db.Where("menu_item_id = ? AND language = ?", item.ID, targetLang).
				First(&existingItem).Error; err == nil {
				continue
			}

			translatedTitle, err := ts.aiService.TranslateText(context.Background(), item.Title, "tr", targetLang)
			if err != nil {
				log.Printf("Failed to translate menu item: %v", err)
				continue
			}

			itemTranslation := models.MenuItemTranslation{
				MenuItemID: item.ID,
				Language:   targetLang,
				Title:      translatedTitle,
				URL:        item.URL, // URL usually stays the same
			}

			ts.db.Create(&itemTranslation)
		}
	}

	return nil
}

// TranslateNotification translates notifications
func (ts *AITranslationService) TranslateNotification(notificationID uint, targetLanguages []string) error {
	var notification models.Notification
	if err := ts.db.First(&notification, notificationID).Error; err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	for _, targetLang := range targetLanguages {
		var existing models.NotificationTranslation
		if err := ts.db.Where("notification_id = ? AND language = ?", notificationID, targetLang).
			First(&existing).Error; err == nil {
			continue
		}

		translatedTitle, err := ts.aiService.TranslateText(context.Background(), notification.Title, "tr", targetLang)
		if err != nil {
			log.Printf("Failed to translate notification title: %v", err)
			continue
		}

		translatedMessage, err := ts.aiService.TranslateText(context.Background(), notification.Message, "tr", targetLang)
		if err != nil {
			log.Printf("Failed to translate notification message: %v", err)
			continue
		}

		translation := models.NotificationTranslation{
			NotificationID: notificationID,
			Language:       targetLang,
			Title:          translatedTitle,
			Message:        translatedMessage,
		}

		ts.db.Create(&translation)
	}

	return nil
}

// BulkTranslateAllContent translates all existing content
func (ts *AITranslationService) BulkTranslateAllContent() error {
	log.Println("Starting bulk translation of all content...")

	// Translate all categories
	var categories []models.Category
	ts.db.Find(&categories)
	for _, category := range categories {
		if err := ts.TranslateCategory(category.ID, ts.languages); err != nil {
			log.Printf("Failed to translate category %d: %v", category.ID, err)
		}
		time.Sleep(1 * time.Second) // Rate limiting
	}

	// Translate all tags
	var tags []models.Tag
	ts.db.Find(&tags)
	for _, tag := range tags {
		if err := ts.TranslateTag(tag.ID, ts.languages); err != nil {
			log.Printf("Failed to translate tag %d: %v", tag.ID, err)
		}
		time.Sleep(1 * time.Second)
	}

	// Translate all menus
	var menus []models.Menu
	ts.db.Find(&menus)
	for _, menu := range menus {
		if err := ts.TranslateMenu(menu.ID, ts.languages); err != nil {
			log.Printf("Failed to translate menu %d: %v", menu.ID, err)
		}
		time.Sleep(2 * time.Second)
	}

	log.Println("Bulk translation completed!")
	return nil
}

// ScheduledTranslationWorker runs translation jobs in background
func (ts *AITranslationService) ScheduledTranslationWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute) // Every 30 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ts.ProcessPendingTranslations()
		}
	}
}

// ProcessPendingTranslations processes translation queue
func (ts *AITranslationService) ProcessPendingTranslations() {
	log.Println("Processing pending translation jobs...")

	// Get pending translation jobs from the queue
	pendingJobs, err := ts.GetPendingJobs()
	if err != nil {
		log.Printf("Error fetching pending jobs: %v", err)
		return
	}

	if len(pendingJobs) == 0 {
		log.Println("No pending translation jobs found")
		return
	}

	log.Printf("Found %d pending translation jobs", len(pendingJobs))

	for _, job := range pendingJobs {
		// Update job status to processing
		if err := ts.UpdateJobStatus(job.ID, "processing"); err != nil {
			log.Printf("Error updating job %d status: %v", job.ID, err)
			continue
		}

		// Process the translation job
		go ts.processTranslationJob(job)
		time.Sleep(1 * time.Second) // Rate limiting
	}
}

// GetPendingJobs retrieves pending translation jobs from the queue
func (ts *AITranslationService) GetPendingJobs() ([]models.TranslationQueue, error) {
	var jobs []models.TranslationQueue
	err := ts.db.Where("status = ?", "pending").
		Order("priority DESC, created_at ASC").
		Limit(10). // Process maximum 10 jobs at a time
		Find(&jobs).Error
	return jobs, err
}

// CreateTranslationJob creates a new translation job in the queue
func (ts *AITranslationService) CreateTranslationJob(entityType string, entityID uint, sourceLang, targetLang string, priority int) error {
	// Check if job already exists
	var existingJob models.TranslationQueue
	if err := ts.db.Where("entity_type = ? AND entity_id = ? AND target_language = ? AND status IN ?",
		entityType, entityID, targetLang, []string{"pending", "processing"}).
		First(&existingJob).Error; err == nil {
		return nil // Job already exists
	}

	// If source language is empty, default to Turkish
	if sourceLang == "" {
		sourceLang = "tr"
	}

	// Default priority if not specified
	if priority <= 0 || priority > 3 {
		priority = 1
	}

	job := models.TranslationQueue{
		EntityType: entityType,
		EntityID:   entityID,
		TargetLang: targetLang,
		SourceLang: sourceLang,
		Status:     "pending",
		Priority:   priority,
	}

	// Validate the job before creating it
	// Note: validation is handled by the BeforeCreate hook in models.TranslationQueue

	return ts.db.Create(&job).Error
}

// UpdateJobStatus updates the status of a translation job
func (ts *AITranslationService) UpdateJobStatus(jobID uint, status string) error {
	// First, let's try a simple direct update to debug the issue
	var job models.TranslationQueue
	if err := ts.db.First(&job, jobID).Error; err != nil {
		log.Printf("Job with ID %d not found: %v", jobID, err)
		return fmt.Errorf("job with ID %d not found: %w", jobID, err)
	}

	// Update the job fields directly
	job.Status = status

	// Note: TranslationQueue model doesn't have ProcessedAt field
	// Timestamps are handled by UpdatedAt field automatically

	// Save the updated job
	result := ts.db.Save(&job)

	if result.Error != nil {
		log.Printf("Failed to update job status for job ID %d to %s: %v", jobID, status, result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Printf("No rows affected when updating job ID %d - job may not exist", jobID)
		return fmt.Errorf("job with ID %d not found", jobID)
	}

	log.Printf("Successfully updated job %d status to %s", jobID, status)
	return nil
}

// processTranslationJob processes a single translation job
func (ts *AITranslationService) processTranslationJob(job models.TranslationQueue) {
	log.Printf("Processing translation job %d: %s %d to %s", job.ID, job.EntityType, job.EntityID, job.TargetLang)

	// Update job status to processing
	if err := ts.UpdateJobStatus(job.ID, "processing"); err != nil {
		log.Printf("Error updating job %d status: %v", job.ID, err)
		return
	}

	var err error
	switch job.EntityType {
	case "article":
		err = ts.processArticleTranslation(job)
	case "category":
		err = ts.TranslateCategory(job.EntityID, []string{job.TargetLang})
	case "tag":
		err = ts.TranslateTag(job.EntityID, []string{job.TargetLang})
	case "menu":
		err = ts.TranslateMenu(job.EntityID, []string{job.TargetLang})
	case "notification":
		err = ts.TranslateNotification(job.EntityID, []string{job.TargetLang})
	default:
		err = fmt.Errorf("unsupported entity type: %s", job.EntityType)
	}

	if err != nil {
		log.Printf("Translation job %d failed: %v", job.ID, err)
		// Update status to failed and set detailed error message
		if updateErr := ts.UpdateJobStatus(job.ID, "failed"); updateErr != nil {
			log.Printf("Failed to update job status to failed: %v", updateErr)
		}
		errorMsg := err.Error()

		// Add more context to common errors
		if errorMsg == "unsupported data" {
			errorMsg = "Validation failed: check source and target languages match supported language codes"
		}

		ts.db.Model(&models.TranslationQueue{}).Where("id = ?", job.ID).
			Updates(map[string]interface{}{"error_message": errorMsg})
	} else {
		log.Printf("Translation job %d completed successfully", job.ID)
		// UpdateJobStatus will handle setting processed_at for completed status
		if updateErr := ts.UpdateJobStatus(job.ID, "completed"); updateErr != nil {
			log.Printf("Failed to update job status to completed: %v", updateErr)
		}
	}
}

// processArticleTranslation handles article translation specifically
func (ts *AITranslationService) processArticleTranslation(job models.TranslationQueue) error {
	// Get the article
	var article models.Article
	if err := ts.db.First(&article, job.EntityID).Error; err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	// Check if translation already exists
	var existing models.ArticleTranslation
	if err := ts.db.Where("article_id = ? AND language = ?", article.ID, job.TargetLang).
		First(&existing).Error; err == nil {
		log.Printf("Translation already exists for article %d in %s", article.ID, job.TargetLang)
		return nil
	}

	// Translate title
	translatedTitle, err := ts.aiService.TranslateText(context.Background(), article.Title, job.SourceLang, job.TargetLang)
	if err != nil {
		return fmt.Errorf("failed to translate title: %w", err)
	}

	// Translate content
	translatedContent, err := ts.aiService.TranslateText(context.Background(), article.Content, job.SourceLang, job.TargetLang)
	if err != nil {
		return fmt.Errorf("failed to translate content: %w", err)
	}

	// Translate summary if it exists
	translatedSummary := ""
	if article.Summary != "" {
		var summaryErr error
		translatedSummary, summaryErr = ts.aiService.TranslateText(context.Background(), article.Summary, job.SourceLang, job.TargetLang)
		if summaryErr != nil {
			log.Printf("Warning: Failed to translate summary for article %d: %v", article.ID, summaryErr)
			// Continue with empty summary rather than failing the whole translation
		}
	}

	// Create the translation
	translation := models.ArticleTranslation{
		ArticleID:       job.EntityID,
		Language:        job.TargetLang,
		Title:           translatedTitle,
		Slug:            generateSlug(translatedTitle),
		Summary:         translatedSummary,
		Content:         translatedContent,
		TranslationType: "openai",             // Use valid translation source
		Status:          "machine_translated", // Use valid translation status
		IsActive:        true,
		// Quality:          calculateTranslationQuality(article.Content, translatedContent), // Remove quality calculation for now
	}

	if err := ts.db.Create(&translation).Error; err != nil {
		return fmt.Errorf("failed to save translation: %w", err)
	}

	log.Printf("Successfully created translation for article %d in %s", job.EntityID, job.TargetLang)
	return nil
}

// TranslateArticle translates an article to target languages
func (ts *AITranslationService) TranslateArticle(articleID uint, targetLanguages []string) error {
	// Get the article
	var article models.Article
	if err := ts.db.First(&article, articleID).Error; err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	sourceLang := article.Language
	if sourceLang == "" {
		sourceLang = "tr" // Default source language
	}

	for _, targetLang := range targetLanguages {
		// Skip if same as source language
		if targetLang == sourceLang {
			continue
		}

		// Check if translation already exists
		var existing models.ArticleTranslation
		if err := ts.db.Where("article_id = ? AND language = ?", articleID, targetLang).
			First(&existing).Error; err == nil {
			log.Printf("Translation already exists for article %d in %s", articleID, targetLang)
			continue
		}

		// Translate title
		translatedTitle, err := ts.aiService.TranslateText(context.Background(), article.Title, sourceLang, targetLang)
		if err != nil {
			log.Printf("Failed to translate title for article %d: %v", articleID, err)
			continue
		}

		// Translate content
		translatedContent, err := ts.aiService.TranslateText(context.Background(), article.Content, sourceLang, targetLang)
		if err != nil {
			log.Printf("Failed to translate content for article %d: %v", articleID, err)
			continue
		}

		// Translate summary if it exists
		translatedSummary := ""
		if article.Summary != "" {
			var summaryErr error
			translatedSummary, summaryErr = ts.aiService.TranslateText(context.Background(), article.Summary, sourceLang, targetLang)
			if summaryErr != nil {
				log.Printf("Warning: Failed to translate summary for article %d: %v", articleID, summaryErr)
				// Continue with empty summary rather than failing the whole translation
			}
		}

		// Translate meta title if it exists
		translatedMetaTitle := ""
		if article.MetaTitle != "" {
			var metaTitleErr error
			translatedMetaTitle, metaTitleErr = ts.aiService.TranslateText(context.Background(), article.MetaTitle, sourceLang, targetLang)
			if metaTitleErr != nil {
				log.Printf("Warning: Failed to translate meta title for article %d: %v", articleID, metaTitleErr)
			}
		}

		// Translate meta description if it exists
		translatedMetaDesc := ""
		if article.MetaDesc != "" {
			var metaDescErr error
			translatedMetaDesc, metaDescErr = ts.aiService.TranslateText(context.Background(), article.MetaDesc, sourceLang, targetLang)
			if metaDescErr != nil {
				log.Printf("Warning: Failed to translate meta description for article %d: %v", articleID, metaDescErr)
			}
		}

		// Create the translation
		translation := models.ArticleTranslation{
			ArticleID:       articleID,
			Language:        targetLang,
			Title:           translatedTitle,
			Slug:            generateSlug(translatedTitle),
			Summary:         translatedSummary,
			Content:         translatedContent,
			MetaTitle:       translatedMetaTitle,
			MetaDescription: translatedMetaDesc,
			TranslationType: "openai",
			Status:          "machine_translated",
			IsActive:        true,
		}

		if err := ts.db.Create(&translation).Error; err != nil {
			log.Printf("Failed to save translation for article %d in %s: %v", articleID, targetLang, err)
			continue
		}

		log.Printf("Successfully created translation for article %d in %s", articleID, targetLang)
	}

	return nil
}

// QueueTranslationsForArticle creates translation jobs for all supported languages
func (ts *AITranslationService) QueueTranslationsForArticle(articleID uint, requestedBy uint) error {
	// Get the article to determine source language
	var article models.Article
	if err := ts.db.First(&article, articleID).Error; err != nil {
		return fmt.Errorf("article not found: %v", err)
	}

	sourceLang := article.Language
	if sourceLang == "" {
		sourceLang = "tr" // Default source language
	}

	// Create jobs for all target languages except the source language
	for _, targetLang := range ts.languages {
		if targetLang == sourceLang {
			continue // Skip source language
		}

		// Check if translation already exists
		var existing models.ArticleTranslation
		if err := ts.db.Where("article_id = ? AND language = ?", articleID, targetLang).
			First(&existing).Error; err == nil {
			continue // Translation already exists
		}

		// Create the translation job
		if err := ts.CreateTranslationJob("article", articleID, sourceLang, targetLang, 1); err != nil {
			log.Printf("Failed to create translation job for article %d to %s: %v", articleID, targetLang, err)
		}
	}

	return nil
}

// Helper function to generate slug (same as in article_translation.go)
func generateSlug(title string) string {
	slug := strings.ToLower(title)

	replacements := map[string]string{
		"ı": "i", "ğ": "g", "ü": "u", "ş": "s", "ö": "o", "ç": "c",
		"İ": "i", "Ğ": "g", "Ü": "u", "Ş": "s", "Ö": "o", "Ç": "c",
		"á": "a", "à": "a", "ä": "a", "â": "a", "ā": "a",
		"é": "e", "è": "e", "ë": "e", "ê": "e", "ē": "e",
		"í": "i", "ì": "i", "ï": "i", "î": "i", "ī": "i",
		"ó": "o", "ò": "o", "ô": "o", "ō": "o",
		"ú": "u", "ù": "u", "û": "u", "ū": "u",
		"ý": "y", "ÿ": "y", "ñ": "n", "ß": "ss",
	}

	for old, new := range replacements {
		slug = strings.ReplaceAll(slug, old, new)
	}

	slug = strings.ReplaceAll(slug, " ", "-")

	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()

	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "untitled"
	}

	return slug
}
