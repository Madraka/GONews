package handlers

import (
	"fmt"
	"net/http"
	"time"

	"news/internal/auth"
	"news/internal/cache"
	"news/internal/database"
	"news/internal/dto"
	"news/internal/middleware"
	"news/internal/models"
	"news/internal/validators"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterWithSecurity handles user registration with enhanced security
// @Summary Register a new user with enhanced security
// @Description Register a new user with comprehensive security validation
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body dto.UserRegistrationDTO true "User Registration Data"
// @Success 201 {object} dto.UserResponseDTO
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/register [post]
func RegisterWithSecurity(c *gin.Context) {
	var registrationDTO dto.UserRegistrationDTO
	if err := c.ShouldBindJSON(&registrationDTO); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request payload: " + err.Error()})
		return
	}

	// Use the password validator for enhanced security
	passwordValidator := validators.NewPasswordValidator()
	if err := passwordValidator.Validate(registrationDTO.Password); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Validate username length from the binding
	// Validate role using existing model validation
	user := models.User{
		Username:  registrationDTO.Username,
		Email:     registrationDTO.Email,
		Password:  registrationDTO.Password,
		FirstName: registrationDTO.FirstName,
		LastName:  registrationDTO.LastName,
		Role:      registrationDTO.Role,
	}

	// Validate role
	if !user.ValidateRole() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid role"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user: " + err.Error()})
		return
	}

	// Transform to response DTO without sensitive information
	response := dto.UserResponseDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}

	c.JSON(http.StatusCreated, response)
}

// LoginWithSecurity handles user login with enhanced security
// @Summary Login with enhanced security
// @Description Login with username and password to receive JWT tokens with enhanced security
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body dto.UserLoginDTO true "Login Credentials"
// @Success 200 {object} auth.TokenPair
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/login [post]
func LoginWithSecurity(c *gin.Context) {
	var loginDTO dto.UserLoginDTO
	if err := c.ShouldBindJSON(&loginDTO); err != nil {
		// Record failed login attempt due to invalid payload
		database.DB.Create(&models.LoginAttempt{
			Username:      loginDTO.Username,
			IP:            c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			Success:       false,
			FailureReason: "invalid request payload",
			Timestamp:     time.Now(),
		})
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request payload: " + err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", loginDTO.Username).First(&user).Error; err != nil {
		// Record failed login attempt due to unknown user
		database.DB.Create(&models.LoginAttempt{
			Username:      loginDTO.Username,
			IP:            c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			Success:       false,
			FailureReason: "invalid username or password",
			Timestamp:     time.Now(),
		})
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDTO.Password)); err != nil {
		// Record failed login attempt due to wrong password
		database.DB.Create(&models.LoginAttempt{
			Username:      loginDTO.Username,
			UserAgent:     c.GetHeader("User-Agent"),
			IP:            c.ClientIP(),
			Success:       false,
			FailureReason: "invalid username or password",
			Timestamp:     time.Now(),
		})
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	// Generate token pair using the token manager
	tokenManager := auth.NewTokenManager(
		[]byte(middleware.GetJWTSecret()),
		24*time.Hour,
		7*24*time.Hour,
		cache.GetRedisClient(),
	)

	// Generate token pair
	tokenPair, err := tokenManager.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token: " + err.Error()})
		return
	}

	accessToken := tokenPair.AccessToken
	refreshToken := tokenPair.RefreshToken

	// Record successful login attempt
	database.DB.Create(&models.LoginAttempt{
		Username:  loginDTO.Username,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Success:   true,
		Timestamp: time.Now(),
	})

	// Create user session record
	if err := CreateUserSession(user.ID, tokenPair.TokenID, c.ClientIP(), c.GetHeader("User-Agent"), tokenPair.ExpiresAt); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Warning: Failed to create user session: %v\n", err)
	}

	// Set secure cookie with refresh token
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(7*24*time.Hour.Seconds()), // 7 days
		"/",
		"",   // domain
		true, // secure
		true, // httpOnly
	)

	// Set CSRF token as cookie
	c.SetCookie(
		"csrf_token",
		tokenPair.CSRFToken,
		int(24*time.Hour.Seconds()), // 24 hours
		"/",
		"",    // domain
		true,  // secure
		false, // not httpOnly to allow JavaScript access
	)

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	database.DB.Save(&user)

	// Return a TokenResponse object to match what the tests expect
	c.JSON(http.StatusOK, models.TokenResponse{
		Token:     accessToken,
		CSRFToken: tokenPair.CSRFToken,
		ExpiresIn: int(24 * time.Hour.Seconds()),
		TokenType: "Bearer",
	})
}

