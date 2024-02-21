package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

// GetStatisticCountByName retrieves the count of a specific statistic by its name.
// It takes a name of type tables.StatisticType as a parameter.
// It returns the count of the statistic and any error encountered.
func (d *DB) GetStatisticCountByName(name tables.StatisticType) (count uint64, err error) {
	statistic := &tables.Statistic{}
	err = d.Where("name = ?", name).First(statistic).Error
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
func (d *DB) IncrementStatistic(name tables.StatisticType, count uint64) error {
	statistic := &tables.Statistic{}
	err := d.Where("name = ?", name).First(statistic).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if statistic.Id == 0 {
		statistic.Name = name
		statistic.Count = count
		if err := d.Create(statistic).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Delete(statistic)
			}),
		}).Error
	}
	statistic.Count += count
	if err := d.Save(statistic).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			statistic.Count -= count
			return tx.Save(statistic)
		}),
	}).Error
}

// SetStatistic sets the count of a specific statistic to a given amount.
func (d *DB) SetStatistic(name tables.StatisticType, count uint64) error {
	statistic := &tables.Statistic{}
	err := d.Where("name = ?", name).First(statistic).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if statistic.Id == 0 {
		statistic.Name = name
		statistic.Count = count
		if err := d.Create(statistic).Error; err != nil {
			return err
		}
		return d.Create(&tables.UndoLog{
			Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
				return tx.Delete(statistic)
			}),
		}).Error
	}
	if statistic.Count == count {
		return nil
	}
	oldCont := statistic.Count
	statistic.Count = count
	if err := d.Save(statistic).Error; err != nil {
		return err
	}
	return d.Create(&tables.UndoLog{
		Sql: d.ToSQL(func(tx *gorm.DB) *gorm.DB {
			statistic.Count = oldCont
			return tx.Save(statistic)
		}),
	}).Error
}
