package handlers

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"news/internal/database"
	"news/internal/models"
	"news/internal/queue"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// GetTranslationProgress godoc
// @Summary Get translation progress
// @Description Get translation progress for all content types and languages
// @Tags Translation
// @Produce json
// @Security Bearer
// @Success 200 {array} models.TranslationProgress
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/progress [get]
func GetTranslationProgress(c *gin.Context) {
	var progress []models.TranslationProgress

	entityTypes := []string{"article", "category", "tag", "menu", "notification"}
	languages := []string{"en", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"}

	for _, entityType := range entityTypes {
		var totalCount int64
		var translatedCount int64

		switch entityType {
		case "article":
			database.DB.Model(&models.Article{}).Count(&totalCount)
			database.DB.Model(&models.ArticleTranslation{}).
				Distinct("article_id").Count(&translatedCount)
		case "category":
			database.DB.Model(&models.Category{}).Count(&totalCount)
			database.DB.Model(&models.CategoryTranslation{}).
				Distinct("category_id").Count(&translatedCount)
		case "tag":
			database.DB.Model(&models.Tag{}).Count(&totalCount)
			database.DB.Model(&models.TagTranslation{}).
				Distinct("tag_id").Count(&translatedCount)
		case "menu":
			database.DB.Model(&models.Menu{}).Count(&totalCount)
			database.DB.Model(&models.MenuTranslation{}).
				Distinct("menu_id").Count(&translatedCount)
		case "notification":
			database.DB.Model(&models.Notification{}).Count(&totalCount)
			database.DB.Model(&models.NotificationTranslation{}).
				Distinct("notification_id").Count(&translatedCount)
		}

		completionRate := float64(0)
		if totalCount > 0 {
			completionRate = float64(translatedCount) / float64(totalCount) * 100
		}

		progress = append(progress, models.TranslationProgress{
			EntityType:         entityType,
			TotalEntities:      int(totalCount),
			TranslatedCount:    int(translatedCount),
			PendingCount:       int(totalCount - translatedCount),
			CompletionRate:     completionRate,
			AvailableLanguages: languages,
		})
	}

	c.JSON(http.StatusOK, progress)
}

// BulkTranslateContent godoc
// @Summary Bulk translate content
// @Description Queue bulk translation jobs for specified content
// @Tags Translation
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.BulkTranslationRequest true "Bulk translation request"
// @Success 200 {object} models.BulkTranslationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/bulk [post]
func BulkTranslateContent(c *gin.Context) {
	var request models.BulkTranslationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Ensure we have target languages
	if len(request.TargetLanguages) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Target languages are required"})
		return
	}

	// Validate entity type
	validEntityTypes := map[string]bool{
		"article":      true,
		"category":     true,
		"tag":          true,
		"menu":         true,
		"notification": true,
	}

	if !validEntityTypes[request.EntityType] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid entity type"})
		return
	}

	// Validate target languages
	validLanguages := map[string]bool{
		"en": true, // English
		"tr": true, // Turkish
		"es": true, // Spanish
		"fr": true, // French
		"de": true, // German
		"ar": true, // Arabic
		"zh": true, // Chinese
		"ru": true, // Russian
		"ja": true, // Japanese
		"ko": true, // Korean
	}

	for _, lang := range request.TargetLanguages {
		if !validLanguages[lang] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: fmt.Sprintf("Invalid target language: %s. Supported languages: %v",
					lang, []string{"en", "tr", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"}),
			})
			return
		}
	}

	// Set default source language if not provided
	sourceLanguage := request.SourceLanguage
	if sourceLanguage == "" {
		sourceLanguage = "tr" // default to Turkish as source
		// Set the source language on the request object to avoid validation issues
		request.SourceLanguage = sourceLanguage
	}

	// Validate source language
	if !validLanguages[sourceLanguage] {
		supportedLanguages := []string{"en", "tr", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Invalid source language: %s. Supported languages: %v",
				sourceLanguage, supportedLanguages),
		})
		return
	}

	// Make sure source and target languages are different
	for _, targetLang := range request.TargetLanguages {
		if targetLang == sourceLanguage {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: fmt.Sprintf("Target language '%s' cannot be the same as source language", targetLang),
			})
			return
		}
	}

	// If no specific entity IDs provided, get all
	var entityIDs []uint
	if len(request.EntityIDs) == 0 {
		switch request.EntityType {
		case "article":
			var articles []models.Article
			database.DB.Select("id").Find(&articles)
			for _, article := range articles {
				entityIDs = append(entityIDs, article.ID)
			}
		case "category":
			var categories []models.Category
			database.DB.Select("id").Find(&categories)
			for _, category := range categories {
				entityIDs = append(entityIDs, category.ID)
			}
		case "tag":
			var tags []models.Tag
			database.DB.Select("id").Find(&tags)
			for _, tag := range tags {
				entityIDs = append(entityIDs, tag.ID)
			}
		case "menu":
			var menus []models.Menu
			database.DB.Select("id").Find(&menus)
			for _, menu := range menus {
				entityIDs = append(entityIDs, menu.ID)
			}
		case "notification":
			var notifications []models.Notification
			database.DB.Select("id").Find(&notifications)
			for _, notification := range notifications {
				entityIDs = append(entityIDs, notification.ID)
			}
		}
	} else {
		entityIDs = request.EntityIDs
	}

	var queuedJobs int
	var skippedJobs int
	var failedJobs int
	var jobIDs []uint

	// Create translation queue entries
	for _, entityID := range entityIDs {
		for _, targetLang := range request.TargetLanguages {
			// Check if translation already exists (unless force retranslate)
			if !request.ForceRetranslate {
				exists := false

				switch request.EntityType {
				case "article":
					var translation models.ArticleTranslation
					err := database.DB.Where("article_id = ? AND language = ?", entityID, targetLang).
						First(&translation).Error
					exists = err == nil
				case "category":
					var translation models.CategoryTranslation
					err := database.DB.Where("category_id = ? AND language = ?", entityID, targetLang).
						First(&translation).Error
					exists = err == nil
				case "tag":
					var translation models.TagTranslation
					err := database.DB.Where("tag_id = ? AND language = ?", entityID, targetLang).
						First(&translation).Error
					exists = err == nil
				case "menu":
					var translation models.MenuTranslation
					err := database.DB.Where("menu_id = ? AND language = ?", entityID, targetLang).
						First(&translation).Error
					exists = err == nil
				case "notification":
					var translation models.NotificationTranslation
					err := database.DB.Where("notification_id = ? AND language = ?", entityID, targetLang).
						First(&translation).Error
					exists = err == nil
				}

				if exists {
					skippedJobs++
					continue
				}
			}

			// Get queue manager
			queueManager := queue.GetGlobalQueueManager()
			if queueManager == nil {
				log.Printf("Queue manager not available for entity %d, language %s", entityID, targetLang)
				failedJobs++
				continue
			}

			// Map priority to queue priority
			var queuePriority queue.JobPriority
			switch request.Priority {
			case 3:
				queuePriority = queue.PriorityHigh
			case 2:
				queuePriority = queue.PriorityNormal
			case 1:
				queuePriority = queue.PriorityLow
			default:
				queuePriority = queue.PriorityNormal
			}

			// Enqueue translation job using Redis queue
			if err := queueManager.EnqueueTranslationJob(request.EntityType, entityID, sourceLanguage, targetLang, queuePriority); err != nil {
				log.Printf("Failed to enqueue translation job for entity %d, language %s: %v", entityID, targetLang, err)
				failedJobs++
				continue
			}

			queuedJobs++
			// Note: Redis jobs don't have database IDs, using entity info instead
			jobIDs = append(jobIDs, entityID)
		}
	}

	// Estimate completion time (rough calculation)
	estimatedTime := "5-30 minutes"
	if queuedJobs > 100 {
		estimatedTime = "1-2 hours"
	} else if queuedJobs > 50 {
		estimatedTime = "30-60 minutes"
	}

	response := models.BulkTranslationResponse{
		QueuedJobs:    queuedJobs,
		SkippedJobs:   skippedJobs,
		FailedJobs:    failedJobs,
		EstimatedTime: estimatedTime,
		JobIDs:        jobIDs,
	}

	if queuedJobs == 0 && failedJobs > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Failed to queue any jobs. %d jobs failed validation.", failedJobs),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTranslationQueue godoc
