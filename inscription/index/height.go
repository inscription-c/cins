// Package index provides the implementation of Height type and its associated methods.
package index

// Height is a struct that represents a height in the blockchain.
// It contains a single field, height, which is a uint32.
type Height struct {
	height uint32 // The height in the blockchain
}

// NewHeight is a function that creates a Height.
// It takes a uint32 as a parameter and returns a pointer to a Height.
// The function creates a Height with the given height and returns it.
func NewHeight(height uint32) *Height {
	return &Height{
		height: height, // Set the height of the Height
	}
}

// N is a method that gets the height of a Height.
// It takes no parameters and returns a uint32.
// The method returns the height of the Height.
func (h *Height) N() uint32 {
	return h.height // Return the height of the Height
}

// Subsidy is a method that gets the subsidy of a Height.
// It takes no parameters and returns a uint64.
// The method creates an Epoch from the Height and gets the subsidy of the Epoch.
// It then returns the subsidy.
func (h *Height) Subsidy() uint64 {
	return NewEpochFrom(h).Subsidy() // Return the subsidy of the Epoch created from the Height
}
