package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

// SetOutpointToSatRange sets the satoshi range for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding satoshi ranges.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToSatRange(list []*tables.OutpointSatRange) (err error) {
	return d.CreateInBatches(&list, 10_000).Error
}

type OutpointToSatRangesResp struct {
	List []*tables.OutpointSatRange
	Err  error
}

// OutpointToSatRanges returns the satoshi ranges for a given outpoint.
func (d *DB) OutpointToSatRanges(outpoint string) (list []*tables.OutpointSatRange, err error) {
	err = d.Where("outpoint = ?", outpoint).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// DelSatRangesByOutpoint deletes the satoshi ranges for a given outpoint.
func (d *DB) DelSatRangesByOutpoint(outpoint string) (list []*tables.OutpointSatRange, err error) {
	list, err = d.OutpointToSatRanges(outpoint)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return
	}
	err = d.Where("outpoint = ?", outpoint).Delete(&tables.OutpointSatRange{}).Error
	return
}
