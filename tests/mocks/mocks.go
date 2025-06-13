package mocks

import (
	"news/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of the user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(offset, limit int) ([]*models.User, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockArticleRepository is a mock implementation of the article repository
type MockArticleRepository struct {
	mock.Mock
}

func (m *MockArticleRepository) Create(article *models.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

func (m *MockArticleRepository) GetByID(id int) (*models.Article, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

func (m *MockArticleRepository) Update(article *models.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

func (m *MockArticleRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockArticleRepository) List(offset, limit int) ([]*models.Article, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*models.Article), args.Error(1)
}

func (m *MockArticleRepository) GetByCategoryID(categoryID, offset, limit int) ([]*models.Article, error) {
	args := m.Called(categoryID, offset, limit)
	return args.Get(0).([]*models.Article), args.Error(1)
}

func (m *MockArticleRepository) Search(query string, offset, limit int) ([]*models.Article, error) {
	args := m.Called(query, offset, limit)
	return args.Get(0).([]*models.Article), args.Error(1)
}

// MockArticleTranslationRepository is a mock implementation of the article translation repository
type MockArticleTranslationRepository struct {
	mock.Mock
}

func (m *MockArticleTranslationRepository) CreateTranslation(translation *models.ArticleTranslation) (*models.ArticleTranslation, error) {
	args := m.Called(translation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArticleTranslation), args.Error(1)
}

func (m *MockArticleTranslationRepository) GetTranslationByID(id uint) (*models.ArticleTranslation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArticleTranslation), args.Error(1)
}

func (m *MockArticleTranslationRepository) GetTranslationByLanguage(articleID uint, language string) (*models.ArticleTranslation, error) {
	args := m.Called(articleID, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArticleTranslation), args.Error(1)
}

func (m *MockArticleTranslationRepository) UpdateTranslation(translationID uint, updateData map[string]interface{}) (*models.ArticleTranslation, error) {
	args := m.Called(translationID, updateData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ArticleTranslation), args.Error(1)
}

func (m *MockArticleTranslationRepository) GetTranslationsByArticleID(articleID uint) ([]models.ArticleTranslation, error) {
	args := m.Called(articleID)
	return args.Get(0).([]models.ArticleTranslation), args.Error(1)
}

func (m *MockArticleTranslationRepository) GetTranslationStatsForArticle(articleID uint) (*models.TranslationStats, error) {
	args := m.Called(articleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TranslationStats), args.Error(1)
}

// MockCacheService is a mock implementation of the cache service
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Set(key string, value interface{}, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCacheService) Clear() error {
	args := m.Called()
	return args.Error(0)
}

// MockAIService is a mock implementation of the AI service
type MockAIService struct {
	mock.Mock
}

func (m *MockAIService) AnalyzeText(text string) (map[string]interface{}, error) {
	args := m.Called(text)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAIService) SummarizeText(text string) (string, error) {
	args := m.Called(text)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) TranslateText(text, sourceLang, targetLang string) (string, error) {
	args := m.Called(text, sourceLang, targetLang)
	return args.String(0), args.Error(1)
}

func (m *MockAIService) ExtractKeywords(text string) ([]string, error) {
	args := m.Called(text)
	return args.Get(0).([]string), args.Error(1)
}

// MockEmailService is a mock implementation of the email service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendWelcomeEmail(email, username string) error {
	args := m.Called(email, username)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(email, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

func (m *MockEmailService) SendTranslationCompleteEmail(email string, translationID int) error {
	args := m.Called(email, translationID)
	return args.Error(0)
}

// MockQueueService is a mock implementation of the queue service
type MockQueueService struct {
	mock.Mock
}

func (m *MockQueueService) Enqueue(queueName string, payload interface{}) error {
	args := m.Called(queueName, payload)
	return args.Error(0)
}

func (m *MockQueueService) Dequeue(queueName string) (interface{}, error) {
	args := m.Called(queueName)
	return args.Get(0), args.Error(1)
}

func (m *MockQueueService) GetQueueSize(queueName string) (int, error) {
	args := m.Called(queueName)
	return args.Int(0), args.Error(1)
}

func (m *MockQueueService) ClearQueue(queueName string) error {
	args := m.Called(queueName)
	return args.Error(0)
}
