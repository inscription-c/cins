package index

type Rarity uint8

const (
	RarityCommon Rarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
	RarityMythic
)

func NewRarityFromSat(sat *Sat) Rarity {
	r := Rarity(0)
	degree := NewDegreeFromSat(sat)
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
	return r
}
