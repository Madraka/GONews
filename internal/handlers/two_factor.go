package handlers

import (
	"fmt"
	"net/http"
	"time"

	"news/internal/auth"
	"news/internal/database"
	"news/internal/json"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// TwoFactorHandler handles 2FA-related operations
type TwoFactorHandler struct {
	totpManager *auth.TOTPManager
}

// NewTwoFactorHandler creates a new two-factor authentication handler
func NewTwoFactorHandler() *TwoFactorHandler {
	return &TwoFactorHandler{
		totpManager: auth.NewTOTPManager(),
	}
}

// Setup2FARequest represents the request to set up 2FA
type Setup2FARequest struct {
	TOTPCode string `json:"totp_code" binding:"required" example:"123456"`
}

// Setup2FAResponse represents the response for 2FA setup
type Setup2FAResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Verify2FARequest represents the request to verify 2FA
type Verify2FARequest struct {
	TOTPCode   string `json:"totp_code,omitempty" example:"123456"`
	BackupCode string `json:"backup_code,omitempty" example:"abcd-efgh-ijkl"`
}

// Setup2FA initiates 2FA setup for a user
// @Summary Setup Two-Factor Authentication
// @Description Generate TOTP secret and QR code for 2FA setup
// @Tags Two-Factor Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} Setup2FAResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /2fa/setup [post]
func (h *TwoFactorHandler) Setup2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	username, _ := c.Get("username")

	// Check if user already has 2FA enabled
	var existingTOTP models.UserTOTP
	if err := database.DB.Where("user_id = ?", userID).First(&existingTOTP).Error; err == nil {
		if existingTOTP.Enabled {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "2FA is already enabled for this user"})
			return
		}
	}

	// Generate new secret
	secret := h.totpManager.GenerateSecret()

	// Generate QR code URL
	qrURL := h.totpManager.GetQRCodeURL(secret, username.(string), "News Aggregation Service")

	// Generate backup codes (10 codes)
	backupCodes := generateBackupCodes(10)
	backupCodesJSON, _ := json.Marshal(backupCodes)

	// Save or update TOTP record (not enabled yet)
	userTOTP := models.UserTOTP{
		UserID:      userID.(uint),
		Secret:      secret,
		BackupCodes: string(backupCodesJSON),
		Enabled:     false,
	}

	if err := database.DB.Where("user_id = ?", userID).First(&existingTOTP).Error; err != nil {
		// Create new record
		if err := database.DB.Create(&userTOTP).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to setup 2FA"})
			return
		}
	} else {
		// Update existing record
		existingTOTP.Secret = secret
		existingTOTP.BackupCodes = string(backupCodesJSON)
		existingTOTP.Enabled = false
		if err := database.DB.Save(&existingTOTP).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to setup 2FA"})
			return
		}
	}

	c.JSON(http.StatusOK, Setup2FAResponse{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	})
}

// Enable2FA enables 2FA after verifying TOTP code
// @Summary Enable Two-Factor Authentication
// @Description Enable 2FA by verifying the TOTP code from authenticator app
// @Tags Two-Factor Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body Setup2FARequest true "TOTP verification"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /2fa/enable [post]
func (h *TwoFactorHandler) Enable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request Setup2FARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	// Get user's TOTP record
	var userTOTP models.UserTOTP
	if err := database.DB.Where("user_id = ?", userID).First(&userTOTP).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "2FA setup not found. Please setup 2FA first"})
		return
	}

	if userTOTP.Enabled {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "2FA is already enabled"})
		return
	}

	// Verify TOTP code
	if !h.totpManager.ValidateTOTP(userTOTP.Secret, request.TOTPCode, time.Now()) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid TOTP code"})
		return
	}

	// Enable 2FA
	now := time.Now()
	userTOTP.Enabled = true
	userTOTP.ActivatedAt = &now
	if err := database.DB.Save(&userTOTP).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to enable 2FA"})
		return
	}

	// Log security event
	database.DB.Create(&models.SecurityEvent{
		UserID:    userID.(uint),
		EventType: "2fa_enabled",
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Timestamp: time.Now(),
		Severity:  "info",
	})

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "2FA enabled successfully"})
}

