package dao

import (
	"errors"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Inscription is a struct that embeds tables.Inscriptions and util.SatPoint.
type Inscription struct {
	*tables.Inscriptions
	*tables.SatPointToSequenceNum
}

// NextSequenceNumber retrieves the next sequence number for inscriptions.
// It returns the next sequence number as an uint64 and any error encountered.
func (d *DB) NextSequenceNumber() (num int64, err error) {
	ins := &tables.Inscriptions{}
	if err = d.Where("sequence_num>0").Order("sequence_num desc").First(&ins).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return 1, nil
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
	err = d.Model(&tables.Inscriptions{}).Where("tx_id=? and `index`=?", outpoint.Hash.String(), outpoint.Index).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// DeleteInscriptionById deletes an inscription by its outpoint.
// It returns the sequence number of the deleted inscription and any error encountered.
func (d *DB) DeleteInscriptionById(inscriptionId *tables.InscriptionId) (sequenceNum int64, err error) {
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
	return d.Create(ins).Error
}

func (d *DB) DeleteMockInscriptions() error {
	return d.Where("sequence_num < 0").Delete(&tables.Inscriptions{}).Error
}

// GetInscriptionByInscriptionNum retrieves an inscription by its sequence number.
// It returns the inscription and any error encountered.
func (d *DB) GetInscriptionByInscriptionNum(inscriptionNum int64) (ins tables.Inscriptions, err error) {
	err = d.Where("inscription_num = ?", inscriptionNum).First(&ins).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// GetInscriptionBySequenceNum retrieves an inscription by its sequence number.
// It returns the inscription and any error encountered.
func (d *DB) GetInscriptionBySequenceNum(sequenceNum int64) (ins tables.Inscriptions, err error) {
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

func (d *DB) InscriptionsNum() (total int64, err error) {
	err = d.Model(&tables.Inscriptions{}).Count(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) InscriptionsStoredData() (total uint64, err error) {
	err = d.Model(&tables.Inscriptions{}).Select("sum(content_size)").Scan(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) InscriptionsTotalFees() (total uint64, err error) {
	err = d.Model(&tables.Inscriptions{}).Select("sum(fee)").Scan(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) FirstInscriptionByOwner(owner string) (firts tables.Inscriptions, err error) {
	err = d.Where("owner=?", owner).First(&firts).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type FindProtocolsParams struct {
	Page            int
	Size            int
	TxId            string
	Offset          uint32
	InscriptionNum  *int64
	Owner           string
	Ticker          string
	Order           string
	Types           []string
	InscriptionType string
}

func (d *DB) SearchInscriptions(params *FindProtocolsParams) (list []*tables.Inscriptions, total int64, err error) {
	db := d.Model(&tables.Inscriptions{})
	if params.InscriptionType != "" || params.Ticker != "" {
		db = db.Joins("JOIN protocol ON inscriptions.sequence_num=protocol.sequence_num")
		if params.InscriptionType != "" {
			db = db.Where("protocol.protocol=?", params.InscriptionType)
		} else if params.Ticker != "" {
			db = db.Where("protocol.protocol=? and protocol.ticker=?", constants.ProtocolCBRC20, params.Ticker)
		}
	}
	if params.TxId != "" {
		db = db.Where("inscriptions.tx_id=? and inscriptions.offset=?", params.TxId, params.Offset)
	}
	if params.InscriptionNum != nil {
		db = db.Where("inscriptions.inscription_num=?", *params.InscriptionNum)
	}
	if params.Owner != "" {
		db = db.Where("inscriptions.owner=?", params.Owner)
	}
	if len(params.Types) > 0 {
		db = db.Where("inscriptions.media_type in (?)", params.Types)
	}
	switch params.Order {
	case "newest":
		db = db.Order("inscriptions.id desc")
	case "oldest":
		db = db.Order("inscriptions.id asc")
	}

	if err = db.Count(&total).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	err = db.Offset((params.Page - 1) * params.Size).Limit(params.Size).Find(&list).Error
	return
}
