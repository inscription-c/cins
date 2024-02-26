package dao

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
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
	old := &tables.BlockInfo{}
	err := d.Where("height = ?", block.Height).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if old.Id > 0 {
		block.Id = old.Id
		block.CreatedAt = old.CreatedAt
		if err := d.Save(block).Error; err != nil {
			return err
		}
		sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Save(old)
		})
		return d.AddUndoLog(block.Height, sql)
	}
	if err := d.Create(block).Error; err != nil {
		return err
	}
	sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Delete(block)
	})
	return d.AddUndoLog(block.Height, sql)
}

// DeleteBlockInfoByHeight deletes a block info from the database by height.
func (d *DB) DeleteBlockInfoByHeight(height uint32) (info tables.BlockInfo, err error) {
	if err = d.Clauses(clause.Returning{}).Where("height = ?", height).Delete(&info).Error; err != nil {
		return
	}
	if info.Id > 0 {
		sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Create(&info)
		})
		sql = strings.ReplaceAll(sql, "<binary>", "0x"+hex.EncodeToString(info.Header))
		err = d.AddUndoLog(height, sql)
	}
	return
}
