package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SetOutpointToSatRange sets the satoshi range for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding satoshi ranges.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToSatRange(satRanges ...*tables.OutpointSatRange) (err error) {
	return d.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "outpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"sat_range"}),
	}).CreateInBatches(&satRanges, 10_000).Error
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
func (d *DB) DelSatRangesByOutpoint(outpoint string) (satRange tables.OutpointSatRange, err error) {
	satRange, err = d.OutpointToSatRanges(outpoint)
	if err != nil {
		return
	}
	err = d.Where("outpoint = ?", outpoint).Delete(&tables.OutpointSatRange{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
