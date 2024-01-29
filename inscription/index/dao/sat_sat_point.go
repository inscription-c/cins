package dao

import "github.com/inscription-c/insc/inscription/index/tables"

func (d *DB) SatToSatPointCreate(satSatPoint *tables.SatSatPoint) error {
	return d.Create(satSatPoint).Error
}
