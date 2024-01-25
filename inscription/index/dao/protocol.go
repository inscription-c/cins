package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
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
	err = db.Select("sum(amount)").Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).Scan(&total).Error
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
	err = db.Select("count(distinct to)").Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).Scan(&total).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) FindTokenPageByTicker(protocol, ticker, operator string, page, pageSize int) (list []*tables.Protocol, err error) {
	err = d.Where("protocol=? and ticker=? and operator=?", protocol, ticker, operator).
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
