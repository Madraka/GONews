package repositories

import (
	"errors"
	"fmt"
	"strings"

	"news/internal/models"

	"gorm.io/gorm"
)

// PageTemplateRepository handles database operations for page templates
type PageTemplateRepository struct {
	db *gorm.DB
}

// NewPageTemplateRepository creates a new page template repository
func NewPageTemplateRepository(db *gorm.DB) *PageTemplateRepository {
	return &PageTemplateRepository{db: db}
}

// Create creates a new page template
func (r *PageTemplateRepository) Create(template *models.PageTemplate) error {
	if err := r.validateTemplate(template); err != nil {
		return err
	}

	return r.db.Create(template).Error
}

// GetByID retrieves a page template by ID
func (r *PageTemplateRepository) GetByID(id uint) (*models.PageTemplate, error) {
	var template models.PageTemplate
	if err := r.db.Preload("Creator").First(&template, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("page template with ID %d not found", id)
		}
		return nil, err
	}

	return &template, nil
}

// GetList retrieves a paginated list of page templates
func (r *PageTemplateRepository) GetList(filters PageTemplateFilters) ([]models.PageTemplate, int64, error) {
	var templates []models.PageTemplate
	var total int64

	query := r.db.Model(&models.PageTemplate{}).Preload("Creator")

	// Apply filters
	query = r.applyTemplateFilters(query, filters)

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
		query = query.Order("usage_count DESC, rating DESC, updated_at DESC")
	}

	if err := query.Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// GetPublicTemplates retrieves public templates
func (r *PageTemplateRepository) GetPublicTemplates(limit int, offset int) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate
	err := r.db.Preload("Creator").
		Where("is_public = ?", true).
		Order("usage_count DESC, rating DESC").
		Limit(limit).Offset(offset).
		Find(&templates).Error

	return templates, err
}

// GetByCategory retrieves templates by category
func (r *PageTemplateRepository) GetByCategory(category string, limit int, offset int) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate
	err := r.db.Preload("Creator").
		Where("category = ? AND is_public = ?", category, true).
		Order("usage_count DESC, rating DESC").
		Limit(limit).Offset(offset).
		Find(&templates).Error

	return templates, err
}

// GetByCreator retrieves templates created by a specific user
func (r *PageTemplateRepository) GetByCreator(creatorID uint, includePrivate bool) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate
	query := r.db.Where("creator_id = ?", creatorID)

	if !includePrivate {
		query = query.Where("is_public = ?", true)
	}

	err := query.Order("updated_at DESC").Find(&templates).Error
	return templates, err
}

// GetPopularTemplates retrieves most popular templates
func (r *PageTemplateRepository) GetPopularTemplates(limit int) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate
	err := r.db.Preload("Creator").
		Where("is_public = ?", true).
		Order("usage_count DESC, rating DESC").
		Limit(limit).
		Find(&templates).Error

	return templates, err
}

// GetFeaturedTemplates retrieves featured/premium templates
func (r *PageTemplateRepository) GetFeaturedTemplates(limit int) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate
	err := r.db.Preload("Creator").
		Where("is_premium = ? AND is_public = ?", true, true).
		Order("rating DESC, usage_count DESC").
		Limit(limit).
		Find(&templates).Error

	return templates, err
}

// SearchTemplates searches templates by name, description, and tags
func (r *PageTemplateRepository) SearchTemplates(query string, limit int, offset int) ([]models.PageTemplate, int64, error) {
	var templates []models.PageTemplate
	var total int64

	searchQuery := r.db.Model(&models.PageTemplate{}).
		Preload("Creator").
		Where("is_public = ?", true)

	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		// Search in name and description using LIKE, and in tags using JSON text search
		searchQuery = searchQuery.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(tags::text) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results
	if err := searchQuery.Order("usage_count DESC, rating DESC").
		Limit(limit).Offset(offset).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update updates a page template
func (r *PageTemplateRepository) Update(template *models.PageTemplate) error {
	if err := r.validateTemplate(template); err != nil {
		return err
	}

	return r.db.Save(template).Error
}

// Delete soft deletes a page template
func (r *PageTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.PageTemplate{}, id).Error
}

