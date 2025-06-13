package handlers

import (
	"net/http"
	"strconv"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMenuItems godoc
// @Summary Get menu items
// @Description Retrieve menu items for a specific menu
// @Tags MenuItem
// @Produce json
// @Param menu_id query int false "Filter by menu ID"
// @Param parent_id query int false "Filter by parent ID (0 for root items)"
// @Success 200 {array} models.MenuItem
// @Failure 500 {object} models.ErrorResponse
// @Router /menu-items [get]
func GetMenuItems(c *gin.Context) {
	menuIDStr := c.Query("menu_id")
	parentIDStr := c.Query("parent_id")

	query := database.DB.Preload("Menu").Preload("Category").Preload("Children", "is_active = ?", true)

	if menuIDStr != "" {
		menuID, err := strconv.Atoi(menuIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid menu_id"})
			return
		}
		query = query.Where("menu_id = ?", menuID)
	}

	if parentIDStr != "" {
		if parentIDStr == "0" {
			query = query.Where("parent_id IS NULL")
		} else {
			parentID, err := strconv.Atoi(parentIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid parent_id"})
				return
			}
			query = query.Where("parent_id = ?", parentID)
		}
	}

	var items []models.MenuItem
	if err := query.Where("is_active = ?", true).Order("sort_order ASC, title ASC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch menu items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetMenuItem godoc
// @Summary Get menu item by ID
// @Description Retrieve a single menu item by its ID
// @Tags MenuItem
// @Produce json
// @Param id path int true "Menu Item ID"
// @Success 200 {object} models.MenuItem
// @Failure 404 {object} models.ErrorResponse
// @Router /menu-items/{id} [get]
func GetMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := database.DB.Preload("Menu").Preload("Category").Preload("Parent").
		Preload("Children", "is_active = ?", true).
		First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// CreateMenuItem godoc
// @Summary Create a new menu item
// @Description Create a new menu item (admin only)
// @Tags MenuItem
// @Accept json
// @Produce json
// @Security Bearer
// @Param item body models.MenuItem true "Menu item data"
// @Success 201 {object} models.MenuItem
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/menu-items [post]
func CreateMenuItem(c *gin.Context) {
	var item models.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate menu exists
	var menu models.Menu
	if err := database.DB.First(&menu, item.MenuID).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Menu not found"})
		return
	}

	// Validate parent item exists if parent_id is provided
	if item.ParentID != nil {
		var parent models.MenuItem
		if err := database.DB.First(&parent, *item.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Parent menu item not found"})
			return
		}
		// Ensure parent belongs to the same menu
		if parent.MenuID != item.MenuID {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Parent item must belong to the same menu"})
			return
		}
	}

	// Validate category exists if category_id is provided
	if item.CategoryID != nil {
		var category models.Category
		if err := database.DB.First(&category, *item.CategoryID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Category not found"})
			return
		}
	}

	// Validate target
	validTargets := map[string]bool{
		"_self":   true,
		"_blank":  true,
		"_parent": true,
		"_top":    true,
	}
	if item.Target == "" {
		item.Target = "_self"
	} else if !validTargets[item.Target] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid target. Must be one of: _self, _blank, _parent, _top"})
		return
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create menu item"})
		return
	}

	// Load relations for response
	database.DB.Preload("Menu").Preload("Category").Preload("Parent").First(&item, item.ID)

	c.JSON(http.StatusCreated, item)
}

// UpdateMenuItem godoc
// @Summary Update a menu item
// @Description Update an existing menu item (admin only)
// @Tags MenuItem
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Menu Item ID"
// @Param item body models.MenuItem true "Menu item data"
// @Success 200 {object} models.MenuItem
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/menu-items/{id} [put]
func UpdateMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu item not found"})
		return
	}

	var updateData models.MenuItem
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Update fields
	if updateData.Title != "" {
		item.Title = updateData.Title
	}
	if updateData.URL != "" {
		item.URL = updateData.URL
	}
	if updateData.Icon != "" {
		item.Icon = updateData.Icon
	}
	if updateData.Target != "" {
		validTargets := map[string]bool{
			"_self":   true,
			"_blank":  true,
			"_parent": true,
			"_top":    true,
		}
		if !validTargets[updateData.Target] {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid target. Must be one of: _self, _blank, _parent, _top"})
			return
		}
		item.Target = updateData.Target
	}

	// Handle parent_id update
	if updateData.ParentID != nil {
		// Validate parent item exists
		var parent models.MenuItem
		if err := database.DB.First(&parent, *updateData.ParentID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Parent menu item not found"})
			return
		}
		// Ensure parent belongs to the same menu
		if parent.MenuID != item.MenuID {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Parent item must belong to the same menu"})
			return
		}
		// Prevent circular reference
		if *updateData.ParentID == item.ID {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Menu item cannot be its own parent"})
			return
		}
		item.ParentID = updateData.ParentID
	}

	// Handle category_id update
	if updateData.CategoryID != nil {
		var category models.Category
		if err := database.DB.First(&category, *updateData.CategoryID).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Category not found"})
			return
		}
		item.CategoryID = updateData.CategoryID
	}

	// Update other fields
	item.SortOrder = updateData.SortOrder
	item.IsActive = updateData.IsActive

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update menu item"})
		return
	}

	// Load relations for response
	database.DB.Preload("Menu").Preload("Category").Preload("Parent").
		Preload("Children", "is_active = ?", true).First(&item, item.ID)

	c.JSON(http.StatusOK, item)
}

// DeleteMenuItem godoc
// @Summary Delete a menu item
// @Description Soft delete a menu item and all its children (admin only)
// @Tags MenuItem
// @Produce json
// @Security Bearer
// @Param id path int true "Menu Item ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/menu-items/{id} [delete]
func DeleteMenuItem(c *gin.Context) {
	id := c.Param("id")

	var item models.MenuItem
	if err := database.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Menu item not found"})
		return
	}

	// Start transaction to delete item and all its children
	tx := database.DB.Begin()

	// Recursively delete all children
	if err := deleteMenuItemChildren(tx, item.ID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete menu item children"})
		return
	}

	// Delete the item itself
	if err := tx.Delete(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete menu item"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusNoContent, nil)
}

// ReorderMenuItems godoc
// @Summary Reorder menu items
// @Description Update sort order for multiple menu items (admin only)
// @Tags MenuItem
// @Accept json
// @Produce json
// @Security Bearer
// @Param items body []map[string]interface{} true "Array of items with id and sort_order"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/menu-items/reorder [put]
func ReorderMenuItems(c *gin.Context) {
	var items []map[string]interface{}
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	tx := database.DB.Begin()

	for _, itemData := range items {
		id, ok := itemData["id"].(float64)
		if !ok {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid item ID"})
			return
		}

		sortOrder, ok := itemData["sort_order"].(float64)
		if !ok {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid sort order"})
			return
		}

		if err := tx.Model(&models.MenuItem{}).Where("id = ?", uint(id)).
			Update("sort_order", int(sortOrder)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update menu item order"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Menu items reordered successfully"})
}

// Helper function to recursively delete menu item children
func deleteMenuItemChildren(tx *gorm.DB, parentID uint) error {
	var children []models.MenuItem
	if err := tx.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		// Recursively delete grandchildren
		if err := deleteMenuItemChildren(tx, child.ID); err != nil {
			return err
		}
		// Delete the child
		if err := tx.Delete(&child).Error; err != nil {
			return err
		}
	}

	return nil
}
