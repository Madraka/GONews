package processors

import (
	"context"
	"fmt"
	"log"

	"news/internal/queue"
	"news/internal/services"
)

// TranslationProcessor handles translation jobs for Redis queue
type TranslationProcessor struct {
	translationService *services.AITranslationService
}

// NewTranslationProcessor creates a new translation processor
func NewTranslationProcessor(translationService *services.AITranslationService) *TranslationProcessor {
	return &TranslationProcessor{
		translationService: translationService,
	}
}

// ProcessJob processes a translation job
func (tp *TranslationProcessor) ProcessJob(ctx context.Context, job *queue.Job) error {
	log.Printf("Processing translation job %s", job.ID)

	// Extract parameters from job payload
	entityType, ok := job.Payload["entity_type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid entity_type in job payload")
	}

	entityIDFloat, ok := job.Payload["entity_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid entity_id in job payload")
	}
	entityID := uint(entityIDFloat)

	sourceLang, ok := job.Payload["source_lang"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid source_lang in job payload")
	}

	targetLang, ok := job.Payload["target_lang"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid target_lang in job payload")
	}

	// Process based on entity type
	switch entityType {
	case "article":
		return tp.processArticleTranslation(entityID, sourceLang, targetLang)
	case "category":
		return tp.translationService.TranslateCategory(entityID, []string{targetLang})
	case "tag":
		return tp.translationService.TranslateTag(entityID, []string{targetLang})
	case "menu":
		return tp.translationService.TranslateMenu(entityID, []string{targetLang})
	case "notification":
		return tp.translationService.TranslateNotification(entityID, []string{targetLang})
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

// GetJobTypes returns the job types this processor handles
func (tp *TranslationProcessor) GetJobTypes() []string {
	return []string{"translation", "translate_article", "translate_category", "translate_tag"}
}

// processArticleTranslation handles article translation specifically
func (tp *TranslationProcessor) processArticleTranslation(entityID uint, sourceLang, targetLang string) error {
	// Use existing translation service method
	return tp.translationService.TranslateArticle(entityID, []string{targetLang})
}

// Helper function to create translation jobs
func CreateTranslationJob(entityType string, entityID uint, sourceLang, targetLang string, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "translation",
		Priority: priority,
		Payload: map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
			"source_lang": sourceLang,
			"target_lang": targetLang,
		},
		MaxAttempts: 3,
	}
}
