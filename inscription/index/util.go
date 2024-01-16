package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dotbitHQ/insc/constants"
	"strings"
)

func IsEmptyHash(h chainhash.Hash) bool {
	empty, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), empty.String())
}

type Height struct {
	Height int64
}

func (h *Height) Subsidy() int64 {
	if h.Height < 33 {
		return (50 * constants.OneBtc) >> h.Height
	}
	return 0
}
