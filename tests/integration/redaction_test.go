package integration

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"news/internal/cache"
	"news/internal/database"
	"news/internal/middleware"
	"news/internal/models"
	"news/internal/routes"
	"news/tests/testutil"
)

// RedactionIntegrationTestSuite defines the integration test suite for redaction functionality
type RedactionIntegrationTestSuite struct {
	suite.Suite
	server                   *testutil.TestServer
	testData                 *testutil.TestData
	testDB                   *testutil.TestDB
	originalRedactionSetting string
}

// SetupSuite runs before all tests in the suite
func (suite *RedactionIntegrationTestSuite) SetupSuite() {
	// Store original redaction setting
	suite.originalRedactionSetting = os.Getenv("NEWS_REDACTION_ENABLED")

	// Setup test mode
	testutil.SetupGinTestMode()

	// Initialize test data
	suite.testData = testutil.NewTestData()

	// Setup test database
	suite.testDB = testutil.SetupTestDB(suite.T())

	// Set the global database connection for the application
	database.DB = suite.testDB.DB

	// Enable test mode for middleware
	middleware.SetTestMode(true)

	// Initialize cache in test mode
	cache.SetTestMode(true)
	if err := cache.InitRedis(); err != nil {
		suite.T().Logf("Warning: Failed to initialize Redis: %v", err)
	}

	// Create and configure the router
	router := suite.setupRouter()
	suite.server = testutil.NewTestServer(router)
}

// setupRouter creates a router with all application routes
func (suite *RedactionIntegrationTestSuite) setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Register all application routes
	routes.RegisterRoutes(router)

	return router
}

// TearDownSuite runs after all tests in the suite
func (suite *RedactionIntegrationTestSuite) TearDownSuite() {
	// Restore original redaction setting
	if suite.originalRedactionSetting != "" {
		if err := os.Setenv("NEWS_REDACTION_ENABLED", suite.originalRedactionSetting); err != nil {
			suite.T().Logf("Warning: Failed to restore redaction setting: %v", err)
		}
	} else {
		if err := os.Unsetenv("NEWS_REDACTION_ENABLED"); err != nil {
			suite.T().Logf("Warning: Failed to unset redaction setting: %v", err)
		}
	}

	if suite.server != nil {
		suite.server.Close()
	}
	if suite.testDB != nil {
		if err := suite.testDB.Close(); err != nil {
			log.Printf("Warning: Failed to close test database: %v", err)
		}
	}
}

// SetupTest runs before each test
func (suite *RedactionIntegrationTestSuite) SetupTest() {
	// Clean database before each test
	suite.testDB.Cleanup(suite.T())
}

// createTestArticleWithSensitiveData creates a test article containing sensitive information
func (suite *RedactionIntegrationTestSuite) createTestArticleWithSensitiveData() *models.Article {
	// Create test user first
	user := &models.User{
		Username:  "testuser",
		Email:     "testuser@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
		Status:    "active",
	}
	err := database.DB.Create(user).Error
	require.NoError(suite.T(), err)

	// Create test category
	category := &models.Category{
		Name:        "Test Category",
		Slug:        "test-category",
		Description: "Test category description",
		IsActive:    true,
	}
	err = database.DB.Create(category).Error
	require.NoError(suite.T(), err)

	// Create article with sensitive data
	article := &models.Article{
		Title:   "Test Article with Sensitive Data",
		Slug:    "test-article-sensitive-data",
		Summary: "Article containing email addresses and phone numbers for testing redaction",
		Content: `# Test Article with Sensitive Data

This article contains various types of sensitive information that should be redacted:

## Contact Information
- Email: john.doe@company.com
- Phone: +1-555-123-4567
- Alternative email: support@example.org
- Mobile: (555) 987-6543

## Personal Information
- SSN: 123-45-6789
- Credit Card: 4532-1234-5678-9012

## More Contact Details
Contact us at info@testcompany.com or call 555.444.3333 for more information.
You can also reach our support team at help@support.net.

Phone numbers: 1-800-555-0199, (212) 555-7890, 555 123 4567
`,
		ContentType:   "legacy",
		AuthorID:      user.ID,
		Status:        "published",
		Language:      "en",
		AllowComments: true,
		MetaTitle:     "Test Article - Redaction Testing",
		MetaDesc:      "Test article for redaction functionality with emails like test@example.com",
	}
	err = database.DB.Create(article).Error
	require.NoError(suite.T(), err)

	// Associate with category
	err = database.DB.Model(article).Association("Categories").Append(category)
	require.NoError(suite.T(), err)

	return article
}

