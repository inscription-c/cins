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
	// Delete the OutpointValue records for the given outpoints
	err = d.DB.Where("outpoint in (?)", outpoints).Delete(&tables.OutpointValue{}).Error
	return
}

// SetOutpointToValue sets the values for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding values.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToValue(values map[string]int64) (err error) {
	// Create a slice to hold the outpoint values
	list := make([]*tables.OutpointValue, 0, len(values))

	// Iterate over the values map
	for outpoint, value := range values {
		// Append a new OutpointValue to the list for each outpoint in the map
		list = append(list, &tables.OutpointValue{
			Outpoint: outpoint,
			Value:    value,
		})
	}

	// Create the OutpointValue records in the database
	return d.DB.CreateInBatches(&list, 10_000).Error
}
