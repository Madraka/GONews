package handlers

import (
	"net/http"
	"strconv"

	"news/internal/json"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// @Summary Get articles with pagination (Cache Optimized)
// @Description Retrieve a list of articles with pagination using cached JSON
// @Tags Articles
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10, max: 50)"
// @Param category query string false "Filter by category"
// @Success 200 {object} models.PaginatedResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles [get]
func GetArticles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get cached JSON from service with smart redaction
	cachedJSON, err := services.GetArticlesWithPaginationCachedSmart(offset, limit, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add headers to indicate redaction status if enabled
	if json.IsRedactionEnabled() {
		c.Header("X-Content-Redacted", "true")
		c.Header("X-Redaction-Version", "1.0")
	}

	// Return raw JSON directly (ZERO marshal overhead!)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, cachedJSON)
}

// @Summary Get a single article by ID (Cache Optimized)
// @Description Retrieve a single article by its ID using cached JSON
// @Tags Articles
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} models.Article
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id} [get]
func GetArticleById(c *gin.Context) {
	id := c.Param("id")

	// Get cached JSON from service with smart redaction
	cachedJSON, err := services.GetArticleByIdCachedSmart(id)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	// Add headers to indicate redaction status if enabled
	if json.IsRedactionEnabled() {
		c.Header("X-Content-Redacted", "true")
		c.Header("X-Redaction-Version", "1.0")
	}

	// Return raw JSON directly (ZERO marshal overhead!)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, cachedJSON)
}

// @Summary Create a new article
// @Description Add a new article to the database
// @Tags Articles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param article body models.Article true "Article data"
// @Success 201 {object} models.Article
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /admin/articles [post]
func CreateArticle(c *gin.Context) {
	var articleInput struct {
		Title         string   `json:"title" binding:"required"`
		Content       string   `json:"content" binding:"required"`
		CategoryIDs   []uint   `json:"category_ids"`
		TagIDs        []uint   `json:"tag_ids"`
		FeaturedImage string   `json:"featured_image"`
		Gallery       []string `json:"gallery"` // Array of image URLs
		Status        string   `json:"status"`
		MetaTitle     string   `json:"meta_title"`
		MetaDesc      string   `json:"meta_description"`
		Language      string   `json:"language"`
	}

	if err := c.ShouldBindJSON(&articleInput); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Additional validation
	if len(articleInput.Title) < 3 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Title must be at least 3 characters"})
		return
	}

	if len(articleInput.Content) < 10 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Content must be at least 10 characters"})
		return
	}

	// Get author ID from token claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get author from token"})
		return
	}

	// Create article
	article := models.Article{
		Title:         articleInput.Title,
		Content:       articleInput.Content,
		AuthorID:      userID.(uint),
		FeaturedImage: articleInput.FeaturedImage,
		Status:        articleInput.Status,
		MetaTitle:     articleInput.MetaTitle,
		MetaDesc:      articleInput.MetaDesc,
		Language:      articleInput.Language,
	}

	// Handle Gallery field - convert array to JSON string or set empty array
	if len(articleInput.Gallery) > 0 {
		galleryJSON, err := json.Marshal(articleInput.Gallery)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid gallery format"})
			return
		}
		article.Gallery = datatypes.JSON(galleryJSON)
	} else {
		article.Gallery = datatypes.JSON("[]") // Empty JSON array
	}

	if article.Status == "" {
		article.Status = "draft"
	}

	if article.Language == "" {
		article.Language = "tr"
	}

	createdArticle, err := services.CreateArticle(article)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdArticle)
}

// @Summary Update an article
// @Description Update an existing article in the database (admin only)
// @Tags Articles
// @Accept json
// @Produce json
// @Param id path int true "Article ID"
// @Param article body models.Article true "Article data"
// @Success 200 {object} models.Article
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /admin/articles/{id} [put]
func UpdateArticle(c *gin.Context) {
	id := c.Param("id")

	// Get existing article first
	existingArticle, err := services.GetArticleById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Parse update data with custom struct to handle Gallery as array
	var updateInput struct {
		Title         string   `json:"title"`
		Content       string   `json:"content"`
		FeaturedImage string   `json:"featured_image"`
		Gallery       []string `json:"gallery"` // Array of image URLs
		Status        string   `json:"status"`
		MetaTitle     string   `json:"meta_title"`
		MetaDesc      string   `json:"meta_description"`
		Language      string   `json:"language"`
	}

	if err := c.ShouldBindJSON(&updateInput); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Update fields with validation
	if updateInput.Title != "" {
		if len(updateInput.Title) < 3 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Title must be at least 3 characters"})
			return
		}
		existingArticle.Title = updateInput.Title
	}

	if updateInput.Content != "" {
		if len(updateInput.Content) < 10 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Content must be at least 10 characters"})
			return
		}
		existingArticle.Content = updateInput.Content
	}

	if updateInput.FeaturedImage != "" {
		existingArticle.FeaturedImage = updateInput.FeaturedImage
	}

	if updateInput.Status != "" {
		existingArticle.Status = updateInput.Status
	}

	if updateInput.MetaTitle != "" {
		existingArticle.MetaTitle = updateInput.MetaTitle
	}

	if updateInput.MetaDesc != "" {
		existingArticle.MetaDesc = updateInput.MetaDesc
	}

	// Handle Gallery field updates - convert array to JSON or set empty array
	if len(updateInput.Gallery) > 0 {
		galleryJSON, err := json.Marshal(updateInput.Gallery)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid gallery format"})
			return
		}
		existingArticle.Gallery = datatypes.JSON(galleryJSON)
	} else if updateInput.Gallery != nil && len(updateInput.Gallery) == 0 {
		// Explicitly set empty array if Gallery is provided as empty array
		existingArticle.Gallery = datatypes.JSON("[]")
	}

	updatedArticle, err := services.UpdateArticle(id, existingArticle)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedArticle)
}

