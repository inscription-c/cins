package index

import (
	"github.com/inscription-c/insc/constants"
	"github.com/shopspring/decimal"
)

const (
	SupplySat     = 2099999997690000
	LastSupplySat = SupplySat - 1
)

type Amount float64

func AmountToSat(amount float64) Sat {
	return Amount(amount).Sat()
}

type Sat uint64

func (a Amount) Sat() Sat {
	return Sat(decimal.NewFromFloat(float64(a)).
		Mul(decimal.NewFromInt(constants.OneBtc)).IntPart())
}

func (s *Sat) Degree() *Degree {
	return NewDegreeFromSat(s)
}

func (s *Sat) Rarity() Rarity {
	return NewRarityFromSat(s)
}

func (s *Sat) NineBall() bool {
	return uint64(*s) >= 50*constants.OneBtc*9 && uint64(*s) < 50*constants.OneBtc*10
}

func (s *Sat) Coin() bool {
	return uint64(*s)%constants.OneBtc == 0
}

func (s *Sat) Height() *Height {
	start := s.Epoch().StartingHeight()
	h := uint64(s.EpochPosition()) / uint64(s.Epoch().Subsidy())
	return &Height{height: start.N() + uint32(h)}
}

func (s *Sat) Epoch() *Epoch {
	return NewEpochFromSat(*s)
}

func (s *Sat) EpochPosition() Sat {
	e := s.Epoch()
	startSat := e.StartingSat()
	return *s - startSat
}

func (s *Sat) Third() uint64 {
	return uint64(s.EpochPosition()) % uint64(s.Epoch().Subsidy())
}
