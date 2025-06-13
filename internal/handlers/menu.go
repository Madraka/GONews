package handlers

import (
	"net/http"
	"strings"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMenus godoc
// @Summary Get all menus
// @Description Retrieve all menus with their items
// @Tags Menu
// @Produce json
// @Param location query string false "Filter by location (header, footer, sidebar)"
// @Param active query bool false "Filter by active status" default(true)
// @Success 200 {array} models.Menu
// @Failure 500 {object} models.ErrorResponse
// @Router /menus [get]
func GetMenus(c *gin.Context) {
	location := c.Query("location")
	activeParam := c.DefaultQuery("active", "true")
	active := activeParam == "true"

	query := database.DB.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Order("sort_order ASC")
	}).Preload("Items.Category").Preload("Items.Children", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_active = ?", true).Order("sort_order ASC")
	})

	if location != "" {
		query = query.Where("location = ?", location)
	}
	if active {
		query = query.Where("is_active = ?", true)
	}

	var menus []models.Menu
	if err := query.Order("name ASC").Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch menus"})
		return
	}

	c.JSON(http.StatusOK, menus)
}

// GetMenuBySlug godoc
// @Summary Get menu by slug
// @Description Retrieve a single menu by its slug with items
// @Tags Menu
// @Produce json
// @Param slug path string true "Menu slug"
// @Success 200 {object} models.Menu
// @Failure 404 {object} models.ErrorResponse
// @Router /menus/{slug} [get]
func GetMenuBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var menu models.Menu
	if err := database.DB.Where("slug = ? AND is_active = ?", slug, true).
		Preload("Items", "is_active = ? AND parent_id IS NULL", true).
		Preload("Items.Category").
		Preload("Items.Children", "is_active = ?", true).
		Order("Items.sort_order ASC").
		First(&menu).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu not found"})
		return
	}

	c.JSON(http.StatusOK, menu)
}

// CreateMenu godoc
// @Summary Create a new menu
// @Description Create a new navigation menu (admin only)
// @Tags Menu
// @Accept json
// @Produce json
// @Security Bearer
// @Param menu body models.Menu true "Menu data"
// @Success 201 {object} models.Menu
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/menus [post]
func CreateMenu(c *gin.Context) {
	var menu models.Menu
	if err := c.ShouldBindJSON(&menu); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Generate slug if not provided
	if menu.Slug == "" {
		menu.Slug = generateSlug(menu.Name)
	}

	// Validate location
	validLocations := map[string]bool{
		"header":  true,
		"footer":  true,
		"sidebar": true,
	}
	if !validLocations[menu.Location] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid location. Must be one of: header, footer, sidebar"})
		return
	}

	if err := database.DB.Create(&menu).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Menu with this slug already exists"})
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create menu"})
		return
	}

	c.JSON(http.StatusCreated, menu)
}

// UpdateMenu godoc
// @Summary Update a menu
// @Description Update an existing menu (admin only)
// @Tags Menu
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Menu ID"
// @Param menu body models.Menu true "Menu data"
// @Success 200 {object} models.Menu
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/menus/{id} [put]
func UpdateMenu(c *gin.Context) {
	id := c.Param("id")

	var menu models.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu not found"})
		return
	}

	var updateData models.Menu
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update fields
	if updateData.Name != "" {
		menu.Name = updateData.Name
	}
	if updateData.Slug != "" {
		menu.Slug = updateData.Slug
	}
	if updateData.Location != "" {
		validLocations := map[string]bool{
			"header":  true,
			"footer":  true,
			"sidebar": true,
		}
		if !validLocations[updateData.Location] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid location. Must be one of: header, footer, sidebar"})
			return
		}
		menu.Location = updateData.Location
	}

	// Handle boolean field - check if it was explicitly set
	menu.IsActive = updateData.IsActive

	if err := database.DB.Save(&menu).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Menu with this slug already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update menu"})
		return
	}

	// Load items for response
	database.DB.Preload("Items", "is_active = ?", true).
		Preload("Items.Category").
		Preload("Items.Children", "is_active = ?", true).
		First(&menu, menu.ID)

	c.JSON(http.StatusOK, menu)
}

// DeleteMenu godoc
// @Summary Delete a menu
// @Description Soft delete a menu and all its items (admin only)
// @Tags Menu
// @Produce json
// @Security Bearer
// @Param id path int true "Menu ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/menus/{id} [delete]
func DeleteMenu(c *gin.Context) {
	id := c.Param("id")

	var menu models.Menu
	if err := database.DB.First(&menu, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu not found"})
		return
	}

	// Start transaction to delete menu and all its items
	tx := database.DB.Begin()

	// Delete all menu items first
	if err := tx.Where("menu_id = ?", menu.ID).Delete(&models.MenuItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete menu items"})
		return
	}

	// Delete the menu
	if err := tx.Delete(&menu).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete menu"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusNoContent, nil)
}
