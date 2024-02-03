package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeleteAllBySatPoint deletes all records by a given SatPoint.
// It takes a SatPoint as a parameter.
// It returns any error encountered during the operation.
func (d *DB) DeleteAllBySatPoint(satpoint *tables.SatPointToSequenceNum) error {
	return d.Where("outpoint = ? AND offset = ?", satpoint.Outpoint, satpoint.Offset).Delete(satpoint).Error
}

// SetSatPointToSequenceNum sets a SatPoint to a sequence number in the database.
// It takes a SatPoint and a sequence number as parameters.
// It returns any error encountered during the operation.
func (d *DB) SetSatPointToSequenceNum(satPoint *tables.SatPointToSequenceNum, sequenceNum uint64) error {
	return d.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "outpoint"}, {Name: "offset"}},
		DoUpdates: clause.AssignmentColumns([]string{"sequence_num"}),
	}).Create(&tables.SatPointToSequenceNum{
		Outpoint:    satPoint.Outpoint,
		Offset:      satPoint.Offset,
		SequenceNum: sequenceNum,
	}).Error
}

// InscriptionsByOutpoint retrieves inscriptions by a given outpoint.
// It takes an outpoint as a parameter.
// It returns a list of inscriptions and any error encountered.
func (d *DB) InscriptionsByOutpoint(outpoint string) (res []*Inscription, err error) {
	satpoints := make([]*tables.SatPointToSequenceNum, 0)
	err = d.DB.Where("outpoint = ?", outpoint).Find(&satpoints).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	if len(satpoints) == 0 {
		return
	}

	satpointMap := make(map[uint64]*tables.SatPointToSequenceNum)
	sequenceNums := make([]uint64, 0, len(satpoints))
	for _, satpoint := range satpoints {
		sequenceNums = append(sequenceNums, satpoint.SequenceNum)
		satpointMap[satpoint.SequenceNum] = satpoint
	}

	list := make([]*tables.Inscriptions, 0, len(sequenceNums))
	err = d.DB.Where("sequence_num in (?)", sequenceNums).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	res = make([]*Inscription, 0, len(list))
	for _, ins := range list {
		satpoint := satpointMap[ins.SequenceNum]
		res = append(res, &Inscription{
			Inscriptions: ins,
			SatPointToSequenceNum: &tables.SatPointToSequenceNum{
				Outpoint: satpoint.Outpoint,
				Offset:   satpoint.Offset,
			},
		})
	}
	return
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