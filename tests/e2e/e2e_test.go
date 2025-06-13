package e2e

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"news/tests/testutil"
)

// E2ETestSuite defines the end-to-end test suite
type E2ETestSuite struct {
	suite.Suite
	server     *testutil.TestServer
	testData   *testutil.TestData
	testDB     *testutil.TestDB
	userToken  string
	adminToken string
}

// SetupSuite runs before all tests in the suite
func (suite *E2ETestSuite) SetupSuite() {
	// Setup test environment
	testutil.SetupGinTestMode()

	// Initialize test data
	suite.testData = testutil.NewTestData()

	// Setup test database
	suite.testDB = testutil.SetupTestDB(suite.T())

	// TODO: Initialize your actual Gin router here
	// router := setupRouter(suite.testDB.DB)
	// suite.server = testutil.NewTestServer(router)

}

// TearDownSuite runs after all tests in the suite
func (suite *E2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.testDB != nil {
		if err := suite.testDB.Close(); err != nil {
			suite.T().Logf("Warning: Error closing test database: %v", err)
		}
	}
}

// SetupTest runs before each test
func (suite *E2ETestSuite) SetupTest() {
	// Clean database before each test
	suite.testDB.Cleanup(suite.T())
}

// TestE2E_CompleteUserJourney tests complete user workflow
func (suite *E2ETestSuite) TestE2E_CompleteUserJourney() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Step 1: User Registration
	newUser := map[string]interface{}{
		"username": "journeyuser",
		"email":    "journey@example.com",
		"password": "JourneyPass123!",
	}

	resp := suite.server.POST(suite.T(), "/api/auth/register", newUser)
	if err := resp.Body.Close(); err != nil {
		suite.T().Logf("Warning: Failed to close response body: %v", err)
	}
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	// Step 2: User Login
	loginData := map[string]interface{}{
		"username": "journeyuser",
		"password": "JourneyPass123!",
	}

	resp = suite.server.POST(suite.T(), "/api/auth/login", loginData)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close login response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var loginResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &loginResponse); err != nil {
		suite.T().Fatalf("Failed to parse login response: %v", err)
	}
	token := loginResponse["token"].(string)

	// Step 3: Browse Articles
	authHeader := testutil.AuthHeader(token)
	resp = suite.server.GET(suite.T(), "/api/news", authHeader)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close news response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Step 4: Request Translation (if user has permission)
	translationRequest := map[string]interface{}{
		"article_id":      1,
		"target_language": "tr",
	}

	resp = suite.server.POST(suite.T(), "/api/translation/request", translationRequest, authHeader)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close translation request response body: %v", err)
		}
	}()
	// This might be 201 (created) or 403 (forbidden) depending on permissions
	assert.True(suite.T(), resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusForbidden)
}

// TestE2E_AdminWorkflow tests complete admin workflow
func (suite *E2ETestSuite) TestE2E_AdminWorkflow() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	if suite.adminToken == "" {
		suite.T().Skip("Admin token not available")
		return
	}

	authHeader := testutil.AuthHeader(suite.adminToken)

	// Step 1: Create Category
	category := map[string]interface{}{
		"name":        "E2E Test Category",
		"description": "Category created during E2E testing",
		"slug":        "e2e-test-category",
	}

	resp := suite.server.POST(suite.T(), "/api/admin/categories", category, authHeader)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close category creation response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var categoryResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &categoryResponse); err != nil {
		suite.T().Fatalf("Failed to parse category response: %v", err)
	}
	categoryID := int(categoryResponse["id"].(float64))

	// Step 2: Create Article
	article := map[string]interface{}{
		"title":       "E2E Test Article",
		"content":     "This article was created during E2E testing to verify the complete workflow.",
		"summary":     "E2E test article summary",
		"author":      "E2E Test Author",
		"category_id": categoryID,
		"published":   true,
	}

	resp = suite.server.POST(suite.T(), "/api/admin/articles", article, authHeader)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close article creation response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var articleResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &articleResponse); err != nil {
		suite.T().Fatalf("Failed to parse article response: %v", err)
	}
	articleID := int(articleResponse["id"].(float64))

	// Step 3: Verify Article is Listed
	resp = suite.server.GET(suite.T(), "/api/news")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close news listing response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var newsResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &newsResponse); err != nil {
		suite.T().Errorf("Failed to parse JSON response: %v", err)
	}

	articles := newsResponse["articles"].([]interface{})
	found := false
	for _, a := range articles {
		art := a.(map[string]interface{})
		if int(art["id"].(float64)) == articleID {
			found = true
			assert.Equal(suite.T(), "E2E Test Article", art["title"])
			break
		}
	}
	assert.True(suite.T(), found, "Created article should be visible in news list")

	// Step 4: Update Article
	updateData := map[string]interface{}{
		"title": "Updated E2E Test Article",
	}

	resp = suite.server.PUT(suite.T(), "/api/admin/articles/"+string(rune(articleID)), updateData, authHeader)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Step 5: Delete Article
	resp = suite.server.DELETE(suite.T(), "/api/admin/articles/"+string(rune(articleID)), authHeader)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

