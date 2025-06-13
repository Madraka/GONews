package handlers

import (
	"net/http"
	"strconv"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
)

// BookmarkArticle godoc
// @Summary Bookmark an article
// @Description Add or remove an article from user's bookmarks
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 {object} BookmarkResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /articles/{id}/bookmark [post]
func BookmarkArticle(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Verify article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Check if already bookmarked
	var existingBookmark models.Bookmark
	err = database.DB.Where("user_id = ? AND article_id = ?", userID, articleID).First(&existingBookmark).Error

	var isBookmarked bool
	if err == nil {
		// Already bookmarked - remove it
		database.DB.Delete(&existingBookmark)
		isBookmarked = false
	} else {
		// Not bookmarked - add it
		bookmark := models.Bookmark{
			UserID:    userID.(uint),
			ArticleID: uint(articleID),
		}
		database.DB.Create(&bookmark)
		isBookmarked = true
	}

	// Get total bookmark count for this article
	var count int64
	database.DB.Model(&models.Bookmark{}).Where("article_id = ?", articleID).Count(&count)

	response := BookmarkResponse{
		IsBookmarked: isBookmarked,
		TotalCount:   int(count),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserBookmarks godoc
// @Summary Get user's bookmarks
// @Description Retrieve all bookmarked articles for the authenticated user
// @Tags Bookmarks
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /user/bookmarks [get]
func GetUserBookmarksPage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get total count
	var total int64
	database.DB.Model(&models.Bookmark{}).Where("user_id = ?", userID).Count(&total)

	// Get bookmarks with article details
	var bookmarks []models.Bookmark
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Article").
		Preload("Article.Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch bookmarks"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       bookmarks,
		TotalItems: int(total), // Renamed from Total
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1, // Renamed from HasPrevious
	}

	c.JSON(http.StatusOK, response)
}

// VoteArticle godoc
// @Summary Vote on an article
// @Description Like or dislike an article
// @Tags Votes
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param vote body VoteRequest true "Vote data"
// @Success 200 {object} VoteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /articles/{id}/vote [post]
func VoteArticle(c *gin.Context) {
	articleIDStr := c.Param("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid article ID"})
		return
	}

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

	// Verify article exists
	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Article not found"})
		return
	}

	// Check if user already voted
	var existingVote models.Vote
	err = database.DB.Where("user_id = ? AND article_id = ?", userID, articleID).First(&existingVote).Error

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
			ArticleID: &[]uint{uint(articleID)}[0],
			Type:      request.Type,
		}
		database.DB.Create(&vote)
	}

	// Get vote counts
	var likes, dislikes int64
	database.DB.Model(&models.Vote{}).Where("article_id = ? AND type = ?", articleID, "like").Count(&likes)
	database.DB.Model(&models.Vote{}).Where("article_id = ? AND type = ?", articleID, "dislike").Count(&dislikes)

	response := VoteResponse{
		Likes:    int(likes),
		Dislikes: int(dislikes),
	}

	c.JSON(http.StatusOK, response)
}

// FollowUser godoc
// @Summary Follow a user
// @Description Follow or unfollow another user
// @Tags Social
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path int true "User ID to follow"
// @Success 200 {object} FollowResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{user_id}/follow [post]
func FollowUser(c *gin.Context) {
	targetUserIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not authenticated"})
		return
	}

	// Can't follow yourself
	if uint(targetUserID) == userID.(uint) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "You cannot follow yourself"})
		return
	}

	// Verify target user exists
	var targetUser models.User
	if err := database.DB.First(&targetUser, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	// Check if already following
	var existingFollow models.Follow
	err = database.DB.Where("follower_id = ? AND following_id = ?", userID, targetUserID).First(&existingFollow).Error

	var isFollowing bool
	if err == nil {
		// Already following - unfollow
		database.DB.Delete(&existingFollow)
		isFollowing = false
	} else {
		// Not following - follow
		follow := models.Follow{
			FollowerID:  userID.(uint),
			FollowingID: uint(targetUserID),
		}
		database.DB.Create(&follow)
		isFollowing = true
	}

	// Get follower count for target user
	var followerCount int64
	database.DB.Model(&models.Follow{}).Where("following_id = ?", targetUserID).Count(&followerCount)

	response := FollowResponse{
		IsFollowing:   isFollowing,
		FollowerCount: int(followerCount),
	}

	c.JSON(http.StatusOK, response)
}

// GetUserFollowers godoc
// @Summary Get user's followers
// @Description Get list of users following the specified user
// @Tags Social
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{user_id}/followers [get]
func GetUserFollowers(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Verify user exists
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	// Get total count
	var total int64
	database.DB.Model(&models.Follow{}).Where("following_id = ?", userID).Count(&total)

	// Get followers
	var follows []models.Follow
	if err := database.DB.Where("following_id = ?", userID).
		Preload("Follower").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch followers"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       follows,
		TotalItems: int(total), // Renamed from Total
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1, // Renamed from HasPrevious
	}

	c.JSON(http.StatusOK, response)
}

// GetUserFollowing godoc
// @Summary Get users followed by user
// @Description Get list of users that the specified user is following
// @Tags Social
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.PaginatedResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /users/{user_id}/following [get]
func GetUserFollowing(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Verify user exists
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	// Get total count
	var total int64
	database.DB.Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&total)

	// Get following
	var follows []models.Follow
	if err := database.DB.Where("follower_id = ?", userID).
		Preload("Following").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&follows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch following"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := models.PaginatedResponse{
		Data:       follows,
		TotalItems: int(total), // Renamed from Total
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1, // Renamed from HasPrevious
	}

	c.JSON(http.StatusOK, response)
}

// Response structures
type BookmarkResponse struct {
	IsBookmarked bool `json:"is_bookmarked"`
	TotalCount   int  `json:"total_count"`
}

type FollowResponse struct {
	IsFollowing   bool `json:"is_following"`
	FollowerCount int  `json:"follower_count"`
}
