package models

// UITranslationResponse represents a UI translation response
type UITranslationResponse struct {
	Language    string `json:"language"`
	MessageID   string `json:"message_id"`
	Translation string `json:"translation"`
}

// SupportedLanguagesResponse represents supported languages response
type SupportedLanguagesResponse struct {
	DefaultLanguage      string   `json:"default_language"`
	SupportedLanguages   []string `json:"supported_languages"`
	AITranslationEnabled bool     `json:"ai_translation_enabled"`
}

// AITranslationRequest represents a request for AI translation
type AITranslationRequest struct {
	EntityType      string   `json:"entity_type" binding:"required"`
	EntityID        uint     `json:"entity_id" binding:"required"`
	SourceLanguage  string   `json:"source_language,omitempty"`
	TargetLanguages []string `json:"target_languages" binding:"required"`
	Priority        int      `json:"priority,omitempty"` // 1=low, 2=normal, 3=high
}

// TranslationJobResponse represents a translation job response
type TranslationJobResponse struct {
	JobID           string   `json:"job_id"`
	EntityType      string   `json:"entity_type"`
	EntityID        uint     `json:"entity_id"`
	TargetLanguages []string `json:"target_languages"`
	Status          string   `json:"status"`
	Message         string   `json:"message"`
	Progress        int      `json:"progress,omitempty"`
	CreatedAt       string   `json:"created_at,omitempty"`
	CompletedAt     string   `json:"completed_at,omitempty"`
	Error           string   `json:"error,omitempty"`
}
