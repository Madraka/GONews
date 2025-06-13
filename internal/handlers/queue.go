package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"news/internal/queue"

	"github.com/gin-gonic/gin"
)

// QueueStatsResponse represents the queue statistics response
type QueueStatsResponse struct {
	Status string                      `json:"status"`
	Queues map[string]queue.QueueStats `json:"queues"`
}

// GetQueueStats returns statistics for all Redis queues
// @Summary Get queue statistics
// @Description Get real-time statistics for all Redis queues including pending, processing, completed, and failed job counts
// @Tags Queue
// @Produce json
// @Success 200 {object} QueueStatsResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/stats [get]
// @Security BearerAuth
func GetQueueStats(c *gin.Context) {
	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Queue manager not available",
		})
		return
	}

	stats := queueManager.GetQueueStats()

	response := QueueStatsResponse{
		Status: "success",
		Queues: stats,
	}

	c.JSON(http.StatusOK, response)
}

// EnqueueJobRequest represents a request to enqueue a job
type EnqueueJobRequest struct {
	QueueName string                 `json:"queue_name" binding:"required"`
	JobType   string                 `json:"job_type" binding:"required"`
	Priority  int                    `json:"priority"`
	Payload   map[string]interface{} `json:"payload"`
}

// EnqueueJob manually enqueues a job for testing/admin purposes
// @Summary Enqueue a job
// @Description Manually enqueue a job to a specific queue for testing or administrative purposes
// @Tags Queue
// @Accept json
// @Produce json
// @Param request body EnqueueJobRequest true "Job details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/enqueue [post]
// @Security BearerAuth
func EnqueueJob(c *gin.Context) {
	var req EnqueueJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Queue manager not available",
		})
		return
	}

	// Set default priority if not provided
	priority := queue.JobPriority(req.Priority)
	if priority == 0 {
		priority = queue.PriorityNormal
	}

	var err error
	switch req.QueueName {
	case "translations":
		entityType, _ := req.Payload["entity_type"].(string)
		entityIDFloat, _ := req.Payload["entity_id"].(float64)
		entityID := uint(entityIDFloat)
		sourceLang, _ := req.Payload["source_lang"].(string)
		targetLang, _ := req.Payload["target_lang"].(string)

		err = queueManager.EnqueueTranslationJob(entityType, entityID, sourceLang, targetLang, priority)

	case "video_processing":
		videoIDFloat, _ := req.Payload["video_id"].(float64)
		videoID := uint(videoIDFloat)

		err = queueManager.EnqueueVideoJob(req.JobType, videoID, priority)

	case "agent_tasks":
		err = queueManager.EnqueueAgentJob(req.JobType, req.Payload, priority)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported queue name",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue job: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Job enqueued successfully",
	})
}

// GetQueueJobs returns jobs from a specific queue
// @Summary Get queue jobs
// @Description Get jobs from a specific queue with pagination and filtering
// @Tags Queue
// @Produce json
// @Param queue query string true "Queue name (translations, video_processing, agent_tasks, general)"
// @Param status query string false "Filter by status (pending, processing, completed, failed, retrying)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} QueueJobsResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/jobs [get]
// @Security BearerAuth
func GetQueueJobs(c *gin.Context) {
	queueName := c.Query("queue")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Queue name is required"})
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	jobs, total, err := queueManager.GetJobs(queueName, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs: " + err.Error()})
		return
	}

	response := QueueJobsResponse{
		Jobs:       jobs,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		QueueName:  queueName,
	}

	c.JSON(http.StatusOK, response)
}

// GetQueueJob returns details of a specific job
// @Summary Get queue job details
// @Description Get detailed information about a specific job by ID
// @Tags Queue
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} queue.JobStatus
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/jobs/{id} [get]
// @Security BearerAuth
func GetQueueJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	job, err := queueManager.GetJobStatus(jobID)
	if err != nil {
		if err.Error() == "job not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch job: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, job)
}

// RetryQueueJob retries a failed job
// @Summary Retry a failed job
// @Description Retry a failed job by moving it back to pending status
// @Tags Queue
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/jobs/{id}/retry [post]
// @Security BearerAuth
func RetryQueueJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	err := queueManager.RetryJob(jobID)
	if err != nil {
		if err.Error() == "job not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry job: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job retried successfully"})
}

// DeleteQueueJob deletes a job from the queue
// @Summary Delete a queue job
// @Description Delete a job from the queue (only completed, failed, or dead jobs)
// @Tags Queue
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/jobs/{id} [delete]
// @Security BearerAuth
func DeleteQueueJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	err := queueManager.DeleteJob(jobID)
	if err != nil {
		if err.Error() == "job not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else if strings.Contains(err.Error(), "cannot delete") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job: " + err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}

// CleanupCompletedJobs removes old completed jobs
// @Summary Cleanup completed jobs
// @Description Remove completed jobs older than specified hours (default 24h)
// @Tags Queue
// @Produce json
// @Param hours query int false "Hours old to cleanup" default(24)
// @Param queue query string false "Specific queue to cleanup (optional)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/jobs/cleanup [post]
// @Security BearerAuth
func CleanupCompletedJobs(c *gin.Context) {
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))
	if hours < 1 {
		hours = 24
	}

	queueName := c.Query("queue")

	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	cleaned, err := queueManager.CleanupOldJobs(hours, queueName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup jobs: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cleanup completed successfully",
		"cleaned": cleaned,
		"hours":   hours,
	})
}

// GetQueueHealth returns health status of all queues
// @Summary Get queue health status
// @Description Get health and operational status of all Redis queues
// @Tags Queue
// @Produce json
// @Success 200 {object} QueueHealthResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/queue/health [get]
// @Security BearerAuth
func GetQueueHealth(c *gin.Context) {
	queueManager := queue.GetGlobalQueueManager()
	if queueManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Queue manager not available"})
		return
	}

	health, err := queueManager.GetHealthStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get health status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, health)
}

// QueueJobsResponse represents the response for GetQueueJobs
type QueueJobsResponse struct {
	Jobs       []queue.JobStatusInfo `json:"jobs"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	QueueName  string                `json:"queue_name"`
}

// QueueHealthResponse represents the health status response
type QueueHealthResponse struct {
	Status    string                        `json:"status"`
	Timestamp int64                         `json:"timestamp"`
	Queues    map[string]QueueHealthDetails `json:"queues"`
}

// QueueHealthDetails represents health details for a specific queue
type QueueHealthDetails struct {
	Status         string `json:"status"`
	WorkerCount    int    `json:"worker_count"`
	ActiveWorkers  int    `json:"active_workers"`
	PendingJobs    int64  `json:"pending_jobs"`
	ProcessingJobs int64  `json:"processing_jobs"`
	FailedJobs     int64  `json:"failed_jobs"`
	LastProcessed  *int64 `json:"last_processed,omitempty"`
}
