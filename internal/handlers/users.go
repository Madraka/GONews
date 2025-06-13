package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get a user's public profile information
// @Tags Users
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} models.UserProfileResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/users/{username}/profile [get]
func GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	profile := models.UserProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Avatar:    user.Avatar,
		Bio:       user.Bio,
		Website:   user.Website,
		Location:  user.Location,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	var articles []models.Article
	if err := database.DB.Where("author_id = ? AND status = ?", user.ID, "published").Order("created_at DESC").Limit(5).Find(&articles).Error; err == nil {
		profile.RecentArticles = articles
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body models.UpdateUserProfileRequest true "Profile data to update"
// @Success 200 {object} models.User // Refers to user_advanced.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/profile [put]
func UpdateProfile(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}
	userID := userIDAny.(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	var req models.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Email != nil {
		user.Email = *req.Email // Field from user_advanced.go
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Website != nil {
		user.Website = *req.Website
	}
	if req.Location != nil {
		user.Location = *req.Location
	}

	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserArticles godoc
// @Summary Get user's articles
// @Description Get a list of articles published by a specific user
// @Tags Users
// @Produce json
// @Param username path string true "Username"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.PaginatedArticlesResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/users/{username}/articles [get]
func GetUserArticles(c *gin.Context) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	var articles []models.Article
	var total int64

	offset := (page - 1) * limit
	query := database.DB.Model(&models.Article{}).Where("author_id = ? AND status = ?", user.ID, "published")

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to count user articles"})
		return
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch user articles"})
		return
	}

	response := models.PaginatedArticlesResponse{
		Page:     page,
		Limit:    limit,
		Total:    total,
		Articles: articles,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserNotifications godoc
// @Summary Get user notifications
// @Description Retrieve notifications for the authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param unread_only query bool false "Fetch only unread notifications"
// @Success 200 {object} models.PaginatedNotificationsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/notifications [get]
func GetUserNotifications(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}
	userID := userIDAny.(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	unreadOnly := c.Query("unread_only") == "true"

	var notifications []models.Notification
	var total int64

	offset := (page - 1) * limit
	query := database.DB.Model(&models.Notification{}).Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to count notifications"})
		return
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch notifications"})
		return
	}

	response := models.PaginatedNotificationsResponse{
		Page:          page,
		Limit:         limit,
		Total:         total,
		Notifications: notifications,
	}
	c.JSON(http.StatusOK, response)
}

// MarkNotificationRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read for the authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param notification_id path int true "Notification ID"
// @Success 200 {object} models.Notification
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/notifications/{notification_id}/read [patch]
func MarkNotificationRead(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}
	userID := userIDAny.(uint)

	notificationID, err := strconv.Atoi(c.Param("notification_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid notification ID"})
		return
	}

	var notification models.Notification
	if err := database.DB.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Notification not found"})
		return
	}

	notification.IsRead = true
	if err := database.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// MarkAllNotificationsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all unread notifications as read for the authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} MarkAllNotificationsReadResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/notifications/read-all [patch]
func MarkAllNotificationsRead(c *gin.Context) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}
	userID := userIDAny.(uint)

	now := time.Now()
	result := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{"is_read": true, "read_at": &now})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, MarkAllNotificationsReadResponse{
		Message:      "All notifications marked as read",
		UpdatedCount: result.RowsAffected,
	})
}

// Define response struct for MarkAllNotificationsRead
type MarkAllNotificationsReadResponse struct {
	Message      string `json:"message"`
	UpdatedCount int64  `json:"updated_count"`
}
