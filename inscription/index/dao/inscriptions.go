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
	if err = d.DB.Last(&ins).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	return ins.SequenceNum + 1, nil
}

func (d *DB) GetInscriptionById(inscriptionId string) (ins tables.Inscriptions, err error) {
	err = d.DB.Where("inscription_id = ?", inscriptionId).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) DeleteInscriptionById(inscriptionId string) (sequenceNum uint64, err error) {
	ins := &tables.Inscriptions{}
	err = d.DB.Clauses(clause.Returning{}).Where("inscription_id = ?", inscriptionId).Delete(ins).Error
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
	err = d.DB.Where("sequence_num = ?", sequenceNum).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type InscriptionId struct {
	Outpoint string `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset   uint32 `gorm:"column:offset;type:int unsigned;default:0;NOT NULL"`
}

func (d *DB) FindInscriptionsByPage(page, size int) (list []InscriptionId, total int64, err error) {
	db := d.DB.Where("inscription_num>=0").Order("inscription_num asc")
	if err = db.Count(&total).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}
