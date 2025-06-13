package handlers

import (
	"fmt"
	"log"
	"net/http"
	"news/internal/auth"
	"news/internal/cache"
	"news/internal/database"
	"news/internal/middleware"
	"news/internal/models"
	"news/internal/pubsub"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should validate the origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// @Summary WebSocket connection for real-time notifications
// @Description Establishes a WebSocket connection for receiving real-time notifications
// @Tags WebSocket
// @Security BearerAuth
// @Param user_id query int true "User ID for connection"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ws/notifications [get]
func HandleWebSocketNotifications(c *gin.Context) {
	// Handle WebSocket authentication manually since middleware doesn't work with query params
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	// Validate the token manually
	tokenManager := auth.NewTokenManager(
		[]byte(middleware.GetJWTSecret()),
		24*time.Hour,
		7*24*time.Hour,
		cache.GetRedisClient(),
	)

	claims, err := tokenManager.ValidateToken(token)
	if err != nil {
		log.Printf("❌ WebSocket token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Get user ID from database using the username from token
	var user models.User
	if err := database.DB.Where("username = ?", claims.Username).First(&user).Error; err != nil {
		log.Printf("❌ User not found: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Get user ID from query parameter or use the one from token
	userIDParam := c.Query("user_id")
	if userIDParam == "" {
		userIDParam = strconv.FormatUint(uint64(user.ID), 10)
	}

	// Verify that the user is only connecting as themselves (security check)
	requestedUserID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if uint(requestedUserID) != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot connect as different user"})
		return
	}

	// Get user's preferred language (default to English if not specified)
	language := c.DefaultQuery("lang", "en")

	// Validate language is supported
	supportedLanguages := []string{"en", "tr", "es", "fr", "de", "ar", "zh", "ru", "ja", "ko"}
	isSupported := false
	for _, lang := range supportedLanguages {
		if language == lang {
			isSupported = true
			break
		}
	}
	if !isSupported {
		language = "en" // fallback to English
	}

	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	// Register client with notification hub
	pubsub.GetNotificationHub().RegisterClient(uint(userID), conn, language)

	// Handle WebSocket connection lifecycle
	defer func() {
		pubsub.GetNotificationHub().UnregisterClient(uint(userID))
		if err := conn.Close(); err != nil {
			log.Printf("Warning: Failed to close WebSocket connection: %v", err)
		}
	}()

	// Keep connection alive and handle incoming messages
	for {
		// Read messages from client (ping/pong, subscription updates, etc.)
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error for user %d: %v", userID, err)
			}
			break
		}

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			log.Printf("📨 Received message from user %d: %s", userID, string(message))

			// You can handle client-to-server messages here
			// For example: subscription management, preferences, etc.

		case websocket.PingMessage:
			// Respond to ping with pong
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("❌ Failed to send pong to user %d: %v", userID, err)
				return
			}
		}
	}
}

// @Summary Get notification statistics
// @Description Returns statistics about connected users and notification system
// @Tags WebSocket
// @Security BearerAuth
// @Produce json
// @Success 200 {object} NotificationStatsResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ws/stats [get]
func GetNotificationStats(c *gin.Context) {
	stats := NotificationStatsResponse{
		ConnectedUsers: pubsub.GetConnectedUsers(),
		SystemStatus:   "active",
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Send test notification
// @Description Sends a test notification (admin only)
// @Tags WebSocket
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body TestNotificationRequest true "Test notification"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/ws/test [post]
func SendTestNotification(c *gin.Context) {
	// Check if user is admin
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req TestNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Send test notification based on type
	var err error
	switch req.Type {
	case "system_alert":
		err = pubsub.PublishSystemAlert(req.Message, "info")
	case "user_notification":
		if req.UserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required for user notifications"})
			return
		}
		// Send directly to WebSocket client for immediate delivery
		hub := pubsub.GetNotificationHub()
		if hub != nil {
			notification := pubsub.NotificationMessage{
				Type:   "user_notification",
				UserID: req.UserID,
				Data: map[string]interface{}{
					"message": req.Message,
					"title":   "Test Notification",
				},
			}
			hub.SendToUser(req.UserID, notification)
		} else {
			err = fmt.Errorf("notification hub not available")
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
		return
	}

	if err != nil {
		log.Printf("❌ Failed to send test notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test notification sent successfully",
		"type":    req.Type,
	})
}

// @Summary Check user connection status
// @Description Checks if a specific user is connected to WebSocket
// @Tags WebSocket
// @Security BearerAuth
// @Param user_id path int true "User ID"
// @Produce json
// @Success 200 {object} UserConnectionStatusResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/ws/user/{user_id}/status [get]
func GetUserConnectionStatus(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	isConnected := pubsub.IsUserConnected(uint(userID))

	c.JSON(http.StatusOK, UserConnectionStatusResponse{
		UserID:      uint(userID),
		IsConnected: isConnected,
	})
}

// Response types
type NotificationStatsResponse struct {
	ConnectedUsers int    `json:"connected_users"`
	SystemStatus   string `json:"system_status"`
}

type TestNotificationRequest struct {
	Type    string `json:"type" binding:"required"` // "system_alert", "user_notification"
	Message string `json:"message" binding:"required"`
	UserID  uint   `json:"user_id,omitempty"` // Required for user_notification type
}

type UserConnectionStatusResponse struct {
	UserID      uint `json:"user_id"`
	IsConnected bool `json:"is_connected"`
}
