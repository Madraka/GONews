package handlers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"news/internal/database"
	"news/internal/json"
	"news/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto-migrate test tables
	if err := db.AutoMigrate(&models.User{}, &models.Newsletter{}, &models.Menu{},
		&models.MenuItem{}, &models.Setting{}, &models.Media{}, &models.Category{}); err != nil {
		log.Printf("Warning: Failed to auto-migrate test tables: %v", err)
	}

	return db
}

func TestCreateNewsletter(t *testing.T) {
	// Setup
	database.DB = setupTestDB()
	gin.SetMode(gin.TestMode)

	// Create test user
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Role:     "admin",
	}
	database.DB.Create(&user)

	router := gin.New()
	router.POST("/newsletters", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		CreateNewsletter(c)
	})

	newsletter := models.Newsletter{
		Title:   "Test Newsletter",
		Subject: "Test Subject",
		Content: "Test Content",
		Status:  "draft",
	}

	jsonData, _ := json.Marshal(newsletter)
	req := httptest.NewRequest("POST", "/newsletters", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Newsletter
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "Test Newsletter", response.Title)
	assert.Equal(t, "draft", response.Status)
}

func TestCreateMenu(t *testing.T) {
	// Setup
	database.DB = setupTestDB()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/menus", CreateMenu)

	menu := models.Menu{
		Name:     "Test Menu",
		Slug:     "test-menu",
		Location: "header",
		IsActive: true,
	}

	jsonData, _ := json.Marshal(menu)
	req := httptest.NewRequest("POST", "/menus", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Menu
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "Test Menu", response.Name)
	assert.Equal(t, "header", response.Location)
}

func TestCreateSetting(t *testing.T) {
	// Setup
	database.DB = setupTestDB()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/settings", CreateSetting)

	setting := models.Setting{
		Key:         "site_title",
		Value:       "My News Site",
		Type:        "string",
		Description: "The title of the website",
		Group:       "general",
		IsPublic:    true,
	}

	jsonData, _ := json.Marshal(setting)
	req := httptest.NewRequest("POST", "/settings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Setting
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}
	assert.Equal(t, "site_title", response.Key)
	assert.Equal(t, "My News Site", response.Value)
	assert.Equal(t, "string", response.Type)
}

func TestGetPublicSettings(t *testing.T) {
	// Setup
	database.DB = setupTestDB()
	gin.SetMode(gin.TestMode)

	// Create test settings
	publicSetting := models.Setting{
		Key:      "public_setting",
		Value:    "public_value",
		Type:     "string",
		IsPublic: true,
	}
	privateSetting := models.Setting{
		Key:      "private_setting",
		Value:    "private_value",
		Type:     "string",
		IsPublic: false,
	}
	database.DB.Create(&publicSetting)
	database.DB.Create(&privateSetting)

	router := gin.New()
	router.GET("/settings", GetSettings)

	req := httptest.NewRequest("GET", "/settings?public=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var settings []models.Setting
	if err := json.Unmarshal(w.Body.Bytes(), &settings); err != nil {
		t.Errorf("Failed to unmarshal settings response: %v", err)
	}

	// Should only return public settings
	assert.Equal(t, 1, len(settings))
	assert.Equal(t, "public_setting", settings[0].Key)
	assert.True(t, settings[0].IsPublic)
}

func TestCreateMenuItem(t *testing.T) {
	// Setup
	database.DB = setupTestDB()
	gin.SetMode(gin.TestMode)

	// Create test menu first
	menu := models.Menu{
		Name:     "Test Menu",
		Slug:     "test-menu",
		Location: "header",
		IsActive: true,
	}
	database.DB.Create(&menu)

	router := gin.New()
	router.POST("/menu-items", CreateMenuItem)

	menuItem := models.MenuItem{
		MenuID:    menu.ID,
		Title:     "Home",
		URL:       "/",
		Target:    "_self",
		SortOrder: 1,
		IsActive:  true,
	}

	jsonData, _ := json.Marshal(menuItem)
	req := httptest.NewRequest("POST", "/menu-items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.MenuItem
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal menu item response: %v", err)
	}
	assert.Equal(t, "Home", response.Title)
	assert.Equal(t, "/", response.URL)
	assert.Equal(t, menu.ID, response.MenuID)
}
