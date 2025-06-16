package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"news/internal/database"
	"news/internal/models"
	"news/internal/repositories"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gorm.io/gorm"
)

type ArticleTranslationHandlers struct {
	repo               *repositories.ArticleTranslationRepository
	translationService *services.UnifiedTranslationService
}

func NewArticleTranslationHandlers() *ArticleTranslationHandlers {
	return &ArticleTranslationHandlers{
		repo:               repositories.NewArticleTranslationRepository(database.DB),
		translationService: services.GetUnifiedTranslationService(),
	}
}

// CreateTranslationRequest represents the request payload for creating a translation
type CreateTranslationRequest struct {
	Language        string `json:"language" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Content         string `json:"content" binding:"required"`
	Summary         string `json:"summary,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	TranslationType string `json:"translation_type,omitempty"` // "manual", "ai", "professional"
}

// UpdateTranslationRequest represents the request payload for updating a translation
type UpdateTranslationRequest struct {
	Title           string `json:"title,omitempty"`
	Content         string `json:"content,omitempty"`
	Summary         string `json:"summary,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	Status          string `json:"status,omitempty"` // "draft", "published", "under_review"
}

// GetLocalizedArticles godoc
// @Summary Get localized articles
// @Description Retrieve a paginated list of articles in the user's preferred language with fallback
// @Tags Article Translations
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of articles per page" default(10)
// @Param lang query string false "Language code (e.g., 'en', 'tr', 'es')"
// @Success 200 {object} models.PaginatedLocalizedArticlesResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/localized [get]
func (h *ArticleTranslationHandlers) GetLocalizedArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get language from context (set by i18n middleware)
	language, exists := c.Get("language")
	if !exists {
		language = "en" // fallback
	}
	lang := language.(string)

	// Get localized articles with pagination
	articles, total, err := h.repo.GetLocalizedArticlesPaginated(lang, page, limit)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	response := models.PaginatedLocalizedArticlesResponse{
		Page:     page,
		Limit:    limit,
		Total:    total,
		Language: lang,
		Articles: articles,
	}

	c.JSON(http.StatusOK, response)
}

// GetLocalizedArticle godoc
// @Summary Get localized article by ID
// @Description Retrieve a single article in the user's preferred language with fallback
// @Tags Article Translations
// @Accept json
// @Produce json
// @Param id path int true "Article ID"
// @Param lang query string false "Language code (e.g., 'en', 'tr', 'es')"
// @Success 200 {object} models.LocalizedArticle
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/articles/{id}/localized [get]
func (h *ArticleTranslationHandlers) GetLocalizedArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_id", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get language from context (set by i18n middleware)
	language, exists := c.Get("language")
	if !exists {
		language = "en" // fallback
	}
	lang := language.(string)

	// Get localized article
	article, err := h.repo.GetLocalizedArticle(uint(id), lang)
	if err != nil {
		localizer, _ := c.Get("localizer")
		if err == gorm.ErrRecordNotFound {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.articles.not_found", nil)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
		} else {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		}
		return
	}

	// Increment view count for the original article
	database.DB.Model(&models.Article{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1))

	c.JSON(http.StatusOK, article)
}

// CreateArticleTranslation godoc
// @Summary Create a new article translation
// @Description Create a translation for an existing article
// @Tags Article Translations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param translation body CreateTranslationRequest true "Translation data"
// @Success 201 {object} models.ArticleTranslation
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/{id}/translations [post]
func (h *ArticleTranslationHandlers) CreateArticleTranslation(c *gin.Context) {
	idStr := c.Param("id")
	articleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_id", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	var req CreateTranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_payload", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.auth.not_authenticated", nil)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: errMsg})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.auth.invalid_user", nil)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: errMsg})
		return
	}

	// Verify that the article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		localizer, _ := c.Get("localizer")
		if err == gorm.ErrRecordNotFound {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.articles.not_found", nil)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
		} else {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		}
		return
	}

	// Check if translation already exists for this language
	existingTranslation, err := h.repo.GetTranslationByLanguage(uint(articleID), req.Language)
	if err == nil && existingTranslation != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.already_exists", map[string]interface{}{
			"Language": req.Language,
		})
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: errMsg})
		return
	}

	// Set default translation type if not provided
	translationType := req.TranslationType
	if translationType == "" {
		translationType = "manual"
	}

	// Create translation
	translation := &models.ArticleTranslation{
		ArticleID:       uint(articleID),
		Language:        req.Language,
		Title:           req.Title,
		Content:         req.Content,
		Summary:         req.Summary,
		MetaDescription: req.MetaDescription,
		Status:          "draft",
		TranslationType: translationType,
		TranslatedBy:    &userID,
	}

	// Generate slug for the translation
	translation.Slug = h.generateTranslationSlug(req.Title, req.Language)

	// Create the translation
	createdTranslation, err := h.repo.CreateTranslation(translation)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.create_failed", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	localizer, _ := c.Get("localizer")
	successMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "success.translations.created", map[string]interface{}{
		"Language": req.Language,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message":     successMsg,
		"translation": createdTranslation,
	})
}

// UpdateArticleTranslation godoc
// @Summary Update an article translation
// @Description Update an existing article translation
// @Tags Article Translations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param language path string true "Language code"
// @Param translation body UpdateTranslationRequest true "Translation data"
// @Success 200 {object} models.ArticleTranslation
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/{id}/translations/{language} [put]
func (h *ArticleTranslationHandlers) UpdateArticleTranslation(c *gin.Context) {
	idStr := c.Param("id")
	articleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_id", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	language := c.Param("language")
	if language == "" {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.language_required", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	var req UpdateTranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_payload", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get user ID from context (set by auth middleware) - just verify authentication
	_, exists := c.Get("user_id")
	if !exists {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.auth.not_authenticated", nil)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get existing translation
	existingTranslation, err := h.repo.GetTranslationByLanguage(uint(articleID), language)
	if err != nil {
		localizer, _ := c.Get("localizer")
		if err == gorm.ErrRecordNotFound {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.not_found", nil)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
		} else {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		}
		return
	}

	// Prepare update data
	updateData := make(map[string]interface{})
	if req.Title != "" {
		updateData["title"] = req.Title
		updateData["slug"] = h.generateTranslationSlug(req.Title, language)
	}
	if req.Content != "" {
		updateData["content"] = req.Content
	}
	if req.Summary != "" {
		updateData["summary"] = req.Summary
	}
	if req.MetaDescription != "" {
		updateData["meta_description"] = req.MetaDescription
	}
	if req.Status != "" {
		updateData["status"] = req.Status
	}

	// Update translation
	updatedTranslation, err := h.repo.UpdateTranslation(existingTranslation.ID, updateData)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.update_failed", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	localizer, _ := c.Get("localizer")
	successMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "success.translations.updated", map[string]interface{}{
		"Language": language,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     successMsg,
		"translation": updatedTranslation,
	})
}

// DeleteArticleTranslation godoc
// @Summary Delete an article translation
// @Description Delete an existing article translation
// @Tags Article Translations
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param language path string true "Language code"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/{id}/translations/{language} [delete]
func (h *ArticleTranslationHandlers) DeleteArticleTranslation(c *gin.Context) {
	idStr := c.Param("id")
	articleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_id", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	language := c.Param("language")
	if language == "" {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.language_required", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get user ID from context (set by auth middleware) - just verify authentication
	_, exists := c.Get("user_id")
	if !exists {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.auth.not_authenticated", nil)
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get existing translation to verify it exists
	existingTranslation, err := h.repo.GetTranslationByLanguage(uint(articleID), language)
	if err != nil {
		localizer, _ := c.Get("localizer")
		if err == gorm.ErrRecordNotFound {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.not_found", nil)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
		} else {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		}
		return
	}

	// Delete translation
	err = h.repo.DeleteTranslation(existingTranslation.ID)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.translations.delete_failed", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	localizer, _ := c.Get("localizer")
	successMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "success.translations.deleted", map[string]interface{}{
		"Language": language,
	})

	c.JSON(http.StatusOK, models.SuccessResponse{Message: successMsg})
}

// GetArticleTranslations godoc
// @Summary Get all translations for an article
// @Description Retrieve all available translations for a specific article
// @Tags Article Translations
// @Accept json
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} models.ArticleWithTranslations
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/{id}/translations [get]
func (h *ArticleTranslationHandlers) GetArticleTranslations(c *gin.Context) {
	idStr := c.Param("id")
	articleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.invalid_id", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	// Get article with all translations
	articleWithTranslations, err := h.repo.GetArticleWithTranslations(uint(articleID))
	if err != nil {
		localizer, _ := c.Get("localizer")
		if err == gorm.ErrRecordNotFound {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.articles.not_found", nil)
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: errMsg})
		} else {
			errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, articleWithTranslations)
}

// SearchLocalizedArticles godoc
// @Summary Search localized articles
// @Description Search articles in the user's preferred language with fallback
// @Tags Article Translations
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of articles per page" default(10)
// @Param lang query string false "Language code (e.g., 'en', 'tr', 'es')"
// @Success 200 {object} models.PaginatedLocalizedArticlesResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/articles/search/localized [get]
func (h *ArticleTranslationHandlers) SearchLocalizedArticles(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.validation.search_query_required", nil)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: errMsg})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get language from context (set by i18n middleware)
	language, exists := c.Get("language")
	if !exists {
		language = "en" // fallback
	}
	lang := language.(string)

	// Search localized articles
	offset := (page - 1) * limit
	articles, total, err := h.repo.SearchTranslatedArticles(query, lang, offset, limit)
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.search_failed", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	response := models.PaginatedLocalizedArticlesResponse{
		Page:     page,
		Limit:    limit,
		Total:    total,
		Language: lang,
		Articles: articles,
		Query:    query,
	}

	c.JSON(http.StatusOK, response)
}

// GetTranslationStats godoc
// @Summary Get translation statistics
// @Description Get statistics about article translations
// @Tags Article Translations
// @Accept json
// @Produce json
// @Success 200 {object} models.TranslationStats
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/translations/stats [get]
func (h *ArticleTranslationHandlers) GetTranslationStats(c *gin.Context) {
	stats, err := h.repo.GetTranslationStats()
	if err != nil {
		localizer, _ := c.Get("localizer")
		errMsg := h.getLocalizedMessage(localizer.(*i18n.Localizer), "errors.general.database_error", nil)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: errMsg})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper functions

// getLocalizedMessage gets a localized message with fallback to English using unified translation service
func (h *ArticleTranslationHandlers) getLocalizedMessage(localizer *i18n.Localizer, messageID string, templateData map[string]interface{}) string {
	// Try to get language from the localizer context
	language := "en" // default fallback

	// If unified translation service is available, use it
	if h.translationService != nil {
		translatedMessage, err := h.translationService.TranslateUI(language, messageID, templateData)
		if err == nil {
			return translatedMessage
		}
	}

	// Fallback to original localizer method
	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	message, err := localizer.Localize(config)
	if err != nil {
		// Fallback to message ID if translation fails
		return messageID
	}

	return message
}

// generateTranslationSlug generates a URL-friendly slug for the translation
func (h *ArticleTranslationHandlers) generateTranslationSlug(title, language string) string {
	// Simple slug generation - in production you might want to use a more sophisticated approach
	slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "!", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")

	// Add language suffix to avoid conflicts
	return slug + "-" + language
}
