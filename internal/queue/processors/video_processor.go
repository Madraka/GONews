package processors

import (
	"context"
	"fmt"
	"log"

	"news/internal/queue"
	"news/internal/services"
)

// VideoProcessor handles video processing jobs for Redis queue
type VideoProcessor struct {
	videoService *services.VideoProcessingService
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(videoService *services.VideoProcessingService) *VideoProcessor {
	return &VideoProcessor{
		videoService: videoService,
	}
}

// ProcessJob processes a video processing job
func (vp *VideoProcessor) ProcessJob(ctx context.Context, job *queue.Job) error {
	log.Printf("Processing video job %s of type %s", job.ID, job.Type)

	// Extract video ID from job payload
	videoIDFloat, ok := job.Payload["video_id"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid video_id in job payload")
	}
	videoID := uint(videoIDFloat)

	// Process based on job type
	switch job.Type {
	case "video_thumbnail":
		return vp.videoService.GenerateThumbnail(ctx, videoID)

	case "video_transcode":
		return vp.videoService.TranscodeVideo(ctx, videoID)

	case "video_analysis":
		return vp.videoService.AnalyzeVideo(ctx, videoID)

	case "video_tts":
		return vp.videoService.GenerateTextToSpeech(ctx, videoID)

	case "video_processing_complete":
		// Complete video processing workflow
		return vp.processCompleteVideoWorkflow(ctx, videoID)

	default:
		return fmt.Errorf("unsupported video job type: %s", job.Type)
	}
}

// GetJobTypes returns the job types this processor handles
func (vp *VideoProcessor) GetJobTypes() []string {
	return []string{
		"video_thumbnail",
		"video_transcode",
		"video_analysis",
		"video_tts",
		"video_processing_complete",
	}
}

// processCompleteVideoWorkflow handles the complete video processing pipeline
func (vp *VideoProcessor) processCompleteVideoWorkflow(ctx context.Context, videoID uint) error {
	log.Printf("Starting complete video processing workflow for video %d", videoID)

	// This could process multiple steps or coordinate other jobs
	// For now, we'll use the existing ProcessVideo method
	return vp.videoService.ProcessVideo(ctx, videoID)
}

// Helper functions to create video processing jobs

// CreateVideoThumbnailJob creates a thumbnail generation job
func CreateVideoThumbnailJob(videoID uint, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "video_thumbnail",
		Priority: priority,
		Payload: map[string]interface{}{
			"video_id": videoID,
		},
		MaxAttempts: 3,
	}
}

// CreateVideoTranscodeJob creates a video transcoding job
func CreateVideoTranscodeJob(videoID uint, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "video_transcode",
		Priority: priority,
		Payload: map[string]interface{}{
			"video_id": videoID,
		},
		MaxAttempts: 2, // Transcoding is resource intensive, fewer retries
	}
}

// CreateVideoAnalysisJob creates an AI analysis job
func CreateVideoAnalysisJob(videoID uint, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "video_analysis",
		Priority: priority,
		Payload: map[string]interface{}{
			"video_id": videoID,
		},
		MaxAttempts: 3,
	}
}

// CreateVideoTTSJob creates a text-to-speech job
func CreateVideoTTSJob(videoID uint, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "video_tts",
		Priority: priority,
		Payload: map[string]interface{}{
			"video_id": videoID,
		},
		MaxAttempts: 3,
	}
}

// CreateCompleteVideoWorkflowJob creates a job that processes the entire video pipeline
func CreateCompleteVideoWorkflowJob(videoID uint, priority queue.JobPriority) *queue.Job {
	return &queue.Job{
		Type:     "video_processing_complete",
		Priority: priority,
		Payload: map[string]interface{}{
			"video_id": videoID,
		},
		MaxAttempts: 1, // Complete workflow should not retry
	}
}
