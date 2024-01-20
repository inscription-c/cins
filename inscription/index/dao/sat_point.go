package dao

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type SatPoint struct {
	Outpoint wire.OutPoint
	Offset   uint32
}

func NewSatPointFromString(satpoint string) (*SatPoint, error) {
	parts := strings.Split(satpoint, constants.OutpointDelimiter)
	if len(parts) != 3 {
		return nil, errors.New("satpoint should be of the form txid:index:offset")
	}
	hash, err := chainhash.NewHashFromStr(parts[0])
	if err != nil {
		return nil, err
	}
	outputIndex, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid output index: %v", err)
	}
	offset, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid satpoint offset: %v", err)
	}
	return &SatPoint{
		Outpoint: wire.OutPoint{
			Hash:  *hash,
			Index: uint32(outputIndex),
		},
		Offset: uint32(offset),
	}, nil
}

func (s *SatPoint) String() string {
	return fmt.Sprintf("%s:%d", s.Outpoint, s.Offset)
}

func (d *DB) DeleteAllBySatPoint(satpoint *SatPoint) error {
	return d.Where("outpoint = ? AND offset = ?", satpoint.Outpoint.String(), satpoint.Offset).Delete(&tables.SatPoint{}).Error
}

func (d *DB) SetSatPointToSequenceNum(satPoint *SatPoint, sequenceNum uint64) error {
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

	satpointMap := make(map[uint64]*tables.SatPoint)
	ids := make([]uint64, 0, len(satpoints))
	for _, satpoint := range satpoints {
		ids = append(ids, satpoint.SequenceNum)
		satpointMap[satpoint.SequenceNum] = satpoint
	}

	list := make([]*tables.Inscriptions, 0, len(ids))
	err = d.DB.Where("id in (?)", ids).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}

	res = make([]*Inscription, 0, len(list))
	for _, ins := range list {
		res = append(res, &Inscription{
			Inscriptions: ins,
			SatPoint: &SatPoint{
				Outpoint: model.StringToOutpoint(satpointMap[ins.Id].Outpoint).OutPoint,
				Offset:   satpointMap[ins.Id].Offset,
			},
		})
	}
	return
}
