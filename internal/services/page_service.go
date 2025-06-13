package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"news/internal/models"
	"news/internal/repositories"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreatePageRequest represents a request to create a new page
type CreatePageRequest struct {
	Title           string                   `json:"title"`
	Slug            string                   `json:"slug"`
	MetaTitle       string                   `json:"meta_title"`
	MetaDescription string                   `json:"meta_description"`
	Template        string                   `json:"template"`
	Layout          string                   `json:"layout"`
	Status          string                   `json:"status"`
	Language        string                   `json:"language"`
	ParentID        *uint                    `json:"parent_id"`
	SortOrder       int                      `json:"sort_order"`
	FeaturedImage   string                   `json:"featured_image"`
	ExcerptText     string                   `json:"excerpt_text"`
	AuthorID        uint                     `json:"author_id"`
	IsHomepage      bool                     `json:"is_homepage"`
	IsLandingPage   bool                     `json:"is_landing_page"`
	SEOSettings     map[string]interface{}   `json:"seo_settings"`
	PageSettings    map[string]interface{}   `json:"page_settings"`
	LayoutData      map[string]interface{}   `json:"layout_data"`
	ContentBlocks   []CreatePageBlockRequest `json:"content_blocks"`
}

// UpdatePageRequest represents a request to update a page
type UpdatePageRequest struct {
	Title           string                 `json:"title"`
	MetaTitle       string                 `json:"meta_title"`
	MetaDescription string                 `json:"meta_description"`
	Template        string                 `json:"template"`
	Layout          string                 `json:"layout"`
	Status          string                 `json:"status"`
	Language        string                 `json:"language"`
	ParentID        *uint                  `json:"parent_id"`
	SortOrder       *int                   `json:"sort_order"`
	FeaturedImage   string                 `json:"featured_image"`
	ExcerptText     string                 `json:"excerpt_text"`
	SEOSettings     map[string]interface{} `json:"seo_settings"`
	PageSettings    map[string]interface{} `json:"page_settings"`
	LayoutData      map[string]interface{} `json:"layout_data"`
}

// CreatePageBlockRequest represents a request to create a page content block
type CreatePageBlockRequest struct {
	PageID         uint                   `json:"page_id"`
	ContainerID    *uint                  `json:"container_id"`
	BlockType      string                 `json:"block_type"`
	Content        string                 `json:"content"`
	Settings       map[string]interface{} `json:"settings"`
	Styles         map[string]interface{} `json:"styles"`
	Position       int                    `json:"position"`
	IsVisible      bool                   `json:"is_visible"`
	IsContainer    bool                   `json:"is_container"`
	ContainerType  string                 `json:"container_type"`
	GridSettings   map[string]interface{} `json:"grid_settings"`
	ResponsiveData map[string]interface{} `json:"responsive_data"`
}

// UpdatePageBlockRequest represents a request to update a page content block
type UpdatePageBlockRequest struct {
	BlockType      string                 `json:"block_type"`
	Content        string                 `json:"content"`
	Settings       map[string]interface{} `json:"settings"`
	Styles         map[string]interface{} `json:"styles"`
	Position       *int                   `json:"position"`
	IsVisible      *bool                  `json:"is_visible"`
	IsContainer    *bool                  `json:"is_container"`
	ContainerType  string                 `json:"container_type"`
	GridSettings   map[string]interface{} `json:"grid_settings"`
	ResponsiveData map[string]interface{} `json:"responsive_data"`
}

// PageFilter represents filtering options for pages
type PageFilter struct {
	Status   string `json:"status"`
	Template string `json:"template"`
	Language string `json:"language"`
	ParentID *uint  `json:"parent_id"`
	Search   string `json:"search"`
}

