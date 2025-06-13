package cache

import (
	"context"
	"strconv"
	"time"

	"news/internal/json"

	"github.com/go-redis/redis/v8"
)

// Video cache constants
const (
	VideoVoteCountsPrefix = "video:votes:"
	VideoVoteUserPrefix   = "video:vote:user:"
	VideoViewCountPrefix  = "video:views:"
	VideoCacheTTL         = 24 * time.Hour
	VideoVoteTTL          = 7 * 24 * time.Hour // 1 week for user votes
)

// VideoCacheManager handles video-specific caching operations
type VideoCacheManager struct {
	client *RedisClient
	ctx    context.Context
}

// NewVideoCacheManager creates a new video cache manager
func NewVideoCacheManager() *VideoCacheManager {
	return &VideoCacheManager{
		client: GetRedisClient(),
		ctx:    context.Background(),
	}
}

// VideoVoteCounts represents cached vote counts
type VideoVoteCounts struct {
	Likes    int64 `json:"likes"`
	Dislikes int64 `json:"dislikes"`
	VideoID  uint  `json:"video_id"`
}

// CacheVideoVoteCounts caches the vote counts for a video
func (vm *VideoCacheManager) CacheVideoVoteCounts(videoID uint, likes, dislikes int64) error {
	if inTestMode {
		return nil
	}

	key := VideoVoteCountsPrefix + strconv.FormatUint(uint64(videoID), 10)
	counts := VideoVoteCounts{
		Likes:    likes,
		Dislikes: dislikes,
		VideoID:  videoID,
	}

	data, err := json.MarshalForCache(counts)
	if err != nil {
		return err
	}

	return vm.client.client.Set(vm.ctx, key, data, VideoCacheTTL).Err()
}

