package pubsub

import (
	"context"
	"fmt"
	"log"
	"news/internal/cache"
	"news/internal/json"
	"news/internal/models"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// TranslationService interface for dependency injection
type TranslationService interface {
	GetLocalizer(language string) *i18n.Localizer
	Localize(language, messageID string, templateData map[string]interface{}) (string, error)
}

// NotificationHub manages real-time notifications via Redis pub/sub
type NotificationHub struct {
	// Redis client for pub/sub operations
	redisClient *redis.Client
	ctx         context.Context
	cancel      context.CancelFunc

	// Translation service for localized messages
	translationService TranslationService

	// Connected WebSocket clients with their language preferences
	clients map[uint]*ClientConnection

	// Channels for managing connections
	register   chan *ClientConnection
	unregister chan *ClientConnection
	broadcast  chan NotificationMessage

	// Pub/sub subscription
	pubsub *redis.PubSub

	// Shutdown control
	done chan struct{}

	// Ensure Close() can only be called once
	closeOnce sync.Once
}

// ClientConnection represents a connected user
type ClientConnection struct {
	UserID   uint
	Conn     *websocket.Conn
	Language string // User's preferred language
}

// NotificationMessage represents a message to be sent via pub/sub
type NotificationMessage struct {
	Type      string      `json:"type"`
	UserID    uint        `json:"user_id,omitempty"` // For targeted notifications
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Channel   string      `json:"channel"`
}

// Notification channels
const (
	// Global channels - all users receive
	ChannelBreakingNews = "breaking_news"
	ChannelSystemAlert  = "system_alert"

	// User-specific channels - format: "user:{user_id}"
	ChannelUserNotification = "user_notification"
	ChannelUserComment      = "user_comment"
	ChannelUserVote         = "user_vote"
	ChannelUserMention      = "user_mention"

	// News-specific channels - format: "news:{news_id}"
	ChannelNewsUpdate  = "news_update"
	ChannelNewsComment = "news_comment"
)

// Global notification hub instance
var globalHub *NotificationHub

// InitNotificationHub initializes the global notification hub
func InitNotificationHub(translationService TranslationService) error {
	redisClient := cache.GetRedisClient()
	if redisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	globalHub = NewNotificationHub(redisClient.GetClient(), translationService)

	// Subscribe to Redis channels
	channels := []string{
		ChannelBreakingNews,
		ChannelSystemAlert,
		// We'll add more specific channels as needed
	}

	globalHub.pubsub = globalHub.redisClient.Subscribe(globalHub.ctx, channels...)

	// Start the hub and Redis listener
	go globalHub.Run()
	go globalHub.listenToRedis()

	log.Println("Redis pub/sub notification hub initialized successfully")
	return nil
}

// GetNotificationHub returns the global notification hub instance
func GetNotificationHub() *NotificationHub {
	return globalHub
}

// NewNotificationHub creates a new NotificationHub instance
func NewNotificationHub(redisClient *redis.Client, translationService TranslationService) *NotificationHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationHub{
		redisClient:        redisClient,
		ctx:                ctx,
		cancel:             cancel,
		translationService: translationService,
		clients:            make(map[uint]*ClientConnection),
		register:           make(chan *ClientConnection),
		unregister:         make(chan *ClientConnection),
		broadcast:          make(chan NotificationMessage),
		done:               make(chan struct{}),
	}
}

// Run manages WebSocket connections and broadcasts
func (h *NotificationHub) Run() {
	defer close(h.done)

	for {
		select {
		case client := <-h.register:
			h.clients[client.UserID] = client
			log.Printf("📱 User %d connected to notification hub (lang: %s)", client.UserID, client.Language)

			// Send welcome message
			welcome := NotificationMessage{
				Type:      "welcome",
				Data:      map[string]string{"message": "Connected to notifications"},
				Timestamp: time.Now(),
				Channel:   "system",
			}
			h.sendToClient(client.UserID, welcome)

		case client := <-h.unregister:
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				if err := client.Conn.Close(); err != nil {
					log.Printf("Warning: Error closing client connection for user %d: %v", client.UserID, err)
				}
				log.Printf("📱 User %d disconnected from notification hub", client.UserID)
			}

		case message := <-h.broadcast:
			if message.UserID != 0 {
				// Send to specific user
				h.sendToClient(message.UserID, message)
			} else {
				// Broadcast to all connected clients
				h.broadcastToAll(message)
			}

		case <-h.ctx.Done():
			log.Println("🔴 Notification hub Run() shutting down...")
			return
		}
	}
}

