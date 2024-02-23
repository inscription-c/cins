package dao

import (
	"errors"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeleteBySatPoint deletes records by a given SatPoint.
// It takes a SatPoint as a parameter.
// It returns any error encountered during the operation.
func (d *DB) DeleteBySatPoint(satpoint *tables.SatPointToSequenceNum) error {
	if err := d.Clauses(clause.Returning{}).
		Where("outpoint = ? AND offset = ?", satpoint.Outpoint, satpoint.Offset).
		Delete(satpoint).Error; err != nil {
		return err
	}
	if satpoint.Id > 0 {
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Create(satpoint)
			}),
		}).Error
	}
	return nil
}

// SetSatPointToSequenceNum sets a SatPoint to a sequence number in the database.
// It takes a SatPoint and a sequence number as parameters.
// It returns any error encountered during the operation.
func (d *DB) SetSatPointToSequenceNum(satPoint *tables.SatPointToSequenceNum) error {
	old := &tables.SatPointToSequenceNum{}
	err := d.Where("outpoint = ? AND offset = ?", satPoint.Outpoint, satPoint.Offset).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if old.Id > 0 {
		if old.SequenceNum == satPoint.SequenceNum {
			return nil
		}
		satPoint.Id = old.Id
		satPoint.CreatedAt = old.CreatedAt
		if err := d.Save(satPoint).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Save(old)
			}),
		}).Error
	}
	if err := d.Create(satPoint).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Delete(satPoint)
		}),
	}).Error
}

// GetSatPointBySat retrieves a SatSatPoint by a given SAT.
// It takes a SAT as a parameter.
// It returns a SatSatPoint and any error encountered.
func (d *DB) GetSatPointBySat(sat uint64) (res tables.SatSatPoint, err error) {
	err = d.DB.Where("sat = ?", sat).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
