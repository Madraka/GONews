package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"news/internal/cache"
	"news/internal/database"
	"news/internal/models"
	"news/internal/pubsub"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoHandlerCached struct {
	db         *gorm.DB
	videoCache *cache.VideoCacheManager
}

func NewVideoHandlerCached() *VideoHandlerCached {
	return &VideoHandlerCached{
		db:         database.DB,
		videoCache: cache.GetVideoCacheManager(),
	}
}

// VoteVideoCached handles video voting with Redis cache optimization
// @Summary Vote on a video (Cache Optimized)
// @Description Like or dislike a video with real-time cache updates
// @Tags Videos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Param vote body models.VoteRequest true "Vote data (type: like/dislike)"
// @Success 200 {object} models.VoteResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/videos/{id}/vote-cached [post]
func (h *VideoHandlerCached) VoteVideoCached(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	var request models.VoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request format"})
		return
	}

	if request.Type != "like" && request.Type != "dislike" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Vote type must be 'like' or 'dislike'"})
		return
	}

	// Check if video exists
	var video models.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Video not found"})
		return
	}

	// 🚀 CACHE-FIRST APPROACH: Check cached user vote first
	cachedVote, cacheErr := h.videoCache.GetUserVideoVote(userID, uint(videoID))

	var existingVote models.VideoVote
	var dbVoteExists bool

	// If cache miss, check database
	if cacheErr != nil {
		err = h.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&existingVote).Error
		dbVoteExists = (err == nil)

		// Cache the existing vote for future requests
		if dbVoteExists {
			if err := h.videoCache.CacheUserVideoVote(userID, uint(videoID), existingVote.Type); err != nil {
				log.Printf("Warning: Failed to cache user video vote: %v", err)
			}
		}
	} else {
		// Cache hit - we know the user's vote type
		dbVoteExists = true
		existingVote.Type = cachedVote
		existingVote.UserID = userID
		existingVote.VideoID = uint(videoID)
	}

	var voteChange int64 = 0
	var oldVoteType string

	if dbVoteExists {
		oldVoteType = existingVote.Type
		if existingVote.Type == request.Type {
			// Same vote - remove it
			h.db.Delete(&existingVote)
			if err := h.videoCache.RemoveUserVideoVote(userID, uint(videoID)); err != nil {
				log.Printf("Warning: Failed to remove user video vote from cache: %v", err)
			}
			voteChange = -1 // Decrement count
		} else {
			// Different vote - update it
			existingVote.Type = request.Type
			h.db.Save(&existingVote)
			if err := h.videoCache.CacheUserVideoVote(userID, uint(videoID), request.Type); err != nil {
				log.Printf("Warning: Failed to cache updated user video vote: %v", err)
			}
			// No net change in total votes, but we need to update both counts
		}
	} else {
		// New vote
		vote := models.VideoVote{
			UserID:  userID,
			VideoID: uint(videoID),
			Type:    request.Type,
		}
		h.db.Create(&vote)
		if err := h.videoCache.CacheUserVideoVote(userID, uint(videoID), request.Type); err != nil {
			log.Printf("Warning: Failed to cache new user video vote: %v", err)
		}
		voteChange = 1 // Increment count
	}

	// 🚀 CACHE-FIRST COUNT RETRIEVAL: Try to get counts from cache first
	cachedCounts, cacheCountErr := h.videoCache.GetVideoVoteCounts(uint(videoID))

	var likes, dislikes int64

	if cacheCountErr != nil {
		// Cache miss - get from database and cache the results
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "like").Count(&likes)
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "dislike").Count(&dislikes)

		// Cache the counts for future requests
		if err := h.videoCache.CacheVideoVoteCounts(uint(videoID), likes, dislikes); err != nil {
			log.Printf("Warning: Failed to cache video vote counts: %v", err)
		}
	} else {
		// Cache hit - use cached counts and update them
		likes = cachedCounts.Likes
		dislikes = cachedCounts.Dislikes

		// Update cached counts based on vote changes
		if dbVoteExists && oldVoteType != request.Type {
			// Vote type changed - decrease old type, increase new type
			if oldVoteType == "like" {
				likes--
			} else if oldVoteType == "dislike" {
				dislikes--
			}

			if voteChange != -1 { // Only increment if not removing vote
				if request.Type == "like" {
					likes++
				} else if request.Type == "dislike" {
					dislikes++
				}
			}
		} else {
			// New vote or vote removal
			if request.Type == "like" {
				likes += voteChange
			} else if request.Type == "dislike" {
				dislikes += voteChange
			}
		}

		// Update cache with new counts
		if err := h.videoCache.CacheVideoVoteCounts(uint(videoID), likes, dislikes); err != nil {
			log.Printf("Warning: Failed to update cached video vote counts: %v", err)
		}
	}

	// Update cached counts in Video model for dashboard/listing performance
	h.db.Model(&models.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
		"like_count":    likes,
		"dislike_count": dislikes,
	})

	response := models.VoteResponse{
		Likes:    likes,
		Dislikes: dislikes,
	}

	// 📡 REAL-TIME NOTIFICATION: Publish vote notification via Redis pub/sub
	// Using video-specific notification for video votes
	videoNotification := pubsub.NotificationMessage{
		Type: "video_vote",
		Data: map[string]interface{}{
			"video_id":  videoID,
			"user_id":   userID,
			"vote_type": request.Type,
			"timestamp": time.Now(),
		},
	}

	videoChannel := fmt.Sprintf("video_votes:%d", videoID)
	if err := pubsub.PublishNotification(videoChannel, videoNotification); err != nil {
		log.Printf("❌ Failed to publish video vote notification: %v", err)
		// Don't fail the request, just log the error
	}

	c.JSON(http.StatusOK, response)
}

