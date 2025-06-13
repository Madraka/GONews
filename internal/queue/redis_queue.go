package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"news/internal/cache"
	"news/internal/json"

	"github.com/go-redis/redis/v8"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	PriorityLow      JobPriority = 1
	PriorityNormal   JobPriority = 5
	PriorityHigh     JobPriority = 8
	PriorityCritical JobPriority = 10
)

// Job represents a generic queue job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    JobPriority            `json:"priority"`
	Status      JobStatus              `json:"status"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	QueueName      string     `json:"queue_name"`
	PendingJobs    int64      `json:"pending_jobs"`
	ProcessingJobs int64      `json:"processing_jobs"`
	CompletedJobs  int64      `json:"completed_jobs"`
	FailedJobs     int64      `json:"failed_jobs"`
	DeadJobs       int64      `json:"dead_jobs"`
	WorkerCount    int        `json:"worker_count"`
	LastProcessed  *time.Time `json:"last_processed,omitempty"`
}

// RedisQueue represents a Redis-based job queue
type RedisQueue struct {
	client    *redis.Client
	ctx       context.Context
	queueName string
}

// NewRedisQueue creates a new Redis queue instance
func NewRedisQueue(queueName string) *RedisQueue {
	redisClient := cache.GetRedisClient()
	if redisClient == nil {
		log.Printf("Warning: Redis client not available, queue operations will fail")
		return nil
	}

	return &RedisQueue{
		client:    redisClient.GetClient(),
		ctx:       context.Background(),
		queueName: queueName,
	}
}

// Enqueue adds a job to the queue with priority
func (rq *RedisQueue) Enqueue(job *Job) error {
	if rq.client == nil {
		return fmt.Errorf("redis client not available")
	}

	// Set defaults
	if job.ID == "" {
		job.ID = fmt.Sprintf("%s_%d", job.Type, time.Now().UnixNano())
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.ScheduledAt.IsZero() {
		job.ScheduledAt = time.Now()
	}
	if job.MaxAttempts == 0 {
		job.MaxAttempts = 3
	}
	job.Status = JobStatusPending

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to serialize job: %w", err)
	}

	// Add to priority queue (higher priority = higher score)
	score := float64(job.Priority)*1000000 + float64(job.ScheduledAt.Unix())

	pipe := rq.client.Pipeline()

	// Add to sorted set for priority ordering
	pipe.ZAdd(rq.ctx, rq.getQueueKey(), &redis.Z{
		Score:  score,
		Member: job.ID,
	})

	// Store job data
	pipe.HSet(rq.ctx, rq.getJobsKey(), job.ID, jobData)

	// Set TTL for job data (24 hours)
	pipe.Expire(rq.ctx, rq.getJobsKey(), 24*time.Hour)

	// Send notification to wake up blocking workers
	// Push a simple notification to the blocking queue
	pipe.LPush(rq.ctx, rq.getNotificationKey(), job.ID)
	// Keep only the last 100 notifications to prevent memory leak
	pipe.LTrim(rq.ctx, rq.getNotificationKey(), 0, 99)

	_, err = pipe.Exec(rq.ctx)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("Job %s enqueued to queue %s with priority %d", job.ID, rq.queueName, job.Priority)
	return nil
}

// Dequeue retrieves and removes the highest priority job from the queue
func (rq *RedisQueue) Dequeue() (*Job, error) {
	if rq.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	// Get highest priority job (ZRANGE with LIMIT)
	result, err := rq.client.ZPopMax(rq.ctx, rq.getQueueKey()).Result()
	if err == redis.Nil {
		return nil, nil // No jobs available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) == 0 {
		return nil, nil // No jobs available
	}

	jobID := result[0].Member.(string)

	// Get job data
	jobData, err := rq.client.HGet(rq.ctx, rq.getJobsKey(), jobID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	// Update job status
	now := time.Now()
	job.Status = JobStatusProcessing
	job.StartedAt = &now

	// Update job in Redis
	updatedJobData, _ := json.Marshal(job)
	rq.client.HSet(rq.ctx, rq.getJobsKey(), jobID, updatedJobData)

	return &job, nil
}

// BlockingDequeue blocks until a job is available or timeout occurs
// This method uses BLPOP to reduce CPU consumption by blocking until a job is available
func (rq *RedisQueue) BlockingDequeue(timeout time.Duration) (*Job, error) {
	if rq.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	// Use BLPOP on the notification list to wait for new jobs
	notificationKey := rq.getNotificationKey()
	result, err := rq.client.BLPop(rq.ctx, timeout, notificationKey).Result()
	if err == redis.Nil {
		return nil, nil // Timeout occurred, no jobs available
	}
	if err != nil {
		// If context was cancelled, return nil
		if rq.ctx.Err() != nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to wait for job notification: %w", err)
	}

	if len(result) < 2 {
		return nil, nil // No notification received
	}

	// A job notification was received, now get the actual highest priority job
	// Use the non-blocking version to get the job
	return rq.Dequeue()
}

// CompleteJob marks a job as completed
func (rq *RedisQueue) CompleteJob(jobID string, result map[string]interface{}) error {
	return rq.updateJobStatus(jobID, JobStatusCompleted, "", result)
}

// FailJob marks a job as failed and handles retry logic
func (rq *RedisQueue) FailJob(jobID string, errorMsg string) error {
	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	job.Attempts++
	job.ErrorMsg = errorMsg

	// Check if we should retry
	if job.Attempts < job.MaxAttempts {
		// Retry with exponential backoff
		retryDelay := time.Duration(job.Attempts*job.Attempts) * time.Minute
		job.ScheduledAt = time.Now().Add(retryDelay)
		job.Status = JobStatusRetrying

		// Re-enqueue for retry
		return rq.Enqueue(job)
	}

	// Max attempts reached, mark as failed
	return rq.updateJobStatus(jobID, JobStatusFailed, errorMsg, nil)
}

// GetJob retrieves job details
func (rq *RedisQueue) GetJob(jobID string) (*Job, error) {
	if rq.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	jobData, err := rq.client.HGet(rq.ctx, rq.getJobsKey(), jobID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to deserialize job: %w", err)
	}

	return &job, nil
}

// GetStats returns queue statistics
func (rq *RedisQueue) GetStats() (QueueStats, error) {
	stats := QueueStats{
		QueueName: rq.queueName,
	}

	// Get pending jobs count
	pendingCount, err := rq.client.ZCard(rq.ctx, rq.getQueueKey()).Result()
	if err != nil {
		return stats, err
	}
	stats.PendingJobs = pendingCount

	// Get jobs by status from the jobs hash
	jobsMap, err := rq.client.HGetAll(rq.ctx, rq.getJobsKey()).Result()
	if err != nil {
		return stats, err
	}

	for _, jobData := range jobsMap {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		switch job.Status {
		case JobStatusProcessing:
			stats.ProcessingJobs++
		case JobStatusCompleted:
			stats.CompletedJobs++
		case JobStatusFailed:
			stats.FailedJobs++
		}

		// Track last processed time
		if job.CompletedAt != nil && (stats.LastProcessed == nil || job.CompletedAt.After(*stats.LastProcessed)) {
			stats.LastProcessed = job.CompletedAt
		}
	}

	// Get dead letter queue count
	deadLetterQueue := fmt.Sprintf("dead_letter:%s", rq.queueName)
	deadCount, err := rq.client.LLen(rq.ctx, deadLetterQueue).Result()
	if err == nil {
		stats.DeadJobs = deadCount
	}

	return stats, nil
}

// CleanupCompletedJobs removes completed jobs older than the specified duration
func (rq *RedisQueue) CleanupCompletedJobs(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	jobsMap, err := rq.client.HGetAll(rq.ctx, rq.getJobsKey()).Result()
	if err != nil {
		return err
	}

	var toDelete []string
	for jobID, jobData := range jobsMap {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		if job.Status == JobStatusCompleted && job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			toDelete = append(toDelete, jobID)
		}
	}

	if len(toDelete) > 0 {
		return rq.client.HDel(rq.ctx, rq.getJobsKey(), toDelete...).Err()
	}

	return nil
}

// JobStatusInfo represents a job status for API responses
type JobStatusInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Priority    int                    `json:"priority"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   int64                  `json:"created_at"`
	ScheduledAt int64                  `json:"scheduled_at"`
	StartedAt   *int64                 `json:"started_at,omitempty"`
	CompletedAt *int64                 `json:"completed_at,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

// GetJobs returns jobs from the queue with pagination and status filtering
func (rq *RedisQueue) GetJobs(status string, page, limit int) ([]JobStatusInfo, int64, error) {
	if rq.client == nil {
		return nil, 0, fmt.Errorf("redis client not available")
	}

	// Get all jobs
	jobsMap, err := rq.client.HGetAll(rq.ctx, rq.getJobsKey()).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get jobs: %w", err)
	}

	var filteredJobs []Job
	for _, jobData := range jobsMap {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		// Filter by status if provided
		if status == "" || string(job.Status) == status {
			filteredJobs = append(filteredJobs, job)
		}
	}

	total := int64(len(filteredJobs))

	// Apply pagination
	offset := (page - 1) * limit
	if offset >= len(filteredJobs) {
		return []JobStatusInfo{}, total, nil
	}

	end := offset + limit
	if end > len(filteredJobs) {
		end = len(filteredJobs)
	}

	// Convert to JobStatusInfo slice
	result := make([]JobStatusInfo, 0, end-offset)
	for i := offset; i < end; i++ {
		job := filteredJobs[i]
		jobStatus := JobStatusInfo{
			ID:          job.ID,
			Type:        job.Type,
			Status:      string(job.Status),
			Priority:    int(job.Priority),
			Attempts:    job.Attempts,
			MaxAttempts: job.MaxAttempts,
			CreatedAt:   job.CreatedAt.Unix(),
			ScheduledAt: job.ScheduledAt.Unix(),
			ErrorMsg:    job.ErrorMsg,
			Payload:     job.Payload,
		}

		if job.StartedAt != nil {
			startedAt := job.StartedAt.Unix()
			jobStatus.StartedAt = &startedAt
		}

		if job.CompletedAt != nil {
			completedAt := job.CompletedAt.Unix()
			jobStatus.CompletedAt = &completedAt
		}

		result = append(result, jobStatus)
	}

	return result, total, nil
}

// GetJobStatus returns the status of a specific job
func (rq *RedisQueue) GetJobStatus(jobID string) (*JobStatusInfo, error) {
	if rq.client == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	job, err := rq.GetJob(jobID)
	if err != nil {
		return nil, err
	}

	jobStatus := &JobStatusInfo{
		ID:          job.ID,
		Type:        job.Type,
		Status:      string(job.Status),
		Priority:    int(job.Priority),
		Attempts:    job.Attempts,
		MaxAttempts: job.MaxAttempts,
		CreatedAt:   job.CreatedAt.Unix(),
		ScheduledAt: job.ScheduledAt.Unix(),
		ErrorMsg:    job.ErrorMsg,
		Payload:     job.Payload,
	}

	if job.StartedAt != nil {
		startedAt := job.StartedAt.Unix()
		jobStatus.StartedAt = &startedAt
	}

	if job.CompletedAt != nil {
		completedAt := job.CompletedAt.Unix()
		jobStatus.CompletedAt = &completedAt
	}

	return jobStatus, nil
}

// RetryJob retries a failed job
func (rq *RedisQueue) RetryJob(jobID string) error {
	if rq.client == nil {
		return fmt.Errorf("redis client not available")
	}

	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	// Only retry failed jobs
	if job.Status != JobStatusFailed {
		return fmt.Errorf("cannot retry job with status: %s", job.Status)
	}

	// Reset job for retry
	job.Status = JobStatusPending
	job.ErrorMsg = ""
	job.StartedAt = nil
	job.CompletedAt = nil
	job.ScheduledAt = time.Now()

	// Re-add to queue
	score := float64(job.Priority)*1000000 + float64(job.ScheduledAt.Unix())

	pipe := rq.client.Pipeline()

	// Add back to sorted set
	pipe.ZAdd(rq.ctx, rq.getQueueKey(), &redis.Z{
		Score:  score,
		Member: job.ID,
	})

	// Update job data
	jobData, _ := json.Marshal(job)
	pipe.HSet(rq.ctx, rq.getJobsKey(), job.ID, jobData)

	_, err = pipe.Exec(rq.ctx)
	return err
}

// DeleteJob deletes a job from the queue
func (rq *RedisQueue) DeleteJob(jobID string) error {
	if rq.client == nil {
		return fmt.Errorf("redis client not available")
	}

	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	// Only allow deletion of completed, failed, or dead jobs
	if job.Status == JobStatusProcessing || job.Status == JobStatusPending {
		return fmt.Errorf("cannot delete job with status: %s", job.Status)
	}

	// Remove from both queue and jobs hash
	pipe := rq.client.Pipeline()
	pipe.ZRem(rq.ctx, rq.getQueueKey(), jobID)
	pipe.HDel(rq.ctx, rq.getJobsKey(), jobID)

	_, err = pipe.Exec(rq.ctx)
	return err
}

// CleanupOldJobs removes old completed and failed jobs
func (rq *RedisQueue) CleanupOldJobs(hours int) (int, error) {
	if rq.client == nil {
		return 0, fmt.Errorf("redis client not available")
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)

	jobsMap, err := rq.client.HGetAll(rq.ctx, rq.getJobsKey()).Result()
	if err != nil {
		return 0, err
	}

	var toDelete []string
	for jobID, jobData := range jobsMap {
		var job Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		// Delete old completed or failed jobs
		if (job.Status == JobStatusCompleted || job.Status == JobStatusFailed) &&
			job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			toDelete = append(toDelete, jobID)
		}
	}

	if len(toDelete) > 0 {
		pipe := rq.client.Pipeline()
		pipe.HDel(rq.ctx, rq.getJobsKey(), toDelete...)

		// Convert to interface{} slice for ZRem
		toDeleteInterfaces := make([]interface{}, len(toDelete))
		for i, v := range toDelete {
			toDeleteInterfaces[i] = v
		}
		pipe.ZRem(rq.ctx, rq.getQueueKey(), toDeleteInterfaces...)

		_, err = pipe.Exec(rq.ctx)
		if err != nil {
			return 0, err
		}
	}

	return len(toDelete), nil
}

// GetQueueStats returns statistics about the queue
func (rq *RedisQueue) GetQueueStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get queue lengths
	queueKey := rq.getQueueKey()
	jobsKey := rq.getJobsKey()
	deadLetterKey := fmt.Sprintf("dead_letter:%s", rq.queueName)

	// Count jobs by status
	pendingCount, err := rq.client.ZCard(rq.ctx, queueKey).Result()
	if err != nil {
		return nil, err
	}

	totalJobs, err := rq.client.HLen(rq.ctx, jobsKey).Result()
	if err != nil {
		return nil, err
	}

	deadLetterCount, err := rq.client.LLen(rq.ctx, deadLetterKey).Result()
	if err != nil {
		deadLetterCount = 0 // Ignore error for dead letter queue
	}

	// Calculate other statuses
	processingCount := int64(0)
	completedCount := int64(0)
	failedCount := int64(0)

	// Get all jobs and count by status
	jobs, err := rq.client.HGetAll(rq.ctx, jobsKey).Result()
	if err == nil {
		for _, jobData := range jobs {
			var job Job
			if json.Unmarshal([]byte(jobData), &job) == nil {
				switch job.Status {
				case JobStatusProcessing:
					processingCount++
				case JobStatusCompleted:
					completedCount++
				case JobStatusFailed:
					failedCount++
				}
			}
		}
	}

	stats["queue_name"] = rq.queueName
	stats["total_jobs"] = totalJobs
	stats["pending_jobs"] = pendingCount
	stats["processing_jobs"] = processingCount
	stats["completed_jobs"] = completedCount
	stats["failed_jobs"] = failedCount
	stats["dead_letter_jobs"] = deadLetterCount

	return stats, nil
}

// Helper methods
func (rq *RedisQueue) getQueueKey() string {
	return fmt.Sprintf("queue:%s", rq.queueName)
}

func (rq *RedisQueue) getJobsKey() string {
	return fmt.Sprintf("jobs:%s", rq.queueName)
}

func (rq *RedisQueue) updateJobStatus(jobID string, status JobStatus, errorMsg string, result map[string]interface{}) error {
	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	job.Status = status
	if errorMsg != "" {
		job.ErrorMsg = errorMsg
	}

	if status == JobStatusCompleted || status == JobStatusFailed {
		now := time.Now()
		job.CompletedAt = &now
	}

	// Store result in payload if provided
	if result != nil {
		if job.Payload == nil {
			job.Payload = make(map[string]interface{})
		}
		job.Payload["result"] = result
	}

	// Update job in Redis
	jobData, _ := json.Marshal(job)
	return rq.client.HSet(rq.ctx, rq.getJobsKey(), jobID, jobData).Err()
}

// DeadLetterQueue handles failed jobs
func (rq *RedisQueue) MoveToDeadLetter(jobID string) error {
	job, err := rq.GetJob(jobID)
	if err != nil {
		return err
	}

	// Move to dead letter queue
	deadLetterQueue := fmt.Sprintf("dead_letter:%s", rq.queueName)
	jobData, _ := json.Marshal(job)

	pipe := rq.client.Pipeline()
	pipe.LPush(rq.ctx, deadLetterQueue, jobData)
	pipe.HDel(rq.ctx, rq.getJobsKey(), jobID)
	pipe.ZRem(rq.ctx, rq.getQueueKey(), jobID)

	_, err = pipe.Exec(rq.ctx)
	return err
}

// getNotificationKey returns the Redis key for job notifications
func (rq *RedisQueue) getNotificationKey() string {
	return fmt.Sprintf("notifications:%s", rq.queueName)
}
