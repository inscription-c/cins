package dao

import (
	"errors"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
)

// SaveProtocol saves a protocol to the database.
// It returns any error encountered during the operation.
func (d *DB) SaveProtocol(height uint32, protocol *tables.Protocol) error {
	old := &tables.Protocol{}
	err := d.Where("tx_id = ? AND offset = ?", protocol.TxId, protocol.Offset).First(old).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if old.Id > 0 {
		protocol.Id = old.Id
		protocol.CreatedAt = old.CreatedAt
		if err := d.Save(protocol).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Height: height,
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Save(old)
			}),
		}).Error
	}
	if err := d.Create(protocol).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Height: height,
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Delete(protocol)
		}),
	}).Error
}

func (d *DB) DeleteMockProtocol() error {
	return d.Where("sequence_num<0").Delete(&tables.Protocol{}).Error
}

// FindProtocol finds protocols based on the provided parameters.
// It returns a list of matching protocols and any error encountered.
func (d *DB) FindProtocol(protocol, ticker, operator string, tkid ...string) (list []*tables.Protocol, err error) {
	db := d.DB
	if len(tkid) > 0 {
		db = d.Where("tkid=?", tkid[0])
	}
	err = db.Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// SumProtocolAmount counts the total amount for a specific protocol.
// It returns the total amount and any error encountered.
func (d *DB) SumProtocolAmount(protocol, ticker, operator string, tkid ...string) (total uint64, err error) {
	db := d.DB
	if len(tkid) > 0 {
		db = d.Where("tkid=?", tkid[0])
	}
	err = db.Model(&tables.Protocol{}).Select("sum(amount)").Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).Scan(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// GetProtocolByInscriptionId retrieves a protocol by its outpoint.
// It returns the protocol and any error encountered.
func (d *DB) GetProtocolByInscriptionId(inscriptionId *tables.InscriptionId) (p tables.Protocol, err error) {
	err = d.Where("tx_id=? and offset=?", inscriptionId.TxId, inscriptionId.Offset).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// SumAddressNum counts the number of distinct addresses for a specific protocol.
// It returns the total count and any error encountered.
func (d *DB) SumAddressNum(protocol, ticker, operator string, tkid ...string) (total uint64, err error) {
	db := d.DB
	if len(tkid) > 0 {
		db = d.Where("tkid=?", tkid[0])
	}
	err = db.Model(&tables.Protocol{}).Select("count(distinct COALESCE(`to`, miner))").Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).Scan(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// FindTokenPageByTicker retrieves a page of tokens by ticker for a specific protocol.
// It returns a list of tokens and any error encountered.
func (d *DB) FindTokenPageByTicker(protocol, ticker, operator string, page, pageSize int) (list []*tables.Protocol, err error) {
	err = d.Model(&tables.Protocol{}).Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// SumMintAmountByAddress sums the amount for a specific address and protocol.
// It returns a list of protocol amounts and any error encountered.
func (d *DB) SumMintAmountByAddress(address, protocol string, page, pageSize int) (list []*tables.ProtocolAmount, err error) {
	err = d.Model(&tables.Protocol{}).Select("tx_id, ticker, sum(amount) as amount").
		Where("(`to`=? or miner=?) and protocol=? and operator=?", address, address, protocol, constants.OperationMint).
		Group("`to`,miner").Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// FindMintHistoryByTkIdResp is a struct that represents the response for finding mint history by TkId.
type FindMintHistoryByTkIdResp struct {
	tables.InscriptionId `gorm:"embedded" json:"inscription_id"`
	Amount               uint64 `gorm:"column:amount" json:"amount"` // The amount of the mint
	To                   string `gorm:"column:to" json:"to_address"` // The address to which the mint was made
	Miner                string `gorm:"column:miner" json:"miner"`   // The miner of the mint
}

// FindMintHistoryByTkId finds the mint history for a specific TkId and protocol.
// It returns a list of mint histories and any error encountered.
func (d *DB) FindMintHistoryByTkId(tkid, protocol, operator string, page, pageSize int) (list []*FindMintHistoryByTkIdResp, err error) {
	err = d.Model(&tables.Protocol{}).Where("tkid=? and protocol=? and operator=?", tkid, protocol, operator).
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// FindHoldersByTkIdResp is a struct that represents the response for finding holders by TkId.
type FindHoldersByTkIdResp struct {
	Amount  uint64 `gorm:"column:amount" json:"amount"`   // The amount held by the holder
	Address string `gorm:"column:address" json:"address"` // The address of the holder
}

// FindHoldersByTkId finds the holders for a specific TkId and protocol.
// It returns a list of holders and any error encountered.
func (d *DB) FindHoldersByTkId(tkid, protocol, operator string, page, pageSize int) (list []*FindHoldersByTkIdResp, err error) {
	err = d.Model(&tables.Protocol{}).Select("DISTINCT COALESCE(`to`,miner) as address, sum(amount) as amount").
		Where("tkid=? and protocol=? and operator=?", tkid, protocol, operator).
		Group("address").Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
