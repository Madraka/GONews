package handlers

import (
	"log"
	"net/http"
	"news/internal/cache"
	"news/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCacheStats returns cache statistics for monitoring
// @Summary Get cache statistics
// @Description Retrieve detailed cache statistics for monitoring and debugging
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /api/cache/stats [get]
func GetCacheStats(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()

	// Get stats from the current primary cache system
	stats := cacheManager.GetCacheStats()
	efficiency := cacheManager.GetCacheEfficiency()

	c.JSON(http.StatusOK, gin.H{
		"cache_stats":      stats,
		"cache_efficiency": efficiency,
		"migration_status": map[string]interface{}{
			"optimized_primary":   !cacheManager.IsFallbackMode(),
			"fallback_mode":       cacheManager.IsFallbackMode(),
			"ready_for_migration": cacheManager.IsReadyForMigration(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "active",
	})
}

// GetCacheHealth returns cache health status
// @Summary Get cache health
// @Description Check the health status of all cache layers
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/cache/health [get]
func GetCacheHealth(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()
	health := cacheManager.GetCacheHealth()

	// Overall health status
	overallHealthy := health["overall_healthy"].(bool)

	c.JSON(http.StatusOK, gin.H{
		"overall_healthy": overallHealthy,
		"cache_layers":    health,
		"status": func() string {
			if overallHealthy {
				return "all_systems_operational"
			}
			return "degraded_performance"
		}(),
		"migration_info": map[string]interface{}{
			"primary_system":  health["primary_system"],
			"fallback_mode":   cacheManager.IsFallbackMode(),
			"migration_ready": cacheManager.IsReadyForMigration(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// ClearCache clears all cache layers (admin only)
// @Summary Clear cache
// @Description Clear all cache layers (L1 Ristretto + L2 Redis)
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/cache/clear [delete]
// @Security BearerAuth
func ClearCache(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()

	// Clear both cache systems via migration helper
	if err := cacheManager.ClearCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to clear cache",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
		"cleared": []string{"optimized_cache", "standard_cache"},
		"migration_info": map[string]interface{}{
			"primary_system": func() string {
				if cacheManager.IsFallbackMode() {
					return "standard_cache"
				}
				return "optimized_cache"
			}(),
			"fallback_mode": cacheManager.IsFallbackMode(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// WarmCache preloads cache with popular content (admin only)
// @Summary Warm cache
// @Description Preload cache with popular articles, categories, and tags
// @Tags Admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/cache/warm [post]
// @Security BearerAuth
func WarmCache(c *gin.Context) {
	// Method 1: Use migration helper preload (covers both cache systems)
	cacheManager := cache.GetMigrationCacheManager()
	if err := cacheManager.PreloadCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to preload cache via migration helper",
			"details": err.Error(),
		})
		return
	}

	// Method 2: Trigger actual data fetching to populate cache through normal operations
	// Warm popular articles
	_, _, err := services.GetArticlesWithPagination(0, 20, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to warm articles cache",
			"details": err.Error(),
		})
		return
	}

	// Warm categories
	_, err = services.GetCategoriesWithCache(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to warm categories cache",
			"details": err.Error(),
		})
		return
	}

	// Warm tags
	_, err = services.GetTagsWithCache("usage_count", 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to warm tags cache",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache warmed successfully",
		"warmed": []string{
			"articles_page_1_limit_20",
			"categories_hierarchical",
			"tags_popular_50",
		},
		"migration_info": map[string]interface{}{
			"primary_system": func() string {
				if cacheManager.IsFallbackMode() {
					return "standard_cache"
				}
				return "optimized_cache"
			}(),
			"fallback_mode":   cacheManager.IsFallbackMode(),
			"migration_ready": cacheManager.IsReadyForMigration(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetCacheAnalytics returns advanced cache performance analytics
// @Summary Get cache analytics
// @Description Retrieve advanced cache performance analytics and optimization recommendations
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/cache/analytics [get]
func GetCacheAnalytics(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()
	efficiency := cacheManager.GetCacheEfficiency()
	stats := cacheManager.GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"analytics": map[string]interface{}{
			"performance_metrics": efficiency,
			"detailed_stats":      stats,
			"health_score": func() string {
				if eff, ok := efficiency["overall_efficiency"].(string); ok {
					switch eff {
					case "excellent":
						return "A+ (95-100%)"
					case "good":
						return "A (85-94%)"
					case "moderate":
						return "B (70-84%)"
					default:
						return "C (<70%)"
					}
				}
				return "Unknown"
			}(),
			"optimization_suggestions": efficiency["recommendations"],
		},
		"migration_info": map[string]interface{}{
			"primary_system": func() string {
				if cacheManager.IsFallbackMode() {
					return "standard_cache"
				}
				return "optimized_cache"
			}(),
			"fallback_mode":   cacheManager.IsFallbackMode(),
			"migration_ready": cacheManager.IsReadyForMigration(),
			"performance_gains": func() string {
				if !cacheManager.IsFallbackMode() {
					return "50-80% improvement active"
				}
				return "Using fallback - standard performance"
			}(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "active",
	})
}

// PreloadCache triggers cache preloading for popular content
// @Summary Preload cache
// @Description Preload cache with popular content for better performance
// @Tags Monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/cache/preload [post]
func PreloadCache(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()

	if err := cacheManager.PreloadCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to preload cache",
			"details":   err.Error(),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache preloading completed successfully",
		"migration_info": map[string]interface{}{
			"primary_system": func() string {
				if cacheManager.IsFallbackMode() {
					return "standard_cache"
				}
				return "optimized_cache"
			}(),
			"fallback_mode":   cacheManager.IsFallbackMode(),
			"migration_ready": cacheManager.IsReadyForMigration(),
		},
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "success",
	})
}

// PublicCacheWarm provides public cache warming for development and testing
// @Summary Public cache warm
// @Description Warm cache with popular content (public endpoint for development)
// @Tags Cache
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/cache/warm [post]
func PublicCacheWarm(c *gin.Context) {
	cacheManager := cache.GetMigrationCacheManager()

	// Use optimized cache preload if available
	if !cacheManager.IsFallbackMode() {
		optimizedCache := cache.GetOptimizedUnifiedCache()
		if err := optimizedCache.PreloadPopularContent(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to preload optimized cache",
				"details": err.Error(),
			})
			return
		}
	}

	// Trigger normal API calls to populate cache
	go func() {
		// Make some API calls in background to warm cache
		if _, _, err := services.GetArticlesWithPagination(0, 20, ""); err != nil {
			log.Printf("Warning: Failed to warm articles cache: %v", err)
		}
		if _, err := services.GetCategoriesWithCache(true); err != nil {
			log.Printf("Warning: Failed to warm categories cache: %v", err)
		}
		if _, err := services.GetTagsWithCache("usage_count", 50); err != nil {
			log.Printf("Warning: Failed to warm tags cache: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache warming initiated successfully",
		"method":  "optimized_preload + background_api_calls",
		"cache_system": func() string {
			if cacheManager.IsFallbackMode() {
				return "standard_cache"
			}
			return "optimized_cache"
		}(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
