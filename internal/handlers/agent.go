package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/json"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var agentTracer = otel.Tracer("news-api/handlers/agent")

// CreateAgentTask godoc
// @Summary Create Agent Task
// @Description Create a new automated task for n8n integration
// @Tags Agent
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateAgentTaskRequest true "Agent task creation request"
// @Success 201 {object} models.AgentTaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks [post]
func CreateAgentTask(c *gin.Context) {
	_, span := agentTracer.Start(c.Request.Context(), "agent.create_task")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request models.CreateAgentTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate task type
	validTypes := []string{"content_generation", "content_moderation", "content_analysis", "scheduled_summary", "bulk_categorization"}
	isValidType := false
	for _, t := range validTypes {
		if request.TaskType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task type"})
		return
	}

	// Set defaults
	if request.Priority == 0 {
		request.Priority = 1 // medium priority
	}

	span.SetAttributes(
		attribute.String("request.task_type", request.TaskType),
		attribute.Int("request.priority", request.Priority),
		attribute.Int("user.id", int(userID.(uint))),
	)

	// Create agent task
	inputDataJSON, _ := json.Marshal(request.InputData)

	task := models.AgentTask{
		TaskType:    request.TaskType,
		Status:      "pending",
		Priority:    request.Priority,
		InputData:   inputDataJSON,
		RequestedBy: userID.(uint),
		WebhookURL:  request.WebhookURL,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create agent task"})
		return
	}

	// Generate webhook URL for n8n if not provided
	if task.WebhookURL == "" {
		task.WebhookURL = generateWebhookURL(task.ID)
		database.DB.Model(&task).Update("webhook_url", task.WebhookURL)
	}

	response := models.AgentTaskResponse{
		ID:        task.ID,
		TaskType:  task.TaskType,
		Status:    task.Status,
		Priority:  task.Priority,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}

	span.SetAttributes(attribute.Int("response.task_id", int(task.ID)))
	c.JSON(http.StatusCreated, response)
}

// GetAgentTasks godoc
// @Summary Get Agent Tasks
// @Description Retrieve user's agent tasks with pagination
// @Tags Agent
// @Produce json
// @Security Bearer
// @Param status query string false "Filter by status"
// @Param task_type query string false "Filter by task type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks [get]
func GetAgentTasks(c *gin.Context) {
	_, span := agentTracer.Start(c.Request.Context(), "agent.get_tasks")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	taskType := c.Query("task_type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Where("requested_by = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}

	var total int64
	query.Model(&models.AgentTask{}).Count(&total)

	var tasks []models.AgentTask
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch agent tasks"})
		return
	}

	span.SetAttributes(
		attribute.Int("response.count", len(tasks)),
		attribute.Int64("response.total", total),
	)

	response := models.PaginatedResponse{
		Data:       tasks,
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: (int(total) + limit - 1) / limit,
		HasNext:    page < (int(total)+limit-1)/limit,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// GetAgentTask godoc
// @Summary Get Agent Task
// @Description Retrieve a specific agent task by ID
// @Tags Agent
// @Produce json
// @Security Bearer
// @Param id path int true "Task ID"
// @Success 200 {object} models.AgentTaskResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks/{id} [get]
func GetAgentTask(c *gin.Context) {
	_, span := agentTracer.Start(c.Request.Context(), "agent.get_task")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var task models.AgentTask
	if err := database.DB.Where("id = ? AND requested_by = ?", taskID, userID).First(&task).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Parse input data
	var inputData map[string]interface{}
	if len(task.InputData) > 0 {
		if err := json.Unmarshal(task.InputData, &inputData); err != nil {
			log.Printf("Warning: Failed to unmarshal input data: %v", err)
		}
	}

	// Parse output data
	var outputData map[string]interface{}
	if len(task.OutputData) > 0 {
		if err := json.Unmarshal(task.OutputData, &outputData); err != nil {
			log.Printf("Warning: Failed to unmarshal output data: %v", err)
		}
	}

	response := models.AgentTaskResponse{
		ID:          task.ID,
		TaskType:    task.TaskType,
		Status:      task.Status,
		Priority:    task.Priority,
		InputData:   inputData,
		OutputData:  outputData,
		ErrorMsg:    task.ErrorMsg,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		StartedAt:   task.StartedAt,
		CompletedAt: task.CompletedAt,
	}

	span.SetAttributes(attribute.String("response.status", task.Status))
	c.JSON(http.StatusOK, response)
}

// UpdateAgentTask godoc
// @Summary Update Agent Task
// @Description Update the status and result of an agent task (for n8n webhook)
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body models.UpdateAgentTaskRequest true "Task update request"
// @Success 200 {object} models.AgentTaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks/{id} [put]
func UpdateAgentTask(c *gin.Context) {
	_, span := agentTracer.Start(c.Request.Context(), "agent.update_task")
	defer span.End()

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var request models.UpdateAgentTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	var task models.AgentTask
	if err := database.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Update task fields
	updates := map[string]interface{}{}

	if request.Status != "" {
		updates["status"] = request.Status

		// Update timestamps based on status
		now := time.Now()
		if request.Status == "running" && task.StartedAt == nil {
			updates["started_at"] = &now
		} else if (request.Status == "completed" || request.Status == "failed") && task.CompletedAt == nil {
			updates["completed_at"] = &now
		}
	}

	if request.OutputData != nil {
		outputDataJSON, _ := json.Marshal(request.OutputData)
		updates["output_data"] = outputDataJSON
	}

	if request.ErrorMsg != "" {
		updates["error_msg"] = request.ErrorMsg
	}

	if request.Progress >= 0 && request.Progress <= 100 {
		updates["progress"] = request.Progress
	}

	// Update the task
	if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update task"})
		return
	}

	// Fetch updated task
	database.DB.Where("id = ?", taskID).First(&task)

	// Parse input data and output data for response
	var inputData map[string]interface{}
	if len(task.InputData) > 0 {
		if err := json.Unmarshal(task.InputData, &inputData); err != nil {
			log.Printf("Warning: Failed to unmarshal input data: %v", err)
		}
	}

	var outputData map[string]interface{}
	if len(task.OutputData) > 0 {
		if err := json.Unmarshal(task.OutputData, &outputData); err != nil {
			log.Printf("Warning: Failed to unmarshal output data: %v", err)
		}
	}

	response := models.AgentTaskResponse{
		ID:          task.ID,
		TaskType:    task.TaskType,
		Status:      task.Status,
		Priority:    task.Priority,
		InputData:   inputData,
		OutputData:  outputData,
		ErrorMsg:    task.ErrorMsg,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		StartedAt:   task.StartedAt,
		CompletedAt: task.CompletedAt,
	}

	span.SetAttributes(
		attribute.String("response.status", task.Status),
	)

	c.JSON(http.StatusOK, response)
}

