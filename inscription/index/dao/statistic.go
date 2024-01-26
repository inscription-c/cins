package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

// GetStatisticCountByName retrieves the count of a specific statistic by its name.
// It takes a name of type tables.StatisticType as a parameter.
// It returns the count of the statistic and any error encountered.
func (d *DB) GetStatisticCountByName(name tables.StatisticType) (count uint32, err error) {
	statistic := &tables.Statistic{}
	err = d.DB.Where("name = ?", name).First(statistic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	return statistic.Count, nil
}

// IncrementStatistic increments the count of a specific statistic by a given amount.
// It takes a name of type tables.StatisticType and a count of type uint32 as parameters.
// If the statistic does not exist, it creates a new one with the given name and count.
// If the statistic exists, it increments its count by the given amount.
// It returns any error encountered during the operation.
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
