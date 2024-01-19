package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) GetStatisticCountByName(name tables.StatisticType) (count uint64, err error) {
	statistic := &tables.Statistic{}
	if err = d.DB.Where("name = ?", name).First(statistic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	return statistic.Count, nil
}
