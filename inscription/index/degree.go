// Package index provides the implementation of Degree struct and its associated methods.
package index

// CycleEpochs is a constant that represents the number of epochs in a cycle.
const CycleEpochs uint32 = 6

// DiffChangeInterval is a constant that represents the interval at which the difficulty changes.
const DiffChangeInterval uint32 = 2016

// Degree is a struct that represents a degree in the blockchain.
// It contains four fields:
// - hour: the number of hours in the degree, calculated as the height divided by the product of CycleEpochs and SubsidyHalvingInterval.
// - minute: the number of minutes in the degree, calculated as the height modulo SubsidyHalvingInterval.
// - second: the number of seconds in the degree, calculated as the height modulo DiffChangeInterval.
// - third: the third part of the degree, obtained from the Sat struct.
type Degree struct {
	hour   uint32 // The number of hours in the degree
	minute uint32 // The number of minutes in the degree
	second uint32 // The number of seconds in the degree
	third  uint64 // The third part of the degree
}

// NewDegreeFromSat is a function that creates a Degree from a Sat.
// It takes a pointer to a Sat as a parameter and returns a pointer to a Degree.
// The function first gets the height from the Sat.
// It then calculates the hour, minute, and second fields of the Degree using the height.
// It also gets the third field of the Degree from the Sat.
// Finally, it creates a Degree with the calculated fields and returns it.
func NewDegreeFromSat(sat *Sat) *Degree {
	height := sat.Height().N() // Get the height from the Sat
	return &Degree{
		hour:   height / (CycleEpochs * SubsidyHalvingInterval), // Calculate the number of hours in the degree
		minute: height % SubsidyHalvingInterval,                 // Calculate the number of minutes in the degree
		second: height % DiffChangeInterval,                     // Calculate the number of seconds in the degree
		third:  sat.Third(),                                     // Get the third part of the degree from the Sat
	}
}
