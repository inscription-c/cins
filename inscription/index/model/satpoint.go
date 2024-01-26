package model

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/constants"
	"strconv"
	"strings"
)

type SatPoint struct {
	Outpoint wire.OutPoint
	Offset   uint64
}

func FormatSatPoint(outpoint string, sat uint64) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.OutpointDelimiter, sat)
}

func NewSatPointFromString(satpoint string) (*SatPoint, error) {
	parts := strings.Split(satpoint, constants.OutpointDelimiter)
	if len(parts) != 3 {
		return nil, errors.New("satpoint should be of the form txid:index:offset")
	}
	hash, err := chainhash.NewHashFromStr(parts[0])
	if err != nil {
		return nil, err
	}
	outputIndex, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid output index: %v", err)
	}
	offset, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid satpoint offset: %v", err)
	}
	return &SatPoint{
		Outpoint: wire.OutPoint{
			Hash:  *hash,
			Index: uint32(outputIndex),
		},
		Offset: offset,
	}, nil
}

func (s *SatPoint) String() string {
	return fmt.Sprintf("%s%s%d", s.Outpoint, constants.OutpointDelimiter, s.Offset)
}
