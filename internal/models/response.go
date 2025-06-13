package models

// TokenResponse represents the response for successful authentication
type TokenResponse struct {
	Token     string `json:"token"`
	CSRFToken string `json:"csrf_token"`
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}

// Centralized definition of ErrorResponse to avoid redeclaration issues.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// LogoutResponse represents the response when logging out
type LogoutResponse struct {
	Message string `json:"message"`
}

// VideoProcessingJobResponse represents a processing job in API responses
type VideoProcessingJobResponse struct {
	ID            uint   `json:"id"`
	VideoID       uint   `json:"video_id"`
	VideoTitle    string `json:"video_title"`
	JobType       string `json:"job_type"`
	Status        string `json:"status"`
	Progress      int    `json:"progress"`
	ErrorMsg      string `json:"error_msg,omitempty"`
	StartedAt     *int64 `json:"started_at"`      // Unix timestamp
	CompletedAt   *int64 `json:"completed_at"`    // Unix timestamp
	CreatedAt     int64  `json:"created_at"`      // Unix timestamp
	LastUpdatedAt int64  `json:"last_updated_at"` // Unix timestamp
}

// ProcessingJobResponse represents a video processing job response for manual triggers
type ProcessingJobResponse struct {
	Message string `json:"message"`
	VideoID uint   `json:"video_id"`
	Status  string `json:"status"`
	JobID   *uint  `json:"job_id,omitempty"`
}

// PaginatedProcessingJobsResponse represents paginated processing jobs response
type PaginatedProcessingJobsResponse struct {
	Jobs       []VideoProcessingJobResponse `json:"jobs"`
	Pagination PaginationInfo               `json:"pagination"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"total_pages"`
}

// VideoProcessingStatusResponse represents video processing status in API responses
type VideoProcessingStatusResponse struct {
	VideoID            uint                         `json:"video_id"`
	ProcessingStatus   string                       `json:"processing_status"`
	ThumbnailURL       string                       `json:"thumbnail_url"`
	ProcessingProgress int                          `json:"processing_progress"`
	ProcessingError    string                       `json:"processing_error"`
	LastProcessedAt    *int64                       `json:"last_processed_at"` // Unix timestamp
	ProcessingJobs     []VideoProcessingJobResponse `json:"processing_jobs"`
}
