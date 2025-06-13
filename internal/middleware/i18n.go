package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const (
	LanguageContextKey  = "language"
	LocalizerContextKey = "localizer"
	DefaultLanguage     = "en"
)

// SupportedLanguages defines the languages supported by the application
var SupportedLanguages = []string{"en", "tr", "es"}

// I18nMiddleware creates middleware for internationalization
func I18nMiddleware(bundle *i18n.Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Try to get language from query parameter
		lang := c.Query("lang")

		// 2. If not in query, try to get from Accept-Language header
		if lang == "" {
			acceptLanguage := c.GetHeader("Accept-Language")
			lang = parseAcceptLanguage(acceptLanguage)
		}

		// 3. Validate language and fallback to default if invalid
		if !isValidLanguage(lang) {
			lang = DefaultLanguage
		}

		// Create localizer for the detected language
		localizer := i18n.NewLocalizer(bundle, lang)

		// Store language and localizer in context
		c.Set(LanguageContextKey, lang)
		c.Set(LocalizerContextKey, localizer)

		// Add language to response headers for debugging
		c.Header("Content-Language", lang)

		c.Next()
	}
}

// parseAcceptLanguage parses Accept-Language header and returns the best matching language
func parseAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return DefaultLanguage
	}

	// Parse Accept-Language header
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		return DefaultLanguage
	}

	// Find the best matching supported language
	matcher := language.NewMatcher(getSupportedLanguageTags())
	_, index, _ := matcher.Match(tags...)

	if index < len(SupportedLanguages) {
		return SupportedLanguages[index]
	}

	return DefaultLanguage
}

// getSupportedLanguageTags converts supported language strings to language.Tag slice
func getSupportedLanguageTags() []language.Tag {
	tags := make([]language.Tag, len(SupportedLanguages))
	for i, lang := range SupportedLanguages {
		tags[i] = language.MustParse(lang)
	}
	return tags
}

// isValidLanguage checks if the given language is supported
func isValidLanguage(lang string) bool {
	for _, supported := range SupportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetLanguage extracts language from gin context
func GetLanguage(c *gin.Context) string {
	if lang, exists := c.Get(LanguageContextKey); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return DefaultLanguage
}

// GetLocalizer extracts localizer from gin context
func GetLocalizer(c *gin.Context) *i18n.Localizer {
	if localizer, exists := c.Get(LocalizerContextKey); exists {
		if loc, ok := localizer.(*i18n.Localizer); ok {
			return loc
		}
	}
	return nil
}

// LocalizeMessage translates a message using the localizer from context
func LocalizeMessage(c *gin.Context, messageID string, templateData map[string]interface{}) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		return messageID // Fallback to message ID if no localizer
	}

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

// LocalizeError creates a localized error response
func LocalizeError(c *gin.Context, messageID string, templateData map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"error":   true,
		"message": LocalizeMessage(c, messageID, templateData),
		"code":    messageID,
	}
}

// LocalizeSuccess creates a localized success response
func LocalizeSuccess(c *gin.Context, messageID string, templateData map[string]interface{}, data interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"success": true,
		"message": LocalizeMessage(c, messageID, templateData),
	}

	if data != nil {
		response["data"] = data
	}

	return response
}

// RespondWithLocalizedError sends a localized error response
func RespondWithLocalizedError(c *gin.Context, statusCode int, messageID string, templateData map[string]interface{}) {
	c.JSON(statusCode, LocalizeError(c, messageID, templateData))
}

// RespondWithLocalizedSuccess sends a localized success response
func RespondWithLocalizedSuccess(c *gin.Context, statusCode int, messageID string, templateData map[string]interface{}, data interface{}) {
	c.JSON(statusCode, LocalizeSuccess(c, messageID, templateData, data))
}

// LanguageValidationMiddleware validates language parameter in routes
func LanguageValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if lang := c.Param("lang"); lang != "" {
			if !isValidLanguage(lang) {
				RespondWithLocalizedError(c, http.StatusBadRequest, "errors.general.bad_request", map[string]interface{}{
					"Details": fmt.Sprintf("Unsupported language: %s. Supported languages: %s", lang, strings.Join(SupportedLanguages, ", ")),
				})
				c.Abort()
				return
			}
			// Override context language with URL parameter
			c.Set(LanguageContextKey, lang)
		}
		c.Next()
	}
}
