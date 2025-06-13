package handlers

import (
	"log"
	"net/http"
	"strconv"

	"news/internal/database"
	"news/internal/models"
	"news/internal/pubsub"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetComments godoc
// @Summary Get comments for an article
// @Description Retrieve comments for a specific article with threading support
// @Tags Comments
// @Produce json
// @Param article_id path int true "Article ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Comments per page" default(20)
// @Param sort query string false "Sort by: newest, oldest, likes" default(newest)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /articles/{article_id}/comments [get]
func GetComments(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "newest")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Verify article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Build query
	query := database.DB.Where("article_id = ? AND status = ? AND parent_id IS NULL", articleID, "approved").
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", "approved").Order("created_at ASC").Preload("User")
		})

	// Apply sorting
	switch sort {
	case "oldest":
		query = query.Order("created_at ASC")
	case "likes":
		// This would require a vote count subquery - simplified for now
		query = query.Order("created_at DESC")
	default: // newest
		query = query.Order("created_at DESC")
	}

	// Get total count
	var total int64
	database.DB.Model(&models.Comment{}).Where("article_id = ? AND status = ? AND parent_id IS NULL", articleID, "approved").Count(&total)

	// Get comments with pagination
	var comments []models.Comment
	if err := query.Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch comments"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       comments,
		TotalItems: int(total), // Renamed from Total
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1, // Renamed from HasPrevious
	}

	c.JSON(http.StatusOK, response)
}

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a new comment on an article (requires authentication)
// @Tags Comments
// @Accept json
// @Produce json
// @Security Bearer
// @Param article_id path int true "Article ID"
// @Param comment body CreateCommentRequest true "Comment data"
// @Success 201 {object} models.Comment
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /articles/{article_id}/comments [post]
func CreateComment(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	// Get user from JWT
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request CreateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate content
	if len(request.Content) < 5 || len(request.Content) > 1000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Comment must be between 5-1000 characters"})
		return
	}

	// Verify article exists and allows comments
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	if !article.AllowComments {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Comments are disabled for this article"})
		return
	}

	// Verify parent comment if provided
	if request.ParentID != nil {
		var parentComment models.Comment
		if err := database.DB.Where("id = ? AND article_id = ?", *request.ParentID, articleID).First(&parentComment).Error; err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Parent comment not found"})
			return
		}
	}

	// Create comment
	comment := models.Comment{
		ArticleID: uint(articleID),
		UserID:    userID.(uint),
		ParentID:  request.ParentID,
		Content:   request.Content,
		Status:    "approved", // Can be changed to "pending" for moderation
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create comment"})
		return
	}

	// Load user relationship for response
	database.DB.Preload("User").First(&comment, comment.ID)

	// Publish real-time comment notification
	if err := pubsub.PublishCommentNotification(uint(articleID), comment); err != nil {
		log.Printf("Failed to publish comment notification: %v", err)
	}

	c.JSON(http.StatusCreated, comment)
}

// UpdateComment godoc
// @Summary Update a comment
// @Description Update an existing comment (only by comment author)
// @Tags Comments
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Comment ID"
// @Param comment body UpdateCommentRequest true "Comment data"
// @Success 200 {object} models.Comment
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /comments/{id} [put]
func UpdateComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Comment not found"})
		return
	}

	// Check if user owns the comment
	if comment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "You can only edit your own comments"})
		return
	}

	var request UpdateCommentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	// Validate content
	if len(request.Content) < 5 || len(request.Content) > 1000 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Comment must be between 5-1000 characters"})
		return
	}

	// Update comment
	comment.Content = request.Content
	comment.IsEdited = true

	if err := database.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update comment"})
		return
	}

	// Load user relationship for response
	database.DB.Preload("User").First(&comment, comment.ID)

	c.JSON(http.StatusOK, comment)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment (only by comment author or admin)
// @Tags Comments
// @Produce json
// @Security Bearer
// @Param id path int true "Comment ID"
// @Success 204
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /comments/{id} [delete]
func DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	userRole, _ := c.Get("user_role")

	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Comment not found"})
		return
	}

	// Check permissions
	if comment.UserID != userID.(uint) && userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "You can only delete your own comments"})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// VoteComment godoc
// @Summary Vote on a comment
// @Description Like or dislike a comment
// @Tags Comments
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Comment ID"
// @Param vote body VoteRequest true "Vote data"
// @Success 200 {object} VoteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /comments/{id}/vote [post]
func VoteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	var request VoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	if request.Type != "like" && request.Type != "dislike" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Vote type must be 'like' or 'dislike'"})
		return
	}

	// Verify comment exists
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Comment not found"})
		return
	}

	commentIDUint, _ := strconv.ParseUint(commentID, 10, 32)

	// Check if user already voted
	var existingVote models.Vote
	err := database.DB.Where("user_id = ? AND comment_id = ?", userID, commentIDUint).First(&existingVote).Error

	if err == nil {
		// User already voted
		if existingVote.Type == request.Type {
			// Same vote - remove it
			database.DB.Delete(&existingVote)
		} else {
			// Different vote - update it
			existingVote.Type = request.Type
			database.DB.Save(&existingVote)
		}
	} else {
		// New vote
		vote := models.Vote{
			UserID:    userID.(uint),
			CommentID: &[]uint{uint(commentIDUint)}[0],
			Type:      request.Type,
		}
		database.DB.Create(&vote)
	}

	// Get vote counts
	var likes, dislikes int64
	database.DB.Model(&models.Vote{}).Where("comment_id = ? AND type = ?", commentIDUint, "like").Count(&likes)
	database.DB.Model(&models.Vote{}).Where("comment_id = ? AND type = ?", commentIDUint, "dislike").Count(&dislikes)

	response := VoteResponse{
		Likes:    int(likes),
		Dislikes: int(dislikes),
	}

	c.JSON(http.StatusOK, response)
}

// Request/Response structures
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type VoteRequest struct {
	Type string `json:"type" binding:"required"`
}

type VoteResponse struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}
