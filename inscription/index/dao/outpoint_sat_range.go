package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SetOutpointToSatRange sets the satoshi range for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding satoshi ranges.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToSatRange(rangeCache map[string][]*model.SatRange) (err error) {
	list := make([]*tables.OutpointSatRange, 0, len(rangeCache))

	for outpoint, satRanges := range rangeCache {
		for _, satRange := range satRanges {
			list = append(list, &tables.OutpointSatRange{
				Outpoint: outpoint,
				Start:    satRange.Start,
				End:      satRange.End,
			})
		}
	}
	return d.DB.CreateInBatches(&list, 10_000).Error
}

// OutpointToSatRanges returns the satoshi ranges for a given outpoint.
func (d *DB) OutpointToSatRanges(outpoint string) ([]*model.SatRange, error) {
	var list []*tables.OutpointSatRange
	err := d.DB.Where("outpoint = ?", outpoint).Find(&list).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return nil, err
	}

	ranges := make([]*model.SatRange, 0, len(list))
	for _, v := range list {
		ranges = append(ranges, &model.SatRange{
			Start: v.Start,
			End:   v.End,
		})
	}
	return ranges, nil
}

// DelSatPointByOutpoint deletes the satoshi range for a given outpoint.
func (d *DB) DelSatPointByOutpoint(outpoint string) ([]*model.SatRange, error) {
	var list []*tables.OutpointSatRange
	err := d.DB.Clauses(clause.Returning{}).Where("outpoint = ?", outpoint).Delete(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	result := make([]*model.SatRange, 0, len(list))
	if len(list) == 0 {
		return result, nil
	}
	for _, v := range list {
		result = append(result, &model.SatRange{
			Start: v.Start,
			End:   v.End,
		})
	}
	return result, nil
}
