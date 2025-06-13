package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"news/internal/json"
	"news/internal/models"
	"news/internal/tracing"

	"go.opentelemetry.io/otel/attribute"
)

// Global AI service instance
var aiServiceInstance *AIService

// AIService handles AI-powered content generation and assistance
type AIService struct {
	openAIKey   string
	openAIURL   string
	httpClient  *http.Client
	model       string
	maxTokens   int
	temperature float32
}

// GetAIService returns the global AI service instance
func GetAIService() *AIService {
	if aiServiceInstance == nil {
		aiServiceInstance = NewAIService()
	}
	return aiServiceInstance
}

// NewAIService creates a new AI service instance
func NewAIService() *AIService {
	return &AIService{
		openAIKey:   os.Getenv("OPENAI_API_KEY"),
		openAIURL:   "https://api.openai.com/v1/chat/completions",
		model:       getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo"),
		maxTokens:   getEnvIntOrDefault("OPENAI_MAX_TOKENS", 1000),
		temperature: getEnvFloatOrDefault("OPENAI_TEMPERATURE", 0.7),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenAI API request/response structures
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Embedding API structures
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// GenerateHeadlines creates multiple headline suggestions for given content
func (ai *AIService) GenerateHeadlines(ctx context.Context, content string, count int, style string) ([]string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.generate_headlines",
		attribute.Int("content_length", len(content)),
		attribute.Int("requested_count", count),
		attribute.String("style", style))
	defer span.End()

	if ai.openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	stylePrompt := ""
	switch style {
	case "clickbait":
		stylePrompt = "Dikkat çekici ve merak uyandıran"
	case "formal":
		stylePrompt = "Resmi ve ciddi"
	case "casual":
		stylePrompt = "Rahat ve samimi"
	default:
		stylePrompt = "Haber tarzında objektif"
	}

	prompt := fmt.Sprintf(`Aşağıdaki haber içeriği için %d adet %s başlık öner. Her başlığı ayrı satırda yaz:

İçerik:
%s`, count, stylePrompt, content)

	response, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate headlines: %w", err)
	}

	headlines := ai.parseLines(response)
	if len(headlines) > count {
		headlines = headlines[:count]
	}

	span.SetAttributes(attribute.Int("generated_count", len(headlines)))
	return headlines, nil
}

// GenerateContent creates article content based on topic and parameters
func (ai *AIService) GenerateContent(ctx context.Context, topic string, style string, length string, keywords []string, perspective string) (string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.generate_content",
		attribute.String("topic", topic),
		attribute.String("style", style),
		attribute.String("length", length),
		attribute.StringSlice("keywords", keywords))
	defer span.End()

	if ai.openAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	wordCount := 300
	switch length {
	case "short":
		wordCount = 200
	case "long":
		wordCount = 600
	}

	keywordsStr := strings.Join(keywords, ", ")
	prompt := fmt.Sprintf(`%s konusunda %s tarzında %d kelimelik bir haber makalesi yaz.

Gereksinimler:
- Perspektif: %s
- Anahtar kelimeler: %s
- SEO uyumlu
- Türkçe dilbilgisi kurallarına uygun
- Gerçekçi ve güvenilir

Sadece makale içeriğini döndür:

`, topic, style, wordCount, perspective, keywordsStr)

	content, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	span.SetAttributes(attribute.Int("generated_content_length", len(content)))
	return content, nil
}

