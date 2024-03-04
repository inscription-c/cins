package util

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"math"
	"strings"
)

func IsNullOutpoint(point wire.OutPoint) bool {
	h, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), point.Hash.String()) && point.Index == math.MaxUint32
}

func IsNullOutpointString(output string) (bool, error) {
	out, err := wire.NewOutPointFromString(output)
	if err != nil {
		return false, err
	}
	h, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), out.Hash.String()) && out.Index == math.MaxUint32, nil
}

func IsUnBoundOutpointString(output string) (bool, error) {
	out, err := wire.NewOutPointFromString(output)
	if err != nil {
		return false, err
	}
	h, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), out.Hash.String()) && out.Index == 0, nil
}

func NullOutpoint() *wire.OutPoint {
	return &wire.OutPoint{
		Index: math.MaxUint32,
	}
}
