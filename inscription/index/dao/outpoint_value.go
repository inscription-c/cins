package dao

import "github.com/inscription-c/insc/inscription/index/tables"

func (d *DB) GetValueByOutpoint(outpoint string) (value uint64, err error) {
	outpointVal := &tables.OutpointValue{}
	if err = d.DB.Where("outpoint = ?", outpoint).First(outpointVal).Error; err != nil {
		return
	}
	return outpointVal.Value, nil
}