// DeleteAgentTask godoc
// @Summary Delete Agent Task
// @Description Delete an agent task
// @Tags Agent
// @Security Bearer
// @Param id path int true "Task ID"
// @Success 204
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks/{id} [delete]
func DeleteAgentTask(c *gin.Context) {
	_, span := agentTracer.Start(c.Request.Context(), "agent.delete_task")
	defer span.End()

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	result := database.DB.Where("id = ? AND requested_by = ?", taskID, userID).Delete(&models.AgentTask{})
	if result.Error != nil {
		span.RecordError(result.Error)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	span.SetAttributes(attribute.Int("task_id", int(taskID)))
	c.Status(http.StatusNoContent)
}

// ProcessAgentTask godoc
// @Summary Process Agent Task
// @Description Process an agent task using AI services (for internal use)
// @Tags Agent
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} models.ProcessAgentTaskResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/agent/tasks/{id}/process [post]
func ProcessAgentTask(c *gin.Context) {
	ctx, span := agentTracer.Start(c.Request.Context(), "agent.process_task")
	defer span.End()

	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var task models.AgentTask
	if err := database.DB.Where("id = ?", taskID).First(&task).Error; err != nil {
		span.RecordError(err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Update task status to running
	now := time.Now()
	database.DB.Model(&task).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": &now,
	})

	// Parse input data
	var inputData map[string]interface{}
	if len(task.InputData) > 0 {
		if err := json.Unmarshal(task.InputData, &inputData); err != nil {
			log.Printf("Warning: Failed to unmarshal input data: %v", err)
		}
	}

	// Process based on task type
	aiService := services.GetAIService()
	var result map[string]interface{}
	var processingError string

	switch task.TaskType {
	case "content_generation":
		result, processingError = processContentGeneration(ctx, aiService, inputData)
	case "content_moderation":
		result, processingError = processContentModeration(ctx, aiService, inputData)
	case "content_analysis":
		result, processingError = processContentAnalysis(ctx, aiService, inputData)
	case "scheduled_summary":
		result, processingError = processScheduledSummary(ctx, aiService, inputData)
	case "bulk_categorization":
		result, processingError = processBulkCategorization(ctx, aiService, inputData)
	default:
		processingError = "Unknown task type"
	}

	// Update task with result
	updates := map[string]interface{}{
		"completed_at": time.Now(),
	}

	if processingError != "" {
		updates["status"] = "failed"
		updates["error_msg"] = processingError
	} else {
		updates["status"] = "completed"
		resultJSON, _ := json.Marshal(result)
		updates["output_data"] = resultJSON
	}

	database.DB.Model(&task).Updates(updates)

	response := models.ProcessAgentTaskResponse{
		TaskID:      task.ID,
		Status:      updates["status"].(string),
		Result:      result,
		ErrorMsg:    processingError,
		Progress:    100,
		ProcessedAt: time.Now(),
	}

	span.SetAttributes(
		attribute.String("response.status", updates["status"].(string)),
		attribute.String("task.type", task.TaskType),
	)

	c.JSON(http.StatusOK, response)
}

// Helper functions

func generateWebhookURL(taskID uint) string {
	// In production, this should use your actual domain
	return "https://your-api-domain.com/agent/tasks/" + strconv.Itoa(int(taskID))
}

