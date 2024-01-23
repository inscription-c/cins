package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) BlockHash(height ...uint32) (blockHash string, err error) {
	blockInfo := &tables.BlockInfo{}
	if len(height) == 0 {
		if err = d.DB.Last(blockInfo).Error; err != nil {
			return
		}
	} else {
		if err = d.DB.Where("height = ?", height[0]).First(blockInfo).Error; err != nil {
			return
		}
	}

	header, err := blockInfo.LoadHeader()
	if err != nil {
		return "", err
	}
	blockHash = header.BlockHash().String()
	return
}

func (d *DB) BlockCount() (count uint32, err error) {
	block := &tables.BlockInfo{}
	err = d.DB.Order("id desc").First(block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	count = block.Height + 1
	return
}

func (d *DB) SaveBlockInfo(block *tables.BlockInfo) error {
	old := &tables.BlockInfo{}
	err := d.DB.Where("height=?", block.Height).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		block.Id = old.Id
	}
	return d.DB.Save(block).Error
}
