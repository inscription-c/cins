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

func NullOutpoint() *wire.OutPoint {
	return &wire.OutPoint{
		Index: math.MaxUint32,
	}
}