func processContentGeneration(ctx context.Context, aiService *services.AIService, parameters map[string]interface{}) (map[string]interface{}, string) {
	topic, ok := parameters["topic"].(string)
	if !ok {
		return nil, "Missing topic parameter"
	}

	style, _ := parameters["style"].(string)
	if style == "" {
		style = "news"
	}

	length, _ := parameters["length"].(string)
	if length == "" {
		length = "medium"
	}

	keywords, _ := parameters["keywords"].([]interface{})
	keywordStrings := make([]string, len(keywords))
	for i, k := range keywords {
		keywordStrings[i] = k.(string)
	}

	perspective, _ := parameters["perspective"].(string)
	if perspective == "" {
		perspective = "objective"
	}

	content, err := aiService.GenerateContent(ctx, topic, style, length, keywordStrings, perspective)
	if err != nil {
		return nil, err.Error()
	}

	return map[string]interface{}{
		"content":  content,
		"topic":    topic,
		"style":    style,
		"length":   length,
		"keywords": keywordStrings,
	}, ""
}

func processContentModeration(ctx context.Context, aiService *services.AIService, parameters map[string]interface{}) (map[string]interface{}, string) {
	content, ok := parameters["content"].(string)
	if !ok {
		return nil, "Missing content parameter"
	}

	strict, _ := parameters["strict"].(bool)

	isApproved, confidence, reason, categories, severity, err := aiService.ModerateComment(ctx, content, strict)
	if err != nil {
		return nil, err.Error()
	}

	return map[string]interface{}{
		"is_approved": isApproved,
		"confidence":  confidence,
		"reason":      reason,
		"categories":  categories,
		"severity":    severity,
	}, ""
}

func processContentAnalysis(ctx context.Context, aiService *services.AIService, parameters map[string]interface{}) (map[string]interface{}, string) {
	content, ok := parameters["content"].(string)
	if !ok {
		return nil, "Missing content parameter"
	}

	// Perform multiple analysis operations
	summary, err := aiService.SummarizeContent(ctx, content, "short", "paragraph", "en")
	if err != nil {
		return nil, "Failed to summarize: " + err.Error()
	}

	tags, err := aiService.GenerateTags(ctx, content, 10)
	if err != nil {
		return nil, "Failed to generate tags: " + err.Error()
	}

	categories, tagSuggestions, err := aiService.CategorizeContent(ctx, content, []string{"Technology", "Politics", "Sports", "Business", "Entertainment"})
	if err != nil {
		return nil, "Failed to categorize: " + err.Error()
	}

	return map[string]interface{}{
		"summary":         summary,
		"tags":            tags,
		"categories":      categories,
		"tag_suggestions": tagSuggestions,
		"word_count":      len(content) / 5, // Rough estimation
	}, ""
}

func processScheduledSummary(ctx context.Context, aiService *services.AIService, parameters map[string]interface{}) (map[string]interface{}, string) {
	// Get recent articles for summary
	var articles []models.Article
	hours, _ := parameters["hours"].(float64)
	if hours == 0 {
		hours = 24
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	if err := database.DB.Where("created_at >= ?", since).Order("created_at DESC").Limit(50).Find(&articles).Error; err != nil {
		return nil, "Failed to fetch articles: " + err.Error()
	}

	if len(articles) == 0 {
		return map[string]interface{}{
			"summary": "No articles found in the specified time period",
			"count":   0,
		}, ""
	}

	// Combine article content for summarization
	combinedContent := ""
	for _, article := range articles {
		combinedContent += article.Title + ". " + article.Content + "\n\n"
	}

	summary, err := aiService.SummarizeContent(ctx, combinedContent, "medium", "paragraph", "en")
	if err != nil {
		return nil, "Failed to create summary: " + err.Error()
	}

	return map[string]interface{}{
		"summary":       summary,
		"article_count": len(articles),
		"period_hours":  hours,
		"generated_at":  time.Now(),
	}, ""
}

func processBulkCategorization(ctx context.Context, aiService *services.AIService, parameters map[string]interface{}) (map[string]interface{}, string) {
	// Get uncategorized articles
	var articles []models.Article
	limit, _ := parameters["limit"].(float64)
	if limit == 0 {
		limit = 100
	}

	if err := database.DB.Where("category_id IS NULL OR category_id = 0").Limit(int(limit)).Find(&articles).Error; err != nil {
		return nil, "Failed to fetch articles: " + err.Error()
	}

	availableCategories := []string{"Technology", "Politics", "Sports", "Business", "Entertainment", "Health", "Science", "World"}

	results := make([]map[string]interface{}, 0)
	processed := 0

	for _, article := range articles {
		categories, _, err := aiService.CategorizeContent(ctx, article.Content, availableCategories)
		if err != nil {
			continue
		}

		if len(categories) > 0 {
			topCategory := categories[0]
			results = append(results, map[string]interface{}{
				"article_id": article.ID,
				"title":      article.Title,
				"category":   topCategory.Name,
				"confidence": topCategory.Confidence,
			})
		}
		processed++
	}

	return map[string]interface{}{
		"processed_count": processed,
		"categorized":     len(results),
		"results":         results,
	}, ""
}