// listenToRedis listens for Redis pub/sub messages
func (h *NotificationHub) listenToRedis() {
	for {
		select {
		case msg := <-h.pubsub.Channel():
			var notification NotificationMessage
			if err := json.Unmarshal([]byte(msg.Payload), &notification); err != nil {
				log.Printf("❌ Error unmarshalling notification: %v", err)
				continue
			}

			log.Printf("📨 Received notification from Redis: %s -> %s", msg.Channel, notification.Type)

			// Forward to WebSocket clients
			select {
			case h.broadcast <- notification:
				// Successfully sent to broadcast channel
			case <-h.ctx.Done():
				log.Println("⚠️ Cannot broadcast notification: hub is shutting down")
				return
			}

		case <-h.ctx.Done():
			log.Println("🔴 Redis pub/sub listener stopped")
			return
		}
	}
}

// RegisterClient registers a new WebSocket client
func (h *NotificationHub) RegisterClient(userID uint, conn *websocket.Conn, language string) {
	select {
	case h.register <- &ClientConnection{
		UserID:   userID,
		Conn:     conn,
		Language: language,
	}:
		// Successfully sent registration
	case <-h.ctx.Done():
		log.Printf("⚠️ Cannot register client %d: hub is shutting down", userID)
		if err := conn.Close(); err != nil {
			log.Printf("Warning: Error closing connection during shutdown: %v", err)
		}
	}
}

// UnregisterClient removes a WebSocket client
func (h *NotificationHub) UnregisterClient(userID uint) {
	if clientConn, exists := h.clients[userID]; exists {
		select {
		case h.unregister <- &ClientConnection{
			UserID: userID,
			Conn:   clientConn.Conn,
		}:
			// Successfully sent unregistration
		case <-h.ctx.Done():
			log.Printf("⚠️ Cannot unregister client %d: hub is shutting down", userID)
			if err := clientConn.Conn.Close(); err != nil {
				log.Printf("Warning: Error closing connection during unregister: %v", err)
			}
		}
	}
}

// sendToClient sends a message to a specific client
func (h *NotificationHub) sendToClient(userID uint, message NotificationMessage) {
	if clientConn, exists := h.clients[userID]; exists {
		if err := clientConn.Conn.WriteJSON(message); err != nil {
			log.Printf("❌ Error sending message to user %d: %v", userID, err)
			// Remove broken connection
			delete(h.clients, userID)
			if closeErr := clientConn.Conn.Close(); closeErr != nil {
				log.Printf("Warning: Error closing broken connection: %v", closeErr)
			}
		}
	}
}

// broadcastToAll sends a message to all connected clients
func (h *NotificationHub) broadcastToAll(message NotificationMessage) {
	for userID, clientConn := range h.clients {
		if err := clientConn.Conn.WriteJSON(message); err != nil {
			log.Printf("❌ Error broadcasting to user %d: %v", userID, err)
			// Remove broken connection
			delete(h.clients, userID)
			if closeErr := clientConn.Conn.Close(); closeErr != nil {
				log.Printf("Warning: Error closing broken connection: %v", closeErr)
			}
		}
	}
}

// PublishNotification publishes a notification to Redis
func PublishNotification(channel string, notification NotificationMessage) error {
	if cache.IsTestMode() {
		// In test mode, just log the notification
		log.Printf("🧪 TEST MODE: Would publish to %s: %+v", channel, notification)
		return nil
	}

	redisClient := cache.GetRedisClient().GetClient()
	if redisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	notification.Timestamp = time.Now()
	notification.Channel = channel

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	ctx := context.Background()
	return redisClient.Publish(ctx, channel, data).Err()
}

// Convenience functions for different notification types