// PaginatedPagesResponse represents a paginated response for pages
type PaginatedPagesResponse struct {
	Pages      []models.Page `json:"pages"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// PageHierarchyNode represents a node in the page hierarchy
type PageHierarchyNode struct {
	Page     models.Page         `json:"page"`
	Children []PageHierarchyNode `json:"children"`
}

// DuplicatePageRequest represents a request to duplicate a page
type DuplicatePageRequest struct {
	NewTitle       string `json:"new_title"`
	NewSlug        string `json:"new_slug"`
	IncludeBlocks  bool   `json:"include_blocks"`
	CopyAsTemplate bool   `json:"copy_as_template"`
	AuthorID       uint   `json:"author_id"`
}

// ReorderBlocksRequest represents a request to reorder page blocks
type ReorderBlocksRequest struct {
	BlockOrders []BlockOrder `json:"block_orders"`
}

// BlockOrder represents the new order for a block
type BlockOrder struct {
	BlockID  uint `json:"block_id"`
	Position int  `json:"position"`
}

// DuplicateBlockRequest represents a request to duplicate a block
type DuplicateBlockRequest struct {
	NewPosition       *int  `json:"new_position"`
	TargetPageID      *uint `json:"target_page_id"`
	TargetContainerID *uint `json:"target_container_id"`
}

// BlockTypeInfo represents information about a block type
type BlockTypeInfo struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Category    string                 `json:"category"`
	Settings    map[string]interface{} `json:"settings"`
	Styles      map[string]interface{} `json:"styles"`
	IsContainer bool                   `json:"is_container"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// PageService handles business logic for pages
type PageService struct {
	pageRepo     *repositories.PageRepository
	blockRepo    *repositories.PageContentBlockRepository
	templateRepo *repositories.PageTemplateRepository
	db           *gorm.DB
}

// NewPageService creates a new page service
func NewPageService(db *gorm.DB) *PageService {
	return &PageService{
		pageRepo:     repositories.NewPageRepository(db),
		blockRepo:    repositories.NewPageContentBlockRepository(db),
		templateRepo: repositories.NewPageTemplateRepository(db),
		db:           db,
	}
}

// CreatePage creates a new page
func (s *PageService) CreatePage(req CreatePageRequest) (*models.Page, error) {
	// Validate request
	if err := s.validateCreatePageRequest(req); err != nil {
		return nil, err
	}

	page := &models.Page{
		Title:         req.Title,
		Slug:          req.Slug,
		MetaTitle:     req.MetaTitle,
		MetaDesc:      req.MetaDescription,
		Template:      req.Template,
		Layout:        req.Layout,
		Status:        req.Status,
		Language:      req.Language,
		ParentID:      req.ParentID,
		SortOrder:     req.SortOrder,
		FeaturedImage: req.FeaturedImage,
		ExcerptText:   req.ExcerptText,
		AuthorID:      req.AuthorID,
		IsHomepage:    req.IsHomepage,
		IsLandingPage: req.IsLandingPage,
	}

	// Set JSON fields
	if len(req.SEOSettings) > 0 {
		if seoJSON, err := json.Marshal(req.SEOSettings); err == nil {
			page.SEOSettings = datatypes.JSON(seoJSON)
		}
	}

	if len(req.PageSettings) > 0 {
		if settingsJSON, err := json.Marshal(req.PageSettings); err == nil {
			page.PageSettings = datatypes.JSON(settingsJSON)
		}
	}

	if len(req.LayoutData) > 0 {
		if layoutJSON, err := json.Marshal(req.LayoutData); err == nil {
			page.LayoutData = datatypes.JSON(layoutJSON)
		}
	}

	// Create page
	if err := s.pageRepo.Create(page); err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Create content blocks if provided
	if len(req.ContentBlocks) > 0 {
		if err := s.createContentBlocks(page.ID, req.ContentBlocks); err != nil {
			return nil, fmt.Errorf("failed to create content blocks: %w", err)
		}
	}

	// Reload page with relations
	return s.pageRepo.GetByID(page.ID, true)
}

// GetPageByID retrieves a page by ID
func (s *PageService) GetPageByID(id uint, includeBlocks bool) (*models.Page, error) {
	page, err := s.pageRepo.GetByID(id, includeBlocks)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return page, nil
}

// GetPageBySlug retrieves a page by slug
func (s *PageService) GetPageBySlug(slug string, includeBlocks bool) (*models.Page, error) {
	page, err := s.pageRepo.GetBySlug(slug, includeBlocks)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return page, nil
}

// GetPages retrieves pages with pagination and filtering (basit implementasyon)
func (s *PageService) GetPages(page, limit int, filter PageFilter) (*PaginatedPagesResponse, error) {
	offset := (page - 1) * limit

	// Basit query oluştur
	query := s.db.Model(&models.Page{}).Preload("Author")

	// Filter uygula
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Template != "" {
		query = query.Where("template = ?", filter.Template)
	}
	if filter.Language != "" {
		query = query.Where("language = ?", filter.Language)
	}
	if filter.ParentID != nil {
		query = query.Where("parent_id = ?", *filter.ParentID)
	}
	if filter.Search != "" {
		query = query.Where("title ILIKE ? OR meta_desc ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Get pages
	var pages []models.Page
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pages).Error; err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &PaginatedPagesResponse{
		Pages:      pages,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdatePage updates an existing page
func (s *PageService) UpdatePage(id uint, req UpdatePageRequest) (*models.Page, error) {
	page, err := s.pageRepo.GetByID(id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update fields
	if req.Title != "" {
		page.Title = req.Title
	}
	if req.MetaTitle != "" {
		page.MetaTitle = req.MetaTitle
	}
	if req.MetaDescription != "" {
		page.MetaDesc = req.MetaDescription
	}
	if req.Template != "" {
		page.Template = req.Template
	}
	if req.Layout != "" {
		page.Layout = req.Layout
	}
	if req.Status != "" {
		page.Status = req.Status
	}
	if req.Language != "" {
		page.Language = req.Language
	}
	if req.ParentID != nil {
		page.ParentID = req.ParentID
	}
	if req.SortOrder != nil {
		page.SortOrder = *req.SortOrder
	}
	if req.FeaturedImage != "" {
		page.FeaturedImage = req.FeaturedImage
	}
	if req.ExcerptText != "" {
		page.ExcerptText = req.ExcerptText
	}

	// Update JSON fields
	if len(req.SEOSettings) > 0 {
		if seoJSON, err := json.Marshal(req.SEOSettings); err == nil {
			page.SEOSettings = datatypes.JSON(seoJSON)
		}
	}

	if len(req.PageSettings) > 0 {
		if settingsJSON, err := json.Marshal(req.PageSettings); err == nil {
			page.PageSettings = datatypes.JSON(settingsJSON)
		}
	}

	if len(req.LayoutData) > 0 {
		if layoutJSON, err := json.Marshal(req.LayoutData); err == nil {
			page.LayoutData = datatypes.JSON(layoutJSON)
		}
	}

	page.UpdatedAt = time.Now()

	if err := s.pageRepo.Update(page); err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}

	return page, nil
}

// DeletePage deletes a page and its content blocks
func (s *PageService) DeletePage(id uint) error {
	page, err := s.pageRepo.GetByID(id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.pageRepo.Delete(page.ID)
}

// PublishPage publishes a page
func (s *PageService) PublishPage(id uint) (*models.Page, error) {
	page, err := s.pageRepo.GetByID(id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	page.Status = "published"
	now := time.Now()
	page.PublishedAt = &now
	page.UpdatedAt = now

	if err := s.pageRepo.Update(page); err != nil {
		return nil, fmt.Errorf("failed to publish page: %w", err)
	}

	return page, nil
}

// UnpublishPage unpublishes a page
func (s *PageService) UnpublishPage(id uint) (*models.Page, error) {
	page, err := s.pageRepo.GetByID(id, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	page.Status = "draft"
	page.PublishedAt = nil
	page.UpdatedAt = time.Now()

	if err := s.pageRepo.Update(page); err != nil {
		return nil, fmt.Errorf("failed to unpublish page: %w", err)
	}

	return page, nil
}

// GetPageHierarchy retrieves the hierarchical structure of pages (basit implementasyon)
func (s *PageService) GetPageHierarchy(language, status string) ([]PageHierarchyNode, error) {
	query := s.db.Model(&models.Page{}).Preload("Author")

	if language != "" {
		query = query.Where("language = ?", language)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var pages []models.Page
	if err := query.Order("sort_order ASC, created_at ASC").Find(&pages).Error; err != nil {
		return nil, err
	}

	return s.buildHierarchy(pages, nil), nil
}

// DuplicatePage creates a copy of an existing page
func (s *PageService) DuplicatePage(id uint, req DuplicatePageRequest) (*models.Page, error) {
	originalPage, err := s.pageRepo.GetByID(id, req.IncludeBlocks)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Create new page
	newPage := &models.Page{
		Title:         req.NewTitle,
		Slug:          req.NewSlug,
		MetaTitle:     originalPage.MetaTitle,
		MetaDesc:      originalPage.MetaDesc,
		Template:      originalPage.Template,
		Layout:        originalPage.Layout,
		Status:        "draft",
		Language:      originalPage.Language,
		ParentID:      originalPage.ParentID,
		SortOrder:     originalPage.SortOrder,
		FeaturedImage: originalPage.FeaturedImage,
		ExcerptText:   originalPage.ExcerptText,
		AuthorID:      req.AuthorID,
		SEOSettings:   originalPage.SEOSettings,
		PageSettings:  originalPage.PageSettings,
		LayoutData:    originalPage.LayoutData,
	}

	if err := s.pageRepo.Create(newPage); err != nil {
		return nil, fmt.Errorf("failed to create duplicate page: %w", err)
	}

	// Copy content blocks if requested
	if req.IncludeBlocks && len(originalPage.ContentBlocks) > 0 {
		for _, block := range originalPage.ContentBlocks {
			newBlock := &models.PageContentBlock{
				PageID:         newPage.ID,
				ContainerID:    block.ContainerID,
				BlockType:      block.BlockType,
				Content:        block.Content,
				Settings:       block.Settings,
				Styles:         block.Styles,
				Position:       block.Position,
				IsVisible:      block.IsVisible,
				IsContainer:    block.IsContainer,
				ContainerType:  block.ContainerType,
				GridSettings:   block.GridSettings,
				ResponsiveData: block.ResponsiveData,
			}

			if err := s.blockRepo.Create(newBlock); err != nil {
				return nil, fmt.Errorf("failed to duplicate content block: %w", err)
			}
		}
	}

	// Reload page with relations
	return s.pageRepo.GetByID(newPage.ID, req.IncludeBlocks)
}

// IncrementViews increments the view count for a page
func (s *PageService) IncrementViews(id uint) error {
	return s.pageRepo.IncrementViews(id)
}

// Helper methods

func (s *PageService) validateCreatePageRequest(req CreatePageRequest) error {
	if len(req.Title) < 3 {
		return fmt.Errorf("title must be at least 3 characters")
	}

	if req.AuthorID == 0 {
		return fmt.Errorf("author ID is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft":     true,
		"published": true,
		"scheduled": true,
		"archived":  true,
	}
	if !validStatuses[req.Status] {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	return nil
}

func (s *PageService) createContentBlocks(pageID uint, blocks []CreatePageBlockRequest) error {
	for i, blockReq := range blocks {
		block := &models.PageContentBlock{
			PageID:        pageID,
			ContainerID:   blockReq.ContainerID,
			BlockType:     blockReq.BlockType,
			Content:       blockReq.Content,
			Position:      blockReq.Position,
			IsVisible:     blockReq.IsVisible,
			IsContainer:   blockReq.IsContainer,
			ContainerType: blockReq.ContainerType,
		}

		// Set position if not provided
		if block.Position == 0 {
			block.Position = i + 1
		}

		// Set default visibility
		if !blockReq.IsVisible && i == 0 {
			block.IsVisible = true
		}

		// Marshal JSON fields
		if len(blockReq.Settings) > 0 {
			if settingsJSON, err := json.Marshal(blockReq.Settings); err == nil {
				block.Settings = datatypes.JSON(settingsJSON)
			}
		}

		if len(blockReq.Styles) > 0 {
			if stylesJSON, err := json.Marshal(blockReq.Styles); err == nil {
				block.Styles = datatypes.JSON(stylesJSON)
			}
		}

		if len(blockReq.GridSettings) > 0 {
			if gridJSON, err := json.Marshal(blockReq.GridSettings); err == nil {
				block.GridSettings = datatypes.JSON(gridJSON)
			}
		}

		if len(blockReq.ResponsiveData) > 0 {
			if responsiveJSON, err := json.Marshal(blockReq.ResponsiveData); err == nil {
				block.ResponsiveData = datatypes.JSON(responsiveJSON)
			}
		}

		if err := s.blockRepo.Create(block); err != nil {
			return fmt.Errorf("failed to create content block %d: %w", i, err)
		}
	}

	return nil
}

func (s *PageService) buildHierarchy(pages []models.Page, parentID *uint) []PageHierarchyNode {
	var nodes []PageHierarchyNode

	for _, page := range pages {
		if (parentID == nil && page.ParentID == nil) || (parentID != nil && page.ParentID != nil && *page.ParentID == *parentID) {
			node := PageHierarchyNode{
				Page:     page,
				Children: s.buildHierarchy(pages, &page.ID),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}
