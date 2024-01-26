package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Inscription is a struct that embeds tables.Inscriptions and util.SatPoint.
type Inscription struct {
	*tables.Inscriptions
	*model.SatPoint
}

// NextSequenceNumber retrieves the next sequence number for inscriptions.
// It returns the next sequence number as an uint64 and any error encountered.
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

// GetInscriptionById retrieves an inscription by its outpoint.
// It returns the inscription and any error encountered.
func (d *DB) GetInscriptionById(inscriptionId *tables.InscriptionId) (ins tables.Inscriptions, err error) {
	err = d.Where("tx_id=? and offset=?", inscriptionId.TxId, inscriptionId.Offset).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// GetInscriptionByOutpoint retrieves an inscription by its outpoint.
func (d *DB) GetInscriptionByOutpoint(outpoint *model.OutPoint) (list []*tables.InscriptionId, err error) {
	err = d.Where("tx_id=? and index=?", outpoint.Hash.String(), outpoint.Index).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// DeleteInscriptionById deletes an inscription by its outpoint.
// It returns the sequence number of the deleted inscription and any error encountered.
func (d *DB) DeleteInscriptionById(inscriptionId *tables.InscriptionId) (sequenceNum uint64, err error) {
	ins := &tables.Inscriptions{}
	err = d.Clauses(clause.Returning{}).Where("tx_id=? and offset=?", inscriptionId.TxId, inscriptionId.Offset).Delete(ins).Error
	if err != nil {
		return
	}
	sequenceNum = ins.SequenceNum
	return
}

// CreateInscription creates a new inscription in the database.
// It returns any error encountered.
func (d *DB) CreateInscription(ins *tables.Inscriptions) error {
	return d.DB.Create(ins).Error
}

// GetInscriptionBySequenceNum retrieves an inscription by its sequence number.
// It returns the inscription and any error encountered.
func (d *DB) GetInscriptionBySequenceNum(sequenceNum uint64) (ins tables.Inscriptions, err error) {
	err = d.Where("sequence_num = ?", sequenceNum).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// FindInscriptionsByPage retrieves a page of inscription IDs.
// It returns a list of inscription IDs and any error encountered.
func (d *DB) FindInscriptionsByPage(page, size int) (list []*tables.InscriptionId, err error) {
	err = d.Model(&tables.Inscriptions{}).Select("tx_id,offset").
		Offset((page - 1) * size).Limit(size + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}

// FindInscriptionsInBlockPage retrieves a page of inscription IDs in a block.
// It returns a list of inscription IDs and any error encountered.
func (d *DB) FindInscriptionsInBlockPage(height, page, size int) (list []*model.OutPoint, err error) {
	newBlock := &tables.BlockInfo{}
	err = d.Where("height=?", height).First(newBlock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}

	oldBlock := &tables.BlockInfo{}
	if height > 0 {
		err = d.Where("height=?", height-1).First(oldBlock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
	}

	err = d.Model(&tables.Inscriptions{}).
		Where("sequence_num>? and sequence_num<=?", oldBlock.SequenceNum, newBlock.SequenceNum).
		Offset((page - 1) * size).Limit(size + 1).Scan(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}

func (d *DB) FindInscriptionsInBlock(height uint32) (list []*model.OutPoint, err error) {
	newBlock := &tables.BlockInfo{}
	err = d.Where("height=?", height).First(newBlock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}

	oldBlock := &tables.BlockInfo{}
	if height > 0 {
		err = d.Where("height=?", height-1).First(oldBlock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
	}

	err = d.Model(&tables.Inscriptions{}).
		Where("sequence_num>? and sequence_num<=?",
			oldBlock.SequenceNum, newBlock.SequenceNum).Scan(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}
