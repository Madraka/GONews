package config

import (
	"os"
)

// TranslationConfig holds all translation-related configuration
type TranslationConfig struct {
	DefaultLanguage     string   `json:"default_language"`
	FallbackLanguage    string   `json:"fallback_language"`
	SupportedLanguages  []string `json:"supported_languages"`
	EnableAITranslation bool     `json:"enable_ai_translation"`
	CacheEnabled        bool     `json:"cache_enabled"`
	LocalesPath         string   `json:"locales_path"`
}

// GetTranslationConfig returns the global translation configuration
func GetTranslationConfig() *TranslationConfig {
	return &TranslationConfig{
		DefaultLanguage:     getEnvOrDefault("DEFAULT_LANGUAGE", "en"),
		FallbackLanguage:    getEnvOrDefault("FALLBACK_LANGUAGE", "en"),
		SupportedLanguages:  []string{"en", "tr", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"},
		EnableAITranslation: getEnvOrDefault("ENABLE_AI_TRANSLATION", "true") == "true",
		CacheEnabled:        getEnvOrDefault("TRANSLATION_CACHE_ENABLED", "true") == "true",
		LocalesPath:         getEnvOrDefault("LOCALES_PATH", "./locales"),
	}
}

// ValidateLanguage checks if a language is supported
func (tc *TranslationConfig) ValidateLanguage(lang string) bool {
	for _, supported := range tc.SupportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetSourceLanguage returns the source language for AI translations
// This is typically Turkish as it's the original content language
func (tc *TranslationConfig) GetSourceLanguage() string {
	return "tr" // Content is originally in Turkish
}

// IsContentTranslationEnabled checks if content translation is enabled
func (tc *TranslationConfig) IsContentTranslationEnabled() bool {
	return tc.EnableAITranslation
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LanguagePriority defines the priority order for language selection
var LanguagePriority = map[string]int{
	"en": 1, // Primary
	"tr": 2, // Original content language
	"es": 3, // High demand
	"fr": 4,
	"de": 5,
	"ar": 6,
	"zh": 7,
	"ru": 8,
	"ja": 9,
	"ko": 10,
}

// GetLanguagePriority returns the priority of a language
func GetLanguagePriority(lang string) int {
	if priority, exists := LanguagePriority[lang]; exists {
		return priority
	}
	return 999 // Unknown languages have lowest priority
}
