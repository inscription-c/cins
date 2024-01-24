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

func (d *DB) DeleteValueByOutpoint(outpoint string) (err error) {
	err = d.DB.Where("outpoint = ?", outpoint).Delete(&tables.OutpointValue{}).Error
	return
}

func (d *DB) SetOutpointToValue(outpoint string, value int64) (err error) {
	entity := &tables.OutpointValue{
		Outpoint: outpoint,
		Value:    value,
	}
	return d.DB.Where("outpoint=?", outpoint).
		Assign(tables.OutpointValue{
			Outpoint: outpoint,
			Value:    value,
		}).
		FirstOrCreate(entity).Error
}
