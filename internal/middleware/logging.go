package middleware

import (
	"fmt"
	"io"
	"os"
	"time"

	"news/internal/json"

	"github.com/gin-gonic/gin"
)

// Log levels
const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
	LevelFatal   = "FATAL"
)

// Logger is a custom logger instance
type Logger struct {
	Out         io.Writer
	TimeFormat  string
	ServiceName string
	LogLevel    string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	Status      int                    `json:"status,omitempty"`
	Latency     float64                `json:"latency,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	ServiceName string                 `json:"service,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	// Get log level from environment, default to INFO
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = LevelInfo
	}

	return &Logger{
		Out:         os.Stdout,
		TimeFormat:  time.RFC3339,
		ServiceName: "news-api",
		LogLevel:    logLevel,
	}
}

// LoggingMiddleware is a middleware function that logs request details
func LoggingMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health check and metrics requests to reduce excessive logs
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Get request ID
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get response status
		status := c.Writer.Status()

		// Skip detailed logging for successful static file requests
		if status == 200 && (c.Request.Method == "GET" &&
			(c.ContentType() == "application/javascript" ||
				c.ContentType() == "text/css" ||
				c.ContentType() == "image/png" ||
				c.ContentType() == "image/jpeg" ||
				c.ContentType() == "image/svg+xml")) {
			return
		}

		// Get user ID if authenticated
		var userID string
		if user, exists := c.Get("user_id"); exists {
			userID = fmt.Sprintf("%v", user)
		}

		// Log request details
		entry := LogEntry{
			Timestamp:   time.Now().Format(logger.TimeFormat),
			Level:       getLogLevel(status),
			Message:     fmt.Sprintf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, status, latency),
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			Status:      status,
			Latency:     latency.Seconds(),
			RequestID:   requestID,
			ClientIP:    c.ClientIP(),
			ServiceName: logger.ServiceName,
			UserID:      userID,
		}

		// Add errors if any
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Log the entry
		logger.Log(entry)
	}
}

// GetLogLevelPriority returns numeric priority for log level
func GetLogLevelPriority(level string) int {
	switch level {
	case LevelDebug:
		return 0
	case LevelInfo:
		return 1
	case LevelWarning:
		return 2
	case LevelError:
		return 3
	case LevelFatal:
		return 4
	default:
		return 1 // Default to INFO
	}
}

// Log writes a log entry if the level is above or equal to the configured level
func (l *Logger) Log(entry LogEntry) {
	// Skip logging if the entry level is below the configured level
	if GetLogLevelPriority(entry.Level) < GetLogLevelPriority(l.LogLevel) {
		return
	}

	// Filter out noisy trace export errors which occur frequently
	if entry.Level == LevelError && entry.Message != "" {
		if entry.Error != "" && len(entry.Error) > 12 && entry.Error[:12] == "traces export" {
			// Only log trace export errors at debug level
			if l.LogLevel != LevelDebug {
				return
			}
		}
	}

	// Filter out excessive health and metrics endpoint requests if they were missed by the middleware
	if entry.Path == "/health" || entry.Path == "/metrics" {
		// Skip these completely unless we're at debug level
		if l.LogLevel != LevelDebug {
			return
		}
	}

	jsonEntry, _ := json.Marshal(entry)
	if _, err := fmt.Fprintln(l.Out, string(jsonEntry)); err != nil {
		// Fallback to stderr if we can't write to the configured output
		fmt.Fprintf(os.Stderr, "Warning: Failed to write log entry: %v\n", err)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, ctx ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(l.TimeFormat),
		Level:       LevelDebug,
		Message:     msg,
		ServiceName: l.ServiceName,
	}
	if len(ctx) > 0 {
		entry.Context = ctx[0]
	}
	l.Log(entry)
}

// Info logs an info message
func (l *Logger) Info(msg string, ctx ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(l.TimeFormat),
		Level:       LevelInfo,
		Message:     msg,
		ServiceName: l.ServiceName,
	}
	if len(ctx) > 0 {
		entry.Context = ctx[0]
	}
	l.Log(entry)
}

// Warning logs a warning message
func (l *Logger) Warning(msg string, ctx ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(l.TimeFormat),
		Level:       LevelWarning,
		Message:     msg,
		ServiceName: l.ServiceName,
	}
	if len(ctx) > 0 {
		entry.Context = ctx[0]
	}
	l.Log(entry)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, ctx ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(l.TimeFormat),
		Level:       LevelError,
		Message:     msg,
		ServiceName: l.ServiceName,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if len(ctx) > 0 {
		entry.Context = ctx[0]
	}
	l.Log(entry)
}

// Fatal logs a fatal error message and exits
func (l *Logger) Fatal(msg string, err error, ctx ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:   time.Now().Format(l.TimeFormat),
		Level:       LevelFatal,
		Message:     msg,
		ServiceName: l.ServiceName,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	if len(ctx) > 0 {
		entry.Context = ctx[0]
	}
	l.Log(entry)
	os.Exit(1)
}

// getLogLevel determines the log level based on status code
func getLogLevel(status int) string {
	switch {
	case status >= 500:
		return LevelError
	case status >= 400:
		return LevelWarning
	default:
		return LevelInfo
	}
}
