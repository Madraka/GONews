package handlers

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMedia godoc
// @Summary Get all media files
// @Description Retrieve all media files with pagination and filtering
// @Tags Media
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param mime_type query string false "Filter by MIME type (image, video, audio, document)"
// @Param uploaded_by query int false "Filter by uploader user ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /media [get]
func GetMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	mimeTypeFilter := c.Query("mime_type")
	uploadedByStr := c.Query("uploaded_by")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Preload("Uploader")

	// Filter by MIME type category
	if mimeTypeFilter != "" {
		switch mimeTypeFilter {
		case "image":
			query = query.Where("mime_type LIKE ?", "image/%")
		case "video":
			query = query.Where("mime_type LIKE ?", "video/%")
		case "audio":
			query = query.Where("mime_type LIKE ?", "audio/%")
		case "document":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ? OR mime_type LIKE ?",
				"application/pdf", "application/msword", "application/vnd.ms-excel")
		}
	}

	// Filter by uploader
	if uploadedByStr != "" {
		uploadedBy, err := strconv.Atoi(uploadedByStr)
		if err == nil {
			query = query.Where("uploaded_by = ?", uploadedBy)
		}
	}

	var media []models.Media
	var total int64

	// Get total count
	if err := query.Model(&models.Media{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to count media files"})
		return
	}

	// Get media with pagination
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch media files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media": media,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetMediaByID godoc
// @Summary Get media file by ID
// @Description Retrieve a single media file by its ID
// @Tags Media
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} models.Media
// @Failure 404 {object} models.ErrorResponse
// @Router /media/{id} [get]
func GetMediaByID(c *gin.Context) {
	id := c.Param("id")

	var media models.Media
	if err := database.DB.Preload("Uploader").First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Media file not found"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// UploadMedia godoc
// @Summary Upload a media file
// @Description Upload a new media file (authenticated users)
// @Tags Media
// @Accept multipart/form-data
// @Produce json
// @Security Bearer
// @Param file formData file true "Media file to upload"
// @Param alt_text formData string false "Alternative text for the media"
// @Param caption formData string false "Caption for the media"
// @Success 201 {object} models.Media
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse
// @Router /media/upload [post]
func UploadMedia(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to parse form data"})
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "No file provided"})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: Failed to close uploaded file: %v", err)
		}
	}()

	// Validate file size (10MB max)
	if fileHeader.Size > 10*1024*1024 {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{Error: "File too large. Maximum size is 10MB"})
		return
	}

	// Validate MIME type
	mimeType := fileHeader.Header.Get("Content-Type")
	if !isAllowedMimeType(mimeType) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "File type not allowed"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	fileName := uuid.New().String() + ext

	// Create upload directory
	uploadDir := "uploads/" + time.Now().Format("2006/01")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create upload directory"})
		return
	}

	// Full file path
	filePath := filepath.Join(uploadDir, fileName)

	// Save file to disk
	if err := saveFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save file"})
		return
	}

	// Create media record
	media := models.Media{
		FileName:     fileName,
		OriginalName: fileHeader.Filename,
		MimeType:     mimeType,
		Size:         fileHeader.Size,
		Path:         filePath,
		URL:          "/" + filePath, // Relative URL
		AltText:      c.PostForm("alt_text"),
		Caption:      c.PostForm("caption"),
		UploadedBy:   userID.(uint),
	}

	if err := database.DB.Create(&media).Error; err != nil {
		// Remove file if database operation fails
		if removeErr := os.Remove(filePath); removeErr != nil {
			log.Printf("Warning: Failed to remove file after database error: %v", removeErr)
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save media record"})
		return
	}

	// Load uploader relation
	database.DB.Preload("Uploader").First(&media, media.ID)

	c.JSON(http.StatusCreated, media)
}

// UpdateMedia godoc
// @Summary Update media metadata
// @Description Update media file metadata (alt text, caption)
// @Tags Media
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Media ID"
// @Param media body models.Media true "Media metadata"
// @Success 200 {object} models.Media
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /media/{id} [put]
func UpdateMedia(c *gin.Context) {
	id := c.Param("id")

	var media models.Media
	if err := database.DB.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Media file not found"})
		return
	}

	// Check if user owns the media or is admin
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	if media.UploadedBy != userID.(uint) && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to update this media"})
		return
	}

	var updateData models.Media
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update metadata fields only
	if updateData.AltText != "" {
		media.AltText = updateData.AltText
	}
	if updateData.Caption != "" {
		media.Caption = updateData.Caption
	}

	if err := database.DB.Save(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update media"})
		return
	}

	// Load uploader relation
	database.DB.Preload("Uploader").First(&media, media.ID)

	c.JSON(http.StatusOK, media)
}

// DeleteMedia godoc
// @Summary Delete a media file
// @Description Delete a media file and its record (admin or owner only)
// @Tags Media
// @Produce json
// @Security Bearer
// @Param id path int true "Media ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /media/{id} [delete]
func DeleteMedia(c *gin.Context) {
	id := c.Param("id")

	var media models.Media
	if err := database.DB.First(&media, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Media file not found"})
		return
	}

	// Check if user owns the media or is admin
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	if media.UploadedBy != userID.(uint) && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Not authorized to delete this media"})
		return
	}

	// Delete file from disk
	if err := os.Remove(media.Path); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("Warning: Failed to delete file %s: %v\n", media.Path, err)
	}

	// Delete from database
	if err := database.DB.Delete(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete media record"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetMediaStats godoc
// @Summary Get media statistics
// @Description Get statistics about uploaded media (admin only)
// @Tags Media
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/media/stats [get]
func GetMediaStats(c *gin.Context) {
	var stats struct {
		TotalFiles    int64 `json:"total_files"`
		TotalSize     int64 `json:"total_size"`
		ImageCount    int64 `json:"image_count"`
		VideoCount    int64 `json:"video_count"`
		DocumentCount int64 `json:"document_count"`
		OtherCount    int64 `json:"other_count"`
	}

	// Get total files and size
	database.DB.Model(&models.Media{}).Count(&stats.TotalFiles)
	database.DB.Model(&models.Media{}).Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize)

	// Get counts by type
	database.DB.Model(&models.Media{}).Where("mime_type LIKE ?", "image/%").Count(&stats.ImageCount)
	database.DB.Model(&models.Media{}).Where("mime_type LIKE ?", "video/%").Count(&stats.VideoCount)
	database.DB.Model(&models.Media{}).Where("mime_type LIKE ? OR mime_type LIKE ? OR mime_type LIKE ?",
		"application/pdf", "application/msword", "application/vnd.ms-excel").Count(&stats.DocumentCount)

	stats.OtherCount = stats.TotalFiles - stats.ImageCount - stats.VideoCount - stats.DocumentCount

	c.JSON(http.StatusOK, stats)
}

// Helper functions

func isAllowedMimeType(mimeType string) bool {
	allowedTypes := map[string]bool{
		// Images
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
		// Videos
		"video/mp4":  true,
		"video/webm": true,
		"video/ogg":  true,
		// Audio
		"audio/mpeg": true,
		"audio/wav":  true,
		"audio/ogg":  true,
		// Documents
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain": true,
		"text/csv":   true,
	}
	return allowedTypes[mimeType]
}

func saveFile(file multipart.File, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("Warning: Failed to close file %s: %v", filePath, err)
		}
	}()

	_, err = io.Copy(dst, file)
	return err
}
