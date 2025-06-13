package repositories

import (
	"errors"
	"fmt"

	"news/internal/models"

	"gorm.io/gorm"
)

// PageContentBlockRepository handles database operations for page content blocks
type PageContentBlockRepository struct {
	db *gorm.DB
}

// NewPageContentBlockRepository creates a new page content block repository
func NewPageContentBlockRepository(db *gorm.DB) *PageContentBlockRepository {
	return &PageContentBlockRepository{db: db}
}

// Create creates a new content block
func (r *PageContentBlockRepository) Create(block *models.PageContentBlock) error {
	if err := r.validateBlock(block); err != nil {
		return err
	}

	// Set position if not provided
	if block.Position == 0 {
		block.Position = r.getNextPosition(block.PageID, block.ContainerID)
	}

	return r.db.Create(block).Error
}

// GetByID retrieves a content block by ID
func (r *PageContentBlockRepository) GetByID(id uint) (*models.PageContentBlock, error) {
	var block models.PageContentBlock
	if err := r.db.Preload("ChildBlocks", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_visible = ?", true).Order("position ASC")
	}).First(&block, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("content block with ID %d not found", id)
		}
		return nil, err
	}

	return &block, nil
}

// GetByPageID retrieves all content blocks for a page
func (r *PageContentBlockRepository) GetByPageID(pageID uint, includeHidden bool) ([]models.PageContentBlock, error) {
	var blocks []models.PageContentBlock
	query := r.db.Where("page_id = ? AND container_id IS NULL", pageID)

	if !includeHidden {
		query = query.Where("is_visible = ?", true)
	}

	err := query.Preload("ChildBlocks", func(db *gorm.DB) *gorm.DB {
		childQuery := db.Order("position ASC")
		if !includeHidden {
			childQuery = childQuery.Where("is_visible = ?", true)
		}
		return childQuery
	}).Order("position ASC").Find(&blocks).Error

	return blocks, err
}

// GetByContainerID retrieves all child blocks for a container
func (r *PageContentBlockRepository) GetByContainerID(containerID uint, includeHidden bool) ([]models.PageContentBlock, error) {
	var blocks []models.PageContentBlock
	query := r.db.Where("container_id = ?", containerID)

	if !includeHidden {
		query = query.Where("is_visible = ?", true)
	}

	err := query.Order("position ASC").Find(&blocks).Error
	return blocks, err
}

// Update updates a content block
func (r *PageContentBlockRepository) Update(block *models.PageContentBlock) error {
	if err := r.validateBlock(block); err != nil {
		return err
	}

	return r.db.Save(block).Error
}

// Delete soft deletes a content block and its children
func (r *PageContentBlockRepository) Delete(id uint) error {
	// First, delete all child blocks
	if err := r.db.Where("container_id = ?", id).Delete(&models.PageContentBlock{}).Error; err != nil {
		return err
	}

	// Then delete the block itself
	return r.db.Delete(&models.PageContentBlock{}, id).Error
}

// UpdatePosition updates the position of a content block
func (r *PageContentBlockRepository) UpdatePosition(id uint, newPosition int) error {
	return r.db.Model(&models.PageContentBlock{}).
		Where("id = ?", id).
		Update("position", newPosition).Error
}

// UpdateVisibility updates the visibility of a content block
func (r *PageContentBlockRepository) UpdateVisibility(id uint, isVisible bool) error {
	return r.db.Model(&models.PageContentBlock{}).
		Where("id = ?", id).
		Update("is_visible", isVisible).Error
}

// ReorderBlocks reorders content blocks within a container or page
func (r *PageContentBlockRepository) ReorderBlocks(pageID uint, containerID *uint, blockIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, blockID := range blockIDs {
			var updates map[string]interface{}
			if containerID != nil {
				updates = map[string]interface{}{
					"position":     i + 1,
					"container_id": *containerID,
				}
			} else {
				updates = map[string]interface{}{
					"position":     i + 1,
					"container_id": nil,
				}
			}

			if err := tx.Model(&models.PageContentBlock{}).
				Where("id = ? AND page_id = ?", blockID, pageID).
				Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkCreate creates multiple content blocks in a transaction
func (r *PageContentBlockRepository) BulkCreate(blocks []models.PageContentBlock) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range blocks {
			if err := r.validateBlock(&blocks[i]); err != nil {
				return err
			}

			if blocks[i].Position == 0 {
				blocks[i].Position = r.getNextPosition(blocks[i].PageID, blocks[i].ContainerID)
			}
		}

		return tx.Create(&blocks).Error
	})
}

