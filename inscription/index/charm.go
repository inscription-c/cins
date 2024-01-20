package index

type Charm uint16

var (
	CharmCoin          Charm = 0
	CharmCursed        Charm = 1
	CharmEpic          Charm = 2
	CharmLegendary     Charm = 3
	CharmLost          Charm = 4
	CharmNineBall      Charm = 5
	CharmRare          Charm = 6
	CharmReInscription Charm = 7
	CharmUnbound       Charm = 8
	CharmUncommon      Charm = 9
	CharmVindicated    Charm = 10
)

func (c *Charm) Set(charms *uint16) {
	*charms |= uint16(*c)
}

func (c *Charm) IsSet(charms uint16) bool {
	return charms&uint16(*c) != 0
}
