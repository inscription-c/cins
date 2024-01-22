// Package index provides the implementation of Charm type and its associated methods.
package index

// Charm is a type that represents a charm in the blockchain.
type Charm uint16

// Define various types of charms as constants of type Charm.
var (
	CharmCoin          Charm = 0  // Represents a coin charm
	CharmCursed        Charm = 1  // Represents a cursed charm
	CharmEpic          Charm = 2  // Represents an epic charm
	CharmLegendary     Charm = 3  // Represents a legendary charm
	CharmLost          Charm = 4  // Represents a lost charm
	CharmNineBall      Charm = 5  // Represents a nine ball charm
	CharmRare          Charm = 6  // Represents a rare charm
	CharmReInscription Charm = 7  // Represents a re-inscription charm
	CharmUnbound       Charm = 8  // Represents an unbound charm
	CharmUncommon      Charm = 9  // Represents an uncommon charm
	CharmVindicated    Charm = 10 // Represents a vindicated charm
)

// Set is a method that sets a charm in a set of charms.
// It takes a pointer to a uint16 as a parameter.
// The method sets the bit at the position of the charm in the set of charms.
func (c *Charm) Set(charms *uint16) {
	*charms |= uint16(*c) // Set the bit at the position of the charm in the set of charms
}

// IsSet is a method that checks if a charm is set in a set of charms.
// It takes a uint16 as a parameter and returns a bool.
// The method checks if the bit at the position of the charm in the set of charms is set.
// If the bit is set, it returns true. Otherwise, it returns false.
func (c *Charm) IsSet(charms uint16) bool {
	return charms&uint16(*c) != 0 // Check if the bit at the position of the charm in the set of charms is set
}
