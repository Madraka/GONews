package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ContentSuggestion represents AI-generated content suggestions
type ContentSuggestion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Type       string    `gorm:"size:50;not null" json:"type"` // headline, content, summary, tags
	Input      string    `gorm:"type:text;not null" json:"input"`
	Suggestion string    `gorm:"type:text;not null" json:"suggestion"`
	Context    string    `gorm:"type:json" json:"context"` // Additional context data
	Confidence float64   `gorm:"type:decimal(3,2)" json:"confidence"`
	UserID     uint      `gorm:"index" json:"user_id"`
	ArticleID  *uint     `gorm:"index" json:"article_id"`
	Used       bool      `gorm:"default:false" json:"used"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	User    User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// ModerationResult represents AI content moderation results
type ModerationResult struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ContentType string     `gorm:"size:20;not null" json:"content_type"` // comment, article
	ContentID   uint       `gorm:"not null" json:"content_id"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	IsApproved  bool       `gorm:"default:false" json:"is_approved"`
	Confidence  float64    `gorm:"type:decimal(3,2)" json:"confidence"`
	Reason      string     `gorm:"type:text" json:"reason"`
	Categories  string     `gorm:"type:json" json:"categories"`           // JSON array of flagged categories
	Severity    string     `gorm:"size:20;default:'low'" json:"severity"` // low, medium, high, critical
	ReviewedBy  *uint      `gorm:"index" json:"reviewed_by"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	Reviewer *User `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

// ValidateSeverity validates moderation severity
func (mr *ModerationResult) ValidateSeverity() bool {
	validSeverities := []string{"low", "medium", "high", "critical"}
	for _, severity := range validSeverities {
		if mr.Severity == severity {
			return true
		}
	}
	return false
}

// ContentAnalysis represents AI content analysis results
type ContentAnalysis struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ArticleID    uint      `gorm:"not null;index" json:"article_id"`
	ReadingLevel string    `gorm:"size:20" json:"reading_level"` // beginner, intermediate, advanced
	Sentiment    string    `gorm:"size:20" json:"sentiment"`     // positive, negative, neutral
	Keywords     string    `gorm:"type:json" json:"keywords"`    // JSON array of extracted keywords
	Categories   string    `gorm:"type:json" json:"categories"`  // JSON array of suggested categories
	Tags         string    `gorm:"type:json" json:"tags"`        // JSON array of suggested tags
	Summary      string    `gorm:"type:text" json:"summary"`
	ReadTime     int       `gorm:"default:0" json:"read_time"`       // estimated read time in minutes
	Quality      float64   `gorm:"type:decimal(3,2)" json:"quality"` // quality score 0-1
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// AgentTask represents tasks for the n8n Agent API
type AgentTask struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TaskType    string         `gorm:"size:50;not null" json:"task_type"`                // generate_headlines, moderate_content, etc.
	Status      string         `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, processing, completed, failed
	Priority    int            `gorm:"default:0" json:"priority"`                        // 0=low, 1=medium, 2=high, 3=urgent
	InputData   datatypes.JSON `gorm:"not null" json:"input_data" swaggertype:"object"`
	OutputData  datatypes.JSON `json:"output_data" swaggertype:"object"`
	ErrorMsg    string         `gorm:"type:text" json:"error_msg"`
	WebhookURL  string         `gorm:"size:500" json:"webhook_url"` // n8n callback URL
	RetryCount  int            `gorm:"default:0" json:"retry_count"`
	MaxRetries  int            `gorm:"default:3" json:"max_retries"`
	RequestedBy uint           `gorm:"index" json:"requested_by"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:RequestedBy" json:"user,omitempty"`
}

// ValidateStatus validates agent task status
func (at *AgentTask) ValidateStatus() bool {
	validStatuses := []string{"pending", "processing", "completed", "failed"}
	for _, status := range validStatuses {
		if at.Status == status {
			return true
		}
	}
	return false
}

// ValidateTaskType validates agent task type
func (at *AgentTask) ValidateTaskType() bool {
	validTypes := []string{
		"generate_headlines", "generate_content", "improve_content",
		"moderate_comment", "summarize_content", "categorize_content",
		"generate_tags", "analyze_content", "batch_moderate",
	}
	for _, taskType := range validTypes {
		if at.TaskType == taskType {
			return true
		}
	}
	return false
}

