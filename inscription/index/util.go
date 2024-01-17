package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"strings"
)

func IsEmptyHash(hash chainhash.Hash) bool {
	h, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), hash.String())
}