// ImproveContent provides suggestions to improve existing content
func (ai *AIService) ImproveContent(ctx context.Context, content string, goals []string, targetLevel string) (string, []models.ContentImprovement, float64, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.improve_content",
		attribute.Int("content_length", len(content)),
		attribute.StringSlice("goals", goals),
		attribute.String("target_level", targetLevel))
	defer span.End()

	if ai.openAIKey == "" {
		return "", nil, 0, fmt.Errorf("OpenAI API key not configured")
	}

	goalsStr := strings.Join(goals, ", ")
	prompt := fmt.Sprintf(`Aşağıdaki metni şu hedefler doğrultusunda iyileştir: %s
Hedef seviye: %s

JSON formatında yanıt ver:
{
  "improved_content": "İyileştirilmiş metin",
  "suggestions": [
    {
      "type": "grammar",
      "original": "orijinal metin",
      "suggestion": "önerilen değişiklik",
      "explanation": "açıklama",
      "impact": "high"
    }
  ],
  "quality_score": 0.85
}

Metin:
%s`, goalsStr, targetLevel, content)

	response, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return "", nil, 0, fmt.Errorf("failed to improve content: %w", err)
	}

	// Parse the response
	var result struct {
		ImprovedContent string                      `json:"improved_content"`
		Suggestions     []models.ContentImprovement `json:"suggestions"`
		QualityScore    float64                     `json:"quality_score"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback: return original content with default values
		return content, []models.ContentImprovement{}, 0.5, nil
	}

	span.SetAttributes(
		attribute.Float64("quality_score", result.QualityScore),
		attribute.Int("suggestions_count", len(result.Suggestions)))

	return result.ImprovedContent, result.Suggestions, result.QualityScore, nil
}

// ModerateComment checks if a comment is appropriate
func (ai *AIService) ModerateComment(ctx context.Context, comment string, strict bool) (bool, float64, string, []models.ModerationCategory, string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.moderate_comment",
		attribute.Int("comment_length", len(comment)),
		attribute.Bool("strict_mode", strict))
	defer span.End()

	if ai.openAIKey == "" {
		return false, 0, "", nil, "", fmt.Errorf("OpenAI API key not configured")
	}

	strictnessLevel := "normal"
	if strict {
		strictnessLevel = "katı"
	}

	prompt := fmt.Sprintf(`Aşağıdaki yorumu %s modda analiz et. JSON formatında yanıt ver:

{
  "is_approved": true,
  "confidence": 0.95,
  "reason": "Nezaketli ve uygun dil kullanımı",
  "categories": [
    {
      "category": "safe",
      "confidence": 0.95,
      "severity": "low"
    }
  ],
  "severity": "low"
}

Kategoriler: safe, spam, offensive, hate_speech, inappropriate, violence
Şiddet seviyeleri: low, medium, high, critical

Yorum:
%s`, strictnessLevel, comment)

	response, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return false, 0, "", nil, "", fmt.Errorf("failed to moderate comment: %w", err)
	}

	var result struct {
		IsApproved bool                        `json:"is_approved"`
		Confidence float64                     `json:"confidence"`
		Reason     string                      `json:"reason"`
		Categories []models.ModerationCategory `json:"categories"`
		Severity   string                      `json:"severity"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback: flag for manual review
		return false, 0.5, "Analiz edilemedi, manuel inceleme gerekli",
			[]models.ModerationCategory{{Category: "unknown", Confidence: 0.5, Severity: "medium"}},
			"medium", nil
	}

	span.SetAttributes(
		attribute.Bool("is_approved", result.IsApproved),
		attribute.Float64("confidence", result.Confidence),
		attribute.String("severity", result.Severity))

	return result.IsApproved, result.Confidence, result.Reason, result.Categories, result.Severity, nil
}

// SummarizeContent creates a summary of the given content
func (ai *AIService) SummarizeContent(ctx context.Context, content string, length string, style string, language string) (string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.summarize_content",
		attribute.Int("content_length", len(content)),
		attribute.String("length", length),
		attribute.String("style", style),
		attribute.String("language", language))
	defer span.End()

	if ai.openAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	lengthDesc := "orta uzunlukta"
	switch length {
	case "short":
		lengthDesc = "kısa"
	case "long":
		lengthDesc = "uzun"
	}

	styleDesc := "paragraf halinde"
	switch style {
	case "bullet":
		styleDesc = "madde madde"
	case "highlight":
		styleDesc = "ana noktalar şeklinde"
	}

	langDesc := "Türkçe"
	if language == "en" {
		langDesc = "İngilizce"
	}

	prompt := fmt.Sprintf(`Aşağıdaki metni %s %s %s olarak özetle:

Metin:
%s`, langDesc, lengthDesc, styleDesc, content)

	summary, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to summarize content: %w", err)
	}

	span.SetAttributes(attribute.Int("summary_length", len(summary)))
	return summary, nil
}