// AIUsageStats represents AI service usage statistics
type AIUsageStats struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	ServiceType  string    `gorm:"size:50;not null" json:"service_type"`
	RequestCount int       `gorm:"default:0" json:"request_count"`
	TokensUsed   int       `gorm:"default:0" json:"tokens_used"`
	Date         time.Time `gorm:"type:date;not null;index" json:"date"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AI Request/Response DTOs

// GenerateHeadlinesRequest represents the request structure for headline generation
type GenerateHeadlinesRequest struct {
	Content string `json:"content" binding:"required" example:"News article content"`
	Count   int    `json:"count" binding:"min=1,max=10" example:"5"`
	Style   string `json:"style" example:"news"` // news, clickbait, formal, casual
}

// GenerateHeadlinesResponse represents the response for headline generation
type GenerateHeadlinesResponse struct {
	Headlines []string `json:"headlines"`
	Count     int      `json:"count"`
}

// GenerateContentRequest represents the request for content generation
type GenerateContentRequest struct {
	Topic       string   `json:"topic" binding:"required" example:"AI in journalism"`
	Style       string   `json:"style" example:"news"`    // news, blog, formal, casual
	Length      string   `json:"length" example:"medium"` // short, medium, long
	Keywords    []string `json:"keywords" example:"AI,journalism,technology"`
	Perspective string   `json:"perspective" example:"objective"` // objective, positive, critical
}

// GenerateContentResponse represents the response for content generation
type GenerateContentResponse struct {
	Content  string   `json:"content"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
	ReadTime int      `json:"read_time"`
}

// ImproveContentRequest represents the request for content improvement
type ImproveContentRequest struct {
	Content     string   `json:"content" binding:"required"`
	Goals       []string `json:"goals"`        // clarity, engagement, seo, readability
	TargetLevel string   `json:"target_level"` // beginner, intermediate, advanced
}

// ImproveContentResponse represents the response for content improvement
type ImproveContentResponse struct {
	ImprovedContent string               `json:"improved_content"`
	Suggestions     []ContentImprovement `json:"suggestions"`
	QualityScore    float64              `json:"quality_score"`
}

// ContentImprovement represents individual improvement suggestions
type ContentImprovement struct {
	Type        string `json:"type"` // grammar, style, structure, seo
	Original    string `json:"original"`
	Suggestion  string `json:"suggestion"`
	Explanation string `json:"explanation"`
	Impact      string `json:"impact"` // low, medium, high
}

// ModerateContentRequest represents the request for content moderation
type ModerateContentRequest struct {
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"` // comment, article, message
	Strict      bool   `json:"strict"`       // enable strict moderation
}

// ModerateContentResponse represents the response for content moderation
type ModerateContentResponse struct {
	IsApproved bool                 `json:"is_approved"`
	Confidence float64              `json:"confidence"`
	Reason     string               `json:"reason"`
	Categories []ModerationCategory `json:"categories"`
	Severity   string               `json:"severity"`
}

// ModerationCategory represents flagged content categories
type ModerationCategory struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Severity   string  `json:"severity"`
}

// SummarizeContentRequest represents the request for content summarization
type SummarizeContentRequest struct {
	Content  string `json:"content" binding:"required"`
	Length   string `json:"length"`   // short, medium, long
	Style    string `json:"style"`    // bullet, paragraph, highlight
	Language string `json:"language"` // tr, en
}

// SummarizeContentResponse represents the response for content summarization
type SummarizeContentResponse struct {
	Summary   string   `json:"summary"`
	KeyPoints []string `json:"key_points"`
	WordCount int      `json:"word_count"`
	Reduction float64  `json:"reduction"` // percentage reduction
}

// CategorizeContentRequest represents the request for content categorization
type CategorizeContentRequest struct {
	Content string   `json:"content" binding:"required"`
	Options []string `json:"options"` // available categories to choose from
}

// CategorizeContentResponse represents the response for content categorization
type CategorizeContentResponse struct {
	Categories []CategorySuggestion `json:"categories"`
	Tags       []TagSuggestion      `json:"tags"`
}

// CategorySuggestion represents a suggested category
type CategorySuggestion struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// TagSuggestion represents a suggested tag
type TagSuggestion struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Relevance  string  `json:"relevance"` // high, medium, low
}

// Agent API DTOs