// TestRedaction_DisabledInDevelopment tests that redaction is disabled when NEWS_REDACTION_ENABLED=false
func (suite *RedactionIntegrationTestSuite) TestRedaction_DisabledInDevelopment() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Set redaction to disabled
	if err := os.Setenv("NEWS_REDACTION_ENABLED", "false"); err != nil {
		suite.T().Logf("Warning: Failed to set redaction environment variable: %v", err)
	}

	// Create test article with sensitive data
	article := suite.createTestArticleWithSensitiveData()

	// Test secure articles list endpoint
	resp := suite.server.GET(suite.T(), "/api/articles/secure?limit=1")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close secure articles response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &response); err != nil {
		suite.T().Errorf("Failed to parse secure articles response: %v", err)
	}

	// Check that redaction headers are NOT present (redaction disabled)
	assert.Empty(suite.T(), resp.Header.Get("X-Content-Redacted"))
	assert.Empty(suite.T(), resp.Header.Get("X-Redaction-Version"))

	// Check that sensitive data is still visible
	data := response["data"].([]interface{})
	if len(data) > 0 {
		firstArticle := data[0].(map[string]interface{})
		content := firstArticle["content"].(string)

		// Content should contain unredacted sensitive information
		assert.Contains(suite.T(), content, "john.doe@company.com")
		assert.Contains(suite.T(), content, "+1-555-123-4567")
	}

	// Test secure single article endpoint
	resp2 := suite.server.GET(suite.T(), fmt.Sprintf("/api/articles/%d/secure", article.ID))
	defer func() {
		if err := resp2.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close secure article response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp2.StatusCode)

	// Check headers
	assert.Empty(suite.T(), resp2.Header.Get("X-Content-Redacted"))
	assert.Empty(suite.T(), resp2.Header.Get("X-Redaction-Version"))
}

// TestRedaction_EnabledInProduction tests that redaction works when NEWS_REDACTION_ENABLED=true
func (suite *RedactionIntegrationTestSuite) TestRedaction_EnabledInProduction() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Set redaction to enabled
	if err := os.Setenv("NEWS_REDACTION_ENABLED", "true"); err != nil {
		suite.T().Logf("Warning: Failed to set redaction environment variable: %v", err)
	}

	// Create test article with sensitive data
	article := suite.createTestArticleWithSensitiveData()

	// Test secure articles list endpoint
	resp := suite.server.GET(suite.T(), "/api/articles/secure?limit=1")
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Check that redaction headers are present
	assert.Equal(suite.T(), "true", resp.Header.Get("X-Content-Redacted"))
	assert.Equal(suite.T(), "1.0", resp.Header.Get("X-Redaction-Version"))

	// Parse response
	var response map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &response); err != nil {
		suite.T().Errorf("Failed to parse redacted articles response: %v", err)
	}

	// Check that sensitive data is redacted
	data := response["data"].([]interface{})
	require.Greater(suite.T(), len(data), 0, "Should have at least one article")

	firstArticle := data[0].(map[string]interface{})
	content := firstArticle["content"].(string)

	// Email addresses should be redacted
	assert.NotContains(suite.T(), content, "john.doe@company.com")
	assert.NotContains(suite.T(), content, "support@example.org")
	assert.NotContains(suite.T(), content, "info@testcompany.com")
	assert.Contains(suite.T(), content, "[EMAIL REDACTED]")

	// Phone numbers should be redacted
	assert.NotContains(suite.T(), content, "+1-555-123-4567")
	assert.NotContains(suite.T(), content, "(555) 987-6543")
	assert.NotContains(suite.T(), content, "555.444.3333")
	assert.Contains(suite.T(), content, "[PHONE REDACTED]")

	// Author email should also be redacted
	author := firstArticle["author"].(map[string]interface{})
	authorEmail := author["email"].(string)
	assert.Equal(suite.T(), "[EMAIL REDACTED]", authorEmail)

	// Test secure single article endpoint
	resp2 := suite.server.GET(suite.T(), fmt.Sprintf("/api/articles/%d/secure", article.ID))
	defer func() {
		if err := resp2.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close secure article response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp2.StatusCode)

	// Check headers
	assert.Equal(suite.T(), "true", resp2.Header.Get("X-Content-Redacted"))
	assert.Equal(suite.T(), "1.0", resp2.Header.Get("X-Redaction-Version"))

	// Parse single article response
	var singleResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp2, &singleResponse); err != nil {
		suite.T().Errorf("Failed to parse single article response: %v", err)
	}

	singleContent := singleResponse["content"].(string)
	assert.Contains(suite.T(), singleContent, "[EMAIL REDACTED]")
	assert.Contains(suite.T(), singleContent, "[PHONE REDACTED]")
}

