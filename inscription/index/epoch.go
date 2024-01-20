package index

import "github.com/inscription-c/insc/constants"

const SubsidyHalvingInterval uint32 = 210_000

type Epoch struct {
	epoch uint32
}

var FirstPostSubsidy = Epoch{
	epoch: 33,
}

var EpochStartingStats = []Sat{
	0,
	1050000000000000,
	1575000000000000,
	1837500000000000,
	1968750000000000,
	2034375000000000,
	2067187500000000,
	2083593750000000,
	2091796875000000,
	2095898437500000,
	2097949218750000,
	2098974609270000,
	2099487304530000,
	2099743652160000,
	2099871825870000,
	2099935912620000,
	2099967955890000,
	2099983977420000,
	2099991988080000,
	2099995993410000,
	2099997995970000,
	2099998997250000,
	2099999497890000,
	2099999748210000,
	2099999873370000,
	2099999935950000,
	2099999967240000,
	2099999982780000,
	2099999990550000,
	2099999994330000,
	2099999996220000,
	2099999997060000,
	2099999997480000,
	SupplySat,
}

func NewEpochFrom(height *Height) *Epoch {
	return &Epoch{
		epoch: uint32(height.N() / SubsidyHalvingInterval),
	}
}

func NewEpochFromSat(sat Sat) *Epoch {
	e := &Epoch{}
	if sat < EpochStartingStats[1] {
		e.epoch = 0
	} else if sat < EpochStartingStats[2] {
		e.epoch = 1
	} else if sat < EpochStartingStats[3] {
		e.epoch = 2
	} else if sat < EpochStartingStats[4] {
		e.epoch = 3
	} else if sat < EpochStartingStats[5] {
		e.epoch = 4
	} else if sat < EpochStartingStats[6] {
		e.epoch = 5
	} else if sat < EpochStartingStats[7] {
		e.epoch = 6
	} else if sat < EpochStartingStats[8] {
		e.epoch = 7
	} else if sat < EpochStartingStats[9] {
		e.epoch = 8
	} else if sat < EpochStartingStats[10] {
		e.epoch = 9
	} else if sat < EpochStartingStats[11] {
		e.epoch = 10
	} else if sat < EpochStartingStats[12] {
		e.epoch = 11
	} else if sat < EpochStartingStats[13] {
		e.epoch = 12
	} else if sat < EpochStartingStats[14] {
		e.epoch = 13
	} else if sat < EpochStartingStats[15] {
		e.epoch = 14
	} else if sat < EpochStartingStats[16] {
		e.epoch = 15
	} else if sat < EpochStartingStats[17] {
		e.epoch = 16
	} else if sat < EpochStartingStats[18] {
		e.epoch = 17
	} else if sat < EpochStartingStats[19] {
		e.epoch = 18
	} else if sat < EpochStartingStats[20] {
		e.epoch = 19
	} else if sat < EpochStartingStats[21] {
		e.epoch = 20
	} else if sat < EpochStartingStats[22] {
		e.epoch = 21
	} else if sat < EpochStartingStats[23] {
		e.epoch = 22
	} else if sat < EpochStartingStats[24] {
		e.epoch = 23
	} else if sat < EpochStartingStats[25] {
		e.epoch = 24
	} else if sat < EpochStartingStats[26] {
		e.epoch = 25
	} else if sat < EpochStartingStats[27] {
		e.epoch = 26
	} else if sat < EpochStartingStats[28] {
		e.epoch = 27
	} else if sat < EpochStartingStats[29] {
		e.epoch = 28
	} else if sat < EpochStartingStats[30] {
		e.epoch = 29
	} else if sat < EpochStartingStats[31] {
		e.epoch = 30
	} else if sat < EpochStartingStats[32] {
		e.epoch = 31
	} else if sat < EpochStartingStats[33] {
		e.epoch = 32
	} else {
		e.epoch = 33
	}
	return e
}

func (e *Epoch) Subsidy() uint32 {
	if e.epoch < FirstPostSubsidy.epoch {
		return (50 * constants.OneBtc) >> e.epoch
	}
	return 0
}

func (e *Epoch) StartingHeight() Height {
	return Height{height: e.epoch * SubsidyHalvingInterval}
}

func (e *Epoch) StartingSat() Sat {
	if e.epoch > 33 {
		return SupplySat
	}
	return EpochStartingStats[e.epoch]
}