// IncrementUsage increments template usage count
func (r *PageTemplateRepository) IncrementUsage(id uint) error {
	return r.db.Model(&models.PageTemplate{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
}

// UpdateRating updates template rating
func (r *PageTemplateRepository) UpdateRating(id uint, rating float64) error {
	return r.db.Model(&models.PageTemplate{}).
		Where("id = ?", id).
		Update("rating", rating).Error
}

// GetTemplateCategories retrieves all available template categories
func (r *PageTemplateRepository) GetTemplateCategories() ([]TemplateCategory, error) {
	var categories []TemplateCategory
	err := r.db.Model(&models.PageTemplate{}).
		Select("category, COUNT(*) as count").
		Where("is_public = ?", true).
		Group("category").
		Having("category != ''").
		Order("count DESC").
		Scan(&categories).Error

	return categories, err
}

// GetTemplatesByTags retrieves templates by tags
func (r *PageTemplateRepository) GetTemplatesByTags(tags []string, limit int, offset int) ([]models.PageTemplate, error) {
	var templates []models.PageTemplate

	if len(tags) == 0 {
		return templates, nil
	}

	query := r.db.Preload("Creator").Where("is_public = ?", true)

	// Build OR conditions for each tag using PostgreSQL JSON operators
	var conditions []string
	var values []interface{}

	for _, tag := range tags {
		// Use JSON contains operator to check if tag exists in the JSON array
		conditions = append(conditions, "tags::jsonb @> ?")
		values = append(values, fmt.Sprintf(`["%s"]`, tag))
	}

	query = query.Where(strings.Join(conditions, " OR "), values...)

	err := query.Order("usage_count DESC, rating DESC").
		Limit(limit).Offset(offset).
		Find(&templates).Error

	return templates, err
}

// DuplicateTemplate creates a copy of an existing template
func (r *PageTemplateRepository) DuplicateTemplate(templateID uint, newCreatorID uint, newName string) (*models.PageTemplate, error) {
	var originalTemplate models.PageTemplate
	if err := r.db.First(&originalTemplate, templateID).Error; err != nil {
		return nil, err
	}

	newTemplate := originalTemplate
	newTemplate.ID = 0
	newTemplate.Name = newName
	newTemplate.CreatorID = newCreatorID
	newTemplate.UsageCount = 0
	newTemplate.Rating = 0
	newTemplate.IsPublic = false
	newTemplate.IsPremium = false

	if err := r.db.Create(&newTemplate).Error; err != nil {
		return nil, err
	}

	return &newTemplate, nil
}

// Helper methods

// validateTemplate validates template data
func (r *PageTemplateRepository) validateTemplate(template *models.PageTemplate) error {
	if template.Name == "" {
		return errors.New("template name is required")
	}

	if template.CreatorID == 0 {
		return errors.New("template creator is required")
	}

	if len(template.BlockStructure) == 0 {
		return errors.New("template block structure is required")
	}

	return nil
}

// applyTemplateFilters applies filters to template query
func (r *PageTemplateRepository) applyTemplateFilters(query *gorm.DB, filters PageTemplateFilters) *gorm.DB {
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.CreatorID > 0 {
		query = query.Where("creator_id = ?", filters.CreatorID)
	}

	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}

	if filters.IsPremium != nil {
		query = query.Where("is_premium = ?", *filters.IsPremium)
	}

	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
			searchTerm, searchTerm,
		)
	}

	if filters.MinRating > 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}

	return query
}

// PageTemplateFilters represents filters for template listing
type PageTemplateFilters struct {
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
	Category  string  `json:"category"`
	CreatorID uint    `json:"creator_id"`
	IsPublic  *bool   `json:"is_public"`
	IsPremium *bool   `json:"is_premium"`
	Search    string  `json:"search"`
	MinRating float64 `json:"min_rating"`
	SortBy    string  `json:"sort_by"`
	SortOrder string  `json:"sort_order"`
}

// SetDefaults sets default values for template filters
func (f *PageTemplateFilters) SetDefaults() {
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

// TemplateCategory represents template category with count
type TemplateCategory struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}
