package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"news/internal/dto"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateContentBlock godoc
// @Summary Create a new content block
// @Description Add a new content block to an article
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param block body models.ArticleContentBlock true "Content block data"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks [post]
func CreateContentBlock(c *gin.Context) {
	articleID := c.Param("id")

	var block models.ArticleContentBlock
	if err := c.ShouldBindJSON(&block); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Validate required fields
	if block.BlockType == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Block type is required"})
		return
	}

	// Create the block
	createdBlock, err := services.AddContentBlock(articleID, block)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// GetContentBlocks godoc
// @Summary Get content blocks for an article
// @Description Retrieve all content blocks for a specific article
// @Tags Content Blocks
// @Produce json
// @Param id path int true "Article ID"
// @Param type query string false "Filter by block type"
// @Param visible query bool false "Filter by visibility (default: true)"
// @Success 200 {array} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks [get]
func GetContentBlocks(c *gin.Context) {
	articleID := c.Param("id")
	blockType := c.Query("type")
	visibleStr := c.DefaultQuery("visible", "true")

	log.Printf("DEBUG: [REALTIME UPDATE TEST] GetContentBlocks called - articleID: %s, blockType: %s, visible: %s", articleID, blockType, visibleStr)

	visible, err := strconv.ParseBool(visibleStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid visible parameter"})
		return
	}

	var blocks []models.ArticleContentBlock

	if blockType != "" {
		// Get blocks by type
		blocks, err = services.GetContentBlocksByType(articleID, blockType)
		log.Printf("DEBUG: GetContentBlocksByType returned %d blocks, err: %v", len(blocks), err)
	} else {
		// Get article with blocks
		article, err := services.GetArticleWithBlocks(articleID)
		if err != nil {
			log.Printf("DEBUG: GetArticleWithBlocks failed - err: %v", err)
			if err == services.ErrNotFound {
				c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
			} else {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			}
			return
		}

		log.Printf("DEBUG: GetArticleWithBlocks returned article with %d blocks", len(article.ContentBlocks))

		// Filter by visibility if needed
		if visible {
			for _, block := range article.ContentBlocks {
				if block.IsVisible {
					blocks = append(blocks, block)
				}
			}
		} else {
			blocks = article.ContentBlocks
		}

		log.Printf("DEBUG: After filtering, %d blocks remain", len(blocks))
	}

	if err != nil {
		log.Printf("DEBUG: Final error check - err: %v", err)
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	log.Printf("DEBUG: Returning %d blocks", len(blocks))
	c.JSON(http.StatusOK, blocks)
}

// UpdateContentBlock godoc
// @Summary Update a content block
// @Description Update an existing content block
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param block_id path int true "Content Block ID"
// @Param block body dto.UpdateContentBlockRequest true "Content block update data"
// @Success 200 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/blocks/{block_id} [put]
func UpdateContentBlock(c *gin.Context) {
	blockIDStr := c.Param("block_id")
	blockID, err := strconv.ParseUint(blockIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	var updateData dto.UpdateContentBlockRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Prepare update map
	updateMap := make(map[string]interface{})
	if updateData.Content != nil {
		updateMap["content"] = *updateData.Content
	}
	if updateData.Settings != nil {
		updateMap["settings"] = *updateData.Settings
	}
	if updateData.Position != nil {
		updateMap["position"] = *updateData.Position
	}
	if updateData.IsVisible != nil {
		updateMap["is_visible"] = *updateData.IsVisible
	}

	// Update the block
	updatedBlock, err := services.UpdateContentBlock(uint(blockID), updateMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedBlock)
}

// DeleteContentBlock godoc
// @Summary Delete a content block
// @Description Delete an existing content block
// @Tags Content Blocks
// @Security Bearer
// @Param block_id path int true "Content Block ID"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/blocks/{block_id} [delete]
func DeleteContentBlock(c *gin.Context) {
	blockIDStr := c.Param("block_id")
	blockID, err := strconv.ParseUint(blockIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid block ID"})
		return
	}

	if err := services.DeleteContentBlock(uint(blockID)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderContentBlocks godoc
// @Summary Reorder content blocks
// @Description Update the position of multiple content blocks
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param blocks body dto.ReorderBlocksRequest true "Block positions"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/reorder [post]
func ReorderContentBlocks(c *gin.Context) {
	articleID := c.Param("id")

	var request dto.ReorderBlocksRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if err := services.ReorderContentBlocks(articleID, request.BlockPositions); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Content blocks reordered successfully"})
}

// MigrateArticleToBlocks godoc
// @Summary Migrate article to blocks
// @Description Convert a legacy article to use content blocks
// @Tags Content Blocks
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 {object} models.Article
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/migrate-to-blocks [post]
func MigrateArticleToBlocks(c *gin.Context) {
	articleID := c.Param("id")

	migratedArticle, err := services.MigrateArticleToBlocks(articleID)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, migratedArticle)
}

// UpdateArticleBlocks godoc
// @Summary Update all content blocks for an article
// @Description Replace all content blocks for an article
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param blocks body dto.UpdateArticleBlocksRequest true "Content blocks data"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks [put]
func UpdateArticleBlocks(c *gin.Context) {
	articleID := c.Param("id")

	var request dto.UpdateArticleBlocksRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if err := services.UpdateArticleBlocks(articleID, request.Blocks); err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Article content blocks updated successfully"})
}

// DetectEmbeds godoc
// @Summary Detect embeddable URLs in content
// @Description Analyze content and suggest automatic embed blocks
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Param request body dto.DetectEmbedsRequest true "Content to analyze"
// @Success 200 {object} dto.DetectEmbedsResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/content-blocks/detect-embeds [post]
func DetectEmbeds(c *gin.Context) {
	var request dto.DetectEmbedsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if request.Content == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Content is required"})
		return
	}

	// Detect embeddable URLs
	servicesSuggestions := services.AutoDetectEmbeds(request.Content)

	// Convert services suggestions to DTO suggestions
	var suggestions []dto.EmbedSuggestion
	for _, suggestion := range servicesSuggestions {
		suggestions = append(suggestions, dto.EmbedSuggestion{
			URL:         suggestion.URL,
			EmbedType:   suggestion.EmbedType,
			Title:       suggestion.Title,
			Description: suggestion.Description,
			Settings:    suggestion.Settings,
			Preview:     suggestion.Preview,
		})
	}

	response := dto.DetectEmbedsResponse{
		Suggestions: suggestions,
		Count:       len(suggestions),
	}

	c.JSON(http.StatusOK, response)
}

// CreateEmbedFromURL godoc
// @Summary Create embed block from URL
// @Description Automatically create an embed block from a detected URL
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateEmbedRequest true "URL and embed settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/embed [post]
func CreateEmbedFromURL(c *gin.Context) {
	articleID := c.Param("article_id")

	var request dto.CreateEmbedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if request.URL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "URL is required"})
		return
	}

	// Create embed block from URL
	embedBlock, err := services.CreateEmbedFromURL(request.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Cannot create embed from URL: " + err.Error()})
		return
	}

	// Override title if provided
	if request.Title != "" {
		embedBlock.Content = request.Title
	}

	// Add the block to article
	createdBlock, err := services.AddContentBlock(articleID, *embedBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// AnalyzeURL godoc
// @Summary Analyze URL for embed compatibility
// @Description Check if a URL can be embedded and return embed information
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Param request body dto.AnalyzeURLRequest true "URL to analyze"
// @Success 200 {object} dto.AnalyzeURLResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/content-blocks/analyze-url [post]
func AnalyzeURL(c *gin.Context) {
	var request dto.AnalyzeURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	if request.URL == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "URL is required"})
		return
	}

	// Analyze URL
	detector := services.GetEmbedDetector()
	suggestion := detector.AnalyzeURL(request.URL)

	if suggestion == nil {
		response := dto.AnalyzeURLResponse{
			IsEmbeddable: false,
			URL:          request.URL,
			Message:      "URL is not embeddable or not supported",
		}
		c.JSON(http.StatusOK, response)
		return
	}

	response := dto.AnalyzeURLResponse{
		IsEmbeddable: true,
		URL:          request.URL,
		EmbedType:    suggestion.EmbedType,
		Title:        suggestion.Title,
		Description:  suggestion.Description,
		Settings:     suggestion.Settings,
		Preview:      suggestion.Preview,
		Message:      "URL is embeddable",
	}

	c.JSON(http.StatusOK, response)
}

