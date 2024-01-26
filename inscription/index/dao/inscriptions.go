package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Inscription is a struct that embeds tables.Inscriptions and util.SatPoint.
type Inscription struct {
	*tables.Inscriptions
	*util.SatPoint
}

// NextSequenceNumber retrieves the next sequence number for inscriptions.
// It returns the next sequence number as a uint64 and any error encountered.
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
func (d *DB) GetInscriptionById(outpoint string) (ins tables.Inscriptions, err error) {
	err = d.Where("outpoint = ?", outpoint).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// DeleteInscriptionById deletes an inscription by its outpoint.
// It returns the sequence number of the deleted inscription and any error encountered.
func (d *DB) DeleteInscriptionById(outpoint string) (sequenceNum uint64, err error) {
	ins := &tables.Inscriptions{}
	err = d.Clauses(clause.Returning{}).Where("outpoint = ?", outpoint).Delete(ins).Error
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

// InscriptionId is a struct that represents the ID of an inscription.
type InscriptionId struct {
	Outpoint string `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL"`
	Offset   uint32 `gorm:"column:offset;type:int unsigned;default:0;NOT NULL"`
}

// FindInscriptionsByPage retrieves a page of inscription IDs.
// It returns a list of inscription IDs and any error encountered.
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

// FindInscriptionsInBlock retrieves a page of inscription IDs in a block.
// It returns a list of inscription IDs and any error encountered.
func (d *DB) FindInscriptionsInBlock(height, page, size int) (list []*InscriptionId, err error) {
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

	err = d.Where("sequence_num>=? and sequence_num<=?",
		oldBlock.SequenceNum, newBlock.SequenceNum).
		Offset((page - 1) * size).Limit(size + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return
}