// @Summary Get translation queue
// @Description Get pending and processing translation jobs
// @Tags Translation
// @Produce json
// @Security Bearer
// @Param status query string false "Filter by status (pending, processing, completed, failed)"
// @Param entity_type query string false "Filter by entity type"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} models.TranslationQueue
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/queue [get]
func GetTranslationQueue(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	// Get queue manager
	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Queue manager not available"})
		return
	}

	// Get jobs from translations queue
	jobs, total, err := queueManager.GetJobs("translations", status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch translation jobs: " + err.Error()})
		return
	}

	// Create a response structure similar to the old database-based queue
	type TranslationQueueResponse struct {
		Jobs       []queue.JobStatusInfo `json:"jobs"`
		TotalCount int64                 `json:"total_count"`
		Page       int                   `json:"page"`
		Limit      int                   `json:"limit"`
	}

	response := TranslationQueueResponse{
		Jobs:       jobs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}

	c.JSON(http.StatusOK, response)
}

// ProcessTranslationQueue godoc
// @Summary Process translation queue
// @Description Manually trigger processing of pending translation jobs
// @Tags Translation
// @Produce json
// @Security Bearer
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/process [post]
func ProcessTranslationQueue(c *gin.Context) {
	// Get queue manager
	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Queue manager not available"})
		return
	}

	// Get pending jobs count from translations queue
	jobs, total, err := queueManager.GetJobs("translations", "pending", 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get pending jobs: " + err.Error()})
		return
	}

	// The Redis queue system processes jobs automatically via worker pools
	// This endpoint now just reports the status
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: fmt.Sprintf("Translation queue is being processed automatically. Found %d pending jobs out of %d total jobs.", len(jobs), total),
	})
}

