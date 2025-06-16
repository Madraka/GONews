package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// UnifiedTranslationHandler handles all translation operations using the unified translation service
type UnifiedTranslationHandler struct {
	translationService *services.UnifiedTranslationService
}

// NewUnifiedTranslationHandler creates a new unified translation handler
func NewUnifiedTranslationHandler() *UnifiedTranslationHandler {
	return &UnifiedTranslationHandler{
		translationService: services.GetUnifiedTranslationService(),
	}
}

// TranslateUI godoc
// @Summary Translate UI message
// @Description Translate a UI message to specified language
// @Tags Translation
// @Accept json
// @Produce json
// @Param language path string true "Target language code (e.g., 'en', 'tr', 'es')"
// @Param message_id query string true "Message ID to translate"
// @Param template_data body map[string]interface{} false "Template data for message interpolation"
// @Success 200 {object} models.UITranslationResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/translations/ui/{language} [post]
func (h *UnifiedTranslationHandler) TranslateUI(c *gin.Context) {
	language := c.Param("language")
	messageID := c.Query("message_id")

	if language == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Language and message_id are required",
		})
		return
	}

	var templateData map[string]interface{}
	if err := c.ShouldBindJSON(&templateData); err != nil {
		templateData = make(map[string]interface{})
	}

	if h.translationService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Translation service not available",
		})
		return
	}

	translatedMessage, err := h.translationService.TranslateUI(language, messageID, templateData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("Translation failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.UITranslationResponse{
		Language:    language,
		MessageID:   messageID,
		Translation: translatedMessage,
	})
}

// TranslateContent godoc
// @Summary Translate dynamic content
// @Description Translate dynamic content (article, category, tag, etc.) to specified language
// @Tags Translation
// @Accept json
// @Produce json
// @Param entity_type path string true "Entity type (article, category, tag, menu)"
// @Param entity_id path int true "Entity ID"
// @Param language path string true "Target language code"
// @Success 200 {object} interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/translations/content/{entity_type}/{entity_id}/{language} [get]
func (h *UnifiedTranslationHandler) TranslateContent(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")
	language := c.Param("language")

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid entity ID",
		})
		return
	}

	if h.translationService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Translation service not available",
		})
		return
	}

	translatedContent, err := h.translationService.TranslateContent(entityType, uint(entityID), language)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: fmt.Sprintf("%s not found", entityType),
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: fmt.Sprintf("Translation failed: %v", err),
			})
		}
		return
	}

	c.JSON(http.StatusOK, translatedContent)
}

// GetSupportedLanguages godoc
// @Summary Get supported languages
// @Description Get list of supported languages for the application
// @Tags Translation
// @Produce json
// @Success 200 {object} models.SupportedLanguagesResponse
// @Router /api/translations/languages [get]
func (h *UnifiedTranslationHandler) GetSupportedLanguages(c *gin.Context) {
	if h.translationService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Translation service not available",
		})
		return
	}

	config := h.translationService.GetConfig()

	response := models.SupportedLanguagesResponse{
		DefaultLanguage:      config.DefaultLanguage,
		SupportedLanguages:   config.SupportedLanguages,
		AITranslationEnabled: config.EnableAITranslation,
	}

	c.JSON(http.StatusOK, response)
}

// RequestAITranslation godoc
// @Summary Request AI translation
// @Description Request AI-powered translation for content
// @Tags Translation
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.AITranslationRequest true "AI translation request"
// @Success 202 {object} models.TranslationJobResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/translations/ai [post]
func (h *UnifiedTranslationHandler) RequestAITranslation(c *gin.Context) {
	var request models.AITranslationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	if h.translationService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Translation service not available",
		})
		return
	}

	// Validate request
	if request.EntityType == "" || request.EntityID == 0 || len(request.TargetLanguages) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Entity type, entity ID, and target languages are required",
		})
		return
	}

	jobID, err := h.translationService.RequestAITranslation(
		request.EntityType,
		request.EntityID,
		request.SourceLanguage,
		request.TargetLanguages,
		request.Priority,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: fmt.Sprintf("Failed to queue AI translation: %v", err),
		})
		return
	}

	c.JSON(http.StatusAccepted, models.TranslationJobResponse{
		JobID:           jobID,
		EntityType:      request.EntityType,
		EntityID:        request.EntityID,
		TargetLanguages: request.TargetLanguages,
		Status:          "queued",
		Message:         "AI translation job queued successfully",
	})
}

// GetTranslationStatus godoc
// @Summary Get translation job status
// @Description Get status of a translation job
// @Tags Translation
// @Produce json
// @Param job_id path string true "Translation job ID"
// @Success 200 {object} models.TranslationJobResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/translations/status/{job_id} [get]
func (h *UnifiedTranslationHandler) GetTranslationStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	if h.translationService == nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Translation service not available",
		})
		return
	}

	status, err := h.translationService.GetTranslationJobStatus(jobID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Translation job not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error: fmt.Sprintf("Failed to get translation status: %v", err),
			})
		}
		return
	}

	c.JSON(http.StatusOK, status)
}
