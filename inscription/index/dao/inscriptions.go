package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Inscription struct {
	*tables.Inscriptions
	*util.SatPoint
}

func (d *DB) NextSequenceNumber() (num uint64, err error) {
	ins := &tables.Inscriptions{}
	if err = d.Last(&ins).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	return ins.SequenceNum + 1, nil
}

func (d *DB) GetInscriptionById(outpoint string) (ins tables.Inscriptions, err error) {
	err = d.Where("outpoint = ?", outpoint).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) DeleteInscriptionById(outpoint string) (sequenceNum uint64, err error) {
	ins := &tables.Inscriptions{}
	err = d.Clauses(clause.Returning{}).Where("outpoint = ?", outpoint).Delete(ins).Error
	if err != nil {
		return
	}
	sequenceNum = ins.SequenceNum
	return
}

func (d *DB) CreateInscription(ins *tables.Inscriptions) error {
	return d.DB.Create(ins).Error
}

func (d *DB) GetInscriptionBySequenceNum(sequenceNum uint64) (ins tables.Inscriptions, err error) {
	err = d.Where("sequence_num = ?", sequenceNum).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type InscriptionId struct {
	Outpoint string `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset   uint32 `gorm:"column:offset;type:int unsigned;default:0;NOT NULL"`
}

func (d *DB) FindInscriptionsByPage(page, size int) (list []*InscriptionId, err error) {
	err = d.Where("inscription_num>=0").
		Order("inscription_num asc").
		Offset((page - 1) * size).Limit(size + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}

func (d *DB) FindInscriptionsInBlock(height, page, size int) (list []*InscriptionId, err error) {
	newBlock := &tables.BlockInfo{}
	err = d.Where("height=?", height).First(newBlock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}

	oldBlock := &tables.BlockInfo{}
	err = d.Where("height=?", height-1).First(oldBlock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}

	err = d.Where("sequence_num>=? and sequence_num<=?",
		oldBlock.SequenceNum, newBlock.SequenceNum).
		Offset((page - 1) * size).Limit(size + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}
