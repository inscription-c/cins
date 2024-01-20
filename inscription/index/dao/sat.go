package dao

import "github.com/inscription-c/insc/inscription/index/tables"

func (d *DB) SaveSatToSequenceNumber(sat, sequenceNum uint64) error {
	s := &tables.Sat{
		Sat:         sat,
		SequenceNum: sequenceNum,
	}
	return d.Create(s).Error
}
