package services

import (
	"encoding/json"
	"errors"

	"news/internal/models"
	"news/internal/repositories"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PageContentBlockService handles business logic for page content blocks
type PageContentBlockService struct {
	blockRepo *repositories.PageContentBlockRepository
	pageRepo  *repositories.PageRepository
	db        *gorm.DB
}

// NewPageContentBlockService creates a new page content block service
func NewPageContentBlockService(db *gorm.DB) *PageContentBlockService {
	return &PageContentBlockService{
		blockRepo: repositories.NewPageContentBlockRepository(db),
		pageRepo:  repositories.NewPageRepository(db),
		db:        db,
	}
}

// CreateBlock creates a new page content block
func (s *PageContentBlockService) CreateBlock(pageID uint, req CreatePageBlockRequest) (*models.PageContentBlock, error) {
	// Validate page exists
	if _, err := s.pageRepo.GetByID(pageID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	block := &models.PageContentBlock{
		PageID:        pageID,
		ContainerID:   req.ContainerID,
		BlockType:     req.BlockType,
		Content:       req.Content,
		Position:      req.Position,
		IsVisible:     req.IsVisible,
		IsContainer:   req.IsContainer,
		ContainerType: req.ContainerType,
	}

	// Set JSON fields
	if len(req.Settings) > 0 {
		if settingsJSON, err := json.Marshal(req.Settings); err == nil {
			block.Settings = datatypes.JSON(settingsJSON)
		}
	}

	if len(req.Styles) > 0 {
		if stylesJSON, err := json.Marshal(req.Styles); err == nil {
			block.Styles = datatypes.JSON(stylesJSON)
		}
	}

	if len(req.GridSettings) > 0 {
		if gridJSON, err := json.Marshal(req.GridSettings); err == nil {
			block.GridSettings = datatypes.JSON(gridJSON)
		}
	}

	if len(req.ResponsiveData) > 0 {
		if responsiveJSON, err := json.Marshal(req.ResponsiveData); err == nil {
			block.ResponsiveData = datatypes.JSON(responsiveJSON)
		}
	}

	if err := s.blockRepo.Create(block); err != nil {
		return nil, err
	}

	return block, nil
}

// GetBlockByID retrieves a page content block by ID
func (s *PageContentBlockService) GetBlockByID(id uint) (*models.PageContentBlock, error) {
	block, err := s.blockRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return block, nil
}

// UpdateBlock updates an existing page content block
func (s *PageContentBlockService) UpdateBlock(id uint, req UpdatePageBlockRequest) (*models.PageContentBlock, error) {
	block, err := s.blockRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Update fields
	if req.BlockType != "" {
		block.BlockType = req.BlockType
	}
	if req.Content != "" {
		block.Content = req.Content
	}
	if req.Position != nil {
		block.Position = *req.Position
	}
	if req.IsVisible != nil {
		block.IsVisible = *req.IsVisible
	}
	if req.IsContainer != nil {
		block.IsContainer = *req.IsContainer
	}
	if req.ContainerType != "" {
		block.ContainerType = req.ContainerType
	}

	// Update JSON fields
	if len(req.Settings) > 0 {
		if settingsJSON, err := json.Marshal(req.Settings); err == nil {
			block.Settings = datatypes.JSON(settingsJSON)
		}
	}

	if len(req.Styles) > 0 {
		if stylesJSON, err := json.Marshal(req.Styles); err == nil {
			block.Styles = datatypes.JSON(stylesJSON)
		}
	}

	if len(req.GridSettings) > 0 {
		if gridJSON, err := json.Marshal(req.GridSettings); err == nil {
			block.GridSettings = datatypes.JSON(gridJSON)
		}
	}

	if len(req.ResponsiveData) > 0 {
		if responsiveJSON, err := json.Marshal(req.ResponsiveData); err == nil {
			block.ResponsiveData = datatypes.JSON(responsiveJSON)
		}
	}

	if err := s.blockRepo.Update(block); err != nil {
		return nil, err
	}

	return block, nil
}

// DeleteBlock deletes a page content block
func (s *PageContentBlockService) DeleteBlock(id uint) error {
	block, err := s.blockRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	return s.blockRepo.Delete(block.ID)
}

// ReorderBlocks reorders page content blocks
func (s *PageContentBlockService) ReorderBlocks(pageID uint, req ReorderBlocksRequest) error {
	// Validate page exists
	if _, err := s.pageRepo.GetByID(pageID, false); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, order := range req.BlockOrders {
			if err := tx.Model(&models.PageContentBlock{}).
				Where("id = ? AND page_id = ?", order.BlockID, pageID).
				Update("position", order.Position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DuplicateBlock duplicates a page content block
func (s *PageContentBlockService) DuplicateBlock(id uint, req DuplicateBlockRequest) (*models.PageContentBlock, error) {
	// Get original block
	original, err := s.blockRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Create new block
	newBlock := &models.PageContentBlock{
		PageID:         original.PageID,
		ContainerID:    original.ContainerID,
		BlockType:      original.BlockType,
		Content:        original.Content,
		Settings:       original.Settings,
		Styles:         original.Styles,
		IsVisible:      original.IsVisible,
		IsContainer:    original.IsContainer,
		ContainerType:  original.ContainerType,
		GridSettings:   original.GridSettings,
		ResponsiveData: original.ResponsiveData,
	}

	// Override with request parameters
	if req.TargetPageID != nil {
		newBlock.PageID = *req.TargetPageID
	}
	if req.TargetContainerID != nil {
		newBlock.ContainerID = req.TargetContainerID
	}
	if req.NewPosition != nil {
		newBlock.Position = *req.NewPosition
	} else {
		newBlock.Position = original.Position + 1
	}

	if err := s.blockRepo.Create(newBlock); err != nil {
		return nil, err
	}

	return newBlock, nil
}

// GetBlocksByPageID retrieves all content blocks for a page
func (s *PageContentBlockService) GetBlocksByPageID(pageID uint) ([]models.PageContentBlock, error) {
	return s.blockRepo.GetByPageID(pageID, false) // Only return visible blocks by default
}

// ValidateBlock validates a block configuration
func (s *PageContentBlockService) ValidateBlock(req CreatePageBlockRequest) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	// Validate required fields
	if req.BlockType == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "Block type is required")
	}

	// Validate block type
	validTypes := []string{"text", "image", "video", "gallery", "code", "quote", "list", "table", "divider", "spacer", "button", "form", "map", "social", "container", "row", "column"}
	isValidType := false
	for _, vt := range validTypes {
		if req.BlockType == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid block type")
	}

	// Validate container type if it's a container
	if req.IsContainer && req.ContainerType == "" {
		result.Warnings = append(result.Warnings, "Container type not specified, using default")
	}

	// Validate position
	if req.Position < 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Position must be non-negative")
	}

	return result
}