// TranslateEntity godoc
// @Summary Translate single entity
// @Description Translate a single entity to specified languages
// @Tags Translation
// @Accept json
// @Produce json
// @Security Bearer
// @Param entity_type path string true "Entity type (article, category, tag, menu, notification)"
// @Param entity_id path int true "Entity ID"
// @Param languages body []string true "Target languages"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/{entity_type}/{entity_id} [post]
func TranslateEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid entity ID"})
		return
	}

	var targetLanguages []string
	if err := c.ShouldBindJSON(&targetLanguages); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Get AI service
	aiService := services.GetAIService()

	// Create AI translation service
	translationService := services.NewAITranslationService(aiService)

	// Translate based on entity type
	switch entityType {
	case "article":
		err = translationService.TranslateArticle(uint(entityID), targetLanguages)
	case "category":
		err = translationService.TranslateCategory(uint(entityID), targetLanguages)
	case "tag":
		err = translationService.TranslateTag(uint(entityID), targetLanguages)
	case "menu":
		err = translationService.TranslateMenu(uint(entityID), targetLanguages)
	case "notification":
		err = translationService.TranslateNotification(uint(entityID), targetLanguages)
	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid entity type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Translation completed successfully",
	})
}

// GetLocalizedContent godoc
// @Summary Get localized content
// @Description Get content in specified language with fallback to default
// @Tags Translation
// @Produce json
// @Param entity_type path string true "Entity type (category, tag, menu, notification)"
// @Param entity_id path int true "Entity ID"
// @Param lang query string false "Language code" default(tr)
// @Success 200 {object} interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /content/{entity_type}/{entity_id} [get]
func GetLocalizedContent(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")
	lang := c.DefaultQuery("lang", "tr")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid entity ID"})
		return
	}

	switch entityType {
	case "category":
		var category models.Category
		if err := database.DB.First(&category, entityID).Error; err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Category not found"})
			return
		}

		// Try to get translation
		var translation models.CategoryTranslation
		if err := database.DB.Where("category_id = ? AND language = ?", entityID, lang).
			First(&translation).Error; err == nil {
			// Return localized version
			localized := models.LocalizedCategory{
				ID:          category.ID,
				Name:        translation.Name,
				Slug:        translation.Slug,
				Description: translation.Description,
				MetaTitle:   translation.MetaTitle,
				MetaDesc:    translation.MetaDesc,
				Color:       category.Color,
				Language:    lang,
				CreatedAt:   category.CreatedAt,
				UpdatedAt:   category.UpdatedAt,
			}
			c.JSON(http.StatusOK, localized)
			return
		}

		// Fallback to original
		localized := models.LocalizedCategory{
			ID:          category.ID,
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
			Color:       category.Color,
			Language:    "tr", // original language
			CreatedAt:   category.CreatedAt,
			UpdatedAt:   category.UpdatedAt,
		}
		c.JSON(http.StatusOK, localized)

	case "tag":
		var tag models.Tag
		if err := database.DB.First(&tag, entityID).Error; err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Tag not found"})
			return
		}

		// Try to get translation
		var translation models.TagTranslation
		if err := database.DB.Where("tag_id = ? AND language = ?", entityID, lang).
			First(&translation).Error; err == nil {
			// Return localized version
			localized := models.LocalizedTag{
				ID:          tag.ID,
				Name:        translation.Name,
				Slug:        translation.Slug,
				Description: translation.Description,
				Color:       tag.Color,
				UsageCount:  tag.UsageCount,
				Language:    lang,
				CreatedAt:   tag.CreatedAt,
				UpdatedAt:   tag.UpdatedAt,
			}
			c.JSON(http.StatusOK, localized)
			return
		}

		// Fallback to original
		localized := models.LocalizedTag{
			ID:          tag.ID,
			Name:        tag.Name,
			Slug:        tag.Slug,
			Description: tag.Description,
			Color:       tag.Color,
			UsageCount:  tag.UsageCount,
			Language:    "tr",
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		}
		c.JSON(http.StatusOK, localized)

	case "menu":
		var menu models.Menu
		if err := database.DB.Preload("Items").First(&menu, entityID).Error; err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu not found"})
			return
		}

		// Try to get translation
		var translation models.MenuTranslation
		if err := database.DB.Where("menu_id = ? AND language = ?", entityID, lang).
			First(&translation).Error; err == nil {
			// Return localized version
			localized := models.LocalizedMenu{
				ID:          menu.ID,
				Name:        translation.Name,
				Slug:        menu.Slug,
				Location:    menu.Location,
				Description: translation.Description,
				Language:    lang,
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
					Language:  "tr", // default
					CreatedAt: item.CreatedAt,
					UpdatedAt: item.UpdatedAt,
				}

				if err := database.DB.Where("menu_item_id = ? AND language = ?", item.ID, lang).
					First(&itemTranslation).Error; err == nil {
					localizedItem.Title = itemTranslation.Title
					localizedItem.URL = itemTranslation.URL
					localizedItem.Language = lang
				}

				localized.Items = append(localized.Items, localizedItem)
			}

			c.JSON(http.StatusOK, localized)
			return
		}

		// Fallback to original
		localized := models.LocalizedMenu{
			ID:        menu.ID,
			Name:      menu.Name,
			Slug:      menu.Slug,
			Location:  menu.Location,
			Language:  "tr",
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
				Language:  "tr",
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			}
			localized.Items = append(localized.Items, localizedItem)
		}

		c.JSON(http.StatusOK, localized)

	default:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid entity type"})
		return
	}
}