// PublishBreakingNews publishes a breaking news notification to all users
func PublishBreakingNews(article models.Article) error {
	notification := NotificationMessage{
		Type: "breaking_news",
		Data: map[string]interface{}{
			"id":        article.ID,
			"title":     article.Title,
			"content":   article.Content,
			"category":  article.Categories,
			"image_url": article.FeaturedImage,
		},
	}
	return PublishNotification(ChannelBreakingNews, notification)
}

// PublishUserNotification publishes a notification to a specific user
func PublishUserNotification(userID uint, notificationType string, data interface{}) error {
	channel := fmt.Sprintf("%s:%d", ChannelUserNotification, userID)
	notification := NotificationMessage{
		Type:   notificationType,
		UserID: userID,
		Data:   data,
	}
	return PublishNotification(channel, notification)
}

// PublishCommentNotification publishes a comment notification
func PublishCommentNotification(newsID uint, comment models.Comment) error {
	// Notify news author and other commenters
	channel := fmt.Sprintf("%s:%d", ChannelNewsComment, newsID)
	notification := NotificationMessage{
		Type: "new_comment",
		Data: map[string]interface{}{
			"news_id":    newsID,
			"comment_id": comment.ID,
			"content":    comment.Content,
			"author":     comment.User.Username,
			"created_at": comment.CreatedAt,
		},
	}
	return PublishNotification(channel, notification)
}

// PublishVoteNotification publishes a vote notification
func PublishVoteNotification(videoID uint, userID uint, voteType string) error {
	notification := NotificationMessage{
		Type: "video_vote",
		Data: map[string]interface{}{
			"video_id":  videoID,
			"user_id":   userID,
			"vote_type": voteType,
			"timestamp": time.Now(),
		},
	}

	// Publish to video-specific channel
	channel := fmt.Sprintf("video_votes:%d", videoID)
	return PublishNotification(channel, notification)
}

// PublishSystemAlert publishes a system-wide alert
func PublishSystemAlert(message string, alertType string) error {
	notification := NotificationMessage{
		Type: "system_alert",
		Data: map[string]interface{}{
			"message":    message,
			"alert_type": alertType,
			"timestamp":  time.Now(),
		},
	}
	return PublishNotification(ChannelSystemAlert, notification)
}

// Enhanced notification functions with multi-language support

// PublishNewsViewUpdate publishes a news view count update
func PublishNewsViewUpdate(newsID uint, viewCount int64) error {
	notification := NotificationMessage{
		Type: "news_view_update",
		Data: map[string]interface{}{
			"news_id":    newsID,
			"view_count": viewCount,
		},
	}
	return PublishNotification(fmt.Sprintf("news:%d", newsID), notification)
}

// PublishProfileUpdate publishes a profile update notification
func PublishProfileUpdate(userID uint, updateType string) error {
	notification := NotificationMessage{
		Type:   "profile_update",
		UserID: userID,
		Data: map[string]interface{}{
			"update_type": updateType,
		},
	}
	return PublishNotification(fmt.Sprintf("user:%d", userID), notification)
}

// PublishPasswordChangeAlert publishes a password change security alert
func PublishPasswordChangeAlert(userID uint, ipAddress string) error {
	notification := NotificationMessage{
		Type:   "security_alert",
		UserID: userID,
		Data: map[string]interface{}{
			"alert_type": "password_change",
			"ip_address": ipAddress,
			"timestamp":  time.Now(),
		},
	}
	return PublishNotification(fmt.Sprintf("user:%d", userID), notification)
}

// PublishMaintenanceNotice publishes a maintenance notice to all users
func PublishMaintenanceNotice(startTime, endTime time.Time, description string) error {
	notification := NotificationMessage{
		Type: "maintenance_notice",
		Data: map[string]interface{}{
			"start_time":  startTime,
			"end_time":    endTime,
			"description": description,
		},
	}
	return PublishNotification(ChannelSystemAlert, notification)
}

// PublishTrendingUpdate publishes trending news updates
func PublishTrendingUpdate(trendingArticles []models.Article) error {
	notification := NotificationMessage{
		Type: "trending_update",
		Data: map[string]interface{}{
			"trending_articles": trendingArticles,
			"update_time":       time.Now(),
		},
	}
	return PublishNotification(ChannelBreakingNews, notification)
}

