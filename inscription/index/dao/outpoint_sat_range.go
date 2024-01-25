package dao

import (
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
)

func (d *DB) SetOutpointToSatRange(rangeCache map[string]model.SatRange) (err error) {
	list := make([]*tables.OutpointSatRange, 0, len(rangeCache))
	for outpoint, satRange := range rangeCache {
		list = append(list, &tables.OutpointSatRange{
			Outpoint: outpoint,
			Start:    satRange.Start,
			End:      satRange.End,
		})
	}
	return d.DB.Create(&list).Error
}
