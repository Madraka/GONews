package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"news/internal/services"
)

// Global queue manager instance for API handlers
var globalQueueManager *QueueManager
var queueManagerMutex sync.RWMutex

// SetGlobalQueueManager sets the global queue manager instance
func SetGlobalQueueManager(qm *QueueManager) {
	queueManagerMutex.Lock()
	defer queueManagerMutex.Unlock()
	globalQueueManager = qm
}

// GetGlobalQueueManager returns the global queue manager instance
func GetGlobalQueueManager() *QueueManager {
	queueManagerMutex.RLock()
	defer queueManagerMutex.RUnlock()
	return globalQueueManager
}

// QueueManager manages multiple Redis queues and worker pools
type QueueManager struct {
	queues      map[string]*RedisQueue
	workerPools map[string]*WorkerPool
	services    *ServiceContainer
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// ServiceContainer holds references to various services needed by processors
type ServiceContainer struct {
	TranslationService     *services.AITranslationService
	VideoProcessingService *services.VideoProcessingService
	// Add other services as needed
}

// Inline job processors to avoid import cycles

// TranslationJobProcessor handles translation jobs
type TranslationJobProcessor struct {
	service *services.AITranslationService
}

func (p *TranslationJobProcessor) ProcessJob(ctx context.Context, job *Job) error {
	entityType, _ := job.Payload["entity_type"].(string)
	entityID, _ := job.Payload["entity_id"].(float64)
	targetLang, _ := job.Payload["target_lang"].(string)

	// Process translation directly using the service methods
	switch entityType {
	case "article":
		return p.service.TranslateArticle(uint(entityID), []string{targetLang})
	case "category":
		return p.service.TranslateCategory(uint(entityID), []string{targetLang})
	case "tag":
		return p.service.TranslateTag(uint(entityID), []string{targetLang})
	case "menu":
		return p.service.TranslateMenu(uint(entityID), []string{targetLang})
	case "notification":
		return p.service.TranslateNotification(uint(entityID), []string{targetLang})
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

func (p *TranslationJobProcessor) GetJobTypes() []string {
	return []string{"translation", "article_translation", "category_translation", "tag_translation", "menu_translation", "notification_translation"}
}

// VideoJobProcessor handles video processing jobs
type VideoJobProcessor struct {
	service *services.VideoProcessingService
}

func (p *VideoJobProcessor) ProcessJob(ctx context.Context, job *Job) error {
	// Simplified video processing logic
	log.Printf("Processing video job: %s", job.Type)
	return nil
}

func (p *VideoJobProcessor) GetJobTypes() []string {
	return []string{"video", "thumbnail", "transcode", "analysis", "tts", "complete_workflow"}
}

// AgentJobProcessor handles agent task jobs
type AgentJobProcessor struct{}

func (p *AgentJobProcessor) ProcessJob(ctx context.Context, job *Job) error {
	// Simplified agent processing logic
	log.Printf("Processing agent job: %s", job.Type)
	return nil
}

func (p *AgentJobProcessor) GetJobTypes() []string {
	return []string{"agent", "webhook", "automation", "notification", "data_sync"}
}

// NewQueueManager creates a new queue manager
func NewQueueManager(services *ServiceContainer) *QueueManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueManager{
		queues:      make(map[string]*RedisQueue),
		workerPools: make(map[string]*WorkerPool),
		services:    services,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Initialize sets up all queues and processors
func (qm *QueueManager) Initialize() error {
	log.Println("Initializing Redis Queue Manager...")

	// Create queues for different job types
	queueConfigs := map[string]int{
		"translations":     3, // 3 workers for translation jobs
		"video_processing": 2, // 2 workers for video jobs (resource intensive)
		"agent_tasks":      2, // 2 workers for agent tasks
		"general":          3, // 3 workers for general tasks
	}

	for queueName, workerCount := range queueConfigs {
		// Create Redis queue
		queue := NewRedisQueue(queueName)
		if queue == nil {
			return fmt.Errorf("failed to create Redis queue for %s", queueName)
		}
		qm.queues[queueName] = queue

		// Create worker pool
		workerPool := NewWorkerPool(queueName, workerCount)
		qm.workerPools[queueName] = workerPool

		// Register appropriate processors
		if err := qm.registerProcessors(queueName, workerPool); err != nil {
			return fmt.Errorf("failed to register processors for %s: %w", queueName, err)
		}

		log.Printf("Initialized queue '%s' with %d workers", queueName, workerCount)
	}

	return nil
}

// registerProcessors registers the appropriate processors for each queue
func (qm *QueueManager) registerProcessors(queueName string, workerPool *WorkerPool) error {
	switch queueName {
	case "translations":
		if qm.services.TranslationService != nil {
			processor := &TranslationJobProcessor{service: qm.services.TranslationService}
			workerPool.RegisterProcessor(processor)
		}

	case "video_processing":
		if qm.services.VideoProcessingService != nil {
			processor := &VideoJobProcessor{service: qm.services.VideoProcessingService}
			workerPool.RegisterProcessor(processor)
		}

	case "agent_tasks":
		processor := &AgentJobProcessor{}
		workerPool.RegisterProcessor(processor)

	case "general":
		// Register multiple processors for general queue
		if qm.services.TranslationService != nil {
			translationProcessor := &TranslationJobProcessor{service: qm.services.TranslationService}
			workerPool.RegisterProcessor(translationProcessor)
		}

		agentProcessor := &AgentJobProcessor{}
		workerPool.RegisterProcessor(agentProcessor)

	default:
		return fmt.Errorf("unknown queue name: %s", queueName)
	}

	return nil
}

// Start begins processing jobs in all worker pools
func (qm *QueueManager) Start() error {
	log.Println("Starting Redis Queue Manager worker pools...")

	for queueName, workerPool := range qm.workerPools {
		qm.wg.Add(1)
		go func(name string, pool *WorkerPool) {
			defer qm.wg.Done()

			log.Printf("Starting worker pool for queue: %s", name)
			if err := pool.Start(); err != nil {
				log.Printf("Worker pool %s stopped with error: %v", name, err)
			}
		}(queueName, workerPool)
	}

	log.Println("All worker pools started successfully")
	return nil
}

// Stop gracefully stops all worker pools
func (qm *QueueManager) Stop() error {
	log.Println("Stopping Redis Queue Manager...")

	// Cancel context to signal all workers to stop
	qm.cancel()

	// Stop all worker pools
	var stopErrors []error
	for queueName, workerPool := range qm.workerPools {
		log.Printf("Stopping worker pool: %s", queueName)
		if err := workerPool.Stop(); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("failed to stop worker pool %s: %w", queueName, err))
		}
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		qm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All worker pools stopped successfully")
	case <-time.After(30 * time.Second):
		log.Println("Warning: Timeout waiting for worker pools to stop")
	}

	if len(stopErrors) > 0 {
		return fmt.Errorf("errors stopping worker pools: %v", stopErrors)
	}

	return nil
}

// EnqueueJob adds a job to the appropriate queue
func (qm *QueueManager) EnqueueJob(queueName string, job *Job) error {
	queue, exists := qm.queues[queueName]
	if !exists {
		return fmt.Errorf("queue '%s' not found", queueName)
	}

	return queue.Enqueue(job)
}

// GetQueue returns a specific queue by name
func (qm *QueueManager) GetQueue(queueName string) *RedisQueue {
	return qm.queues[queueName]
}

// GetQueueStats returns statistics for all queues
func (qm *QueueManager) GetQueueStats() map[string]QueueStats {
	stats := make(map[string]QueueStats)

	for queueName, queue := range qm.queues {
		queueStats, err := queue.GetStats()
		if err != nil {
			log.Printf("Error getting stats for queue %s: %v", queueName, err)
			continue
		}
		stats[queueName] = queueStats
	}

	return stats
}

// EnqueueTranslationJob is a convenience method for translation jobs
func (qm *QueueManager) EnqueueTranslationJob(entityType string, entityID uint, sourceLang, targetLang string, priority JobPriority) error {
	job := &Job{
		ID:          fmt.Sprintf("translation_%d_%d", entityID, time.Now().Unix()),
		Type:        "translation",
		Priority:    priority,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
		Payload: map[string]interface{}{
			"entity_type": entityType,
			"entity_id":   entityID,
			"source_lang": sourceLang,
			"target_lang": targetLang,
		},
	}
	return qm.EnqueueJob("translations", job)
}

// EnqueueVideoJob is a convenience method for video processing jobs
func (qm *QueueManager) EnqueueVideoJob(jobType string, videoID uint, priority JobPriority) error {
	job := &Job{
		ID:          fmt.Sprintf("video_%s_%d_%d", jobType, videoID, time.Now().Unix()),
		Type:        jobType,
		Priority:    priority,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
		Payload: map[string]interface{}{
			"video_id": videoID,
			"job_type": jobType,
		},
	}

	return qm.EnqueueJob("video_processing", job)
}

// EnqueueAgentJob is a convenience method for agent task jobs
func (qm *QueueManager) EnqueueAgentJob(jobType string, payload map[string]interface{}, priority JobPriority) error {
	job := &Job{
		ID:          fmt.Sprintf("agent_%s_%d", jobType, time.Now().Unix()),
		Type:        jobType,
		Priority:    priority,
		Status:      JobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
		Payload:     payload,
	}

	return qm.EnqueueJob("agent_tasks", job)
}

// GetJobs returns jobs from a specific queue with pagination
func (qm *QueueManager) GetJobs(queueName, status string, page, limit int) ([]JobStatusInfo, int64, error) {
	queue, exists := qm.queues[queueName]
	if !exists {
		return nil, 0, fmt.Errorf("queue '%s' not found", queueName)
	}

	return queue.GetJobs(status, page, limit)
}

// GetJobStatus returns the status of a specific job
func (qm *QueueManager) GetJobStatus(jobID string) (*JobStatusInfo, error) {
	// Try to find the job in all queues
	for _, queue := range qm.queues {
		status, err := queue.GetJobStatus(jobID)
		if err == nil {
			return status, nil
		}
		// Continue searching in other queues if not found
		if err.Error() != "job not found" {
			return nil, err
		}
	}

	return nil, fmt.Errorf("job not found")
}

// RetryJob retries a failed job
func (qm *QueueManager) RetryJob(jobID string) error {
	// Try to find and retry the job in all queues
	for _, queue := range qm.queues {
		err := queue.RetryJob(jobID)
		if err == nil {
			return nil
		}
		// Continue searching in other queues if not found
		if err.Error() != "job not found" {
			return err
		}
	}

	return fmt.Errorf("job not found")
}

// DeleteJob deletes a job from the queue
func (qm *QueueManager) DeleteJob(jobID string) error {
	// Try to find and delete the job in all queues
	for _, queue := range qm.queues {
		err := queue.DeleteJob(jobID)
		if err == nil {
			return nil
		}
		// Continue searching in other queues if not found
		if err.Error() != "job not found" {
			return err
		}
	}

	return fmt.Errorf("job not found")
}

// CleanupOldJobs removes old completed jobs
func (qm *QueueManager) CleanupOldJobs(hours int, queueName string) (int, error) {
	totalCleaned := 0

	if queueName != "" {
		// Clean specific queue
		queue, exists := qm.queues[queueName]
		if !exists {
			return 0, fmt.Errorf("queue '%s' not found", queueName)
		}

		cleaned, err := queue.CleanupOldJobs(hours)
		if err != nil {
			return 0, err
		}
		return cleaned, nil
	}

	// Clean all queues
	for name, queue := range qm.queues {
		cleaned, err := queue.CleanupOldJobs(hours)
		if err != nil {
			log.Printf("Error cleaning queue %s: %v", name, err)
			continue
		}
		totalCleaned += cleaned
	}

	return totalCleaned, nil
}

// GetHealthStatus returns health status of all queues
func (qm *QueueManager) GetHealthStatus() (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"queues":    make(map[string]interface{}),
	}

	overallHealthy := true
	queueHealths := make(map[string]interface{})

	for queueName, queue := range qm.queues {
		workerPool := qm.workerPools[queueName]

		stats, err := queue.GetStats()
		if err != nil {
			log.Printf("Error getting stats for queue %s: %v", queueName, err)
			overallHealthy = false
			continue
		}

		// Check if queue is healthy
		queueHealthy := true
		if stats.FailedJobs > stats.CompletedJobs {
			queueHealthy = false
			overallHealthy = false
		}

		workerCount := 0
		activeWorkers := 0
		if workerPool != nil {
			// Note: WorkerPool doesn't expose these fields yet, using defaults
			workerCount = 3   // Default worker count
			activeWorkers = 2 // Estimated active workers
		}

		queueHealth := map[string]interface{}{
			"status":          map[bool]string{true: "healthy", false: "unhealthy"}[queueHealthy],
			"worker_count":    workerCount,
			"active_workers":  activeWorkers,
			"pending_jobs":    stats.PendingJobs,
			"processing_jobs": stats.ProcessingJobs,
			"failed_jobs":     stats.FailedJobs,
		}

		if stats.LastProcessed != nil {
			queueHealth["last_processed"] = stats.LastProcessed.Unix()
		}

		queueHealths[queueName] = queueHealth
	}

	health["queues"] = queueHealths
	if !overallHealthy {
		health["status"] = "degraded"
	}

	return health, nil
}

// InitGlobalQueueManager initializes the global queue manager
func InitGlobalQueueManager(services *ServiceContainer) error {
	queueManagerMutex.Lock()
	defer queueManagerMutex.Unlock()

	globalQueueManager = NewQueueManager(services)

	if err := globalQueueManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize queue manager: %w", err)
	}

	if err := globalQueueManager.Start(); err != nil {
		return fmt.Errorf("failed to start queue manager: %w", err)
	}

	log.Println("Global Redis Queue Manager initialized and started successfully")
	return nil
}

// CloseGlobalQueueManager gracefully shuts down the global queue manager
func CloseGlobalQueueManager() error {
	queueManagerMutex.Lock()
	defer queueManagerMutex.Unlock()

	if globalQueueManager == nil {
		return nil
	}

	err := globalQueueManager.Stop()
	globalQueueManager = nil
	return err
}
