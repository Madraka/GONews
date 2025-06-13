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

// Package-level page service functions following the same pattern as articles

// GetPages godoc
// @Summary Get all pages
// @Description Retrieve a list of pages with pagination and filtering
// @Tags Pages
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Param template query string false "Filter by template"
// @Param language query string false "Filter by language"
// @Param parent_id query int false "Filter by parent ID"
// @Param search query string false "Search in title and content"
// @Success 200 {object} services.PaginatedPagesResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pages [get]
func GetPages(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	template := c.Query("template")
	language := c.Query("language")
	search := c.Query("search")

	var parentID *uint
	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if pid, err := strconv.ParseUint(parentIDStr, 10, 32); err == nil {
			pidUint := uint(pid)
			parentID = &pidUint
		}
	}

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Create service and get pages
	pageService := services.NewPageService(database.DB)

	filter := services.PageFilter{
		Status:   status,
		Template: template,
		Language: language,
		ParentID: parentID,
		Search:   search,
	}

	result, err := pageService.GetPages(page, limit, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPageByID godoc
// @Summary Get a page by ID
// @Description Retrieve a single page by its ID
// @Tags Pages
// @Produce json
// @Param id path int true "Page ID"
// @Param include_blocks query bool false "Include content blocks" default(false)
// @Success 200 {object} models.Page
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pages/{id} [get]
func GetPageByID(c *gin.Context) {
	id := c.Param("id")
	includeBlocks := c.Query("include_blocks") == "true"

	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	pageService := services.NewPageService(database.DB)
	page, err := pageService.GetPageByID(uint(pageID), includeBlocks)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

// GetPageBySlug godoc
// @Summary Get a page by slug
// @Description Retrieve a single page by its slug
// @Tags Pages
// @Produce json
// @Param slug path string true "Page slug"
// @Param include_blocks query bool false "Include content blocks" default(false)
// @Success 200 {object} models.Page
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pages/slug/{slug} [get]
func GetPageBySlug(c *gin.Context) {
	slug := c.Param("slug")
	includeBlocks := c.Query("include_blocks") == "true"

	pageService := services.NewPageService(database.DB)
	page, err := pageService.GetPageBySlug(slug, includeBlocks)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

// GetPageHierarchy godoc
// @Summary Get page hierarchy
// @Description Retrieve page hierarchy (parent-child relationships)
// @Tags Pages
// @Produce json
// @Success 200 {array} models.Page
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pages/hierarchy [get]
func GetPageHierarchy(c *gin.Context) {
	pageRepo := repositories.NewPageRepository(database.DB)
	pages, err := pageRepo.GetPageHierarchy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, pages)
}

// GetPageBlocks godoc
// @Summary Get content blocks for a page
// @Description Retrieve all content blocks for a specific page
// @Tags Pages
// @Produce json
// @Param id path int true "Page ID"
// @Param include_hidden query bool false "Include hidden blocks" default(false)
// @Success 200 {array} models.PageContentBlock
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/pages/{id}/blocks [get]
func GetPageBlocks(c *gin.Context) {
	id := c.Param("id")
	includeHidden := c.Query("include_hidden") == "true"

	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	blockRepo := repositories.NewPageContentBlockRepository(database.DB)
	blocks, err := blockRepo.GetByPageID(uint(pageID), includeHidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, blocks)
}

// CreatePage godoc
// @Summary Create a new page
// @Description Create a new page with content blocks
// @Tags Pages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page body services.CreatePageRequest true "Page data"
// @Success 201 {object} models.Page
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages [post]
func CreatePage(c *gin.Context) {
	var req services.CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Get author ID from token claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get author from token"})
		return
	}

	// Validate required fields
	if len(req.Title) < 3 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Title must be at least 3 characters"})
		return
	}

	// Set defaults
	if req.Template == "" {
		req.Template = "default"
	}
	if req.Layout == "" {
		req.Layout = "container"
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Language == "" {
		req.Language = "tr"
	}

	// Set author ID
	req.AuthorID = userID.(uint)

	// Create page using service
	pageService := services.NewPageService(database.DB)
	page, err := pageService.CreatePage(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create page: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, page)
}

// UpdatePage godoc
// @Summary Update a page
// @Description Update an existing page
// @Tags Pages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Param page body services.UpdatePageRequest true "Page data"
// @Success 200 {object} models.Page
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id} [put]
func UpdatePage(c *gin.Context) {
	id := c.Param("id")
	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	var req services.UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	pageService := services.NewPageService(database.DB)
	page, err := pageService.UpdatePage(uint(pageID), req)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

// DeletePage godoc
// @Summary Delete a page
// @Description Delete a page by ID
// @Tags Pages
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id} [delete]
func DeletePage(c *gin.Context) {
	id := c.Param("id")
	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	pageService := services.NewPageService(database.DB)
	err = pageService.DeletePage(uint(pageID))
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishPage godoc
// @Summary Publish a page
// @Description Change page status to published
// @Tags Pages
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Success 200 {object} models.Page
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id}/publish [post]
func PublishPage(c *gin.Context) {
	id := c.Param("id")
	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	pageService := services.NewPageService(database.DB)
	page, err := pageService.PublishPage(uint(pageID))
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

// UnpublishPage godoc
// @Summary Unpublish a page
// @Description Change page status to draft
// @Tags Pages
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Success 200 {object} models.Page
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id}/unpublish [post]
func UnpublishPage(c *gin.Context) {
	id := c.Param("id")
	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	pageService := services.NewPageService(database.DB)
	page, err := pageService.UnpublishPage(uint(pageID))
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

// DuplicatePage godoc
// @Summary Duplicate a page
// @Description Create a copy of an existing page
// @Tags Pages
// @Security BearerAuth
// @Param id path int true "Page ID"
// @Param title query string false "Title for the duplicated page"
// @Success 201 {object} models.Page
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/pages/{id}/duplicate [post]
func DuplicatePage(c *gin.Context) {
	id := c.Param("id")
	newTitle := c.Query("title")

	pageID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid page ID"})
		return
	}

	// Get author ID from token claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get author from token"})
		return
	}

	// Create duplicate request
	req := services.DuplicatePageRequest{
		NewTitle:       newTitle,
		NewSlug:        "", // Will be auto-generated if empty
		IncludeBlocks:  true,
		CopyAsTemplate: false,
		AuthorID:       userID.(uint),
	}

	pageService := services.NewPageService(database.DB)
	page, err := pageService.DuplicatePage(uint(pageID), req)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, page)
}
