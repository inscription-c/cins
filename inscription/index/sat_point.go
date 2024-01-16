package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"strconv"
	"strings"
)

type SatPoint struct {
	Outpoint *wire.OutPoint
	Offset   int64
}

func NewSatPointFromString(satpoint string) (*SatPoint, error) {
	parts := strings.Split(satpoint, ":")
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
	offset, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid satpoint offset: %v", err)
	}
	return &SatPoint{
		Outpoint: &wire.OutPoint{
			Hash:  *hash,
			Index: uint32(outputIndex),
		},
		Offset: offset,
	}, nil
}

func (s *SatPoint) String() string {
	return fmt.Sprintf("%s:%d", s.Outpoint, s.Offset)
}
