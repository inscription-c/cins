// Package index provides the implementation of Charm type and its associated methods.
package index

// Charm is a type that represents a charm in the blockchain.
type Charm uint16

type Charms []Charm

// CharmsAll is a slice of Charm.
var CharmsAll = Charms{
	CharmCoin,
	CharmCursed,
	CharmEpic,
	CharmLegendary,
	CharmLost,
	CharmNineBall,
	CharmRare,
	CharmReInscription,
	CharmUnbound,
	CharmUncommon,
	CharmVindicated,
}

// Titles is a method that generates a list of titles for a given set of charms.
// It takes an uint16 representing a set of charms as a parameter.
// The method iterates over the Charms slice and checks if each charm is set in the given set of charms.
// If a charm is set, its title is appended to the list of titles.
// The method returns a slice of strings representing the titles of the set charms.
func (cs Charms) Titles(charms uint16) []string {
	titles := make([]string, 0, len(cs))
	for _, c := range cs {
		if !c.IsSet(charms) {
			continue
		}
		titles = append(titles, c.Title())
	}
	return titles
}

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

// Flag is a method that returns the flag for a charm.
func (c *Charm) Flag() uint16 {
	return 1 << *c
}

// Set is a method that sets a charm in a set of charms.
// It takes a pointer to a uint16 as a parameter.
// The method sets the bit at the position of the charm in the set of charms.
func (c *Charm) Set(charms *uint16) {
	*charms |= c.Flag()
}

// IsSet is a method that checks if a charm is set in a set of charms.
// It takes a uint16 as a parameter and returns a bool.
// The method checks if the bit at the position of the charm in the set of charms is set.
// If the bit is set, it returns true. Otherwise, it returns false.
func (c *Charm) IsSet(charms uint16) bool {
	return charms&c.Flag() != 0
}

// Icon is a method that returns the corresponding icon for a charm.
// It uses a switch statement to determine the charm type and returns a string representing the icon.
// If the charm type is not recognized, it returns an empty string.
func (c *Charm) Icon() string {
	switch *c {
	case CharmCoin:
		return "🪙"
	case CharmCursed:
		return "👹"
	case CharmEpic:
		return "🪻"
	case CharmLegendary:
		return "🌝"
	case CharmLost:
		return "🤔"
	case CharmNineBall:
		return "9️⃣"
	case CharmRare:
		return "🧿"
	case CharmReInscription:
		return "♻️"
	case CharmUnbound:
		return "🔓"
	case CharmUncommon:
		return "🌱"
	case CharmVindicated:
		return "️‍❤️‍🔥"
	}
	return ""
}

// Title is a method that returns the titles of all charms in the Charms slice.
// It returns a slice of strings.
func (c *Charm) Title() string {
	switch *c {
	case CharmCoin:
		return "coin"
	case CharmCursed:
		return "cursed"
	case CharmEpic:
		return "epic"
	case CharmLegendary:
		return "legendary"
	case CharmLost:
		return "lost"
	case CharmNineBall:
		return "nineball"
	case CharmRare:
		return "rare"
	case CharmReInscription:
		return "reinscription"
	case CharmUnbound:
		return "unbound"
	case CharmUncommon:
		return "uncommon"
	case CharmVindicated:
		return "vindicated"
	}
	return ""
}

// TitleToCharm is a function that returns a pointer to a Charm for a given title.
func TitleToCharm(title string) *Charm {
	for _, v := range CharmsAll {
		if v.Title() == title {
			newCharm := v
			return &newCharm
		}
	}
	return nil
}
