package dao

import (
	"encoding/hex"
	"errors"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
	"strings"
)

// SetOutpointToSatRange sets the satoshi range for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding satoshi ranges.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToSatRange(height uint32, satRanges ...*tables.OutpointSatRange) (err error) {
	if err := d.CreateInBatches(&satRanges, 10_000).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Height: height,
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Delete(&satRanges)
		}),
	}).Error
}

// OutpointToSatRanges returns the satoshi ranges for a given outpoint.
func (d *DB) OutpointToSatRanges(outpoint string) (satRange tables.OutpointSatRange, err error) {
	err = d.Where("outpoint = ?", outpoint).First(&satRange).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// DelSatRangesByOutpoint deletes the satoshi ranges for a given outpoint.
func (d *DB) DelSatRangesByOutpoint(height uint32, outpoint string) (satRange tables.OutpointSatRange, err error) {
	satRange, err = d.OutpointToSatRanges(outpoint)
	if err != nil {
		return
	}
	if satRange.Id == 0 {
		return
	}

	if err = d.Delete(&satRange).Error; err != nil {
		return
	}

	sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Create(&satRange)
	})
	sql = strings.ReplaceAll(sql, "<binary>", "0x"+hex.EncodeToString(satRange.SatRange))
	err = d.Create(&tables.UndoLog{
		Height: height,
		Sql:    sql,
	}).Error
	return
}