// GetVideoStatsCached retrieves video stats with cache optimization
// @Summary Get video statistics (Cache Optimized)
// @Description Get video like/dislike counts with cache
// @Tags Videos
// @Produce json
// @Param id path int true "Video ID"
// @Success 200 {object} models.VoteResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/videos/{id}/stats-cached [get]
func (h *VideoHandlerCached) GetVideoStatsCached(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	// Try cache first
	cachedCounts, cacheErr := h.videoCache.GetVideoVoteCounts(uint(videoID))

	var likes, dislikes int64

	if cacheErr != nil {
		// Cache miss - get from database
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "like").Count(&likes)
		h.db.Model(&models.VideoVote{}).Where("video_id = ? AND type = ?", videoID, "dislike").Count(&dislikes)

		// Cache for future requests
		if err := h.videoCache.CacheVideoVoteCounts(uint(videoID), likes, dislikes); err != nil {
			log.Printf("Warning: Failed to cache video vote counts: %v", err)
		}
	} else {
		// Cache hit
		likes = cachedCounts.Likes
		dislikes = cachedCounts.Dislikes
	}

	response := models.VoteResponse{
		Likes:    likes,
		Dislikes: dislikes,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserVideoVoteCached checks if user has voted on a video (with cache)
// @Summary Get user vote status (Cache Optimized)
// @Description Check if authenticated user has voted on a video
// @Tags Videos
// @Produce json
// @Security BearerAuth
// @Param id path int true "Video ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/videos/{id}/my-vote-cached [get]
func (h *VideoHandlerCached) GetUserVideoVoteCached(c *gin.Context) {
	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid video ID"})
		return
	}

	userID := getVideoUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Authentication required"})
		return
	}

	// Try cache first
	cachedVote, cacheErr := h.videoCache.GetUserVideoVote(userID, uint(videoID))

	var hasVoted bool
	var voteType string

	if cacheErr != nil {
		// Cache miss - check database
		var vote models.VideoVote
		err = h.db.Where("user_id = ? AND video_id = ?", userID, videoID).First(&vote).Error
		if err == nil {
			hasVoted = true
			voteType = vote.Type
			// Cache for future requests
			if err := h.videoCache.CacheUserVideoVote(userID, uint(videoID), voteType); err != nil {
				log.Printf("Warning: Failed to cache user video vote: %v", err)
			}
		}
	} else {
		// Cache hit
		hasVoted = true
		voteType = cachedVote
	}

	response := map[string]interface{}{
		"has_voted": hasVoted,
		"vote_type": voteType,
	}

	c.JSON(http.StatusOK, response)
}
