package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/cache"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// TranslationHandler handles translation-related API endpoints
type TranslationHandler struct {
	translationService *services.AITranslationService
	translationCache   cache.TranslationCache
}

// NewTranslationHandler creates a new translation handler
func NewTranslationHandler(translationService *services.AITranslationService, translationCache cache.TranslationCache) *TranslationHandler {
	return &TranslationHandler{
		translationService: translationService,
		translationCache:   translationCache,
	}
}

// GetUITranslation retrieves UI translation
// @Summary Get UI Translation
// @Description Get translated UI text by key and language
// @Tags translation
// @Accept json
// @Produce json
// @Param language path string true "Language code"
// @Param key path string true "Translation key"
// @Success 200 {object} map[string]interface{}
// @Router /api/translations/ui/{language}/{key} [get]
func (h *TranslationHandler) GetUITranslation(c *gin.Context) {
	language := c.Param("language")
	key := c.Param("key")

	if language == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Language and key are required"})
		return
	}

	translation, err := h.translationCache.GetUITranslation(language, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":         key,
		"language":    language,
		"translation": translation,
	})
}

// GetArticleTranslation retrieves article with translations
// @Summary Get Article Translation
// @Description Get article with localized content
// @Tags translation
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Param language query string false "Language code" default(en)
// @Success 200 {object} models.LocalizedArticle
// @Router /api/translations/articles/{id} [get]
func (h *TranslationHandler) GetArticleTranslation(c *gin.Context) {
	idStr := c.Param("id")
	language := c.DefaultQuery("language", "en")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.translationCache.GetArticleTranslation(uint(id), language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

// GetSEOTranslation retrieves SEO settings with translations
// @Summary Get SEO Translation
// @Description Get SEO settings with localized content
// @Tags translation
// @Accept json
// @Produce json
// @Param type path string true "Entity type (page, article)"
// @Param id path string true "Entity ID"
// @Param language query string false "Language code" default(en)
// @Success 200 {object} models.LocalizedSEOSettings
// @Router /api/translations/seo/{type}/{id} [get]
func (h *TranslationHandler) GetSEOTranslation(c *gin.Context) {
	entityType := c.Param("type")
	idStr := c.Param("id")
	language := c.DefaultQuery("language", "en")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	seoSettings, err := h.translationCache.GetSEOTranslation(entityType, uint(id), language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SEO settings not found"})
		return
	}

	c.JSON(http.StatusOK, seoSettings)
}

// GetFormTranslation retrieves form field translations
// @Summary Get Form Translation
// @Description Get form field translations
// @Tags translation
// @Accept json
// @Produce json
// @Param form path string true "Form key"
// @Param field path string true "Field key"  
// @Param language query string false "Language code" default(en)
// @Success 200 {object} models.FormTranslation
// @Router /api/translations/forms/{form}/{field} [get]
func (h *TranslationHandler) GetFormTranslation(c *gin.Context) {
	formKey := c.Param("form")
	fieldKey := c.Param("field")
	language := c.DefaultQuery("language", "en")

	translation, err := h.translationCache.GetFormTranslation(formKey, fieldKey, language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form translation not found"})
		return
	}

	c.JSON(http.StatusOK, translation)
}

// GetErrorTranslation retrieves error message translations
// @Summary Get Error Translation
// @Description Get error message translations
// @Tags translation
// @Accept json
// @Produce json
// @Param code path string true "Error code"
// @Param language query string false "Language code" default(en)
// @Success 200 {object} models.ErrorMessageTranslation
// @Router /api/translations/errors/{code} [get]
func (h *TranslationHandler) GetErrorTranslation(c *gin.Context) {
	errorCode := c.Param("code")
	language := c.DefaultQuery("language", "en")

	translation, err := h.translationCache.GetErrorTranslation(errorCode, language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Error translation not found"})
		return
	}

	c.JSON(http.StatusOK, translation)
}

// GetCommentTranslation retrieves comment with translation
// @Summary Get Comment Translation
// @Description Get comment with localized content
// @Tags translation
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param language query string false "Language code" default(en)
// @Success 200 {object} models.LocalizedComment
// @Router /api/translations/comments/{id} [get]
func (h *TranslationHandler) GetCommentTranslation(c *gin.Context) {
	idStr := c.Param("id")
	language := c.DefaultQuery("language", "en")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := h.translationService.GetLocalizedComment(uint(id), language)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// TranslateArticle triggers article translation
// @Summary Translate Article
// @Description Trigger article translation to target languages
// @Tags translation
// @Accept json
// @Produce json
// @Param id path string true "Article ID"
// @Param body body TranslationRequest true "Translation request"
// @Success 200 {object} map[string]interface{}  
// @Router /api/translations/articles/{id}/translate [post]
func (h *TranslationHandler) TranslateArticle(c *gin.Context) {
	idStr := c.Param("id")
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var req TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.translationService.TranslateArticle(uint(id), req.TargetLanguages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache for the article
	h.translationCache.InvalidateEntityTranslations("article", uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Article translation started",
		"article_id": id,
		"languages": req.TargetLanguages,
	})
}

// TranslateComment triggers comment translation
// @Summary Translate Comment
// @Description Trigger comment translation to target languages
// @Tags translation
// @Accept json
// @Produce json
// @Param id path string true "Comment ID"
// @Param body body TranslationRequest true "Translation request"
// @Success 200 {object} map[string]interface{}
// @Router /api/translations/comments/{id}/translate [post]
func (h *TranslationHandler) TranslateComment(c *gin.Context) {
	idStr := c.Param("id")
	
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var req TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.translationService.TranslateComment(uint(id), req.TargetLanguages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment translation started",
		"comment_id": id,
		"languages": req.TargetLanguages,
	})
}

// BulkTranslateComments triggers bulk comment translation
// @Summary Bulk Translate Comments
// @Description Trigger bulk comment translation to target languages
// @Tags translation
// @Accept json
// @Produce json
// @Param body body BulkTranslationRequest true "Bulk translation request"
// @Success 200 {object} map[string]interface{}
// @Router /api/translations/comments/bulk-translate [post]
func (h *TranslationHandler) BulkTranslateComments(c *gin.Context) {
	var req BulkTranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.translationService.BulkTranslateComments(req.CommentIDs, req.TargetLanguages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk comment translation started",
		"comment_count": len(req.CommentIDs),
		"languages": req.TargetLanguages,
	})
}

// GetCacheStats returns translation cache statistics
// @Summary Get Cache Stats
// @Description Get translation cache statistics
// @Tags translation
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/translations/cache/stats [get]
func (h *TranslationHandler) GetCacheStats(c *gin.Context) {
	stats := h.translationCache.GetCacheStats()
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
		"timestamp": h.GetCurrentTime(),
	})
}

// InvalidateCache invalidates translation cache
// @Summary Invalidate Cache
// @Description Invalidate translation cache by pattern
// @Tags translation
// @Accept json
// @Produce json
// @Param pattern query string true "Cache pattern to invalidate"
// @Success 200 {object} map[string]interface{}
// @Router /api/translations/cache/invalidate [post]
func (h *TranslationHandler) InvalidateCache(c *gin.Context) {
	pattern := c.Query("pattern")
	if pattern == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pattern is required"})
		return
	}

	err := h.translationCache.InvalidateCache(pattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache invalidated",
		"pattern": pattern,
	})
}

// Request/Response structures
type TranslationRequest struct {
	TargetLanguages []string `json:"target_languages" binding:"required"`
}

type BulkTranslationRequest struct {
	CommentIDs      []uint   `json:"comment_ids" binding:"required"`
	TargetLanguages []string `json:"target_languages" binding:"required"`
}

// Helper method
func (h *TranslationHandler) GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