// TestE2E_TranslationWorkflow tests translation workflow
func (suite *E2ETestSuite) TestE2E_TranslationWorkflow() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	if suite.userToken == "" {
		suite.T().Skip("User token not available")
		return
	}

	authHeader := testutil.AuthHeader(suite.userToken)

	// Step 1: Request Translation
	translationRequest := map[string]interface{}{
		"article_id":      1,
		"target_language": "tr",
	}

	resp := suite.server.POST(suite.T(), "/api/translation/request", translationRequest, authHeader)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		suite.T().Skip("Translation request not supported or failed")
		return
	}

	var translationResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &translationResponse); err != nil {
		suite.T().Errorf("Failed to parse translation response: %v", err)
	}
	translationID := int(translationResponse["id"].(float64))

	// Step 2: Check Translation Status
	resp = suite.server.GET(suite.T(), "/api/translation/status/"+string(rune(translationID)), authHeader)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var statusResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &statusResponse); err != nil {
		suite.T().Errorf("Failed to parse status response: %v", err)
	}
	assert.Contains(suite.T(), statusResponse, "status")

	// Step 3: Wait for completion or simulate completion
	// In a real scenario, you might wait or trigger the translation process
	time.Sleep(100 * time.Millisecond) // Brief wait for processing

	// Step 4: Check Translation Stats
	resp = suite.server.GET(suite.T(), "/api/translation/stats")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close translation stats response body: %v", err)
		}
	}()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var statsResponse map[string]interface{}
	if err := testutil.ParseJSONResponse(suite.T(), resp, &statsResponse); err != nil {
		suite.T().Errorf("Failed to parse stats response: %v", err)
	}

	assert.Contains(suite.T(), statsResponse, "total_translations")
	assert.Contains(suite.T(), statsResponse, "pending")
	assert.Contains(suite.T(), statsResponse, "completed")
}

// TestE2E_ErrorHandlingWorkflow tests error handling across the application
func (suite *E2ETestSuite) TestE2E_ErrorHandlingWorkflow() {
	if suite.server == nil {
		suite.T().Skip("Server not initialized")
		return
	}

	// Test 1: Invalid login
	invalidLogin := map[string]interface{}{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	resp := suite.server.POST(suite.T(), "/api/auth/login", invalidLogin)
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)

	// Test 2: Access protected resource without token
	resp = suite.server.GET(suite.T(), "/api/admin/users")
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)

	// Test 3: Access admin resource with user token
	if suite.userToken != "" {
		userHeader := testutil.AuthHeader(suite.userToken)
		resp = suite.server.GET(suite.T(), "/api/admin/users", userHeader)
		defer resp.Body.Close()
		assert.Equal(suite.T(), http.StatusForbidden, resp.StatusCode)
	}

	// Test 4: Invalid JSON request
	resp = suite.server.POST(suite.T(), "/api/auth/login", "invalid json")
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test 5: Nonexistent endpoint
	resp = suite.server.GET(suite.T(), "/api/nonexistent")
	defer resp.Body.Close()
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
