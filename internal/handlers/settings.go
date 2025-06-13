package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// GetSettings godoc
// @Summary Get all settings
// @Description Retrieve all system settings or filter by group/public status
// @Tags Settings
// @Produce json
// @Param group query string false "Filter by settings group"
// @Param public query bool false "Filter by public settings only"
// @Success 200 {array} models.Setting
// @Failure 500 {object} models.ErrorResponse
// @Router /settings [get]
func GetSettings(c *gin.Context) {
	group := c.Query("group")
	publicParam := c.Query("public")

	query := database.DB.Model(&models.Setting{})

	if group != "" {
		query = query.Where("\"group\" = ?", group)
	}

	// If public parameter is set to true, only return public settings
	if publicParam == "true" {
		query = query.Where("is_public = ?", true)
	}

	var settings []models.Setting
	if err := query.Order("\"group\" ASC, \"key\" ASC").Find(&settings).Error; err != nil {
		// Debug: print the actual error
		fmt.Printf("Settings query error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetSettingByKey godoc
// @Summary Get setting by key
// @Description Retrieve a single setting by its key
// @Tags Settings
// @Produce json
// @Param key path string true "Setting key"
// @Success 200 {object} models.Setting
// @Failure 404 {object} models.ErrorResponse
// @Router /settings/{key} [get]
func GetSettingByKey(c *gin.Context) {
	key := c.Param("key")

	var setting models.Setting
	query := database.DB.Where("key = ?", key)

	// For non-admin users, only return public settings
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		query = query.Where("is_public = ?", true)
	}

	if err := query.First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// GetSettingGroups godoc
// @Summary Get setting groups
// @Description Retrieve all available setting groups
// @Tags Settings
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} models.ErrorResponse
// @Router /settings/groups [get]
func GetSettingGroups(c *gin.Context) {
	var groups []string
	if err := database.DB.Model(&models.Setting{}).
		Distinct("group").
		Where("group IS NOT NULL AND group != ''").
		Pluck("group", &groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch setting groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// CreateSetting godoc
// @Summary Create a new setting
// @Description Create a new system setting (admin only)
// @Tags Settings
// @Accept json
// @Produce json
// @Security Bearer
// @Param setting body models.Setting true "Setting data"
// @Success 201 {object} models.Setting
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/settings [post]
func CreateSetting(c *gin.Context) {
	var setting models.Setting
	if err := c.ShouldBindJSON(&setting); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate setting type
	if !setting.ValidateSettingType() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid setting type. Must be one of: string, integer, boolean, json, text"})
		return
	}

	// Set default type if not provided
	if setting.Type == "" {
		setting.Type = "string"
	}

	// Validate value based on type
	if err := validateSettingValue(&setting); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := database.DB.Create(&setting).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Setting with this key already exists"})
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create setting"})
		return
	}

	c.JSON(http.StatusCreated, setting)
}

// UpdateSetting godoc
// @Summary Update a setting
// @Description Update an existing setting (admin only)
// @Tags Settings
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Setting ID"
// @Param setting body models.Setting true "Setting data"
// @Success 200 {object} models.Setting
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/settings/{id} [put]
func UpdateSetting(c *gin.Context) {
	id := c.Param("id")

	var setting models.Setting
	if err := database.DB.First(&setting, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Setting not found"})
		return
	}

	var updateData models.Setting
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update fields
	if updateData.Key != "" {
		setting.Key = updateData.Key
	}
	if updateData.Value != "" {
		setting.Value = updateData.Value
	}
	if updateData.Type != "" {
		if !updateData.ValidateSettingType() {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid setting type. Must be one of: string, integer, boolean, json, text"})
			return
		}
		setting.Type = updateData.Type
	}
	if updateData.Description != "" {
		setting.Description = updateData.Description
	}
	if updateData.Group != "" {
		setting.Group = updateData.Group
	}

	// Handle boolean field
	setting.IsPublic = updateData.IsPublic

	// Validate value based on type
	if err := validateSettingValue(&setting); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := database.DB.Save(&setting).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Setting with this key already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// UpdateSettingByKey godoc
// @Summary Update setting by key
// @Description Update a setting value by its key (admin only)
// @Tags Settings
// @Accept json
// @Produce json
// @Security Bearer
// @Param key path string true "Setting key"
// @Param data body map[string]interface{} true "Setting value"
// @Success 200 {object} models.Setting
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/settings/key/{key} [put]
func UpdateSettingByKey(c *gin.Context) {
	key := c.Param("key")

	var setting models.Setting
	if err := database.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Setting not found"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update value if provided
	if value, exists := updateData["value"]; exists {
		if valueStr, ok := value.(string); ok {
			setting.Value = valueStr
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Value must be a string"})
			return
		}
	}

	// Validate value based on type
	if err := validateSettingValue(&setting); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if err := database.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// DeleteSetting godoc
// @Summary Delete a setting
// @Description Soft delete a setting (admin only)
// @Tags Settings
// @Produce json
// @Security Bearer
// @Param id path int true "Setting ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/settings/{id} [delete]
func DeleteSetting(c *gin.Context) {
	id := c.Param("id")

	var setting models.Setting
	if err := database.DB.First(&setting, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Setting not found"})
		return
	}

	if err := database.DB.Delete(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete setting"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// BulkUpdateSettings godoc
// @Summary Bulk update settings
// @Description Update multiple settings at once (admin only)
// @Tags Settings
// @Accept json
// @Produce json
// @Security Bearer
// @Param settings body map[string]string true "Map of setting keys to values"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/settings/bulk [put]
func BulkUpdateSettings(c *gin.Context) {
	var updateData map[string]string
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	tx := database.DB.Begin()
	updatedCount := 0

	for key, value := range updateData {
		var setting models.Setting
		if err := tx.Where("key = ?", key).First(&setting).Error; err != nil {
			// Skip non-existent settings
			continue
		}

		setting.Value = value

		// Validate value based on type
		if err := validateSettingValue(&setting); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid value for setting '" + key + "': " + err.Error(),
			})
			return
		}

		if err := tx.Save(&setting).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update settings"})
			return
		}

		updatedCount++
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"updated": updatedCount,
	})
}

// Helper function to validate setting value based on type
func validateSettingValue(setting *models.Setting) error {
	switch setting.Type {
	case "integer":
		if _, err := strconv.Atoi(setting.Value); err != nil {
			return err
		}
	case "boolean":
		if setting.Value != "true" && setting.Value != "false" {
			return fmt.Errorf("boolean value must be 'true' or 'false'")
		}
		// For json and text types, we accept any string value
		// For string type, we also accept any string value
	}
	return nil
}
