// Package index provides the implementation of the Sat struct and its related functions.
package index

import (
	"github.com/inscription-c/cins/constants"
	"github.com/shopspring/decimal"
)

// Constants representing the total supply of Sat and the last supply of Sat.
const (
	SupplySat     = 2099999997690000
	LastSupplySat = SupplySat - 1
)

// Amount represents a float64 value of an amount.
type Amount float64

// AmountToSat converts a float64 amount to Sat.
func AmountToSat(amount float64) Sat {
	return Amount(amount).Sat()
}

// Sat represents a uint64 value of Sat.
type Sat uint64

func (s *Sat) N() uint64 {
	return uint64(*s)
}

// Sat converts an Amount to Sat.
// It multiplies the Amount by OneBtc and returns the integer part.
func (a Amount) Sat() Sat {
	return Sat(decimal.NewFromFloat(float64(a)).
		Mul(decimal.NewFromInt(int64(constants.OneBtc))).IntPart())
}

// Degree returns the Degree of the Sat.
func (s *Sat) Degree() *Degree {
	return NewDegreeFromSat(s)
}

// Rarity returns the Rarity of the Sat.
func (s *Sat) Rarity() Rarity {
	return NewRarityFromSat(s)
}

// NineBall checks if the Sat is a NineBall.
// It returns true if the Sat is between 50*OneBtc*9 and 50*OneBtc*10, false otherwise.
func (s *Sat) NineBall() bool {
	return uint64(*s) >= 50*constants.OneBtc*9 && uint64(*s) < 50*constants.OneBtc*10
}

// Coin checks if the Sat is a Coin.
// It returns true if the Sat is divisible by OneBtc, false otherwise.
func (s *Sat) Coin() bool {
	return uint64(*s)%constants.OneBtc == 0
}

// Height returns the Height of the Sat.
// It calculates the Height by adding the starting height of the epoch to the epoch position divided by the epoch subsidy.
func (s *Sat) Height() *Height {
	start := s.Epoch().StartingHeight()
	h := uint64(s.EpochPosition()) / uint64(s.Epoch().Subsidy())
	return &Height{height: start.N() + uint32(h)}
}

// Epoch returns the Epoch of the Sat.
func (s *Sat) Epoch() *Epoch {
	return NewEpochFromSat(*s)
}

// EpochPosition returns the position of the Sat in the epoch.
// It calculates the position by subtracting the starting Sat of the epoch from the Sat.
func (s *Sat) EpochPosition() Sat {
	e := s.Epoch()
	startSat := e.StartingSat()
	return *s - startSat
}

// Third returns the third of the Sat.
// It calculates the third by taking the modulus of the epoch position with the epoch subsidy.
func (s *Sat) Third() uint64 {
	return uint64(s.EpochPosition()) % s.Epoch().Subsidy()
}

func (s *Sat) Common() bool {
	epoch := s.Epoch()
	startingSat := epoch.StartingSat()
	return (s.N()-startingSat.N())%epoch.Subsidy() != 0
}
