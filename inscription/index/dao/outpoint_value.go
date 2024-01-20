package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
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
	entity := &tables.OutpointValue{}
	err = d.DB.Where("outpoint=?", outpoint).First(entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		entity.Outpoint = outpoint
		entity.Value = value
		err = d.DB.Create(entity).Error
		return
	}
	return d.DB.Save(entity).Error
}
