package middleware

import (
	"net/http"
	"strings"

	"news/internal/metrics"

	"github.com/gin-gonic/gin"
)

// APIKeyTier defines the tier of an API key and its associated rate limits
type APIKeyTier struct {
	Name                   string
	RateLimit              int
	Burst                  int
	MaxRequestsPerMinute   int
	MaxRequestsPerHour     int
	MaxRequestsPerDay      int
	CacheExpirationMinutes int
	SpecialEndpoints       []string
}

// predefined API key tiers
var (
	BasicTier = APIKeyTier{
		Name:                   "basic",
		RateLimit:              1, // 1 request per second
		Burst:                  5, // Allow bursts up to 5 requests
		MaxRequestsPerMinute:   60,
		MaxRequestsPerHour:     1000,
		MaxRequestsPerDay:      10000,
		CacheExpirationMinutes: 30,
		SpecialEndpoints:       []string{},
	}

	ProTier = APIKeyTier{
		Name:                   "pro",
		RateLimit:              5,  // 5 requests per second
		Burst:                  15, // Allow bursts up to 15 requests
		MaxRequestsPerMinute:   300,
		MaxRequestsPerHour:     5000,
		MaxRequestsPerDay:      50000,
		CacheExpirationMinutes: 60,
		SpecialEndpoints:       []string{"/api/analytics"},
	}

	EnterpriseTier = APIKeyTier{
		Name:                   "enterprise",
		RateLimit:              20, // 20 requests per second
		Burst:                  50, // Allow bursts up to 50 requests
		MaxRequestsPerMinute:   1000,
		MaxRequestsPerHour:     20000,
		MaxRequestsPerDay:      200000,
		CacheExpirationMinutes: 120,
		SpecialEndpoints:       []string{"/api/analytics", "/api/export", "/api/bulk"},
	}

	// Map to look up API key tiers
	apiKeyTiers = make(map[string]APIKeyTier)
)

// InitAPIKeys initializes API key data
func InitAPIKeys() {
	// In a real application, this would likely be loaded from a database
	// or configuration file. Here we're using hardcoded values for demonstration.
	apiKeyTiers["api_key_basic_1234"] = BasicTier
	apiKeyTiers["api_key_pro_5678"] = ProTier
	apiKeyTiers["api_key_enterprise_9012"] = EnterpriseTier
}

// APIKeyAuth middleware verifies and enforces API key tiers
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		// For demo/development, allow some endpoints without API key
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" ||
			strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		tier, exists := apiKeyTiers[apiKey]
		if !exists {
			// Track invalid API key attempt
			metrics.RequestsTotal.WithLabelValues(c.FullPath(), c.Request.Method, "401").Inc()

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Store the tier in the context for use by other middleware (like rate limiting)
		c.Set("api_key_tier", tier.Name)
		c.Set("rate_limit", tier.RateLimit)
		c.Set("burst", tier.Burst)

		// Check if the endpoint is allowed for this tier
		endpoint := c.Request.URL.Path
		if isSpecialEndpoint(endpoint) && !containsEndpoint(tier.SpecialEndpoints, endpoint) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":        "This endpoint requires a higher API key tier",
				"current_tier": tier.Name,
			})
			c.Abort()
			return
		}

		// Add tier information to response headers
		c.Header("X-API-Tier", tier.Name)

		c.Next()
	}
}

// Helper function to check if an endpoint is special (requires higher tier)
func isSpecialEndpoint(endpoint string) bool {
	specialEndpoints := []string{"/api/analytics", "/api/export", "/api/bulk"}
	return containsEndpoint(specialEndpoints, endpoint)
}

// Helper function to check if a slice contains an endpoint
func containsEndpoint(endpoints []string, target string) bool {
	for _, endpoint := range endpoints {
		if strings.HasPrefix(target, endpoint) {
			return true
		}
	}
	return false
}
