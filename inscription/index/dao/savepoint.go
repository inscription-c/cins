package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) ListSavepoint() (list []*tables.SavePoint, err error) {
	err = d.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) OldestSavepoint() (savepoint *tables.SavePoint, err error) {
	err = d.First(&savepoint).Error
	return
}

func (d *DB) DeleteSavepoint() error {
	s := tables.SavePoint{}
	return d.Exec("delete from " + s.TableName()).Error
}
