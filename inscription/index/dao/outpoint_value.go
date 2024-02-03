package dao

import (
	"github.com/inscription-c/insc/inscription/index/tables"
)

// GetValueByOutpoint retrieves the value associated with a given outpoint.
// It returns the value as an int64 and any error encountered.
func (d *DB) GetValueByOutpoint(outpoint string) (value int64, err error) {
	outpointVal := &tables.OutpointValue{}
	// Retrieve the OutpointValue record for the given outpoint
	if err = d.DB.Where("outpoint = ?", outpoint).First(outpointVal).Error; err != nil {
		return
	}
	// Return the value associated with to outpoint
	return outpointVal.Value, nil
}

// DeleteValueByOutpoint deletes the values associated with a list of outpoints.
// It returns any error encountered during the operation.
func (d *DB) DeleteValueByOutpoint(outpoints ...string) (err error) {
	if len(outpoints) == 0 {
		return
	}
	err = d.Where("outpoint in (?)", outpoints).Delete(&tables.OutpointValue{}).Error
	if err != nil {
		return
	}
	return
}

// SetOutpointToValue sets the values for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding values.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToValue(values map[string]int64) (err error) {
	list := make([]*tables.OutpointValue, 0, len(values))
	for outpoint, value := range values {
		list = append(list, &tables.OutpointValue{
			Outpoint: outpoint,
			Value:    value,
		})
	}
	return d.CreateInBatches(&list, 10_000).Error
}
