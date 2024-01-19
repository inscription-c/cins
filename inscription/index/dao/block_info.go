package dao

import (
	"github.com/inscription-c/insc/inscription/index/tables"
)

func (d *DB) BlockHash(height ...uint64) (blockHash string, err error) {
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
