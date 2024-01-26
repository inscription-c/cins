package dao

import "github.com/inscription-c/insc/inscription/index/tables"

// SaveSatToSequenceNumber saves a SAT (Satisfiability) to a sequence number in the database.
// It takes a SAT and a sequence number as parameters, both of type uint64.
// It returns any error encountered during the operation.
func (d *DB) SaveSatToSequenceNumber(sat, sequenceNum uint64) error {
	// Create a new Sat instance with the provided SAT and sequence number
	s := &tables.Sat{
		Sat:         sat,
		SequenceNum: sequenceNum,
	}
	// Save the Sat instance to the database and return any error encountered
	return d.Create(s).Error
}
