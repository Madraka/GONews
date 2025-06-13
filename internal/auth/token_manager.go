package auth

import (
	"errors"
	"time"

	"news/internal/cache"
	"news/internal/models"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
)

// TokenManager handles token generation, validation, and management
type TokenManager struct {
	JWTSecret       []byte
	AccessDuration  time.Duration
	RefreshDuration time.Duration
	RedisClient     *cache.RedisClient
}

// NewTokenManager creates a new token manager
func NewTokenManager(jwtSecret []byte, accessDuration, refreshDuration time.Duration, redisClient *cache.RedisClient) *TokenManager {
	return &TokenManager{
		JWTSecret:       jwtSecret,
		AccessDuration:  accessDuration,
		RefreshDuration: refreshDuration,
		RedisClient:     redisClient,
	}
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	CSRFToken    string `json:"csrf_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	TokenID      string `json:"-"` // For internal use only
	ExpiresAt    int64  `json:"-"` // For internal use only
}

// Claims represents the JWT claims structure
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"tid"` // For blacklisting
	jwt.RegisteredClaims
}

// GenerateTokenPair generates both access and refresh tokens
func (tm *TokenManager) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Generate unique token ID
	tokenID := generateUUID()

	// Access token
	accessToken, err := tm.generateToken(user.Username, user.Role, tokenID, tm.AccessDuration)
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshToken, err := tm.generateToken(user.Username, user.Role, tokenID, tm.RefreshDuration)
	if err != nil {
		return nil, err
	}

	// Generate CSRF token for added security
	csrfToken := generateUUID()

	// Calculate expiry time
	expiresAt := time.Now().Add(tm.AccessDuration).Unix()

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CSRFToken:    csrfToken,
		ExpiresIn:    int(tm.AccessDuration.Seconds()),
		TokenType:    "Bearer",
		TokenID:      tokenID,
		ExpiresAt:    expiresAt,
	}, nil
}

// generateToken creates a signed JWT token
func (tm *TokenManager) generateToken(username, role, tokenID string, duration time.Duration) (string, error) {
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
	return token.SignedString(tm.JWTSecret)
}

// ValidateToken validates a JWT token and returns the claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return tm.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is blacklisted
	if tm.IsTokenBlacklisted(claims.TokenID) {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// BlacklistToken adds a token to the blacklist
func (tm *TokenManager) BlacklistToken(tokenID string, expirationTime time.Time) error {
	duration := time.Until(expirationTime)
	// Use cache RedisClient to store blacklisted token
	return tm.RedisClient.BlacklistToken(tokenID, duration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (tm *TokenManager) IsTokenBlacklisted(tokenID string) bool {
	// Use cache RedisClient to check for blacklisted token
	return tm.RedisClient.IsTokenBlacklisted(tokenID)
}

// RefreshTokens generates new access and refresh tokens using a valid refresh token
func (tm *TokenManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := tm.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user from database to make sure they still exist and permissions haven't changed
	// This would require passing a database reference or using a repository pattern
	// For simplicity, we'll just create a new token pair with the existing claims
	user := &models.User{
		Username: claims.Username,
		Role:     claims.Role,
	}

	// Blacklist the old token ID to prevent reuse
	err = tm.BlacklistToken(claims.TokenID, claims.ExpiresAt.Time)
	if err != nil {
		// Log error but continue - blacklisting failure shouldn't block token refresh
	}

	// Generate new token pair
	return tm.GenerateTokenPair(user)
}

// generateUUID generates a new UUID string
func generateUUID() string {
	id, err := uuid.NewV4()
	if err != nil {
		// In the unlikely event of an error, return a fallback string
		return "fallback-" + time.Now().Format(time.RFC3339Nano)
	}
	return id.String()
}
