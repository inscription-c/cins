package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Inscription struct {
	*tables.Inscriptions
	*SatPoint
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
