package dao

import "C"
import (
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm/clause"
)

// SaveSatToSequenceNumber saves a SAT (Satisfiability) to a sequence number in the database.
// It takes a SAT and a sequence number as parameters, both of type uint64.
// It returns any error encountered during the operation.
func (d *DB) SaveSatToSequenceNumber(sat, sequenceNum uint64) error {
	s := &tables.SatToSequenceNum{
		Sat:         sat,
		SequenceNum: sequenceNum,
	}
	// Save the Sat instance to the database and return any error encountered
	return d.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sat"}},
		DoUpdates: clause.AssignmentColumns([]string{"sequence_num"}),
	}).Create(s).Error
}
