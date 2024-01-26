package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"gorm.io/gorm"
)

func (d *DB) SaveProtocol(protocol *tables.Protocol) error {
	return d.DB.Save(protocol).Error
}

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

func (d *DB) CountProtocolAmount(protocol, ticker, operator string, tkid ...string) (total uint64, err error) {
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

func (d *DB) GetProtocolByOutpoint(outpoint string) (p tables.Protocol, err error) {
	err = d.Where("outpoint=?", outpoint).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) CountToAddress(protocol, ticker, operator string, tkid ...string) (total uint64, err error) {
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

func (d *DB) FindTokenPageByTicker(protocol, ticker, operator string, page, pageSize int) (list []*tables.Protocol, err error) {
	err = d.Model(&tables.Protocol{}).Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type ProtocolAmount struct {
	TkID   string `gorm:"column:tkid;type:varchar(255);index:idx_tkid;default:;NOT NULL" json:"ticker_id"`
	Ticker string `gorm:"column:ticker;type:varchar(255);index:idx_ticker;default:;NOT NULL" json:"ticker"`
	Amount uint64 `gorm:"column:amount;type:bigint unsigned;default:0;NOT NULL" json:"amount"`
}

func (d *DB) SumAmountByToAddress(protocol, to string, page, pageSize int) (list []*ProtocolAmount, err error) {
	err = d.Model(&tables.Protocol{}).Select("tkid, ticker, sum(amount) as amount").
		Where("protocol=? and `to`=?", protocol, to).
		Group("tkid").Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type FindMintHistoryByTkIdResp struct {
	Outpoint *util.OutPoint `gorm:"column:outpoint;type:varchar(255);index:idx_outpoint;default:;NOT NULL" json:"inscription_id"`
	Amount   uint64         `gorm:"column:amount;type:bigint unsigned;default:0;NOT NULL" json:"amount"`
	To       string         `gorm:"column:to;type:varchar(255);index:idx_to;default:;NOT NULL" json:"to_address"`
	Miner    string         `gorm:"column:miner;type:varchar(255);index:idx_miner;default:;NOT NULL" json:"miner"`
}

func (d *DB) FindMintHistoryByTkId(tkid, protocol, operator string, page, pageSize int) (list []*FindMintHistoryByTkIdResp, err error) {
	err = d.Model(&tables.Protocol{}).Where("tkid=? and protocol=? and operator=?", tkid, protocol, operator).
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

type FindHoldersByTkIdResp struct {
	Amount  uint64 `gorm:"column:amount" json:"amount"`
	Address string `gorm:"column:address" json:"address"`
}

func (d *DB) FindHoldersByTkId(tkid, protocol, operator string, page, pageSize int) (list []*FindHoldersByTkIdResp, err error) {
	err = d.Model(&tables.Protocol{}).Select("DISTINCT COALESCE(`to`,miner) as address, sum(amount) as amount").
		Where("tkid=? and protocol=? and operator=?", tkid, protocol, operator).
		Group("address").Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
