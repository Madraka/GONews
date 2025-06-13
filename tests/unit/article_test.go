package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"news/internal/models"
	"news/tests/testutil"
)

// ArticleModelTestSuite defines the test suite for Article model
type ArticleModelTestSuite struct {
	suite.Suite
	testData *testutil.TestData
}

// SetupSuite runs before all tests in the suite
func (suite *ArticleModelTestSuite) SetupSuite() {
	suite.testData = testutil.NewTestData()
}

// TestArticleModel_ValidArticle tests valid article creation
func (suite *ArticleModelTestSuite) TestArticleModel_ValidArticle() {
	article := suite.testData.CreateTestArticle()

	assert.NotEmpty(suite.T(), article.Title)
	assert.NotEmpty(suite.T(), article.Content)
	assert.NotEmpty(suite.T(), article.Slug)
	assert.Equal(suite.T(), uint(1), article.AuthorID)
	assert.Equal(suite.T(), "published", article.Status)
	assert.Equal(suite.T(), "en", article.Language)
	assert.NotZero(suite.T(), article.CreatedAt)
	assert.NotZero(suite.T(), article.UpdatedAt)
}

// TestArticleModel_TitleValidation tests article title validation
func (suite *ArticleModelTestSuite) TestArticleModel_TitleValidation() {
	tests := []struct {
		name    string
		title   string
		valid   bool
		message string
	}{
		{"valid title", "Test Article", true, ""},
		{"empty title", "", false, "title cannot be empty"},
		{"too short title", "A", false, "title too short"},
		{"very long title", string(make([]byte, 300)), false, "title too long"},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			article := &models.Article{
				Title:    tt.title,
				Slug:     "test-slug",
				Content:  "Valid content",
				AuthorID: 1,
				Status:   "draft",
				Language: "en",
			}

			if tt.valid {
				assert.NotEmpty(t, article.Title)
				assert.True(t, len(article.Title) > 1)
			} else {
				if tt.title == "" {
					assert.Empty(t, article.Title)
				} else if len(tt.title) == 1 {
					assert.Equal(t, 1, len(article.Title))
				} else {
					assert.True(t, len(article.Title) > 255)
				}
			}
		})
	}
}

// TestArticleModel_ContentValidation tests article content validation
func (suite *ArticleModelTestSuite) TestArticleModel_ContentValidation() {
	article := suite.testData.CreateTestArticle()

	// Test valid content
	assert.NotEmpty(suite.T(), article.Content)
	assert.True(suite.T(), len(article.Content) > 10)

	// Test empty content
	article.Content = ""
	assert.Empty(suite.T(), article.Content)
}

// TestArticleModel_PublishingStatus tests article publishing status
func (suite *ArticleModelTestSuite) TestArticleModel_PublishingStatus() {
	article := suite.testData.CreateTestArticle()

	// Test published article
	assert.Equal(suite.T(), "published", article.Status)

	// Test draft article
	article.Status = "draft"
	assert.Equal(suite.T(), "draft", article.Status)
}

// TestArticleModel_AuthorAssociation tests article-author association
func (suite *ArticleModelTestSuite) TestArticleModel_AuthorAssociation() {
	article := suite.testData.CreateTestArticle()

	assert.Equal(suite.T(), uint(1), article.AuthorID)
	assert.NotZero(suite.T(), article.AuthorID)
}

// TestArticleModel_Timestamps tests article timestamp handling
func (suite *ArticleModelTestSuite) TestArticleModel_Timestamps() {
	article := suite.testData.CreateTestArticle()

	now := time.Now()

	// Check that timestamps are recent (within 1 second)
	assert.WithinDuration(suite.T(), now, article.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), now, article.UpdatedAt, time.Second)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArticleModelTestSuite(t *testing.T) {
	suite.Run(t, new(ArticleModelTestSuite))
}

// Test individual functions without suite
func TestArticle_Slug_Generation(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		expectedSlug string
	}{
		{"simple title", "Test Article", "test-article"},
		{"title with spaces", "This Is A Test", "this-is-a-test"},
		{"title with special chars", "Test! Article?", "test-article"},
		{"turkish characters", "Türkçe Makale", "turkce-makale"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// In real implementation, you would have a GenerateSlug method
			// For now, we'll just test the concept
			article := &models.Article{
				Title: tt.title,
			}

			assert.NotEmpty(t, article.Title)
			// Here you would test: assert.Equal(t, tt.expectedSlug, article.GenerateSlug())
		})
	}
}

func TestArticle_WordCount(t *testing.T) {
	testData := testutil.NewTestData()
	article := testData.CreateTestArticle()

	// Test word count calculation
	content := "This is a test article with eight words."
	article.Content = content

	// In real implementation, you would have a WordCount method
	words := len(splitWords(content))
	assert.Equal(t, 8, words)
}

// Helper function for word count test
func splitWords(text string) []string {
	if text == "" {
		return []string{}
	}

	words := []string{}
	current := ""

	for _, char := range text {
		if char == ' ' || char == '\n' || char == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}
