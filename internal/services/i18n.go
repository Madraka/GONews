package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"news/internal/json"
)

// I18nService handles internationalization for notifications and messages
type I18nService struct {
	translations map[string]map[string]interface{}
	mutex        sync.RWMutex
	defaultLang  string
}

// Translation data structure
type Translation struct {
	Other string `json:"other"`
}

// Global i18n service instance
var globalI18n *I18nService
var once sync.Once

// InitI18nService initializes the global i18n service
func InitI18nService(localesPath string) error {
	var err error
	once.Do(func() {
		globalI18n = &I18nService{
			translations: make(map[string]map[string]interface{}),
			defaultLang:  "en",
		}
		err = globalI18n.LoadTranslations(localesPath)
	})
	return err
}

// GetI18nService returns the global i18n service instance
func GetI18nService() *I18nService {
	return globalI18n
}

// LoadTranslations loads translation files from the specified directory
func (i *I18nService) LoadTranslations(localesPath string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// Load all JSON files in the locales directory
	files, err := os.ReadDir(localesPath)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract language code from filename (e.g., "en.json" -> "en")
		lang := strings.TrimSuffix(file.Name(), ".json")

		filePath := filepath.Join(localesPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("❌ Failed to read translation file %s: %v", filePath, err)
			continue
		}

		var translations map[string]interface{}
		if err := json.Unmarshal(content, &translations); err != nil {
			log.Printf("❌ Failed to parse translation file %s: %v", filePath, err)
			continue
		}

		i.translations[lang] = translations
		log.Printf("✅ Loaded translations for language: %s", lang)
	}

	return nil
}

// Translate translates a key to the specified language with optional template data
func (i *I18nService) Translate(lang, key string, data interface{}) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// Try the requested language first
	if translation := i.getTranslation(lang, key); translation != "" {
		return i.processTemplate(translation, data)
	}

	// Fall back to default language
	if lang != i.defaultLang {
		if translation := i.getTranslation(i.defaultLang, key); translation != "" {
			return i.processTemplate(translation, data)
		}
	}

	// Return the key itself if no translation found
	return key
}

// getTranslation retrieves a translation for a specific language and key
func (i *I18nService) getTranslation(lang, key string) string {
	langTranslations, exists := i.translations[lang]
	if !exists {
		return ""
	}

	// Navigate through nested keys (e.g., "notifications.welcome.title")
	keys := strings.Split(key, ".")
	current := langTranslations

	for _, k := range keys {
		if next, ok := current[k]; ok {
			switch v := next.(type) {
			case map[string]interface{}:
				// Check if this is a final translation object with "other" key
				if otherValue, hasOther := v["other"]; hasOther {
					if str, isString := otherValue.(string); isString {
						return str
					}
				}
				// Otherwise, continue navigating deeper
				current = v
			default:
				return ""
			}
		} else {
			return ""
		}
	}

	// Check if the final value has an "other" key
	if otherValue, hasOther := current["other"]; hasOther {
		if str, isString := otherValue.(string); isString {
			return str
		}
	}

	return ""
}

// processTemplate processes Go template strings with provided data
func (i *I18nService) processTemplate(text string, data interface{}) string {
	if data == nil {
		return text
	}

	tmpl, err := template.New("notification").Parse(text)
	if err != nil {
		log.Printf("❌ Template parsing error: %v", err)
		return text
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		log.Printf("❌ Template execution error: %v", err)
		return text
	}

	return result.String()
}

// GetSupportedLanguages returns a list of supported language codes
func (i *I18nService) GetSupportedLanguages() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var languages []string
	for lang := range i.translations {
		languages = append(languages, lang)
	}
	return languages
}

// SetDefaultLanguage sets the default fallback language
func (i *I18nService) SetDefaultLanguage(lang string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.defaultLang = lang
}

// Convenience functions for common notification types

// TranslateNotification translates notification title and message
func (i *I18nService) TranslateNotification(lang, notificationType string, data interface{}) (string, string) {
	titleKey := fmt.Sprintf("notifications.%s.title", notificationType)
	messageKey := fmt.Sprintf("notifications.%s.message", notificationType)

	title := i.Translate(lang, titleKey, data)
	message := i.Translate(lang, messageKey, data)

	return title, message
}

// Global convenience function
func TranslateNotification(lang, notificationType string, data interface{}) (string, string) {
	if globalI18n == nil {
		return notificationType, fmt.Sprintf("%v", data)
	}
	return globalI18n.TranslateNotification(lang, notificationType, data)
}

// Translate is a global convenience function
func Translate(lang, key string, data interface{}) string {
	if globalI18n == nil {
		return key
	}
	return globalI18n.Translate(lang, key, data)
}