// Disable2FA disables 2FA for a user
// @Summary Disable Two-Factor Authentication
// @Description Disable 2FA by verifying the TOTP code
// @Tags Two-Factor Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body Verify2FARequest true "TOTP or backup code verification"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /2fa/disable [post]
func (h *TwoFactorHandler) Disable2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request Verify2FARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	// Get user's TOTP record
	var userTOTP models.UserTOTP
	if err := database.DB.Where("user_id = ? AND enabled = ?", userID, true).First(&userTOTP).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "2FA is not enabled for this user"})
		return
	}

	// Verify TOTP code or backup code
	isValid := false
	if request.TOTPCode != "" {
		isValid = h.totpManager.ValidateTOTP(userTOTP.Secret, request.TOTPCode, time.Now())
	} else if request.BackupCode != "" {
		isValid = h.validateBackupCode(userTOTP.BackupCodes, request.BackupCode)
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid TOTP or backup code"})
		return
	}

	// Disable 2FA
	userTOTP.Enabled = false
	userTOTP.ActivatedAt = nil
	if err := database.DB.Save(&userTOTP).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to disable 2FA"})
		return
	}

	// Log security event
	database.DB.Create(&models.SecurityEvent{
		UserID:    userID.(uint),
		EventType: "2fa_disabled",
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Timestamp: time.Now(),
		Severity:  "warning",
	})

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "2FA disabled successfully"})
}

// Verify2FA verifies a 2FA code during login
// @Summary Verify Two-Factor Authentication
// @Description Verify TOTP code or backup code for 2FA
// @Tags Two-Factor Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body Verify2FARequest true "TOTP or backup code verification"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /2fa/verify [post]
func (h *TwoFactorHandler) Verify2FA(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request Verify2FARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	// Get user's TOTP record
	var userTOTP models.UserTOTP
	if err := database.DB.Where("user_id = ? AND enabled = ?", userID, true).First(&userTOTP).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "2FA is not enabled for this user"})
		return
	}

	// Verify TOTP code or backup code
	isValid := false
	if request.TOTPCode != "" {
		isValid = h.totpManager.ValidateTOTP(userTOTP.Secret, request.TOTPCode, time.Now())
	} else if request.BackupCode != "" {
		isValid = h.validateAndConsumeBackupCode(&userTOTP, request.BackupCode)
	}

	if !isValid {
		// Log failed 2FA attempt
		database.DB.Create(&models.SecurityEvent{
			UserID:    userID.(uint),
			EventType: "2fa_failed",
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Timestamp: time.Now(),
			Severity:  "warning",
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid TOTP or backup code"})
		return
	}

	// Log successful 2FA verification
	database.DB.Create(&models.SecurityEvent{
		UserID:    userID.(uint),
		EventType: "2fa_verified",
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Timestamp: time.Now(),
		Severity:  "info",
	})

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "2FA verified successfully"})
}

// Get2FAStatus returns the 2FA status for a user
// @Summary Get Two-Factor Authentication Status
// @Description Get the current 2FA status for the authenticated user
// @Tags Two-Factor Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /2fa/status [get]
func (h *TwoFactorHandler) Get2FAStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var userTOTP models.UserTOTP
	if err := database.DB.Where("user_id = ?", userID).First(&userTOTP).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled":      false,
			"setup_exists": false,
		})
		return
	}

	response := gin.H{
		"enabled":      userTOTP.Enabled,
		"setup_exists": true,
	}

	if userTOTP.ActivatedAt != nil {
		response["activated_at"] = userTOTP.ActivatedAt
	}

	c.JSON(http.StatusOK, response)
}

// generateBackupCodes generates a list of backup codes
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = fmt.Sprintf("%04s-%04s-%04s",
			randomString(4),
			randomString(4),
			randomString(4))
	}
	return codes
}

// randomString generates a random alphanumeric string
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// validateBackupCode validates a backup code without consuming it
func (h *TwoFactorHandler) validateBackupCode(backupCodesJSON, code string) bool {
	var backupCodes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &backupCodes); err != nil {
		return false
	}

	for _, backupCode := range backupCodes {
		if backupCode == code {
			return true
		}
	}
	return false
}

// validateAndConsumeBackupCode validates and removes a backup code
func (h *TwoFactorHandler) validateAndConsumeBackupCode(userTOTP *models.UserTOTP, code string) bool {
	var backupCodes []string
	if err := json.Unmarshal([]byte(userTOTP.BackupCodes), &backupCodes); err != nil {
		return false
	}

	// Find and remove the backup code
	for i, backupCode := range backupCodes {
		if backupCode == code {
			// Remove the used backup code
			backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)

			// Update the backup codes in the database
			updatedCodesJSON, _ := json.Marshal(backupCodes)
			userTOTP.BackupCodes = string(updatedCodesJSON)
			database.DB.Save(userTOTP)

			return true
		}
	}
	return false
}
