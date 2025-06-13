package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"news/internal/database"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
)

var (
	// JWT configuration
	jwtKey          = []byte(getEnvOrDefault("JWT_SECRET", ""))
	tokenDuration   = 24 * time.Hour
	refreshDuration = 7 * 24 * time.Hour

	// Redis client for token blacklist
	redisClient = redis.NewClient(&redis.Options{
		Addr: getEnvOrDefault("REDIS_URL", "localhost:6379"),
	})

	// Test mode blacklist (in-memory for tests)
	testBlacklist = make(map[string]bool)
	isTestMode    = false
)

const (
	AccessTokenExpiry  = time.Hour * 24     // 24 hours
	RefreshTokenExpiry = time.Hour * 24 * 7 // 7 days
)

// Claims represents the JWT claims structure
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"tid"` // For blacklisting
	jwt.RegisteredClaims
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		// In the unlikely event of an error, return a fallback string
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return id.String()
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(username, role string) (string, string, error) {
	// Generate unique token ID
	tokenID := GenerateUUID()

	// Access token
	accessToken, err := generateToken(username, role, tokenID, tokenDuration)
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshToken, err := generateToken(username, role, tokenID, refreshDuration)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateToken(username, role, tokenID string, duration time.Duration) (string, error) {
	if len(jwtKey) == 0 {
		return "", fmt.Errorf("JWT_SECRET is not set")
	}

	expirationTime := time.Now().Add(duration)
	claims := &Claims{
		Username: username,
		Role:     role,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Authenticate middleware with improved security
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ExtractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		// Make sure JWT secret is set
		if len(jwtKey) == 0 {
			// Try to load from environment
			secretKey := getEnvOrDefault("JWT_SECRET", "")
			if secretKey == "" {
				log.Println("ERROR: JWT_SECRET environment variable is not set")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication configuration error"})
				c.Abort()
				return
			}
			jwtKey = []byte(secretKey)
			log.Printf("Loaded JWT secret key from environment, length: %d", len(jwtKey))
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			log.Printf("JWT parse error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("Token not valid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		if IsTokenBlacklisted(claims.TokenID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Store the entire claims object and individual fields for backward compatibility
		c.Set("claims", claims)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("tokenID", claims.TokenID)

		// For backward compatibility with handlers that expect userID
		// Lookup user ID from database if needed
		if isTestMode {
			// In test mode, set a dummy user ID (1)
			c.Set("userID", uint(1))
			c.Set("user_id", uint(1)) // Also set with underscore for compatibility
		} else {
			// In production, get the real user ID from database by username
			var user models.User
			if err := database.DB.Where("username = ?", claims.Username).First(&user).Error; err == nil {
				c.Set("userID", user.ID)
				c.Set("user_id", user.ID) // Also set with underscore for compatibility
			}
		}

		c.Next()
	}
}

// ExtractToken extracts the JWT token from the request header
func ExtractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenID string, expirationTime time.Time) error {
	if isTestMode {
		// Use in-memory map in test mode
		testBlacklist[tokenID] = true
		return nil
	}

	duration := time.Until(expirationTime)
	return redisClient.Set(redisClient.Context(), fmt.Sprintf("blacklist:%s", tokenID), true, duration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func IsTokenBlacklisted(tokenID string) bool {
	if isTestMode {
		// Use in-memory map in test mode
		_, exists := testBlacklist[tokenID]
		return exists
	}

	exists, err := redisClient.Exists(redisClient.Context(), fmt.Sprintf("blacklist:%s", tokenID)).Result()
	return err == nil && exists > 0
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func Authorize(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return Authorize("Admin")
}

func EditorOnly() gin.HandlerFunc {
	return Authorize("Editor")
}

func AuthorOnly() gin.HandlerFunc {
	return Authorize("Author")
}

func UserOnly() gin.HandlerFunc {
	return Authorize("User")
}

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() string {
	if len(jwtKey) == 0 {
		return getEnvOrDefault("JWT_SECRET", "")
	}
	return string(jwtKey)
}

// SetTestMode enables or disables test mode for authentication middleware
func SetTestMode(enabled bool) {
	isTestMode = enabled
	if enabled {
		// Clear the test blacklist when enabling test mode
		testBlacklist = make(map[string]bool)
	}
}

// IsTestMode returns whether the middleware is in test mode
func IsTestMode() bool {
	return isTestMode
}