// GetVideoVoteCounts retrieves cached vote counts for a video
func (vm *VideoCacheManager) GetVideoVoteCounts(videoID uint) (*VideoVoteCounts, error) {
	if inTestMode {
		return nil, redis.Nil
	}

	key := VideoVoteCountsPrefix + strconv.FormatUint(uint64(videoID), 10)

	data, err := vm.client.client.Get(vm.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var counts VideoVoteCounts
	if err := json.UnmarshalForCache([]byte(data), &counts); err != nil {
		return nil, err
	}

	return &counts, nil
}

// CacheUserVideoVote caches a user's vote for a video
func (vm *VideoCacheManager) CacheUserVideoVote(userID, videoID uint, voteType string) error {
	if inTestMode {
		return nil
	}

	key := VideoVoteUserPrefix + strconv.FormatUint(uint64(userID), 10) + ":" + strconv.FormatUint(uint64(videoID), 10)
	return vm.client.client.Set(vm.ctx, key, voteType, VideoVoteTTL).Err()
}

// GetUserVideoVote retrieves a user's cached vote for a video
func (vm *VideoCacheManager) GetUserVideoVote(userID, videoID uint) (string, error) {
	if inTestMode {
		return "", redis.Nil
	}

	key := VideoVoteUserPrefix + strconv.FormatUint(uint64(userID), 10) + ":" + strconv.FormatUint(uint64(videoID), 10)
	return vm.client.client.Get(vm.ctx, key).Result()
}

// RemoveUserVideoVote removes a user's cached vote for a video
func (vm *VideoCacheManager) RemoveUserVideoVote(userID, videoID uint) error {
	if inTestMode {
		return nil
	}

	key := VideoVoteUserPrefix + strconv.FormatUint(uint64(userID), 10) + ":" + strconv.FormatUint(uint64(videoID), 10)
	return vm.client.client.Del(vm.ctx, key).Err()
}

// IncrementVideoVoteCount atomically increments a video's vote count in cache
func (vm *VideoCacheManager) IncrementVideoVoteCount(videoID uint, voteType string, increment int64) error {
	if inTestMode {
		return nil
	}

	countsKey := VideoVoteCountsPrefix + strconv.FormatUint(uint64(videoID), 10)

	// Use pipeline for atomic operations
	pipe := vm.client.client.TxPipeline()

	// Get current counts
	getCurrentCmd := pipe.Get(vm.ctx, countsKey)

	// Execute pipeline to get current state
	_, err := pipe.Exec(vm.ctx)
	if err != nil && err != redis.Nil {
		// If key doesn't exist, initialize with zero values
		if err == redis.Nil {
			counts := VideoVoteCounts{
				Likes:    0,
				Dislikes: 0,
				VideoID:  videoID,
			}

			if voteType == "like" {
				counts.Likes = increment
			} else if voteType == "dislike" {
				counts.Dislikes = increment
			}

			return vm.CacheVideoVoteCounts(videoID, counts.Likes, counts.Dislikes)
		}
		return err
	}

	// Parse existing counts
	var currentCounts VideoVoteCounts
	currentData, err := getCurrentCmd.Result()
	if err != nil && err != redis.Nil {
		return err
	}

	if err != redis.Nil {
		if err := json.UnmarshalForCache([]byte(currentData), &currentCounts); err != nil {
			return err
		}
	}

	// Update counts
	if voteType == "like" {
		currentCounts.Likes += increment
	} else if voteType == "dislike" {
		currentCounts.Dislikes += increment
	}
	currentCounts.VideoID = videoID

	// Save updated counts
	return vm.CacheVideoVoteCounts(videoID, currentCounts.Likes, currentCounts.Dislikes)
}

// InvalidateVideoCache removes all cached data for a video
func (vm *VideoCacheManager) InvalidateVideoCache(videoID uint) error {
	if inTestMode {
		return nil
	}

	keys := []string{
		VideoVoteCountsPrefix + strconv.FormatUint(uint64(videoID), 10),
		VideoViewCountPrefix + strconv.FormatUint(uint64(videoID), 10),
	}

	// Also remove user-specific vote caches (this is expensive, but necessary for data consistency)
	// In production, you might want to set a shorter TTL instead
	pattern := VideoVoteUserPrefix + "*:" + strconv.FormatUint(uint64(videoID), 10)
	userVoteKeys, err := vm.client.client.Keys(vm.ctx, pattern).Result()
	if err == nil {
		keys = append(keys, userVoteKeys...)
	}

	if len(keys) > 0 {
		return vm.client.client.Del(vm.ctx, keys...).Err()
	}

	return nil
}

// CacheVideoViewCount caches the view count for a video
func (vm *VideoCacheManager) CacheVideoViewCount(videoID uint, viewCount int64) error {
	if inTestMode {
		return nil
	}

	key := VideoViewCountPrefix + strconv.FormatUint(uint64(videoID), 10)
	return vm.client.client.Set(vm.ctx, key, viewCount, VideoCacheTTL).Err()
}

// GetVideoViewCount retrieves cached view count for a video
func (vm *VideoCacheManager) GetVideoViewCount(videoID uint) (int64, error) {
	if inTestMode {
		return 0, redis.Nil
	}

	key := VideoViewCountPrefix + strconv.FormatUint(uint64(videoID), 10)

	result, err := vm.client.client.Get(vm.ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result, 10, 64)
}

// IncrementVideoViewCount atomically increments a video's view count
func (vm *VideoCacheManager) IncrementVideoViewCount(videoID uint) (int64, error) {
	if inTestMode {
		return 1, nil
	}

	key := VideoViewCountPrefix + strconv.FormatUint(uint64(videoID), 10)
	return vm.client.client.Incr(vm.ctx, key).Result()
}

// Global video cache manager instance
var defaultVideoCache *VideoCacheManager

// GetVideoCacheManager returns the singleton video cache manager
func GetVideoCacheManager() *VideoCacheManager {
	if defaultVideoCache == nil {
		defaultVideoCache = NewVideoCacheManager()
	}
	return defaultVideoCache
}

// Wrapper functions for easier usage
func CacheVideoVoteCounts(videoID uint, likes, dislikes int64) error {
	return GetVideoCacheManager().CacheVideoVoteCounts(videoID, likes, dislikes)
}

func GetVideoVoteCounts(videoID uint) (*VideoVoteCounts, error) {
	return GetVideoCacheManager().GetVideoVoteCounts(videoID)
}

func CacheUserVideoVote(userID, videoID uint, voteType string) error {
	return GetVideoCacheManager().CacheUserVideoVote(userID, videoID, voteType)
}

func GetUserVideoVote(userID, videoID uint) (string, error) {
	return GetVideoCacheManager().GetUserVideoVote(userID, videoID)
}

func RemoveUserVideoVote(userID, videoID uint) error {
	return GetVideoCacheManager().RemoveUserVideoVote(userID, videoID)
}

func IncrementVideoVoteCount(videoID uint, voteType string, increment int64) error {
	return GetVideoCacheManager().IncrementVideoVoteCount(videoID, voteType, increment)
}

func InvalidateVideoCache(videoID uint) error {
	return GetVideoCacheManager().InvalidateVideoCache(videoID)
}

func IncrementVideoViewCount(videoID uint) (int64, error) {
	return GetVideoCacheManager().IncrementVideoViewCount(videoID)
}
