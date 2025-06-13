package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"news/internal/auth"
	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// SecurityAuditHandler handles security-related user operations
type SecurityAuditHandler struct {
	tokenManager *auth.TokenManager
}

// NewSecurityAuditHandler creates a new security audit handler
func NewSecurityAuditHandler(tokenManager *auth.TokenManager) *SecurityAuditHandler {
	return &SecurityAuditHandler{
		tokenManager: tokenManager,
	}
}

// LoginHistoryResponse represents a login history entry in the API response
type LoginHistoryResponse struct {
	ID            uint      `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"user_agent"`
	Location      string    `json:"location,omitempty"`
	Success       bool      `json:"success"`
	FailureReason string    `json:"failure_reason,omitempty"`
}

// GetUserSessions returns all active sessions for a user
// @Summary Get User Sessions
// @Description Get all active sessions for the authenticated user
// @Tags Security
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "sessions"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /security/sessions [get]
func (h *SecurityAuditHandler) GetUserSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var sessions []models.UserSession
	if err := database.DB.Where("user_id = ? AND active = ?", userID, true).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
	})
}

// RevokeSession revokes a specific user session
// @Summary Revoke User Session
// @Description Revoke a specific active session for the authenticated user
// @Tags Security
// @Security BearerAuth
// @Param session_id path string true "Session ID"
// @Produce json
// @Success 200 {object} map[string]interface{} "message"
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /security/sessions/{session_id} [delete]
func (h *SecurityAuditHandler) RevokeSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	sessionID := c.Param("session_id")

	var session models.UserSession
	if err := database.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Session not found"})
		return
	}

	// Revoke the session
	session.Active = false
	// Set RevokedAt to current time pointer
	now := time.Now()
	session.RevokedAt = &now

	if err := database.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to revoke session"})
		return
	}

	// Add token to blacklist if needed
	if session.TokenID != "" {
		expiry := time.Unix(session.ExpiresAt, 0)
		if err := h.tokenManager.BlacklistToken(session.TokenID, expiry); err != nil {
			log.Printf("Warning: Failed to blacklist token %s: %v", session.TokenID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all active sessions for a user except the current one
// @Summary Revoke All Sessions
// @Description Revoke all active sessions for the authenticated user except the current one
// @Tags Security
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "message and revoked_count"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /security/sessions [delete]
func (h *SecurityAuditHandler) RevokeAllSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	currentTokenID, _ := c.Get("tokenID")

	var sessions []models.UserSession
	if err := database.DB.Where("user_id = ? AND active = ? AND token_id != ?",
		userID, true, currentTokenID).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user sessions"})
		return
	}

	// Revoke all sessions
	now := time.Now()
	for _, session := range sessions {
		// Update session in database
		session.Active = false
		// Set RevokedAt to current time pointer
		session.RevokedAt = &now
		database.DB.Save(&session)

		// Add token to blacklist
		if session.TokenID != "" {
			expiry := time.Unix(session.ExpiresAt, 0)
			if err := h.tokenManager.BlacklistToken(session.TokenID, expiry); err != nil {
				log.Printf("Warning: Failed to blacklist token %s: %v", session.TokenID, err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "All other sessions revoked successfully",
		"revoked_count": len(sessions),
	})
}

// GetLoginHistory returns the user's login history
// @Summary Get Login History
// @Description Get the login history for the authenticated user
// @Tags Security
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "login_history"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /security/login-history [get]
func (h *SecurityAuditHandler) GetLoginHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var loginHistory []models.LoginAttempt
	if err := database.DB.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(50).Find(&loginHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch login history"})
		return
	}

	// Convert to response format
	response := make([]LoginHistoryResponse, len(loginHistory))
	for i, login := range loginHistory {
		response[i] = LoginHistoryResponse{
			ID:            login.ID,
			Timestamp:     login.Timestamp,
			IP:            login.IP,
			UserAgent:     login.UserAgent,
			Location:      login.Location,
			Success:       login.Success,
			FailureReason: login.FailureReason,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"login_history": response,
	})
}

// GetSecurityEvents returns security-related events for a user
// @Summary Get Security Events
// @Description Get security-related events for the authenticated user
// @Tags Security
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "security_events"
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /security/events [get]
func (h *SecurityAuditHandler) GetSecurityEvents(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var securityEvents []models.SecurityEvent
	if err := database.DB.Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(50).Find(&securityEvents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch security events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"security_events": securityEvents,
	})
}

// CreateUserSession creates a new user session record
func CreateUserSession(userID uint, tokenID, ip, userAgent string, expiresAt int64) error {
	// Parse device info from user agent (simplified)
	var device string
	if strings.Contains(userAgent, "Mobile") {
		device = "Mobile"
	} else if strings.Contains(userAgent, "Tablet") {
		device = "Tablet"
	} else {
		device = "Desktop"
	}

	session := models.UserSession{
		UserID:    userID,
		TokenID:   tokenID,
		IP:        ip,
		UserAgent: userAgent,
		Device:    device,
		ExpiresAt: expiresAt,
		Active:    true,
	}

	return database.DB.Create(&session).Error
}

// UpdateUserSession updates session activity
func UpdateUserSession(tokenID string) error {
	return database.DB.Model(&models.UserSession{}).
		Where("token_id = ? AND active = ?", tokenID, true).
		Update("updated_at", time.Now()).Error
}

// RevokeUserSession revokes a user session
func RevokeUserSession(tokenID string) error {
	now := time.Now()
	return database.DB.Model(&models.UserSession{}).
		Where("token_id = ?", tokenID).
		Updates(map[string]interface{}{
			"active":     false,
			"revoked_at": &now,
		}).Error
}
