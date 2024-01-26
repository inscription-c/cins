package dao

import (
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
)

// SetOutpointToSatRange sets the satoshi range for a set of outpoints.
// It takes a map where the keys are outpoints and the values are the corresponding satoshi ranges.
// It returns any error encountered during the operation.
func (d *DB) SetOutpointToSatRange(rangeCache map[string]model.SatRange) (err error) {
	// Create a slice to hold the outpoint satoshi ranges
	list := make([]*tables.OutpointSatRange, 0, len(rangeCache))

	// Iterate over the rangeCache map
	for outpoint, satRange := range rangeCache {
		// Append a new OutpointSatRange to the list for each outpoint in the map
		list = append(list, &tables.OutpointSatRange{
			Outpoint: outpoint,
			Start:    satRange.Start,
			End:      satRange.End,
		})
	}

	// Create the OutpointSatRange records in the database
	return d.DB.Create(&list).Error
}