// LogoutWithSecurity handles user logout with enhanced security
// @Summary Logout with enhanced security
// @Description Securely logout and invalidate tokens
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/logout [post]
func LogoutWithSecurity(c *gin.Context) {
	// Extract the token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "No token provided"})
		return
	}

	// Clean "Bearer " prefix if exists
	tokenString := token
	if len(token) > 7 && token[:7] == "Bearer " {
		tokenString = token[7:]
	}

	// Initialize token manager
	tokenManager := auth.NewTokenManager(
		[]byte(middleware.GetJWTSecret()),
		24*time.Hour,
		7*24*time.Hour,
		cache.GetRedisClient(),
	)

	// Validate and get claims
	claims, err := tokenManager.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid token"})
		return
	}

	// Blacklist the token by its ID using the cache Redis client
	var expirationTime time.Time
	if claims.ExpiresAt != nil {
		expirationTime = claims.ExpiresAt.Time
	} else {
		// Fallback to current time + 24 hours if ExpiresAt is nil
		expirationTime = time.Now().Add(24 * time.Hour)
	}

	// Use the cache Redis client which has proper Docker networking configuration
	duration := time.Until(expirationTime)
	if err := cache.GetRedisClient().BlacklistToken(claims.TokenID, duration); err != nil {
		fmt.Printf("Error blacklisting token: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to securely log out"})
		return
	}

	// Revoke the user session
	if err := RevokeUserSession(claims.TokenID); err != nil {
		fmt.Printf("Warning: Failed to revoke user session: %v\n", err)
		// Don't fail the logout for this
	}

	// Clear cookies
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.SetCookie("csrf_token", "", -1, "/", "", true, false)

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Successfully logged out",
	})
}

// RefreshToken handles token refresh with enhanced security
// @Summary Refresh authentication token
// @Description Use refresh token to get a new access token
// @Tags Auth
// @Produce json
// @Success 200 {object} auth.TokenPair
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Refresh token not found"})
		return
	}

	// Initialize token manager
	tokenManager := auth.NewTokenManager(
		[]byte(middleware.GetJWTSecret()),
		24*time.Hour,
		7*24*time.Hour,
		cache.GetRedisClient(),
	)

	// Validate refresh token
	claims, err := tokenManager.ValidateToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	// Check if token is blacklisted
	if cache.GetRedisClient().IsTokenBlacklisted(claims.TokenID) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Token has been revoked"})
		return
	}

	// Get user info
	var user models.User
	if err := database.DB.Where("username = ?", claims.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "User not found"})
		return
	}

	// Blacklist the old refresh token
	if err := cache.GetRedisClient().BlacklistToken(claims.TokenID, time.Until(claims.ExpiresAt.Time)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error during token refresh process"})
		return
	}

	// Generate new token pair
	tokenPair, err := tokenManager.GenerateTokenPair(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate new tokens"})
		return
	}

	// Set secure cookie with new refresh token
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(7*24*time.Hour.Seconds()),
		"/",
		"",
		true,
		true,
	)

	// Set new CSRF token
	c.SetCookie(
		"csrf_token",
		tokenPair.CSRFToken,
		int(24*time.Hour.Seconds()),
		"/",
		"",
		true,
		false,
	)

	c.JSON(http.StatusOK, tokenPair)
}
