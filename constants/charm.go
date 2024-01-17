package constants

type Charm uint16

var (
	CharmCoin          Charm = 1
	CharmCursed        Charm = 2
	CharmEpic          Charm = 3
	CharmLegendary     Charm = 4
	CharmLost          Charm = 5
	CharmNineBall      Charm = 6
	CharmRare          Charm = 7
	CharmReInscription Charm = 8
	CharmUnbound       Charm = 9
	CharmUncommon      Charm = 10
)

func (c *Charm) Set(charms *uint16) {
	*charms |= uint16(*c)
}

func (c *Charm) IsSet(charms uint16) bool {
	return charms&uint16(*c) != 0
}
