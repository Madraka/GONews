package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// JobProcessor defines the interface for processing jobs
type JobProcessor interface {
	ProcessJob(ctx context.Context, job *Job) error
	GetJobTypes() []string
}

// WorkerPool manages multiple workers for job processing
type WorkerPool struct {
	queue      *RedisQueue
	processors map[string]JobProcessor
	workers    int
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	stopping   bool
	mu         sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(queueName string, workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		queue:      NewRedisQueue(queueName),
		processors: make(map[string]JobProcessor),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterProcessor registers a job processor for specific job types
func (wp *WorkerPool) RegisterProcessor(processor JobProcessor) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for _, jobType := range processor.GetJobTypes() {
		wp.processors[jobType] = processor
		log.Printf("Registered processor for job type: %s", jobType)
	}
}

// Start begins processing jobs with the specified number of workers
func (wp *WorkerPool) Start() error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopping {
		return fmt.Errorf("worker pool is stopping")
	}

	log.Printf("Starting worker pool with %d workers", wp.workers)

	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Start monitoring goroutine
	wp.wg.Add(1)
	go wp.monitor()

	return nil
}

// Stop gracefully stops all workers
func (wp *WorkerPool) Stop() error {
	wp.mu.Lock()
	wp.stopping = true
	wp.mu.Unlock()

	log.Println("Stopping worker pool...")

	// Cancel context to signal workers to stop
	wp.cancel()

	// Wait for all workers to finish
	wp.wg.Wait()

	log.Println("Worker pool stopped")
	return nil
}

// Enqueue adds a job to the queue
func (wp *WorkerPool) Enqueue(job *Job) error {
	if wp.queue == nil {
		return fmt.Errorf("queue not available")
	}
	return wp.queue.Enqueue(job)
}

// GetStats returns worker pool and queue statistics
func (wp *WorkerPool) GetStats() (map[string]interface{}, error) {
	queueStats, err := wp.queue.GetQueueStats()
	if err != nil {
		return nil, err
	}

	wp.mu.RLock()
	processorCount := len(wp.processors)
	isRunning := !wp.stopping
	wp.mu.RUnlock()

	stats := map[string]interface{}{
		"workers":               wp.workers,
		"registered_processors": processorCount,
		"is_running":            isRunning,
		"queue_stats":           queueStats,
	}

	return stats, nil
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(workerID int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", workerID)
	defer log.Printf("Worker %d stopped", workerID)

	for {
		select {
		case <-wp.ctx.Done():
			return
		default:
			wp.processNextJobBlocking(workerID)
		}
	}
}

// processNextJobBlocking attempts to process the next available job using blocking dequeue
func (wp *WorkerPool) processNextJobBlocking(workerID int) {
	// Use blocking dequeue with 30 second timeout to reduce CPU usage
	job, err := wp.queue.BlockingDequeue(30 * time.Second)
	if err != nil {
		log.Printf("Worker %d: Error dequeuing job: %v", workerID, err)
		return
	}

	if job == nil {
		// Timeout occurred or no jobs available, just continue
		return
	}

	log.Printf("Worker %d: Processing job %s (type: %s)", workerID, job.ID, job.Type)

	// Find processor for this job type
	wp.mu.RLock()
	processor, exists := wp.processors[job.Type]
	wp.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("no processor registered for job type: %s", job.Type)
		log.Printf("Worker %d: %v", workerID, err)
		if failErr := wp.queue.FailJob(job.ID, err.Error()); failErr != nil {
			log.Printf("Worker %d: Failed to mark job as failed: %v", workerID, failErr)
		}
		return
	}

	// Process the job
	ctx, cancel := context.WithTimeout(wp.ctx, 10*time.Minute) // 10 minute timeout
	defer cancel()

	err = processor.ProcessJob(ctx, job)
	if err != nil {
		log.Printf("Worker %d: Job %s failed: %v", workerID, job.ID, err)

		// Handle job failure with retry logic
		if retryErr := wp.queue.FailJob(job.ID, err.Error()); retryErr != nil {
			log.Printf("Worker %d: Error handling job failure: %v", workerID, retryErr)
		}
		return
	}

	// Job completed successfully
	log.Printf("Worker %d: Job %s completed successfully", workerID, job.ID)
	if completeErr := wp.queue.CompleteJob(job.ID, nil); completeErr != nil {
		log.Printf("Worker %d: Error marking job as complete: %v", workerID, completeErr)
	}
}

// monitor provides periodic stats and health checks
func (wp *WorkerPool) monitor() {
	defer wp.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-ticker.C:
			wp.logStats()
		}
	}
}

// logStats logs current worker pool statistics
func (wp *WorkerPool) logStats() {
	stats, err := wp.GetStats()
	if err != nil {
		log.Printf("Error getting worker pool stats: %v", err)
		return
	}

	queueStats, ok := stats["queue_stats"].(map[string]interface{})
	if !ok {
		return
	}

	log.Printf("Worker Pool Stats - Workers: %v, Processors: %v, Pending Jobs: %v, Total Jobs: %v",
		stats["workers"],
		stats["registered_processors"],
		queueStats["pending_jobs"],
		queueStats["total_jobs"])
}

// Global worker pool instance for easy access
var defaultWorkerPool *WorkerPool

// InitDefaultWorkerPool initializes the default worker pool
func InitDefaultWorkerPool(queueName string, workers int) error {
	defaultWorkerPool = NewWorkerPool(queueName, workers)
	return nil
}

// GetDefaultWorkerPool returns the default worker pool instance
func GetDefaultWorkerPool() *WorkerPool {
	return defaultWorkerPool
}

// StartDefaultWorkerPool starts the default worker pool
func StartDefaultWorkerPool() error {
	if defaultWorkerPool == nil {
		return fmt.Errorf("default worker pool not initialized")
	}
	return defaultWorkerPool.Start()
}

// StopDefaultWorkerPool stops the default worker pool
func StopDefaultWorkerPool() error {
	if defaultWorkerPool == nil {
		return nil
	}
	return defaultWorkerPool.Stop()
}
