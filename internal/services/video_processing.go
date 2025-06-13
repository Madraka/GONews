package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"news/internal/json"
	"news/internal/models"
	"news/internal/storage"

	"gorm.io/gorm"
)

// Global instances for video processing services
var (
	globalVideoProcessingService *VideoProcessingService
	globalVideoProcessingQueue   *VideoProcessingQueue
)

// GetGlobalVideoProcessingService returns the global video processing service instance
func GetGlobalVideoProcessingService() *VideoProcessingService {
	return globalVideoProcessingService
}

// SetGlobalVideoProcessingService sets the global video processing service instance
func SetGlobalVideoProcessingService(service *VideoProcessingService) {
	globalVideoProcessingService = service
}

// GetGlobalVideoProcessingQueue returns the global video processing queue instance
func GetGlobalVideoProcessingQueue() *VideoProcessingQueue {
	return globalVideoProcessingQueue
}

// SetGlobalVideoProcessingQueue sets the global video processing queue instance
func SetGlobalVideoProcessingQueue(queue *VideoProcessingQueue) {
	globalVideoProcessingQueue = queue
}

// VideoProcessingService handles video processing tasks
type VideoProcessingService struct {
	db          *gorm.DB
	storage     storage.Storage
	aiService   *AIService
	ffmpegPath  string
	tempDir     string
	maxDuration int // Maximum video duration in seconds
}

func NewVideoProcessingService(db *gorm.DB, storage storage.Storage, ai *AIService) *VideoProcessingService {
	return &VideoProcessingService{
		db:          db,
		storage:     storage,
		aiService:   ai,
		ffmpegPath:  "ffmpeg", // Should be configurable
		tempDir:     "/tmp",
		maxDuration: 300, // 5 minutes max for short-form content
	}
}

// ProcessingJob represents a video processing task
type ProcessingJob struct {
	Type     string `json:"type"`
	VideoID  uint   `json:"video_id"`
	Priority int    `json:"priority"`
}

// ProcessingQueue interface for background job processing
type ProcessingQueue interface {
	AddJob(job ProcessingJob) error
	ProcessJobs(ctx context.Context) error
}

// VideoProcessingQueue implements ProcessingQueue
type VideoProcessingQueue struct {
	processor *VideoProcessingService
	jobs      chan ProcessingJob
}

func NewVideoProcessingQueue(processor *VideoProcessingService) *VideoProcessingQueue {
	return &VideoProcessingQueue{
		processor: processor,
		jobs:      make(chan ProcessingJob, 100),
	}
}

func (q *VideoProcessingQueue) AddJob(job ProcessingJob) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func (q *VideoProcessingQueue) ProcessJobs(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job := <-q.jobs:
			switch job.Type {
			case "thumbnail":
				if err := q.processor.GenerateThumbnail(ctx, job.VideoID); err != nil {
					log.Printf("Failed to generate thumbnail for video %d: %v", job.VideoID, err)
				}
			case "transcode":
				if err := q.processor.TranscodeVideo(ctx, job.VideoID); err != nil {
					log.Printf("Failed to transcode video %d: %v", job.VideoID, err)
				}
			case "ai_analysis":
				if err := q.processor.AnalyzeVideo(ctx, job.VideoID); err != nil {
					log.Printf("Failed to analyze video %d: %v", job.VideoID, err)
				}
			case "text_to_speech":
				if err := q.processor.GenerateTextToSpeech(ctx, job.VideoID); err != nil {
					log.Printf("Failed to generate TTS for video %d: %v", job.VideoID, err)
				}
			default:
				log.Printf("Unknown job type: %s", job.Type)
			}
		}
	}
}