// CreateAgentTaskRequest represents the request to create an agent task
type CreateAgentTaskRequest struct {
	TaskType   string                 `json:"task_type" binding:"required"`
	InputData  map[string]interface{} `json:"input_data" binding:"required"`
	WebhookURL string                 `json:"webhook_url"`
	Priority   int                    `json:"priority"`
}

// UpdateAgentTaskRequest represents the request to update an agent task
type UpdateAgentTaskRequest struct {
	Status     string                 `json:"status"`
	OutputData map[string]interface{} `json:"output_data"`
	ErrorMsg   string                 `json:"error_msg"`
	Progress   int                    `json:"progress"`
}

// ProcessAgentTaskResponse represents the response for task processing
type ProcessAgentTaskResponse struct {
	TaskID      uint                   `json:"task_id"`
	Status      string                 `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	Progress    int                    `json:"progress"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// AgentTaskResponse represents the response for agent task operations
type AgentTaskResponse struct {
	ID          uint                   `json:"id"`
	TaskType    string                 `json:"task_type"`
	Status      string                 `json:"status"`
	Priority    int                    `json:"priority"`
	InputData   map[string]interface{} `json:"input_data"`
	OutputData  map[string]interface{} `json:"output_data,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// BatchModerationRequest represents the request for batch content moderation
type BatchModerationRequest struct {
	Contents []BatchModerationItem `json:"contents" binding:"required"`
	Strict   bool                  `json:"strict"`
}

// BatchModerationItem represents a single item in batch moderation
type BatchModerationItem struct {
	ID          string `json:"id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"`
}

// BatchModerationResponse represents the response for batch moderation
type BatchModerationResponse struct {
	Results []BatchModerationResult `json:"results"`
	Summary BatchModerationSummary  `json:"summary"`
}

// BatchModerationResult represents the result for a single item in batch moderation
type BatchModerationResult struct {
	ID         string               `json:"id"`
	IsApproved bool                 `json:"is_approved"`
	Confidence float64              `json:"confidence"`
	Reason     string               `json:"reason"`
	Categories []ModerationCategory `json:"categories"`
	Severity   string               `json:"severity"`
}

// BatchModerationSummary represents the summary of batch moderation results
type BatchModerationSummary struct {
	Total    int `json:"total"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
	Flagged  int `json:"flagged"`
}

// Semantic Search Models

// SemanticSearchRequest represents a semantic search request
type SemanticSearchRequest struct {
	Query     string    `json:"query" binding:"required" example:"AI technology advancements"`
	Lang      string    `json:"lang,omitempty" example:"en"`
	Region    string    `json:"region,omitempty" example:"US"`
	Category  string    `json:"category,omitempty" example:"technology"`
	Limit     int       `json:"limit,omitempty" example:"10"`
	Embedding []float64 `json:"-"` // Internal field, not exposed in JSON
}

// SemanticSearchResponse represents the response from semantic search
type SemanticSearchResponse struct {
	Query   string                 `json:"query"`
	Results []SemanticSearchResult `json:"results"`
	Total   int                    `json:"total"`
	Method  string                 `json:"method"` // "vector" or "fallback"
	Meta    SemanticSearchMeta     `json:"meta"`
}

// SemanticSearchResult represents a single search result
type SemanticSearchResult struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	PublishedAt time.Time `json:"published_at"`
	Score       float64   `json:"score"`
	Lang        string    `json:"lang,omitempty"`
	Region      string    `json:"region,omitempty"`
}

// SemanticSearchMeta contains metadata about the search
type SemanticSearchMeta struct {
	ProcessingTime  string `json:"processing_time"`
	QueryEmbedding  bool   `json:"query_embedding"` // Whether embeddings were generated
	IndexUsed       string `json:"index_used"`
	RateLimitReason string `json:"rate_limit_reason,omitempty"` // Reason for rate limiting (if any)
}

// ElasticSearch Models

// SearchDocument represents a document for ElasticSearch indexing
type SearchDocument struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Lang        string    `json:"lang,omitempty"`
	Region      string    `json:"region,omitempty"`
	Category    string    `json:"category,omitempty"`
	Embedding   []float64 `json:"embedding"`
}

// SearchResult represents a search result from ElasticSearch
type SearchResult struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Summary     string                 `json:"summary"`
	PublishedAt time.Time              `json:"published_at"`
	Score       float64                `json:"score"`
	Lang        string                 `json:"lang,omitempty"`
	Region      string                 `json:"region,omitempty"`
	Source      map[string]interface{} `json:"_source,omitempty"`
}
