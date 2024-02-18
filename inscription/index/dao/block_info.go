package dao

import (
	"errors"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockHeader retrieves the block header for a given block height.
func (d *DB) BlockHeader() (height uint32, header *wire.BlockHeader, err error) {
	blockInfo := &tables.BlockInfo{}
	// Retrieve the last block if no height is provided
	if err = d.DB.Last(blockInfo).Error; err != nil {
		return
	}
	height = blockInfo.Height

	// Load the block header
	header, err = blockInfo.LoadHeader()
	if err != nil {
		return
	}
	return
}

// BlockHash retrieves the block hash for a given block height.
// If no height is provided, it retrieves the hash for the last block.
// It returns the block hash as a string and any error encountered.
func (d *DB) BlockHash(height ...uint32) (blockHash string, err error) {
	blockInfo := &tables.BlockInfo{}
	if len(height) == 0 {
		if err = d.Last(blockInfo).Error; err != nil {
			return
		}
	} else {
		if err = d.Where("height = ?", height[0]).First(blockInfo).Error; err != nil {
			return
		}
	}

	// Load the block header
	header, err := blockInfo.LoadHeader()
	if err != nil {
		return "", err
	}
	blockHash = header.BlockHash().String()
	return
}

// BlockHeight retrieves the height of the last block in the database.
func (d *DB) BlockHeight() (height uint32, err error) {
	block := &tables.BlockInfo{}
	err = d.DB.Order("id desc").First(block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	height = block.Height
	return
}

// BlockCount retrieves the total number of blocks in the database.
// It returns the count as an uint32 and any error encountered.
func (d *DB) BlockCount() (count uint32, err error) {
	block := &tables.BlockInfo{}
	// Retrieve the last block
	err = d.DB.Order("id desc").First(block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	// The count is the height of the last block plus one
	count = block.Height + 1
	return
}

// SaveBlockInfo saves a block info to the database.
// If a block with the same height already exists, it updates the existing record.
// It returns any error encountered.
func (d *DB) SaveBlockInfo(block *tables.BlockInfo) error {
	return d.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "height"}},
		DoUpdates: clause.AssignmentColumns([]string{"sequence_num", "header", "timestamp"}),
	}).Create(block).Error
}

// DeleteBlockInfoByHeight deletes a block info from the database by height.
func (d *DB) DeleteBlockInfoByHeight(height uint32) (info tables.BlockInfo, err error) {
	err = d.Clauses(clause.Returning{}).Where("height = ?", height).Delete(&info).Error
	return
}
