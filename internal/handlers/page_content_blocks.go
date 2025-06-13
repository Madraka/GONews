package handlers

import (
	"net/http"
	"strconv"

	"news/internal/database"
	"news/internal/models"
	"news/internal/repositories"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// CreatePageBlock godoc
// @Summary Create a page content block
// @Description Create a new content block for a page
// @Tags Page Content Blocks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Param block body services.CreatePageBlockRequest true "Content block data"
// @Success 201 {object} models.PageContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id}/blocks [post]
func CreatePageBlock(c *gin.Context) {
	pageIDStr := c.Param("id")
	pageID, err := strconv.ParseUint(pageIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	var req services.CreatePageBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Set page ID from URL parameter
	req.PageID = uint(pageID)

	// Set default visibility
	if !req.IsVisible && c.PostForm("is_visible") == "" {
		req.IsVisible = true
	}

	// Validate required fields
	if req.BlockType == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Block type is required"})
		return
	}

	// Create content block using service
	contentBlockService := services.NewPageContentBlockService(database.DB)
	block, err := contentBlockService.CreateBlock(uint(pageID), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create content block: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, block)
}

// GetPageBlock godoc
// @Summary Get a page content block
// @Description Retrieve a content block by ID
// @Tags Page Content Blocks
// @Produce json
// @Param id path int true "Block ID"
// @Success 200 {object} models.PageContentBlock
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/page-blocks/{id} [get]
func GetPageBlock(c *gin.Context) {
	id := c.Param("id")
	blockID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	blockRepo := repositories.NewPageContentBlockRepository(database.DB)
	block, err := blockRepo.GetByID(uint(blockID))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content block not found"})
		return
	}

	c.JSON(http.StatusOK, block)
}

// UpdatePageBlock godoc
// @Summary Update a page content block
// @Description Update an existing content block
// @Tags Page Content Blocks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Block ID"
// @Param block body services.UpdatePageBlockRequest true "Content block data"
// @Success 200 {object} models.PageContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/page-blocks/{id} [put]
func UpdatePageBlock(c *gin.Context) {
	id := c.Param("id")
	blockID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	var req services.UpdatePageBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	contentBlockService := services.NewPageContentBlockService(database.DB)
	block, err := contentBlockService.UpdateBlock(uint(blockID), req)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content block not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, block)
}

// DeletePageBlock godoc
// @Summary Delete a page content block
// @Description Delete a content block by ID
// @Tags Page Content Blocks
// @Security BearerAuth
// @Param id path int true "Block ID"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/page-blocks/{id} [delete]
func DeletePageBlock(c *gin.Context) {
	id := c.Param("id")
	blockID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	contentBlockService := services.NewPageContentBlockService(database.DB)
	err = contentBlockService.DeleteBlock(uint(blockID))
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content block not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// DuplicatePageBlock godoc
// @Summary Duplicate a page content block
// @Description Create a copy of an existing content block
// @Tags Page Content Blocks
// @Security BearerAuth
// @Param id path int true "Block ID"
// @Param request body services.DuplicateBlockRequest false "Duplication options"
// @Success 201 {object} models.PageContentBlock
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/page-blocks/{id}/duplicate [post]
func DuplicatePageBlock(c *gin.Context) {
	id := c.Param("id")
	blockID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	var req services.DuplicateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use default values if no request body provided
		req = services.DuplicateBlockRequest{}
	}

	contentBlockService := services.NewPageContentBlockService(database.DB)
	block, err := contentBlockService.DuplicateBlock(uint(blockID), req)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Content block not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, block)
}

// ValidatePageBlock godoc
// @Summary Validate a page content block
// @Description Validate content block data
// @Tags Page Content Blocks
// @Accept json
// @Produce json
// @Param block body services.CreatePageBlockRequest true "Content block data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /admin/page-blocks/validate [post]
func ValidatePageBlock(c *gin.Context) {
	var req services.CreatePageBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	contentBlockService := services.NewPageContentBlockService(database.DB)
	result := contentBlockService.ValidateBlock(req)

	c.JSON(http.StatusOK, gin.H{
		"valid":    result.IsValid,
		"errors":   result.Errors,
		"warnings": result.Warnings,
	})
}
