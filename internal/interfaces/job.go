package interfaces

import (
	"context"
)

// JobProcessor interface for processing different types of jobs
type JobProcessor interface {
	ProcessJob(ctx context.Context, job *Job) error
	GetJobType() string
}

// Job represents a generic queue job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   string                 `json:"created_at"`
	ScheduledAt string                 `json:"scheduled_at"`
	StartedAt   *string                `json:"started_at,omitempty"`
	CompletedAt *string                `json:"completed_at,omitempty"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
}