// BulkUpdate updates multiple content blocks in a transaction
func (r *PageContentBlockRepository) BulkUpdate(blocks []models.PageContentBlock) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, block := range blocks {
			if err := r.validateBlock(&block); err != nil {
				return err
			}

			if err := tx.Save(&block).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DuplicateBlock duplicates a content block and its children
func (r *PageContentBlockRepository) DuplicateBlock(blockID uint, newPageID uint, newContainerID *uint) (*models.PageContentBlock, error) {
	var originalBlock models.PageContentBlock
	if err := r.db.Preload("ChildBlocks").First(&originalBlock, blockID).Error; err != nil {
		return nil, err
	}

	var newBlock models.PageContentBlock
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Duplicate the main block
		newBlock = originalBlock
		newBlock.ID = 0
		newBlock.PageID = newPageID
		newBlock.ContainerID = newContainerID
		newBlock.Position = r.getNextPosition(newPageID, newContainerID)

		if err := tx.Create(&newBlock).Error; err != nil {
			return err
		}

		// Duplicate child blocks if it's a container
		if originalBlock.IsContainer && len(originalBlock.ChildBlocks) > 0 {
			for _, childBlock := range originalBlock.ChildBlocks {
				newChildBlock := childBlock
				newChildBlock.ID = 0
				newChildBlock.PageID = newPageID
				newChildBlock.ContainerID = &newBlock.ID

				if err := tx.Create(&newChildBlock).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	return &newBlock, err
}

// GetBlocksByType retrieves all blocks of a specific type for a page
func (r *PageContentBlockRepository) GetBlocksByType(pageID uint, blockType string) ([]models.PageContentBlock, error) {
	var blocks []models.PageContentBlock
	err := r.db.Where("page_id = ? AND block_type = ? AND is_visible = ?", pageID, blockType, true).
		Order("position ASC").
		Find(&blocks).Error

	return blocks, err
}

// GetContainerBlocks retrieves all container blocks for a page
func (r *PageContentBlockRepository) GetContainerBlocks(pageID uint) ([]models.PageContentBlock, error) {
	var blocks []models.PageContentBlock
	err := r.db.Where("page_id = ? AND is_container = ? AND is_visible = ?", pageID, true, true).
		Preload("ChildBlocks", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("position ASC")
		}).
		Order("position ASC").
		Find(&blocks).Error

	return blocks, err
}

// SearchBlocks searches blocks by content
func (r *PageContentBlockRepository) SearchBlocks(pageID uint, searchTerm string) ([]models.PageContentBlock, error) {
	var blocks []models.PageContentBlock
	searchPattern := "%" + searchTerm + "%"

	err := r.db.Where("page_id = ? AND is_visible = ? AND content LIKE ?", pageID, true, searchPattern).
		Order("position ASC").
		Find(&blocks).Error

	return blocks, err
}

// GetBlockAnalytics retrieves performance data for blocks
func (r *PageContentBlockRepository) GetBlockAnalytics(pageID uint) ([]BlockAnalytics, error) {
	var analytics []BlockAnalytics

	err := r.db.Model(&models.PageContentBlock{}).
		Select("block_type, COUNT(*) as count, AVG(CASE WHEN performance_data != '' THEN 1 ELSE 0 END) as avg_performance").
		Where("page_id = ? AND is_visible = ?", pageID, true).
		Group("block_type").
		Scan(&analytics).Error

	return analytics, err
}

// Helper methods

// validateBlock validates content block data
func (r *PageContentBlockRepository) validateBlock(block *models.PageContentBlock) error {
	if block.PageID == 0 {
		return errors.New("page ID is required")
	}

	if block.BlockType == "" {
		return errors.New("block type is required")
	}

	if !block.ValidateBlockType() {
		return fmt.Errorf("invalid block type: %s", block.BlockType)
	}

	return nil
}

// getNextPosition gets the next position for a block in a container or page
func (r *PageContentBlockRepository) getNextPosition(pageID uint, containerID *uint) int {
	var maxPosition int
	query := r.db.Model(&models.PageContentBlock{}).
		Select("COALESCE(MAX(position), 0)").
		Where("page_id = ?", pageID)

	if containerID != nil {
		query = query.Where("container_id = ?", *containerID)
	} else {
		query = query.Where("container_id IS NULL")
	}

	query.Scan(&maxPosition)
	return maxPosition + 1
}

// BlockAnalytics represents analytics data for blocks
type BlockAnalytics struct {
	BlockType      string  `json:"block_type"`
	Count          int     `json:"count"`
	AvgPerformance float64 `json:"avg_performance"`
}
