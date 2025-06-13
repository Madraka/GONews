package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"news/internal/metrics"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Code      string                 `json:"code,omitempty"`
}

// Common error types to standardize responses
var (
	ErrBadRequest         = errors.New("bad_request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not_found")
	ErrMethodNotAllowed   = errors.New("method_not_allowed")
	ErrConflict           = errors.New("conflict")
	ErrTooManyRequests    = errors.New("too_many_requests")
	ErrInternalServer     = errors.New("internal_server_error")
	ErrServiceUnavailable = errors.New("service_unavailable")
	ErrValidation         = errors.New("validation_error")
	ErrDatabaseConnection = errors.New("database_connection_error")
	ErrCacheConnection    = errors.New("cache_connection_error")
	ErrResourceNotFound   = errors.New("resource_not_found")
	ErrInvalidCredentials = errors.New("invalid_credentials")
	ErrInvalidToken       = errors.New("invalid_token")
	ErrExpiredToken       = errors.New("expired_token")
	ErrInsufficientRights = errors.New("insufficient_rights")
)

// HTTP status code mapping
var errorStatusCodes = map[error]int{
	ErrBadRequest:         http.StatusBadRequest,
	ErrUnauthorized:       http.StatusUnauthorized,
	ErrForbidden:          http.StatusForbidden,
	ErrNotFound:           http.StatusNotFound,
	ErrMethodNotAllowed:   http.StatusMethodNotAllowed,
	ErrConflict:           http.StatusConflict,
	ErrTooManyRequests:    http.StatusTooManyRequests,
	ErrInternalServer:     http.StatusInternalServerError,
	ErrServiceUnavailable: http.StatusServiceUnavailable,
	ErrValidation:         http.StatusUnprocessableEntity,

	// Custom errors mapped to standard HTTP codes
	ErrDatabaseConnection: http.StatusServiceUnavailable,
	ErrCacheConnection:    http.StatusServiceUnavailable,
	ErrResourceNotFound:   http.StatusNotFound,
	ErrInvalidCredentials: http.StatusUnauthorized,
	ErrInvalidToken:       http.StatusUnauthorized,
	ErrExpiredToken:       http.StatusUnauthorized,
	ErrInsufficientRights: http.StatusForbidden,
}

// Error descriptions for client-friendly messages
var errorDescriptions = map[error]string{
	ErrBadRequest:         "The request was invalid",
	ErrUnauthorized:       "Authentication is required",
	ErrForbidden:          "You don't have permission to access this resource",
	ErrNotFound:           "The requested resource was not found",
	ErrMethodNotAllowed:   "The HTTP method is not allowed for this resource",
	ErrConflict:           "The request conflicts with the current state of the resource",
	ErrTooManyRequests:    "Too many requests, please try again later",
	ErrInternalServer:     "An internal server error occurred",
	ErrServiceUnavailable: "The service is currently unavailable",
	ErrValidation:         "Validation failed",

	// Custom error descriptions
	ErrDatabaseConnection: "Database connection error",
	ErrCacheConnection:    "Cache connection error",
	ErrResourceNotFound:   "The requested resource was not found",
	ErrInvalidCredentials: "Invalid username or password",
	ErrInvalidToken:       "Invalid authentication token",
	ErrExpiredToken:       "Authentication token has expired",
	ErrInsufficientRights: "Insufficient access rights",
}

// ErrorHandlingMiddleware handles errors in a structured way
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Track the error in metrics
			metrics.RequestsTotal.WithLabelValues(c.FullPath(), c.Request.Method, fmt.Sprintf("%d", c.Writer.Status())).Inc()

			// Create standard error response
			httpStatus := http.StatusInternalServerError
			errorMsg := "An unexpected error occurred"
			var details map[string]interface{}
			var errorCode string

			// Determine the type of error
			switch e := err.Err.(type) {
			case validator.ValidationErrors:
				// Validation errors
				httpStatus = http.StatusUnprocessableEntity
				errorMsg = "Validation failed"
				errorCode = ErrValidation.Error()
				details = make(map[string]interface{})

				for _, fieldErr := range e {
					details[fieldErr.Field()] = fmt.Sprintf(
						"Field validation for '%s' failed on the '%s' tag",
						fieldErr.Field(),
						fieldErr.Tag(),
					)
				}

			case *net.OpError:
				// Network related errors
				httpStatus = http.StatusServiceUnavailable
				errorMsg = "Network operation failed"
				errorCode = "network_error"
				details = map[string]interface{}{
					"operation": e.Op,
					"network":   e.Net,
				}

			default:
				// Check if it's a known error type
				for errType, code := range errorStatusCodes {
					if errors.Is(e, errType) {
						httpStatus = code
						errorMsg = errorDescriptions[errType]
						errorCode = errType.Error()
						break
					}
				}
			}

			// Create the response
			response := ErrorResponse{
				Error:     errorCode,
				Message:   errorMsg,
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().Format(time.RFC3339),
				RequestID: c.GetString("request_id"),
				Code:      errorCode,
			}

			if details != nil {
				response.Details = details
			}

			// Add environment specific information in development mode
			if os.Getenv("GIN_MODE") != "release" {
				// Add stack trace or additional details in development
				if err.Meta != nil {
					if response.Details == nil {
						response.Details = make(map[string]interface{})
					}
					response.Details["meta"] = err.Meta
				}

				// Add the original error message in development
				if response.Details == nil {
					response.Details = make(map[string]interface{})
				}
				response.Details["raw_error"] = err.Err.Error()
			}

			// Return JSON response
			c.JSON(httpStatus, response)

			// Abort further processing
			c.Abort()
		}
	}
}

// NewError creates a new error with a type and optional details
func NewError(errType error, message string, details ...interface{}) error {
	if message == "" && len(details) == 0 {
		return errType
	}

	if message == "" {
		message = errorDescriptions[errType]
	}

	errorMessage := fmt.Sprintf("%s: %s", errType.Error(), message)

	if len(details) > 0 {
		detailsStr := make([]string, len(details))
		for i, detail := range details {
			detailsStr[i] = fmt.Sprintf("%v", detail)
		}
		errorMessage += " (" + strings.Join(detailsStr, ", ") + ")"
	}

	return fmt.Errorf("%w: %s", errType, errorMessage)
}