// TestTranslationSystem godoc
// @Summary Test translation system
// @Description Test endpoint for checking translation system functionality
// @Tags Translation
// @Produce json
// @Security Bearer
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/translations/test [get]
func TestTranslationSystem(c *gin.Context) {
	// Test results map
	testResults := make(map[string]interface{})

	// 1. Check queue manager and pending translation jobs
	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		testResults["queue_manager_error"] = "Queue manager not available"
	} else {
		// Get pending jobs from Redis queue
		jobs, total, err := queueManager.GetJobs("translations", "pending", 1, 10)
		if err != nil {
			testResults["pending_jobs_error"] = err.Error()
		} else {
			testResults["pending_jobs_count"] = len(jobs)
			testResults["total_jobs"] = total

			if len(jobs) > 0 {
				// Get first job for demonstration
				testResults["sample_job"] = jobs[0]
			}
		}
	}

	// 2. Test Redis queue translation
	var article models.Article
	if err := database.DB.First(&article).Error; err == nil {
		testResults["found_article_id"] = article.ID

		// Queue a translation for this article to English using Redis queue
		if queueManager != nil {
			err = queueManager.EnqueueTranslationJob("article", article.ID, "tr", "en", queue.PriorityHigh)
			if err != nil {
				testResults["queue_translation_error"] = err.Error()
			} else {
				testResults["queue_translation_success"] = true
			}
		} else {
			testResults["queue_translation_error"] = "Queue manager not available"
		}
	} else {
		testResults["article_error"] = err.Error()
	}

	// 3. Check JWT token information
	claims, exists := c.Get("claims")
	if exists {
		testResults["auth_claims_exist"] = true
		testResults["auth_claims_type"] = fmt.Sprintf("%T", claims)
	} else {
		testResults["auth_claims_exist"] = false
	}

	role, exists := c.Get("role")
	if exists {
		testResults["auth_role"] = role
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Translation system test completed",
		"results": testResults,
		"environment": map[string]string{
			"go_version": runtime.Version(),
		},
	})
}
