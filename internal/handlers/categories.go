package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// Helper function to generate a slug from a string
func generateSlug(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

// GetCategories godoc
// @Summary Get all categories
// @Description Retrieve all article categories with unified cache support
// @Tags Categories
// @Produce json
// @Param hierarchical query bool false "Return hierarchical structure"
// @Success 200 {array} models.Category
// @Failure 500 {object} models.ErrorResponse
// @Router /categories [get]
func GetCategories(c *gin.Context) {
	hierarchical := c.Query("hierarchical") == "true"

	// Use cached service
	categories, err := services.GetCategoriesWithCache(hierarchical)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryBySlug godoc
// @Summary Get category by slug
// @Description Retrieve a single category by its slug with cache support
// @Tags Categories
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} models.Category
// @Failure 404 {object} models.ErrorResponse
// @Router /categories/{slug} [get]
func GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	// Use cached service
	category, err := services.GetCategoryBySlugWithCache(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new article category with cache invalidation (admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param category body models.Category true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/categories [post]
func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Use cached service for creation
	createdCategory, err := services.CreateCategoryWithCache(category)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update an existing category with cache invalidation (admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Category ID"
// @Param category body models.Category true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var updateData models.Category
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Use cached service for update
	updatedCategory, err := services.UpdateCategoryWithCache(id, updateData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update category"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Soft delete a category with cache invalidation (admin only)
// @Tags Categories
// @Produce json
// @Security Bearer
// @Param id path int true "Category ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	// Use cached service for deletion
	if err := services.DeleteCategoryWithCache(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete category"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetTags godoc
// @Summary Get all tags
// @Description Retrieve all article tags with optional usage count sorting (Cache Optimized)
// @Tags Tags
// @Produce json
// @Param sort query string false "Sort by: name, usage_count" default(name)
// @Param limit query int false "Limit number of results" default(50)
// @Success 200 {array} models.Tag
// @Failure 500 {object} models.ErrorResponse
// @Router /tags [get]
func GetTags(c *gin.Context) {
	sort := c.DefaultQuery("sort", "name")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Use cached service for improved performance
	tags, err := services.GetTagsWithCache(sort, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch tags"})
		return
	}

	c.JSON(http.StatusOK, tags)
}

// GetTagBySlug godoc
// @Summary Get tag by slug
// @Description Retrieve a single tag by its slug with related articles (Cache Optimized)
// @Tags Tags
// @Produce json
// @Param slug path string true "Tag slug"
// @Success 200 {object} models.Tag
// @Failure 404 {object} models.ErrorResponse
// @Router /tags/{slug} [get]
func GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")

	// Use cached service for improved performance
	tag, err := services.GetTagBySlugWithCache(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// CreateTag godoc
// @Summary Create a new tag
// @Description Create a new article tag with cache invalidation (admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param tag body models.Tag true "Tag data"
// @Success 201 {object} models.Tag
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /admin/tags [post]
func CreateTag(c *gin.Context) {
	var tag models.Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Use cached service for creation with cache invalidation
	createdTag, err := services.CreateTagWithCache(tag)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, createdTag)
}

// UpdateTag godoc
// @Summary Update a tag
// @Description Update an existing tag with cache invalidation (admin only)
// @Tags Tags
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Tag ID"
// @Param tag body models.Tag true "Tag data"
// @Success 200 {object} models.Tag
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/tags/{id} [put]
func UpdateTag(c *gin.Context) {
	id := c.Param("id")

	var updateData models.Tag
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Use cached service for update with cache invalidation
	updatedTag, err := services.UpdateTagWithCache(id, updateData)
	if err != nil {
		if err.Error() == "tag not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Tag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update tag"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedTag)
}

// DeleteTag godoc
// @Summary Delete a tag
// @Description Soft delete a tag with cache invalidation (admin only)
// @Tags Tags
// @Produce json
// @Security Bearer
// @Param id path int true "Tag ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/tags/{id} [delete]
func DeleteTag(c *gin.Context) {
	id := c.Param("id")

	// Use cached service for deletion with cache invalidation
	err := services.DeleteTagWithCache(id)
	if err != nil {
		if err.Error() == "tag not found: record not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Tag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete tag"})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
