package repositories

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"news/internal/models"

	"gorm.io/gorm"
)

// PageRepository handles database operations for pages
type PageRepository struct {
	db *gorm.DB
}

// NewPageRepository creates a new page repository
func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{db: db}
}

// Create creates a new page
func (r *PageRepository) Create(page *models.Page) error {
	if err := r.validatePage(page); err != nil {
		return err
	}

	// Generate slug if not provided
	if page.Slug == "" {
		page.Slug = r.generateSlug(page.Title)
	}

	// Ensure slug is unique
	page.Slug = r.ensureUniqueSlug(page.Slug, 0)

	return r.db.Create(page).Error
}

// GetByID retrieves a page by ID with relations
func (r *PageRepository) GetByID(id uint, includeBlocks bool) (*models.Page, error) {
	var page models.Page
	query := r.db.Preload("Author")

	if includeBlocks {
		query = query.Preload("ContentBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		}).Preload("ContentBlocks.ChildBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		})
	}

	if err := query.First(&page, id).Error; err != nil {
		return nil, err
	}

	return &page, nil
}

// GetBySlug retrieves a page by slug
func (r *PageRepository) GetBySlug(slug string, includeBlocks bool) (*models.Page, error) {
	var page models.Page
	query := r.db.Preload("Author")

	if includeBlocks {
		query = query.Preload("ContentBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		}).Preload("ContentBlocks.ChildBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		})
	}

	if err := query.Where("slug = ? AND status != ?", slug, "archived").First(&page).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("page with slug '%s' not found", slug)
		}
		return nil, err
	}

	return &page, nil
}

// GetList retrieves a paginated list of pages
func (r *PageRepository) GetList(filters PageFilters) ([]models.Page, int64, error) {
	var pages []models.Page
	var total int64

	query := r.db.Model(&models.Page{}).Preload("Author")

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (filters.Page - 1) * filters.Limit
	query = query.Offset(offset).Limit(filters.Limit)

	// Apply ordering
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortOrder == "desc" {
			direction = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filters.SortBy, direction))
	} else {
		query = query.Order("updated_at DESC")
	}

	if err := query.Find(&pages).Error; err != nil {
		return nil, 0, err
	}

	return pages, total, nil
}

// Update updates a page
func (r *PageRepository) Update(page *models.Page) error {
	if err := r.validatePage(page); err != nil {
		return err
	}

	return r.db.Save(page).Error
}

// Delete soft deletes a page
func (r *PageRepository) Delete(id uint) error {
	return r.db.Delete(&models.Page{}, id).Error
}

// UpdateStatus updates page status
func (r *PageRepository) UpdateStatus(id uint, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Set published_at when publishing
	if status == "published" {
		updates["published_at"] = time.Now()
	}

	return r.db.Model(&models.Page{}).Where("id = ?", id).Updates(updates).Error
}

// GetHomePage retrieves the homepage
func (r *PageRepository) GetHomePage() (*models.Page, error) {
	var page models.Page
	query := r.db.Preload("Author").
		Preload("ContentBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		}).
		Where("is_homepage = ? AND status = ?", true, "published")

	if err := query.First(&page).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("homepage not found")
		}
		return nil, err
	}

	return &page, nil
}

// GetPageHierarchy retrieves page hierarchy (parent-child relationships)
func (r *PageRepository) GetPageHierarchy() ([]models.Page, error) {
	var pages []models.Page
	err := r.db.Preload("Children").
		Where("parent_id IS NULL").
		Order("sort_order ASC, title ASC").
		Find(&pages).Error

	return pages, err
}

// IncrementViews increments page view count
func (r *PageRepository) IncrementViews(id uint) error {
	return r.db.Model(&models.Page{}).Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}

// GetPublishedPages retrieves only published pages
func (r *PageRepository) GetPublishedPages(limit int, offset int) ([]models.Page, error) {
	var pages []models.Page
	err := r.db.Preload("Author").
		Where("status = ?", "published").
		Order("published_at DESC").
		Limit(limit).Offset(offset).
		Find(&pages).Error

	return pages, err
}

// SearchPages searches pages by title and content
func (r *PageRepository) SearchPages(query string, limit int, offset int) ([]models.Page, int64, error) {
	var pages []models.Page
	var total int64

	searchQuery := r.db.Model(&models.Page{}).
		Preload("Author").
		Where("status = ?", "published")

	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		searchQuery = searchQuery.Where(
			"LOWER(title) LIKE ? OR LOWER(meta_description) LIKE ? OR LOWER(excerpt_text) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results
	if err := searchQuery.Order("published_at DESC").
		Limit(limit).Offset(offset).
		Find(&pages).Error; err != nil {
		return nil, 0, err
	}

	return pages, total, nil
}

// Helper methods

// validatePage validates page data
func (r *PageRepository) validatePage(page *models.Page) error {
	if page.Title == "" {
		return errors.New("page title is required")
	}

	if page.AuthorID == 0 {
		return errors.New("page author is required")
	}

	if !page.ValidateStatus() {
		return errors.New("invalid page status")
	}

	if !page.ValidateTemplate() {
		return errors.New("invalid page template")
	}

	return nil
}

// generateSlug generates a URL-friendly slug from title
func (r *PageRepository) generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic version)
	specialChars := []string{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "=", "+", "[", "]", "{", "}", "|", "\\", ":", ";", "\"", "'", "<", ">", ",", ".", "?", "/"}
	for _, char := range specialChars {
		slug = strings.ReplaceAll(slug, char, "")
	}
	// Remove multiple dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	return slug
}

// ensureUniqueSlug ensures slug uniqueness
func (r *PageRepository) ensureUniqueSlug(baseSlug string, pageID uint) string {
	slug := baseSlug
	counter := 1

	for {
		var count int64
		query := r.db.Model(&models.Page{}).Where("slug = ?", slug)
		if pageID > 0 {
			query = query.Where("id != ?", pageID)
		}
		query.Count(&count)

		if count == 0 {
			break
		}

		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	return slug
}

// applyFilters applies filters to query
func (r *PageRepository) applyFilters(query *gorm.DB, filters PageFilters) *gorm.DB {
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.Template != "" {
		query = query.Where("template = ?", filters.Template)
	}

	if filters.Language != "" {
		query = query.Where("language = ?", filters.Language)
	}

	if filters.AuthorID > 0 {
		query = query.Where("author_id = ?", filters.AuthorID)
	}

	if filters.ParentID != nil {
		if *filters.ParentID == 0 {
			// Root pages only
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", *filters.ParentID)
		}
	}

	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(meta_description) LIKE ?",
			searchTerm, searchTerm,
		)
	}

	return query
}

// PageFilters represents filters for page listing
type PageFilters struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Status    string `json:"status"`
	Template  string `json:"template"`
	Language  string `json:"language"`
	AuthorID  uint   `json:"author_id"`
	ParentID  *uint  `json:"parent_id"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// Default values for filters
func (f *PageFilters) SetDefaults() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
}
