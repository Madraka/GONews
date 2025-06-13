package testutil

import (
	"time"

	"news/internal/models"
)

// TestData provides common test data
type TestData struct{}

// NewTestData creates a new TestData instance
func NewTestData() *TestData {
	return &TestData{}
}

// CreateTestUser creates a test user
func (td *TestData) CreateTestUser() *models.User {
	return &models.User{
		ID:         1,
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "$2a$10$hash", // bcrypt hash for "password"
		FirstName:  "Test",
		LastName:   "User",
		Role:       "user",
		Status:     "active",
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestAdmin creates a test admin user
func (td *TestData) CreateTestAdmin() *models.User {
	return &models.User{
		ID:         2,
		Username:   "admin",
		Email:      "admin@example.com",
		Password:   "$2a$10$hash", // bcrypt hash for "password"
		FirstName:  "Admin",
		LastName:   "User",
		Role:       "admin",
		Status:     "active",
		IsVerified: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateTestCategory creates a test category
func (td *TestData) CreateTestCategory() *models.Category {
	return &models.Category{
		ID:          1,
		Name:        "Technology",
		Description: "Technology news and articles",
		Slug:        "technology",
		Color:       "#007bff",
		Icon:        "tech",
		IsActive:    true,
		SortOrder:   1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreateTestArticle creates a test article
func (td *TestData) CreateTestArticle() *models.Article {
	return &models.Article{
		ID:            1,
		Title:         "Test Article",
		Slug:          "test-article",
		Content:       "This is a test article content",
		Summary:       "Test summary",
		AuthorID:      1,
		Status:        "published",
		Language:      "en",
		Views:         0,
		ReadTime:      5,
		AllowComments: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// CreateTestTranslation creates a test article translation
func (td *TestData) CreateTestTranslation() *models.ArticleTranslation {
	return &models.ArticleTranslation{
		ID:              1,
		ArticleID:       1,
		Language:        "tr",
		Title:           "Test Makale",
		Slug:            "test-makale",
		Summary:         "Test özet",
		Content:         "Bu bir test makale içeriğidir",
		MetaTitle:       "Test Makale Meta",
		MetaDescription: "Test makale meta açıklaması",
		Status:          "draft",
		TranslationType: "manual",
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateLoginRequest creates a test login request
func (td *TestData) CreateLoginRequest() LoginRequest {
	return LoginRequest{
		Username: "testuser",
		Password: "password",
	}
}

// CreateAdminLoginRequest creates an admin login request
func (td *TestData) CreateAdminLoginRequest() LoginRequest {
	return LoginRequest{
		Username: "admin",
		Password: "password",
	}
}

// ArticleCreateRequest represents an article creation request
type ArticleCreateRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	Author     string `json:"author"`
	CategoryID int    `json:"category_id"`
	Published  bool   `json:"published"`
}

// CreateArticleRequest creates a test article creation request
func (td *TestData) CreateArticleRequest() ArticleCreateRequest {
	return ArticleCreateRequest{
		Title:      "New Test Article",
		Content:    "This is new test content",
		Summary:    "New test summary",
		Author:     "Test Author",
		CategoryID: 1,
		Published:  true,
	}
}

// TranslationRequest represents a translation request
type TranslationRequest struct {
	ArticleID      int    `json:"article_id"`
	TargetLanguage string `json:"target_language"`
}

// CreateTranslationRequest creates a test translation request
func (td *TestData) CreateTranslationRequest() TranslationRequest {
	return TranslationRequest{
		ArticleID:      1,
		TargetLanguage: "tr",
	}
}