// PublishLiveStatistics publishes live site statistics
func PublishLiveStatistics(stats map[string]interface{}) error {
	notification := NotificationMessage{
		Type: "live_statistics",
		Data: stats,
	}
	return PublishNotification(ChannelSystemAlert, notification)
}

// PublishVideoProcessingUpdate publishes video processing status updates
func PublishVideoProcessingUpdate(userID uint, videoID uint, status string, progress int) error {
	notification := NotificationMessage{
		Type:   "video_processing",
		UserID: userID,
		Data: map[string]interface{}{
			"video_id": videoID,
			"status":   status,
			"progress": progress,
		},
	}
	return PublishNotification(fmt.Sprintf("user:%d", userID), notification)
}

// PublishFollowNotification publishes follow/unfollow notifications
func PublishFollowNotification(targetUserID uint, followerUserID uint, action string, followerUsername string) error {
	notification := NotificationMessage{
		Type:   "follow_notification",
		UserID: targetUserID,
		Data: map[string]interface{}{
			"follower_id":       followerUserID,
			"follower_username": followerUsername,
			"action":            action, // "follow" or "unfollow"
		},
	}
	return PublishNotification(fmt.Sprintf("user:%d", targetUserID), notification)
}

// PublishCategoryNewsAlert publishes category-specific news alerts to subscribed users
func PublishCategoryNewsAlert(categoryID uint, article models.Article, subscribedUserIDs []uint) error {
	for _, userID := range subscribedUserIDs {
		notification := NotificationMessage{
			Type:   "category_news_alert",
			UserID: userID,
			Data: map[string]interface{}{
				"category_id": categoryID,
				"article":     article,
			},
		}

		// Send to individual user channel
		if err := PublishNotification(fmt.Sprintf("user:%d", userID), notification); err != nil {
			log.Printf("❌ Failed to send category alert to user %d: %v", userID, err)
		}
	}
	return nil
}

// Convenience functions using the global hub for localized notifications

// SendLocalizedNotificationToUser sends a localized notification to a specific user
func SendLocalizedNotificationToUser(userID uint, messageKey string, templateData map[string]interface{}) {
	if globalHub != nil {
		globalHub.SendLocalizedNotification(userID, messageKey, templateData)
	}
}

// BroadcastLocalizedSystemAlert broadcasts a localized system alert to all users
func BroadcastLocalizedSystemAlert(messageKey string, templateData map[string]interface{}) {
	if globalHub != nil {
		globalHub.BroadcastLocalizedNotification(messageKey, templateData)
	}
}

// SendWelcomeMessage sends a localized welcome message to a newly connected user
func SendWelcomeMessage(userID uint, username string) {
	templateData := map[string]interface{}{
		"Username": username,
	}
	SendLocalizedNotificationToUser(userID, "notifications.welcome.message", templateData)
}

// SendBreakingNewsAlert sends a localized breaking news alert
func SendBreakingNewsAlert(newsTitle string, categoryName string) {
	templateData := map[string]interface{}{
		"Title":    newsTitle,
		"Category": categoryName,
	}
	BroadcastLocalizedSystemAlert("notifications.breaking_news.message", templateData)
}

// GetConnectedUsers returns the number of connected users
func GetConnectedUsers() int {
	if globalHub == nil {
		return 0
	}
	return len(globalHub.clients)
}

// IsUserConnected checks if a user is currently connected
func IsUserConnected(userID uint) bool {
	if globalHub == nil {
		return false
	}
	_, exists := globalHub.clients[userID]
	return exists
}

