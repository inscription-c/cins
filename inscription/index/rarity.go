// Package index provides the implementation of Rarity type and its associated methods.
package index

// Rarity is a type that represents a rarity in the blockchain.
type Rarity uint8

// Define various types of rarities as constants of type Rarity.
const (
	RarityCommon    Rarity = iota // Represents a common rarity
	RarityUncommon                // Represents an uncommon rarity
	RarityRare                    // Represents a rare rarity
	RarityEpic                    // Represents an epic rarity
	RarityLegendary               // Represents a legendary rarity
	RarityMythic                  // Represents a mythic rarity
)

// NewRarityFromSat is a function that creates a Rarity from a Sat.
// It takes a pointer to a Sat as a parameter and returns a Rarity.
// The function first creates a Degree from the Sat.
// It then checks the fields of the Degree to determine the rarity.
// If all fields of the Degree are zero, the rarity is Mythic.
// If the minute, second, and third fields are zero, the rarity is Legendary.
// If the minute and third fields are zero, the rarity is Epic.
// If the second and third fields are zero, the rarity is Rare.
// If the third field is zero, the rarity is Uncommon.
// Otherwise, the rarity is Common.
// Finally, it returns the determined rarity.
func NewRarityFromSat(sat *Sat) Rarity {
	r := Rarity(0)                  // Initialize the rarity as Common
	degree := NewDegreeFromSat(sat) // Create a Degree from the Sat
	// Determine the rarity based on the fields of the Degree
	if degree.hour == 0 && degree.minute == 0 && degree.second == 0 && degree.third == 0 {
		r = RarityMythic
	} else if degree.minute == 0 && degree.second == 0 && degree.third == 0 {
		r = RarityLegendary
	} else if degree.minute == 0 && degree.third == 0 {
		r = RarityEpic
	} else if degree.second == 0 && degree.third == 0 {
		r = RarityRare
	} else if degree.third == 0 {
		r = RarityUncommon
	} else {
		r = RarityCommon
	}
	return r // Return the determined rarity
}
