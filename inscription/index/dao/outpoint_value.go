package dao

import (
	"github.com/inscription-c/insc/inscription/index/tables"
)

func (d *DB) GetValueByOutpoint(outpoint string) (value int64, err error) {
	outpointVal := &tables.OutpointValue{}
	if err = d.DB.Where("outpoint = ?", outpoint).First(outpointVal).Error; err != nil {
		return
	}
	return outpointVal.Value, nil
}

func (d *DB) DeleteValueByOutpoint(outpoints ...string) (err error) {
	if len(outpoints) == 0 {
		return
	}
	err = d.DB.Where("outpoint in (?)", outpoints).Delete(&tables.OutpointValue{}).Error
	return
}

func (d *DB) SetOutpointToValue(values map[string]int64) (err error) {
	list := make([]*tables.OutpointValue, 0, len(values))
	for outpoint, value := range values {
		list = append(list, &tables.OutpointValue{
			Outpoint: outpoint,
			Value:    value,
		})
	}
	return d.DB.CreateInBatches(&list, 10_000).Error
}