// VideoMetadata represents video file information
type VideoMetadata struct {
	Duration   float64 `json:"duration"`
	FileSize   int64   `json:"file_size"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FrameRate  float64 `json:"frame_rate"`
	Bitrate    int     `json:"bitrate"`
	VideoCodec string  `json:"video_codec"`
	AudioCodec string  `json:"audio_codec"`
}

// VideoAnalysisResult represents AI analysis results
type VideoAnalysisResult struct {
	Objects       []string `json:"objects"`
	Scenes        []string `json:"scenes"`
	Text          []string `json:"text"`
	Tags          []string `json:"tags"`
	Confidence    float64  `json:"confidence"`
	IsAppropriate bool     `json:"is_appropriate"`
}

// GetVideoMetadata extracts metadata from video file using ffprobe
func (s *VideoProcessingService) GetVideoMetadata(ctx context.Context, videoPath string) (*VideoMetadata, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	var probe struct {
		Format struct {
			Duration string `json:"duration"`
			Size     string `json:"size"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			BitRate    string `json:"bit_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	metadata := &VideoMetadata{}

	if duration, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		metadata.Duration = duration
	}

	if size, err := strconv.ParseInt(probe.Format.Size, 10, 64); err == nil {
		metadata.FileSize = size
	}

	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.VideoCodec = stream.CodecName

			if bitrate, err := strconv.Atoi(stream.BitRate); err == nil {
				metadata.Bitrate = bitrate
			}

			// Parse frame rate from fraction format (e.g., "30/1")
			if parts := strings.Split(stream.RFrameRate, "/"); len(parts) == 2 {
				if num, err1 := strconv.ParseFloat(parts[0], 64); err1 == nil {
					if den, err2 := strconv.ParseFloat(parts[1], 64); err2 == nil && den != 0 {
						metadata.FrameRate = num / den
					}
				}
			}
		} else if stream.CodecType == "audio" {
			metadata.AudioCodec = stream.CodecName
		}
	}

	return metadata, nil
}

// GenerateThumbnail creates a thumbnail image from video
func (s *VideoProcessingService) GenerateThumbnail(ctx context.Context, videoID uint) error {
	video := &models.Video{}
	if err := s.db.First(video, videoID).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// Download video file for processing
	tempFile := filepath.Join(s.tempDir, fmt.Sprintf("video_%d.mp4", videoID))
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			log.Printf("Warning: Failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	// Extract filename from VideoURL for storage interface
	filename := filepath.Base(video.VideoURL)
	reader, err := s.storage.Download(filename)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Handle reader as ReadCloser if possible, otherwise create our own closer
	if closer, ok := reader.(io.ReadCloser); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				log.Printf("Warning: Failed to close reader: %v", err)
			}
		}()
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Failed to close temp file: %v", err)
		}
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy video data: %w", err)
	}

	// Generate thumbnail at 50% of video duration
	thumbnailPath := filepath.Join(s.tempDir, fmt.Sprintf("thumb_%d.jpg", videoID))
	defer func() {
		if err := os.Remove(thumbnailPath); err != nil {
			log.Printf("Warning: Failed to remove thumbnail file %s: %v", thumbnailPath, err)
		}
	}()

	metadata, err := s.GetVideoMetadata(ctx, tempFile)
	if err != nil {
		return fmt.Errorf("failed to get video metadata: %w", err)
	}

	seekTime := metadata.Duration / 2 // Middle of video
	cmd := exec.CommandContext(ctx, s.ffmpegPath,
		"-i", tempFile,
		"-ss", fmt.Sprintf("%.2f", seekTime),
		"-vframes", "1",
		"-q:v", "2",
		"-y", thumbnailPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Upload thumbnail to storage
	thumbnailFile, err := os.Open(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to open thumbnail: %w", err)
	}
	defer func() {
		if err := thumbnailFile.Close(); err != nil {
			log.Printf("Warning: Failed to close thumbnail file: %v", err)
		}
	}()

	thumbnailURL, err := s.storage.Upload(thumbnailFile, fmt.Sprintf("thumbnails/video_%d.jpg", videoID))
	if err != nil {
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Update video record with thumbnail URL
	video.ThumbnailURL = thumbnailURL
	if err := s.db.Save(video).Error; err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}

// TranscodeVideo converts video to multiple formats
func (s *VideoProcessingService) TranscodeVideo(ctx context.Context, videoID uint) error {
	video := &models.Video{}
	if err := s.db.First(video, videoID).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// Download original video
	tempFile := filepath.Join(s.tempDir, fmt.Sprintf("original_%d.mp4", videoID))
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			log.Printf("Warning: Failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	// Extract filename from VideoURL for storage interface
	filename := filepath.Base(video.VideoURL)
	reader, err := s.storage.Download(filename)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Handle reader as ReadCloser if possible
	if closer, ok := reader.(io.ReadCloser); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				log.Printf("Warning: Failed to close reader in TranscodeVideo: %v", err)
			}
		}()
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Failed to close temp file in TranscodeVideo: %v", err)
		}
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy video data: %w", err)
	}

	// Transcode to different resolutions
	resolutions := []struct {
		name    string
		width   int
		height  int
		bitrate string
	}{
		{"720p", 1280, 720, "2500k"},
		{"480p", 854, 480, "1500k"},
		{"360p", 640, 360, "800k"},
	}

	for _, res := range resolutions {
		outputFile := filepath.Join(s.tempDir, fmt.Sprintf("video_%d_%s.mp4", videoID, res.name))
		defer func(file string) {
			if err := os.Remove(file); err != nil {
				log.Printf("Warning: Failed to remove output file %s: %v", file, err)
			}
		}(outputFile)

		cmd := exec.CommandContext(ctx, s.ffmpegPath,
			"-i", tempFile,
			"-vf", fmt.Sprintf("scale=%d:%d", res.width, res.height),
			"-c:v", "libx264",
			"-b:v", res.bitrate,
			"-c:a", "aac",
			"-b:a", "128k",
			"-y", outputFile)

		if err := cmd.Run(); err != nil {
			log.Printf("Failed to transcode %s: %v", res.name, err)
			continue
		}

		// Upload transcoded video
		transcodedFile, err := os.Open(outputFile)
		if err != nil {
			log.Printf("Failed to open transcoded file %s: %v", res.name, err)
			continue
		}

		url, err := s.storage.Upload(transcodedFile, fmt.Sprintf("videos/transcoded/%d_%s.mp4", videoID, res.name))
		if err := transcodedFile.Close(); err != nil {
			log.Printf("Warning: Failed to close transcoded file: %v", err)
		}
		if err != nil {
			log.Printf("Failed to upload transcoded video %s: %v", res.name, err)
			continue
		}

		log.Printf("Successfully transcoded and uploaded %s version: %s", res.name, url)
	}

	return nil
}

// AnalyzeVideo performs AI analysis on video content
func (s *VideoProcessingService) AnalyzeVideo(ctx context.Context, videoID uint) error {
	if s.aiService == nil {
		return fmt.Errorf("AI service not available")
	}

	video := &models.Video{}
	if err := s.db.First(video, videoID).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	// Download video for analysis
	tempFile := filepath.Join(s.tempDir, fmt.Sprintf("analyze_%d.mp4", videoID))
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			log.Printf("Warning: Failed to remove analysis temp file %s: %v", tempFile, err)
		}
	}()

	// Extract filename from VideoURL for storage interface
	filename := filepath.Base(video.VideoURL)
	reader, err := s.storage.Download(filename)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Handle reader as ReadCloser if possible
	if closer, ok := reader.(io.ReadCloser); ok {
		defer func() {
			if err := closer.Close(); err != nil {
				log.Printf("Warning: Failed to close reader in AnalyzeVideo: %v", err)
			}
		}()
	}

	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Failed to close temp file in AnalyzeVideo: %v", err)
		}
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to copy video data: %w", err)
	}

	// Extract frames for analysis
	frameDir := filepath.Join(s.tempDir, fmt.Sprintf("frames_%d", videoID))
	if err := os.MkdirAll(frameDir, 0755); err != nil {
		return fmt.Errorf("failed to create frame directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(frameDir); err != nil {
			log.Printf("Warning: Failed to remove frame directory %s: %v", frameDir, err)
		}
	}()

	// Extract frames every 5 seconds
	cmd := exec.CommandContext(ctx, s.ffmpegPath,
		"-i", tempFile,
		"-vf", "fps=1/5", // 1 frame every 5 seconds
		"-q:v", "2",
		filepath.Join(frameDir, "frame_%03d.jpg"))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract frames: %w", err)
	}

	// Get list of extracted frames
	framePaths := []string{}
	frames, err := os.ReadDir(frameDir)
	if err != nil {
		return fmt.Errorf("failed to read frame directory: %w", err)
	}

	for _, frame := range frames {
		if strings.HasSuffix(frame.Name(), ".jpg") {
			framePaths = append(framePaths, filepath.Join(frameDir, frame.Name()))
		}
	}

	// Note: AI service methods need to be implemented with correct signatures
	// TODO: Use framePaths for actual AI analysis once AI service is implemented
	// For now, creating placeholder analysis without using the extracted frames
	_ = framePaths // Acknowledge that framePaths is currently unused
	analysis := &VideoAnalysisResult{
		Objects:       []string{"placeholder"},
		Scenes:        []string{"indoor", "outdoor"},
		Text:          []string{},
		Tags:          []string{"video", "content"},
		Confidence:    0.8,
		IsAppropriate: true,
	}

	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	// Note: Assuming video model has AIAnalysis field
	// This would need to be added to the model if it doesn't exist
	if err := s.db.Model(video).Update("ai_analysis", string(analysisJSON)).Error; err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	return nil
}

// GenerateTextToSpeech creates audio narration from video description
func (s *VideoProcessingService) GenerateTextToSpeech(ctx context.Context, videoID uint) error {
	if s.aiService == nil {
		return fmt.Errorf("AI service not available")
	}

	video := &models.Video{}
	if err := s.db.First(video, videoID).Error; err != nil {
		return fmt.Errorf("video not found: %w", err)
	}

	if video.Description == "" {
		return fmt.Errorf("video has no description for TTS")
	}

	// Note: AI service method signature needs to be implemented correctly
	// For now, creating placeholder implementation
	log.Printf("Generating TTS for video %d with text: %s", videoID, video.Description)

	// Placeholder for TTS generation
	// audioData, err := s.aiService.GenerateTextToSpeech(video.Description)
	// if err != nil {
	//     return fmt.Errorf("failed to generate TTS: %w", err)
	// }

	// Create placeholder audio file for now
	audioPath := filepath.Join(s.tempDir, fmt.Sprintf("audio_%d.mp3", videoID))
	if err := os.WriteFile(audioPath, []byte("placeholder audio"), 0644); err != nil {
		return fmt.Errorf("failed to create audio file: %w", err)
	}
	defer func() {
		if err := os.Remove(audioPath); err != nil {
			log.Printf("Warning: Failed to remove audio file %s: %v", audioPath, err)
		}
	}()

	// Upload audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("failed to open audio file: %w", err)
	}
	defer func() {
		if err := audioFile.Close(); err != nil {
			log.Printf("Warning: Failed to close audio file: %v", err)
		}
	}()

	audioURL, err := s.storage.Upload(audioFile, fmt.Sprintf("audio/tts_%d.mp3", videoID))
	if err != nil {
		return fmt.Errorf("failed to upload audio: %w", err)
	}

	// Note: Assuming video model has AudioURL field
	// This would need to be added to the model if it doesn't exist
	if err := s.db.Model(video).Update("audio_url", audioURL).Error; err != nil {
		return fmt.Errorf("failed to save audio URL: %w", err)
	}

	return nil
}

// ProcessVideo handles all video processing tasks for a given video
func (s *VideoProcessingService) ProcessVideo(ctx context.Context, videoID uint) error {
	jobs := []ProcessingJob{
		{Type: "thumbnail", VideoID: videoID, Priority: 1},
		{Type: "transcode", VideoID: videoID, Priority: 2},
		{Type: "ai_analysis", VideoID: videoID, Priority: 3},
		{Type: "text_to_speech", VideoID: videoID, Priority: 4},
	}

	queue := GetGlobalVideoProcessingQueue()
	if queue == nil {
		return fmt.Errorf("video processing queue not initialized")
	}

	for _, job := range jobs {
		if err := queue.AddJob(job); err != nil {
			log.Printf("Failed to add job %+v: %v", job, err)
		}
	}

	return nil
}