// TestRedaction_CompareRegularVsSecureEndpoints tests that regular endpoints don't redact while secure ones do
func (suite *RedactionIntegrationTestSuite) TestRedaction_CompareRegularVsSecureEndpoints() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Enable redaction
	if err := os.Setenv("NEWS_REDACTION_ENABLED", "true"); err != nil {
		suite.T().Logf("Warning: Failed to set redaction environment variable: %v", err)
	}

	// Create test article with sensitive data
	article := suite.createTestArticleWithSensitiveData()

	// Test regular articles endpoint (should NOT redact)
	resp1 := suite.server.GET(suite.T(), "/api/articles?limit=1")
	defer func() {
		if err := resp1.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close regular articles response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp1.StatusCode)

	// Should NOT have redaction headers
	assert.Empty(suite.T(), resp1.Header.Get("X-Content-Redacted"))
	assert.Empty(suite.T(), resp1.Header.Get("X-Redaction-Version"))

	var regularResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp1, &regularResponse); err != nil {
		suite.T().Errorf("Failed to parse regular articles response: %v", err)
	}

	// Test secure articles endpoint (should redact)
	resp2 := suite.server.GET(suite.T(), "/api/articles/secure?limit=1")
	defer func() {
		if err := resp2.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close secure articles response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp2.StatusCode)

	// Should have redaction headers
	assert.Equal(suite.T(), "true", resp2.Header.Get("X-Content-Redacted"))
	assert.Equal(suite.T(), "1.0", resp2.Header.Get("X-Redaction-Version"))

	var secureResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp2, &secureResponse); err != nil {
		suite.T().Errorf("Failed to parse secure articles response: %v", err)
	}

	// Compare responses
	regularData := regularResponse["data"].([]interface{})
	secureData := secureResponse["data"].([]interface{})

	if len(regularData) > 0 && len(secureData) > 0 {
		regularArticle := regularData[0].(map[string]interface{})
		secureArticle := secureData[0].(map[string]interface{})

		regularContent := regularArticle["content"].(string)
		secureContent := secureArticle["content"].(string)

		// Regular content should contain sensitive data
		assert.Contains(suite.T(), regularContent, "john.doe@company.com")

		// Secure content should have redacted data
		assert.NotContains(suite.T(), secureContent, "john.doe@company.com")
		assert.Contains(suite.T(), secureContent, "[EMAIL REDACTED]")
	}

	// Test individual article endpoints
	resp3 := suite.server.GET(suite.T(), fmt.Sprintf("/api/articles/%d", article.ID))
	defer func() {
		if err := resp3.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close individual article response body: %v", err)
		}
	}()

	resp4 := suite.server.GET(suite.T(), fmt.Sprintf("/api/articles/%d/secure", article.ID))
	defer func() {
		if err := resp4.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close secure individual article response body: %v", err)
		}
	}()

	assert.Equal(suite.T(), http.StatusOK, resp3.StatusCode)
	assert.Equal(suite.T(), http.StatusOK, resp4.StatusCode)

	// Regular should not have redaction headers
	assert.Empty(suite.T(), resp3.Header.Get("X-Content-Redacted"))

	// Secure should have redaction headers
	assert.Equal(suite.T(), "true", resp4.Header.Get("X-Content-Redacted"))
}

// TestRedaction_OnlyTargetedContent tests that only sensitive content is redacted, not everything
func (suite *RedactionIntegrationTestSuite) TestRedaction_OnlyTargetedContent() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Enable redaction
	if err := os.Setenv("NEWS_REDACTION_ENABLED", "true"); err != nil {
		suite.T().Logf("Warning: Failed to set redaction environment variable: %v", err)
	}

	// Create test article with sensitive data
	article := suite.createTestArticleWithSensitiveData()

	// Test secure single article endpoint
	resp := suite.server.GET(suite.T(), fmt.Sprintf("/api/articles/%d/secure", article.ID))
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	testutil.ParseJSONResponse(suite.T(), resp, &response)

	content := response["content"].(string)

	// Regular text should remain unchanged
	assert.Contains(suite.T(), content, "Test Article with Sensitive Data")
	assert.Contains(suite.T(), content, "Contact Information")
	assert.Contains(suite.T(), content, "Personal Information")

	// Only sensitive patterns should be redacted
	assert.Contains(suite.T(), content, "[EMAIL REDACTED]")
	assert.Contains(suite.T(), content, "[PHONE REDACTED]")

	// Non-sensitive content should remain
	assert.Contains(suite.T(), content, "This article contains")
	assert.Contains(suite.T(), content, "more information")
}

// TestSuite runs the redaction integration test suite
func TestRedactionIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RedactionIntegrationTestSuite))
}
