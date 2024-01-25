package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"gorm.io/gorm"
)

func (d *DB) DeleteAllBySatPoint(satpoint *util.SatPoint) error {
	return d.Where("outpoint = ? AND offset = ?", satpoint.Outpoint.String(), satpoint.Offset).Delete(&tables.SatPoint{}).Error
}

func (d *DB) SetSatPointToSequenceNum(satPoint *util.SatPoint, sequenceNum uint64) error {
	return d.Create(&tables.SatPoint{
		Outpoint:    satPoint.Outpoint.String(),
		Offset:      satPoint.Offset,
		SequenceNum: sequenceNum,
	}).Error
}

func (d *DB) InscriptionsByOutpoint(outpoint string) (res []*Inscription, err error) {
	satpoints := make([]*tables.SatPoint, 0)
	err = d.DB.Where("outpoint = ?", outpoint).Find(&satpoints).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	if len(satpoints) == 0 {
		return
	}

	satpointMap := make(map[uint64]*tables.SatPoint)
	ids := make([]uint64, 0, len(satpoints))
	for _, satpoint := range satpoints {
		ids = append(ids, satpoint.SequenceNum)
		satpointMap[satpoint.SequenceNum] = satpoint
	}

	list := make([]*tables.Inscriptions, 0, len(ids))
	err = d.DB.Where("sequence_num in (?)", ids).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	res = make([]*Inscription, 0, len(list))
	for _, ins := range list {
		res = append(res, &Inscription{
			Inscriptions: ins,
			SatPoint: &util.SatPoint{
				Outpoint: util.StringToOutpoint(satpointMap[ins.Id].Outpoint).OutPoint,
				Offset:   satpointMap[ins.Id].Offset,
			},
		})
	}
	return
}

func (d *DB) GetSatPointBySat(sat uint64) (res tables.SatSatPoint, err error) {
	err = d.DB.Where("sat = ?", sat).First(&res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