// @Summary Delete an article
// @Description Delete an article by ID (admin only)
// @Tags Articles
// @Param id path int true "Article ID"
// @Success 204 "No Content"
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /admin/articles/{id} [delete]
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")

	err := services.DeleteArticle(id)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateArticleWithBlocks godoc
// @Summary Create a new article with content blocks
// @Description Add a new article using modern content blocks system
// @Tags Articles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param article body CreateArticleWithBlocksRequest true "Article with content blocks data"
// @Success 201 {object} models.Article
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/articles/blocks [post]
func CreateArticleWithBlocks(c *gin.Context) {
	var request CreateArticleWithBlocksRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Additional validation
	if len(request.Title) < 3 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Title must be at least 3 characters"})
		return
	}

	if len(request.ContentBlocks) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "At least one content block is required"})
		return
	}

	// Get author ID from token claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get author from token"})
		return
	}

	// Create article structure
	article := models.Article{
		Title:         request.Title,
		Summary:       request.Summary,
		AuthorID:      userID.(uint),
		FeaturedImage: request.FeaturedImage,
		Status:        request.Status,
		MetaTitle:     request.MetaTitle,
		MetaDesc:      request.MetaDesc,
		Language:      request.Language,
		ContentType:   "blocks", // Modern content blocks system
		HasBlocks:     true,
		BlocksVersion: 1,
	}

	// Handle Gallery field
	if len(request.Gallery) > 0 {
		galleryJSON, err := json.Marshal(request.Gallery)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid gallery format"})
			return
		}
		article.Gallery = datatypes.JSON(galleryJSON)
	} else {
		article.Gallery = datatypes.JSON("[]")
	}

	// Set defaults
	if article.Status == "" {
		article.Status = "draft"
	}
	if article.Language == "" {
		article.Language = "tr"
	}

	// Create article with content blocks
	createdArticle, err := services.CreateArticleWithBlocks(article, request.ContentBlocks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdArticle)
}

// GetArticleWithBlocks godoc
// @Summary Get article with content blocks
// @Description Retrieve an article with its content blocks for editing
// @Tags Articles
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} models.Article
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/with-blocks [get]
func GetArticleWithBlocks(c *gin.Context) {
	id := c.Param("id")

	// Get article with blocks
	article, err := services.GetArticleWithBlocks(id)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, article)
}

// GetArticlesWithRedaction handles articles endpoint with redaction capabilities
// @Summary Get articles with pagination and optional redaction
// @Description Retrieve a list of articles with pagination. Redaction applied when NEWS_REDACTION_ENABLED=true
// @Tags Articles
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Number of items per page (default: 10, max: 50)"
// @Param category query string false "Filter by category"
// @Param redact query bool false "Force redaction of sensitive data"
// @Success 200 {object} models.PaginatedResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/secure [get]
func GetArticlesWithRedaction(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")
	forceRedact := c.Query("redact") == "true"

	// Validate parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Calculate offset
	offset := (page - 1) * limit

	var cachedJSON string
	var err error

	// Use redaction if enabled globally or forced by parameter
	if json.IsRedactionEnabled() || forceRedact {
		cachedJSON, err = services.GetArticlesWithPaginationCachedWithRedaction(offset, limit, category)
	} else {
		cachedJSON, err = services.GetArticlesWithPaginationCached(offset, limit, category)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add headers to indicate redaction status
	if json.IsRedactionEnabled() || forceRedact {
		c.Header("X-Content-Redacted", "true")
		c.Header("X-Redaction-Version", "1.0")
	}

	// Return raw JSON directly (ZERO marshal overhead!)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, cachedJSON)
}

// GetArticleByIdWithRedaction handles single article endpoint with redaction capabilities
// @Summary Get a single article by ID with optional redaction
// @Description Retrieve a single article by its ID. Redaction applied when NEWS_REDACTION_ENABLED=true
// @Tags Articles
// @Produce json
// @Param id path int true "Article ID"
// @Param redact query bool false "Force redaction of sensitive data"
// @Success 200 {object} models.Article
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/secure [get]
func GetArticleByIdWithRedaction(c *gin.Context) {
	// Handle both parameter formats for compatibility
	id := c.Param("id")
	if id == "" {
		id = c.Param("article_id")
	}
	forceRedact := c.Query("redact") == "true"

	var cachedJSON string
	var err error

	// Use redaction if enabled globally or forced by parameter
	if json.IsRedactionEnabled() || forceRedact {
		cachedJSON, err = services.GetArticleByIdCachedWithRedaction(id)
	} else {
		cachedJSON, err = services.GetArticleByIdCached(id)
	}

	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	// Add headers to indicate redaction status
	if json.IsRedactionEnabled() || forceRedact {
		c.Header("X-Content-Redacted", "true")
		c.Header("X-Redaction-Version", "1.0")
	}

	// Return raw JSON directly (ZERO marshal overhead!)
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, cachedJSON)
}

// CreateArticleWithBlocksRequest represents the request for creating an article with blocks
type CreateArticleWithBlocksRequest struct {
	Title         string                       `json:"title" binding:"required"`
	Summary       string                       `json:"summary"`
	FeaturedImage string                       `json:"featured_image"`
	Gallery       []string                     `json:"gallery"`
	Status        string                       `json:"status"`
	MetaTitle     string                       `json:"meta_title"`
	MetaDesc      string                       `json:"meta_description"`
	Language      string                       `json:"language"`
	ContentBlocks []models.ArticleContentBlock `json:"content_blocks" binding:"required"`
	CategoryIDs   []uint                       `json:"category_ids"`
	TagIDs        []uint                       `json:"tag_ids"`
}
