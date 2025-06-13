package handlers

import (
	"net/http"
	"strconv"

	"news/internal/dto"
	"news/internal/models"
	"news/internal/services"

	"github.com/gin-gonic/gin"
)

// CreateSocialFeedBlock godoc
// @Summary Create a social media feed block
// @Description Create a dynamic social media feed block
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param request body dto.CreateSocialFeedRequest true "Social feed block settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/social-feed [post]

// Advanced Content Blocks Handlers
// These handlers create specialized content blocks with complex functionality

// CreateSocialFeedBlock godoc
// @Summary Create a social media feed block
// @Description Create a live social media feed block
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param request body dto.CreateSocialFeedRequest true "Social feed settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/social-feed [post]
func CreateSocialFeedBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateSocialFeedRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Validate platform
	allowedPlatforms := map[string]bool{
		"twitter": true, "instagram": true, "linkedin": true, "facebook": true,
	}
	if !allowedPlatforms[request.Platform] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid platform: " + request.Platform})
		return
	}

	// Create social feed block
	advancedService := services.GetAdvancedBlockService()
	socialFeedBlock, err := advancedService.CreateSocialFeedBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *socialFeedBlock)
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

// CreateHeroBlock godoc
// @Summary Create a hero section block
// @Description Create a hero banner with background and CTA buttons
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param request body dto.CreateHeroRequest true "Hero section settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/hero [post]
func CreateHeroBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateHeroRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create hero block
	advancedService := services.GetAdvancedBlockService()
	heroBlock, err := advancedService.CreateHeroBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *heroBlock)
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

// CreateCardGridBlock godoc
// @Summary Create a card grid block
// @Description Create a grid layout with cards
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param request body dto.CreateCardGridRequest true "Card grid settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/card-grid [post]
func CreateCardGridBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateCardGridRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Validate cards
	if len(request.Cards) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "At least one card is required"})
		return
	}

	// Create card grid block
	advancedService := services.GetAdvancedBlockService()
	cardGridBlock, err := advancedService.CreateCardGridBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *cardGridBlock)
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

// CreateSearchBlock godoc
// @Summary Create a search block
// @Description Create a search interface block
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateSearchRequest true "Search block settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/search [post]
func CreateSearchBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateSearchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create search block
	advancedService := services.GetAdvancedBlockService()
	searchBlock, err := advancedService.CreateSearchBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *searchBlock)
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

// CreateCommentsBlock godoc
// @Summary Create a comments block
// @Description Create a comments section block
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param request body dto.CreateCommentsRequest true "Comments block settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/comments [post]
func CreateCommentsBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateCommentsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create comments block
	advancedService := services.GetAdvancedBlockService()
	commentsBlock, err := advancedService.CreateCommentsBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *commentsBlock)
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

// CreateRatingBlock godoc
// @Summary Create a rating/review block
// @Description Create a rating and review block
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateRatingRequest true "Rating block settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/rating [post]
func CreateRatingBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateRatingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create rating block
	advancedService := services.GetAdvancedBlockService()
	ratingBlock, err := advancedService.CreateRatingBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *ratingBlock)
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

// CreateProductBlock godoc
// @Summary Create a product showcase block
// @Description Create a product display block for e-commerce
// @Tags Advanced Content Blocks
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param request body dto.CreateProductRequest true "Product block settings"
// @Success 201 {object} models.ArticleContentBlock
// @Failure 400 {object} models.ErrorResponse
// @Router /api/articles/{id}/blocks/product [post]
func CreateProductBlock(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	var request dto.CreateProductRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format: " + err.Error()})
		return
	}

	// Create product block
	advancedService := services.GetAdvancedBlockService()
	productBlock, err := advancedService.CreateProductBlock(uint(articleID), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Add the block to the article
	createdBlock, err := services.AddContentBlock(articleIDStr, *productBlock)
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
