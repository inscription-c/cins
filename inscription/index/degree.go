package index

const CycleEpochs uint32 = 6
const DiffChangeInterval uint32 = 2016

type Degree struct {
	hour   uint32
	minute uint32
	second uint32
	third  uint64
}

func NewDegreeFromSat(sat *Sat) *Degree {
	height := sat.Height().N()
	return &Degree{
		hour:   height / (CycleEpochs * SubsidyHalvingInterval),
		minute: height % SubsidyHalvingInterval,
		second: height % DiffChangeInterval,
		third:  sat.Third(),
	}
}