// Close gracefully shuts down the notification hub
func Close() error {
	if globalHub == nil {
		return nil
	}

	var closeErr error
	globalHub.closeOnce.Do(func() {
		log.Println("🔴 Shutting down notification hub...")

		// Close Redis pub/sub first to stop incoming messages
		if globalHub.pubsub != nil {
			err := globalHub.pubsub.Close()
			if err != nil {
				log.Printf("❌ Error closing Redis pub/sub: %v", err)
				closeErr = err
			}
		}

		// Cancel context to stop goroutines
		if globalHub.cancel != nil {
			globalHub.cancel()
		}

		// Wait for Run() to finish with longer timeout
		select {
		case <-globalHub.done:
			log.Println("✅ Notification hub Run() stopped")
		case <-time.After(10 * time.Second):
			log.Println("⚠️ Timeout waiting for notification hub to stop")
		}

		// Close all client connections
		for userID, clientConn := range globalHub.clients {
			if clientConn.Conn != nil {
				if err := clientConn.Conn.Close(); err != nil {
					log.Printf("Warning: Failed to close WebSocket connection for user %d: %v", userID, err)
				}
			}
			delete(globalHub.clients, userID)
		}

		// Use a safer channel closing mechanism with defer recovery
		safeCloseChannel := func(ch interface{}, name string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Recovered from panic closing %s channel: %v", name, r)
				}
			}()

			switch c := ch.(type) {
			case chan *ClientConnection:
				select {
				case <-c:
					// Channel already closed
				default:
					close(c)
				}
			case chan NotificationMessage:
				select {
				case <-c:
					// Channel already closed
				default:
					close(c)
				}
			case chan struct{}:
				select {
				case <-c:
					// Channel already closed
				default:
					close(c)
				}
			}
		}

		// Close channels with recovery
		safeCloseChannel(globalHub.register, "register")
		safeCloseChannel(globalHub.unregister, "unregister")
		safeCloseChannel(globalHub.broadcast, "broadcast")

		log.Println("✅ Notification hub shutdown complete")
	})

	// Reset global hub after close operation
	globalHub = nil
	return closeErr
}

// SendToUser sends a message directly to a specific user's WebSocket connection
func (h *NotificationHub) SendToUser(userID uint, message NotificationMessage) {
	message.Timestamp = time.Now()
	h.sendToClient(userID, message)
}

// SendLocalizedNotification sends a localized notification to a specific user
func (h *NotificationHub) SendLocalizedNotification(userID uint, messageKey string, templateData map[string]interface{}) {
	// Get user's language preference from connection
	var userLang string = "en" // default
	if clientConn, exists := h.clients[userID]; exists {
		userLang = clientConn.Language
	}

	// Get translation service instance
	// Note: We'll need to get the initialized translation service from global state
	// For now, we'll create a simple fallback
	localizedMessage := h.translateMessage(messageKey, userLang, templateData)

	// Send localized notification
	notification := NotificationMessage{
		Type: "localized_notification",
		Data: map[string]interface{}{
			"message":  localizedMessage,
			"language": userLang,
			"key":      messageKey,
		},
		Timestamp: time.Now(),
	}

	h.SendToUser(userID, notification)
}

// BroadcastLocalizedNotification sends a localized notification to all connected users
func (h *NotificationHub) BroadcastLocalizedNotification(messageKey string, templateData map[string]interface{}) {
	// Send to each connected user in their preferred language
	for userID, clientConn := range h.clients {
		// Localize the message for user's language
		localizedMessage := h.translateMessage(messageKey, clientConn.Language, templateData)

		// Send localized notification
		notification := NotificationMessage{
			Type: "localized_broadcast",
			Data: map[string]interface{}{
				"message":  localizedMessage,
				"language": clientConn.Language,
				"key":      messageKey,
			},
			Timestamp: time.Now(),
		}

		h.sendToClient(userID, notification)
	}
}

// translateMessage is a helper function to translate messages using the translation service
func (h *NotificationHub) translateMessage(messageKey, language string, templateData map[string]interface{}) string {
	if h.translationService == nil {
		log.Printf("Warning: Translation service not available, returning message key: %s", messageKey)
		return messageKey
	}

	// Try to get localized message
	localizedMessage, err := h.translationService.Localize(language, messageKey, templateData)
	if err != nil {
		log.Printf("Warning: Failed to localize message '%s' for language '%s': %v", messageKey, language, err)
		// Fallback to English if available
		if language != "en" {
			fallbackMessage, fallbackErr := h.translationService.Localize("en", messageKey, templateData)
			if fallbackErr == nil {
				return fallbackMessage
			}
		}
		// Final fallback to message key
		return messageKey
	}

	return localizedMessage
}
