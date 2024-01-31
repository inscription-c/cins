package dao

import (
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm/clause"
)

// SatToSatPoint saves a SAT (Satisfiability) to a sequence number in the database.
func (d *DB) SatToSatPoint(satSatPoint *tables.SatSatPoint) error {
	return d.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sat"}},
		DoUpdates: clause.AssignmentColumns([]string{"outpoint", "offset"}),
	}).Create(satSatPoint).Error
}
