package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"news/internal/json"

	"github.com/stretchr/testify/require"
)

// TestFixtures manages loading test fixtures
type TestFixtures struct {
	Data map[string]interface{}
}

// LoadTestData loads test data from fixtures/test_data.json
func LoadTestData(t *testing.T) *TestFixtures {
	// Get the path to the test data file
	fixturesPath := filepath.Join("fixtures", "test_data.json")

	// Read the file
	data, err := os.ReadFile(fixturesPath)
	require.NoError(t, err, "Failed to read test data file")

	// Parse JSON
	var testData map[string]interface{}
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to parse test data JSON")

	return &TestFixtures{
		Data: testData,
	}
}

// GetTestUsers returns test users from fixtures
func (tf *TestFixtures) GetTestUsers() []map[string]interface{} {
	users, ok := tf.Data["test_users"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = user.(map[string]interface{})
	}

	return result
}

// GetTestArticles returns test articles from fixtures
func (tf *TestFixtures) GetTestArticles() []map[string]interface{} {
	articles, ok := tf.Data["test_articles"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(articles))
	for i, article := range articles {
		result[i] = article.(map[string]interface{})
	}

	return result
}

// GetTestCategories returns test categories from fixtures
func (tf *TestFixtures) GetTestCategories() []map[string]interface{} {
	categories, ok := tf.Data["test_categories"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(categories))
	for i, category := range categories {
		result[i] = category.(map[string]interface{})
	}

	return result
}

// GetAPIEndpoints returns API endpoints for testing
func (tf *TestFixtures) GetAPIEndpoints() []map[string]interface{} {
	endpoints, ok := tf.Data["api_endpoints"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(endpoints))
	for i, endpoint := range endpoints {
		result[i] = endpoint.(map[string]interface{})
	}

	return result
}

// GetErrorScenarios returns error test scenarios
func (tf *TestFixtures) GetErrorScenarios() []map[string]interface{} {
	scenarios, ok := tf.Data["error_scenarios"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(scenarios))
	for i, scenario := range scenarios {
		result[i] = scenario.(map[string]interface{})
	}

	return result
}

// GetPerformanceBenchmarks returns performance benchmarks
func (tf *TestFixtures) GetPerformanceBenchmarks() map[string]interface{} {
	benchmarks, ok := tf.Data["performance_benchmarks"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}

	return benchmarks
}

// GetTestUser returns a specific test user by username
func (tf *TestFixtures) GetTestUser(username string) map[string]interface{} {
	users := tf.GetTestUsers()
	for _, user := range users {
		if user["username"] == username {
			return user
		}
	}
	return nil
}

// GetTestArticle returns a specific test article by ID
func (tf *TestFixtures) GetTestArticle(id int) map[string]interface{} {
	articles := tf.GetTestArticles()
	for _, article := range articles {
		if int(article["id"].(float64)) == id {
			return article
		}
	}
	return nil
}

// GetAITestRequests returns AI test requests
func (tf *TestFixtures) GetAITestRequests() []map[string]interface{} {
	requests, ok := tf.Data["test_ai_requests"].([]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(requests))
	for i, request := range requests {
		result[i] = request.(map[string]interface{})
	}

	return result
}
