package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

func (d *DB) GetStatisticCountByName(name tables.StatisticType) (count uint32, err error) {
	statistic := &tables.Statistic{}
	err = d.DB.Where("name = ?", name).First(statistic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return statistic.Count, nil
}

func (d *DB) IncrementStatistic(name tables.StatisticType, count uint32) error {
	statistic := &tables.Statistic{}
	err := d.DB.Where("name = ?", name).First(statistic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		statistic.Name = name
		statistic.Count = count
		return d.DB.Create(statistic).Error
	}
	statistic.Count += count
	return d.DB.Save(statistic).Error
}
