package handlers

import (
	"net/http"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// GetNewsletters godoc
// @Summary Get all newsletters
// @Description Retrieve all newsletters with pagination
// @Tags Newsletter
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (draft, scheduled, sent)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/newsletters [get]
func GetNewsletters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := database.DB.Preload("Creator")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var newsletters []models.Newsletter
	var total int64

	// Get total count
	if err := query.Model(&models.Newsletter{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to count newsletters"})
		return
	}

	// Get newsletters with pagination
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&newsletters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch newsletters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"newsletters": newsletters,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetNewsletter godoc
// @Summary Get newsletter by ID
// @Description Retrieve a single newsletter by its ID
// @Tags Newsletter
// @Produce json
// @Param id path int true "Newsletter ID"
// @Success 200 {object} models.Newsletter
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/newsletters/{id} [get]
func GetNewsletter(c *gin.Context) {
	id := c.Param("id")

	var newsletter models.Newsletter
	if err := database.DB.Preload("Creator").First(&newsletter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Newsletter not found"})
		return
	}

	c.JSON(http.StatusOK, newsletter)
}

// CreateNewsletter godoc
// @Summary Create a new newsletter
// @Description Create a new newsletter campaign (admin only)
// @Tags Newsletter
// @Accept json
// @Produce json
// @Security Bearer
// @Param newsletter body models.Newsletter true "Newsletter data"
// @Success 201 {object} models.Newsletter
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/newsletters [post]
func CreateNewsletter(c *gin.Context) {
	var newsletter models.Newsletter
	if err := c.ShouldBindJSON(&newsletter); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	newsletter.CreatedBy = userID.(uint)

	// Validate status
	validStatuses := map[string]bool{
		"draft":     true,
		"scheduled": true,
		"sent":      true,
	}
	if newsletter.Status == "" {
		newsletter.Status = "draft"
	} else if !validStatuses[newsletter.Status] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid status"})
		return
	}

	if err := database.DB.Create(&newsletter).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create newsletter"})
		return
	}

	// Load creator relation
	database.DB.Preload("Creator").First(&newsletter, newsletter.ID)

	c.JSON(http.StatusCreated, newsletter)
}

// UpdateNewsletter godoc
// @Summary Update a newsletter
// @Description Update an existing newsletter (admin only)
// @Tags Newsletter
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Newsletter ID"
// @Param newsletter body models.Newsletter true "Newsletter data"
// @Success 200 {object} models.Newsletter
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/newsletters/{id} [put]
func UpdateNewsletter(c *gin.Context) {
	id := c.Param("id")

	var newsletter models.Newsletter
	if err := database.DB.First(&newsletter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Newsletter not found"})
		return
	}

	var updateData models.Newsletter
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update fields
	if updateData.Title != "" {
		newsletter.Title = updateData.Title
	}
	if updateData.Subject != "" {
		newsletter.Subject = updateData.Subject
	}
	if updateData.Content != "" {
		newsletter.Content = updateData.Content
	}
	if updateData.Status != "" {
		validStatuses := map[string]bool{
			"draft":     true,
			"scheduled": true,
			"sent":      true,
		}
		if !validStatuses[updateData.Status] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid status"})
			return
		}
		newsletter.Status = updateData.Status
	}
	if updateData.ScheduledAt != nil {
		newsletter.ScheduledAt = updateData.ScheduledAt
	}

	if err := database.DB.Save(&newsletter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update newsletter"})
		return
	}

	// Load creator relation
	database.DB.Preload("Creator").First(&newsletter, newsletter.ID)

	c.JSON(http.StatusOK, newsletter)
}

// DeleteNewsletter godoc
// @Summary Delete a newsletter
// @Description Soft delete a newsletter (admin only)
// @Tags Newsletter
// @Produce json
// @Security Bearer
// @Param id path int true "Newsletter ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/newsletters/{id} [delete]
func DeleteNewsletter(c *gin.Context) {
	id := c.Param("id")

	var newsletter models.Newsletter
	if err := database.DB.First(&newsletter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Newsletter not found"})
		return
	}

	if err := database.DB.Delete(&newsletter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete newsletter"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SendNewsletter godoc
// @Summary Send a newsletter
// @Description Send a newsletter immediately (admin only)
// @Tags Newsletter
// @Produce json
// @Security Bearer
// @Param id path int true "Newsletter ID"
// @Success 200 {object} models.Newsletter
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/newsletters/{id}/send [post]
func SendNewsletter(c *gin.Context) {
	id := c.Param("id")

	var newsletter models.Newsletter
	if err := database.DB.First(&newsletter, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Newsletter not found"})
		return
	}

	if newsletter.Status == "sent" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Newsletter already sent"})
		return
	}

	// Update status and sent time
	now := time.Now()
	newsletter.Status = "sent"
	newsletter.SentAt = &now

	if err := database.DB.Save(&newsletter).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to send newsletter"})
		return
	}

	// TODO: Implement actual email sending logic here
	// This could integrate with an email service like SendGrid, Mailgun, etc.

	// Load creator relation
	database.DB.Preload("Creator").First(&newsletter, newsletter.ID)

	c.JSON(http.StatusOK, newsletter)
}