// Advanced Block Creation Endpoints

// CreateChartBlock godoc
// @Summary Create a chart block
// @Description Create an interactive chart block with data visualization
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateChartRequest true "Chart data and settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/chart [post]
func CreateChartBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateChartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create chart block
	advancedService := services.GetAdvancedBlockService()
	chartBlock, err := advancedService.CreateChartBlock(uint(articleID), request.ChartData, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *chartBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateMapBlock godoc
// @Summary Create a map block
// @Description Create an interactive map block with markers
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateMapRequest true "Map coordinates and settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/map [post]
func CreateMapBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateMapRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create map block
	advancedService := services.GetAdvancedBlockService()
	mapBlock, err := advancedService.CreateMapBlock(uint(articleID), request.Latitude, request.Longitude, request.Markers, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *mapBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateFAQBlock godoc
// @Summary Create an FAQ block
// @Description Create a frequently asked questions block
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateFAQRequest true "FAQ items and settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/faq [post]
func CreateFAQBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateFAQRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create FAQ block
	advancedService := services.GetAdvancedBlockService()
	faqBlock, err := advancedService.CreateFAQBlock(uint(articleID), request.FAQItems, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *faqBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateNewsletterBlock godoc
// @Summary Create a newsletter signup block
// @Description Create a newsletter subscription block
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateNewsletterRequest true "Newsletter settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/newsletter [post]
func CreateNewsletterBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateNewsletterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create newsletter block
	advancedService := services.GetAdvancedBlockService()
	newsletterBlock, err := advancedService.CreateNewsletterBlock(uint(articleID), request.Title, request.Description, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *newsletterBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateQuizBlock godoc
// @Summary Create a quiz/poll block
// @Description Create an interactive quiz or poll block
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateQuizRequest true "Quiz data and settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/quiz [post]
func CreateQuizBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateQuizRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create quiz block
	advancedService := services.GetAdvancedBlockService()
	quizBlock, err := advancedService.CreateQuizBlock(uint(articleID), request.QuizType, request.Title, request.Questions, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *quizBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateCountdownBlock godoc
// @Summary Create a countdown timer block
// @Description Create a countdown timer block for events
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateCountdownRequest true "Countdown settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/countdown [post]
func CreateCountdownBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateCountdownRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Parse target date
	targetDate, err := time.Parse(time.RFC3339, request.TargetDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid target date format (use RFC3339)"})
		return
	}

	// Create countdown block
	advancedService := services.GetAdvancedBlockService()
	countdownBlock, err := advancedService.CreateCountdownBlock(uint(articleID), targetDate, request.Title, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *countdownBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateNewsTickerBlock godoc
// @Summary Create a news ticker block
// @Description Create a scrolling news ticker block
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateNewsTickerRequest true "News ticker settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/news-ticker [post]
func CreateNewsTickerBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateNewsTickerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create news ticker block
	advancedService := services.GetAdvancedBlockService()
	tickerBlock, err := advancedService.CreateNewsTickerBlock(uint(articleID), request.NewsSource, request.Category, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *tickerBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}

// CreateBreakingNewsBlock godoc
// @Summary Create a breaking news banner block
// @Description Create a breaking news alert banner
// @Tags Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateBreakingNewsRequest true "Breaking news settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/articles/{article_id}/blocks/breaking-news [post]
func CreateBreakingNewsBlock(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateBreakingNewsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create breaking news block
	advancedService := services.GetAdvancedBlockService()
	breakingNewsBlock, err := advancedService.CreateBreakingNewsBanner(uint(articleID), request.Content, request.AlertLevel, request.Position)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *breakingNewsBlock)
	if err != nil {
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, createdBlock)
}
