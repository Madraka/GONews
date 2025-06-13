package repositories

import (
	"encoding/json"
	"fmt"
	"news/internal/database"
	"news/internal/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ArticleContentBlockRepository handles database operations for article content blocks
type ArticleContentBlockRepository struct {
	db *gorm.DB
}

// NewArticleContentBlockRepository creates a new instance of ArticleContentBlockRepository
func NewArticleContentBlockRepository(db *gorm.DB) *ArticleContentBlockRepository {
	return &ArticleContentBlockRepository{db: db}
}

// CreateBlock creates a new content block
func (r *ArticleContentBlockRepository) CreateBlock(block *models.ArticleContentBlock) (*models.ArticleContentBlock, error) {
	// Validate block type
	if !block.ValidateBlockType() {
		return nil, fmt.Errorf("invalid block type: %s", block.BlockType)
	}

	// Set default settings if empty
	if len(block.Settings) == 0 {
		defaultSettings := block.GetDefaultSettings()
		if settingsJSON, err := json.Marshal(defaultSettings); err == nil {
			block.Settings = datatypes.JSON(settingsJSON)
		}
	}

	// Create block
	if err := r.db.Create(block).Error; err != nil {
		return nil, err
	}

	// Load the created block with relationships
	var createdBlock models.ArticleContentBlock
	err := r.db.Preload("Article").First(&createdBlock, block.ID).Error
	if err != nil {
		return nil, err
	}

	return &createdBlock, nil
}

// GetBlockByID retrieves a content block by its ID
func (r *ArticleContentBlockRepository) GetBlockByID(id uint) (*models.ArticleContentBlock, error) {
	var block models.ArticleContentBlock
	err := r.db.Preload("Article").First(&block, id).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetBlocksByArticleID retrieves all content blocks for an article
func (r *ArticleContentBlockRepository) GetBlocksByArticleID(articleID uint) ([]models.ArticleContentBlock, error) {
	var blocks []models.ArticleContentBlock
	err := r.db.Where("article_id = ?", articleID).
		Order("position ASC").
		Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// GetVisibleBlocksByArticleID retrieves only visible content blocks for an article
func (r *ArticleContentBlockRepository) GetVisibleBlocksByArticleID(articleID uint) ([]models.ArticleContentBlock, error) {
	var blocks []models.ArticleContentBlock
	err := r.db.Where("article_id = ? AND is_visible = ?", articleID, true).
		Order("position ASC").
		Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// UpdateBlock updates an existing content block
func (r *ArticleContentBlockRepository) UpdateBlock(blockID uint, updateData map[string]interface{}) (*models.ArticleContentBlock, error) {
	// Get existing block
	var block models.ArticleContentBlock
	if err := r.db.First(&block, blockID).Error; err != nil {
		return nil, err
	}

	// Update block
	if err := r.db.Model(&block).Updates(updateData).Error; err != nil {
		return nil, err
	}

	// Return updated block with relationships
	var updatedBlock models.ArticleContentBlock
	err := r.db.Preload("Article").First(&updatedBlock, blockID).Error
	if err != nil {
		return nil, err
	}

	return &updatedBlock, nil
}

// UpdateBlockContent updates only the content of a block
func (r *ArticleContentBlockRepository) UpdateBlockContent(blockID uint, content string) error {
	return r.db.Model(&models.ArticleContentBlock{}).
		Where("id = ?", blockID).
		Update("content", content).Error
}

// UpdateBlockSettings updates only the settings of a block
func (r *ArticleContentBlockRepository) UpdateBlockSettings(blockID uint, settings models.ArticleContentBlockSettings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return r.db.Model(&models.ArticleContentBlock{}).
		Where("id = ?", blockID).
		Update("settings", datatypes.JSON(settingsJSON)).Error
}

// UpdateBlockPosition updates the position of a block
func (r *ArticleContentBlockRepository) UpdateBlockPosition(blockID uint, position int) error {
	return r.db.Model(&models.ArticleContentBlock{}).
		Where("id = ?", blockID).
		Update("position", position).Error
}

// UpdateBlockVisibility updates the visibility of a block
func (r *ArticleContentBlockRepository) UpdateBlockVisibility(blockID uint, isVisible bool) error {
	return r.db.Model(&models.ArticleContentBlock{}).
		Where("id = ?", blockID).
		Update("is_visible", isVisible).Error
}

// DeleteBlock deletes a content block
func (r *ArticleContentBlockRepository) DeleteBlock(blockID uint) error {
	return r.db.Delete(&models.ArticleContentBlock{}, blockID).Error
}

// DeleteBlocksByArticleID deletes all content blocks for an article
func (r *ArticleContentBlockRepository) DeleteBlocksByArticleID(articleID uint) error {
	return r.db.Where("article_id = ?", articleID).Delete(&models.ArticleContentBlock{}).Error
}

// ReorderBlocks updates the position of multiple blocks in one transaction
func (r *ArticleContentBlockRepository) ReorderBlocks(articleID uint, blockPositions map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for blockID, position := range blockPositions {
			if err := tx.Model(&models.ArticleContentBlock{}).
				Where("id = ? AND article_id = ?", blockID, articleID).
				Update("position", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DuplicateBlock creates a copy of an existing block
func (r *ArticleContentBlockRepository) DuplicateBlock(blockID uint, newPosition int) (*models.ArticleContentBlock, error) {
	// Get original block
	var originalBlock models.ArticleContentBlock
	if err := r.db.First(&originalBlock, blockID).Error; err != nil {
		return nil, err
	}

	// Create new block with same data but different position
	newBlock := models.ArticleContentBlock{
		ArticleID: originalBlock.ArticleID,
		BlockType: originalBlock.BlockType,
		Content:   originalBlock.Content,
		Settings:  originalBlock.Settings,
		Position:  newPosition,
		IsVisible: originalBlock.IsVisible,
	}

	if err := r.db.Create(&newBlock).Error; err != nil {
		return nil, err
	}

	return &newBlock, nil
}

// GetBlocksByType retrieves all blocks of a specific type for an article
func (r *ArticleContentBlockRepository) GetBlocksByType(articleID uint, blockType string) ([]models.ArticleContentBlock, error) {
	var blocks []models.ArticleContentBlock
	err := r.db.Where("article_id = ? AND block_type = ?", articleID, blockType).
		Order("position ASC").
		Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

// CountBlocksByArticleID counts the total number of blocks for an article
func (r *ArticleContentBlockRepository) CountBlocksByArticleID(articleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ArticleContentBlock{}).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, err
}

// CountVisibleBlocksByArticleID counts the visible blocks for an article
func (r *ArticleContentBlockRepository) CountVisibleBlocksByArticleID(articleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ArticleContentBlock{}).
		Where("article_id = ? AND is_visible = ?", articleID, true).
		Count(&count).Error
	return count, err
}

// GetLastPositionForArticle gets the highest position number for an article
func (r *ArticleContentBlockRepository) GetLastPositionForArticle(articleID uint) (int, error) {
	var maxPosition int
	err := r.db.Model(&models.ArticleContentBlock{}).
		Where("article_id = ?", articleID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPosition).Error
	return maxPosition, err
}

// BulkCreateBlocks creates multiple blocks in a single transaction
func (r *ArticleContentBlockRepository) BulkCreateBlocks(blocks []models.ArticleContentBlock) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range blocks {
			// Validate each block
			if !blocks[i].ValidateBlockType() {
				return fmt.Errorf("invalid block type: %s", blocks[i].BlockType)
			}

			// Set default settings if empty
			if len(blocks[i].Settings) == 0 {
				defaultSettings := blocks[i].GetDefaultSettings()
				if settingsJSON, err := json.Marshal(defaultSettings); err == nil {
					blocks[i].Settings = datatypes.JSON(settingsJSON)
				}
			}

			if err := tx.Create(&blocks[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// MigrateArticleToBlocks converts legacy article content to content blocks
func (r *ArticleContentBlockRepository) MigrateArticleToBlocks(articleID uint) error {
	// Get the article
	var article models.Article
	if err := r.db.First(&article, articleID).Error; err != nil {
		return err
	}

	// Skip if already using blocks
	if article.IsUsingBlocks() {
		return nil
	}

	// Generate blocks from legacy content
	blocks := article.MigrateToBlocks()
	if len(blocks) == 0 {
		return nil
	}

	// Start transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the blocks
		for i := range blocks {
			if err := tx.Create(&blocks[i]).Error; err != nil {
				return err
			}
		}

		// Update article to use blocks
		updateData := map[string]interface{}{
			"content_type":   "blocks",
			"has_blocks":     true,
			"blocks_version": 1,
		}

		if err := tx.Model(&article).Updates(updateData).Error; err != nil {
			return err
		}

		return nil
	})
}

// Global repository instance
var ArticleContentBlockRepo *ArticleContentBlockRepository

// InitializeArticleContentBlockRepository initializes the global repository instance
func InitializeArticleContentBlockRepository() {
	ArticleContentBlockRepo = NewArticleContentBlockRepository(database.DB)
}
