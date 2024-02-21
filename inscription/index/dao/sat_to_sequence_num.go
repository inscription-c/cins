package dao

import "C"
import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

// SaveSatToSequenceNumber saves a SAT (Satisfiability) to a sequence number in the database.
// It takes a SAT and a sequence number as parameters, both of type uint64.
// It returns any error encountered during the operation.
func (d *DB) SaveSatToSequenceNumber(sat uint64, sequenceNum int64) error {
	old := &tables.SatToSequenceNum{}
	err := d.Where("sat=?", sat).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if old.Id > 0 {
		if old.SequenceNum == sequenceNum {
			return nil
		}
		oldSequenceNum := old.SequenceNum
		old.SequenceNum = sequenceNum
		if err := d.Save(old).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				old.SequenceNum = oldSequenceNum
				return tx.Save(old)
			}),
		}).Error
	}

	old.Sat = sat
	old.SequenceNum = sequenceNum
	if err := d.Create(old).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Delete(old)
		}),
	}).Error
}
