package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"news/internal/database"
	"news/internal/queue"
	"news/internal/services"
)

func main() {
	log.Println("Starting News Queue Worker...")

	// Initialize database connection
	log.Println("Connecting to database...")
	database.Connect()
	defer func() {
		if db := database.DB; db != nil {
			if sqlDB, err := db.DB(); err == nil {
				if closeErr := sqlDB.Close(); closeErr != nil {
					log.Printf("Warning: Error closing database connection: %v", closeErr)
				}
			} else {
				log.Printf("Warning: Error getting database instance: %v", err)
			}
		}
	}()

	// Initialize AI service
	log.Println("Initializing AI service...")
	aiService := services.GetAIService()
	if aiService == nil {
		log.Fatal("Failed to initialize AI service")
	}

	// Create AI translation service
	translationService := services.NewAITranslationService(aiService)

	log.Println("Initializing queue manager for worker...")

	// Create service container for worker
	serviceContainer := &queue.ServiceContainer{
		TranslationService:     translationService,
		VideoProcessingService: services.GetGlobalVideoProcessingService(),
	}

	// Create queue manager
	queueManager := queue.NewQueueManager(serviceContainer)

	// Initialize and start the queue manager with workers
	log.Println("Initializing queues and workers...")
	if err := queueManager.Initialize(); err != nil {
		log.Fatalf("Failed to initialize queue manager: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the queue manager with all workers
	log.Println("Starting queue workers...")
	if err := queueManager.Start(); err != nil {
		log.Fatalf("Failed to start queue manager: %v", err)
	}

	// Start health check and stats reporting
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Report worker stats
				stats := queueManager.GetQueueStats()
				log.Printf("Worker Stats: %+v", stats)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Worker is running. Press Ctrl+C to stop.")
	<-sigChan

	log.Println("Shutting down worker...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if stopErr := queueManager.Stop(); stopErr != nil {
			log.Printf("Warning: Error stopping queue manager: %v", stopErr)
		}
	}()

	// Wait for shutdown or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Worker stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("Worker shutdown timeout")
	}
}
