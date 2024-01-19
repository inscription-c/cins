package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) NextSequenceNumber() (num uint64, err error) {
	ins := &tables.Inscriptions{}
	if err = d.DB.Last(&ins).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			num = 1
		}
		return 0, err
	}
	return ins.Id + 1, nil
}
