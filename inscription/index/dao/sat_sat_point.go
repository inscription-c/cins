package dao

import (
	"errors"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
)

// SatToSatPoint saves a SAT (Satisfiability) to a sequence number in the database.
func (d *DB) SatToSatPoint(satSatPoint *tables.SatSatPoint) error {
	old := &tables.SatSatPoint{}
	err := d.Where("sat=?", satSatPoint.Sat).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if old.Id > 0 {
		satSatPoint.Id = old.Id
		satSatPoint.CreatedAt = old.CreatedAt
		if err := d.Save(satSatPoint).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Save(old)
			}),
		}).Error
	}
	if err := d.Create(satSatPoint).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Delete(satSatPoint)
		}),
	}).Error
}
