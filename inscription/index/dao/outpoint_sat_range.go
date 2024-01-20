package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) SetOutpointToSatRange(outpoint string, satRange model.SatRange) (err error) {
	entity := &tables.OutpointSatRange{}
	err = d.DB.Where("outpoint=?", outpoint).First(entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		entity.Outpoint = outpoint
		entity.Start = satRange.Start
		entity.End = satRange.End
		err = d.DB.Create(entity).Error
		return
	}
	return d.DB.Save(entity).Error
}