// CategorizeContent suggests categories for the given content
func (ai *AIService) CategorizeContent(ctx context.Context, content string, availableCategories []string) ([]models.CategorySuggestion, []models.TagSuggestion, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.categorize_content",
		attribute.Int("content_length", len(content)),
		attribute.Int("available_categories", len(availableCategories)))
	defer span.End()

	if ai.openAIKey == "" {
		return nil, nil, fmt.Errorf("OpenAI API key not configured")
	}

	categoriesStr := strings.Join(availableCategories, ", ")
	prompt := fmt.Sprintf(`Aşağıdaki haber makalesini analiz et ve JSON formatında yanıt ver:

{
  "categories": [
    {
      "name": "Teknoloji",
      "confidence": 0.95,
      "reason": "İçerik teknoloji ile ilgili"
    }
  ],
  "tags": [
    {
      "name": "yapay zeka",
      "confidence": 0.90,
      "relevance": "high"
    }
  ]
}

Mevcut kategoriler: %s

Makale:
%s`, categoriesStr, content)

	response, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to categorize content: %w", err)
	}

	var result struct {
		Categories []models.CategorySuggestion `json:"categories"`
		Tags       []models.TagSuggestion      `json:"tags"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Fallback: return empty suggestions
		return []models.CategorySuggestion{}, []models.TagSuggestion{}, nil
	}

	span.SetAttributes(
		attribute.Int("categories_count", len(result.Categories)),
		attribute.Int("tags_count", len(result.Tags)))

	return result.Categories, result.Tags, nil
}

// GenerateTags suggests tags for the given content
func (ai *AIService) GenerateTags(ctx context.Context, content string, maxTags int) ([]string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.generate_tags",
		attribute.Int("content_length", len(content)),
		attribute.Int("max_tags", maxTags))
	defer span.End()

	if ai.openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	prompt := fmt.Sprintf(`Aşağıdaki haber makalesi için maksimum %d adet etiket öner. Etiketler:
- İçeriği en iyi temsil etmeli
- SEO uyumlu olmalı
- 1-3 kelime arasında olmalı
- Türkçe olmalı

Her etiketi yeni satırda listele.

Makale:
%s`, maxTags, content)

	response, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tags: %w", err)
	}

	tags := ai.parseLines(response)
	if len(tags) > maxTags {
		tags = tags[:maxTags]
	}

	span.SetAttributes(attribute.Int("generated_tags", len(tags)))
	return tags, nil
}

// TranslateText translates text from source language to target language
func (ai *AIService) TranslateText(ctx context.Context, text, sourceLanguage, targetLanguage string) (string, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.translate_text",
		attribute.Int("text_length", len(text)),
		attribute.String("source_language", sourceLanguage),
		attribute.String("target_language", targetLanguage))
	defer span.End()

	if ai.openAIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	// Language code mapping for better prompts
	langNames := map[string]string{
		"tr": "Türkçe",
		"en": "İngilizce",
		"es": "İspanyolca",
		"fr": "Fransızca",
		"de": "Almanca",
		"ar": "Arapça",
		"zh": "Çince",
		"ru": "Rusça",
		"ja": "Japonca",
		"ko": "Korece",
	}

	sourceLangName := langNames[sourceLanguage]
	if sourceLangName == "" {
		sourceLangName = sourceLanguage
	}

	targetLangName := langNames[targetLanguage]
	if targetLangName == "" {
		targetLangName = targetLanguage
	}

	prompt := fmt.Sprintf(`Aşağıdaki metni %s dilinden %s diline çevir. Sadece çevirinin kendisini döndür, başka açıklama yapma:

%s`, sourceLangName, targetLangName, text)

	translation, err := ai.callOpenAI(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to translate text: %w", err)
	}

	span.SetAttributes(attribute.Int("translation_length", len(translation)))
	return translation, nil
}

// GenerateEmbedding generates embeddings for the given text using OpenAI
func (ai *AIService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.generate_embedding",
		attribute.String("text_length", fmt.Sprintf("%d", len(text))))
	defer span.End()

	if ai.openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Use text-embedding-3-small model for cost efficiency
	embeddingModel := getEnvOrDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")

	request := EmbeddingRequest{
		Input: []string{text},
		Model: embeddingModel,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	embeddingURL := "https://api.openai.com/v1/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", embeddingURL, bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.openAIKey)

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close embedding API response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("embedding API error: %s", string(body)))
		return nil, fmt.Errorf("embedding API error: %s", string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	embedding := embeddingResp.Data[0].Embedding
	span.SetAttributes(
		attribute.Int("embedding_dimensions", len(embedding)),
		attribute.String("model_used", embeddingResp.Model),
		attribute.Int("total_tokens", embeddingResp.Usage.TotalTokens))

	return embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts efficiently
func (ai *AIService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	ctx, span := tracing.StartSpanWithAttributes(ctx, "ai_service.generate_batch_embeddings",
		attribute.Int("batch_size", len(texts)))
	defer span.End()

	if ai.openAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	embeddingModel := getEnvOrDefault("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small")

	request := EmbeddingRequest{
		Input: texts,
		Model: embeddingModel,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to marshal batch embedding request: %w", err)
	}

	embeddingURL := "https://api.openai.com/v1/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", embeddingURL, bytes.NewBuffer(requestBody))
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create batch embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.openAIKey)

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to call batch embedding API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close batch embedding API response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read batch embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("batch embedding API error: %s", string(body)))
		return nil, fmt.Errorf("batch embedding API error: %s", string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.Unmarshal(body, &embeddingResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal batch embedding response: %w", err)
	}

	if len(embeddingResp.Data) != len(texts) {
		return nil, fmt.Errorf("mismatch between input texts (%d) and embedding data (%d)", len(texts), len(embeddingResp.Data))
	}

	embeddings := make([][]float64, len(embeddingResp.Data))
	for _, data := range embeddingResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	span.SetAttributes(
		attribute.Int("embeddings_generated", len(embeddings)),
		attribute.String("model_used", embeddingResp.Model),
		attribute.Int("total_tokens", embeddingResp.Usage.TotalTokens))

	return embeddings, nil
}

// Helper methods

func (ai *AIService) callOpenAI(ctx context.Context, prompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model:       ai.model,
		MaxTokens:   ai.maxTokens,
		Temperature: ai.temperature,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ai.openAIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.openAIKey)

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close OpenAI API response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return strings.TrimSpace(openAIResp.Choices[0].Message.Content), nil
}

func (ai *AIService) parseLines(text string) []string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Remove numbering or bullets
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "* ")
			// Remove numbered lists
			if len(line) > 2 && line[1] == '.' {
				line = strings.TrimSpace(line[2:])
			}
			if line != "" {
				result = append(result, line)
			}
		}
	}
	return result
}

// Environment helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloatOrDefault(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatVal)
		}
	}
	return defaultValue
}
