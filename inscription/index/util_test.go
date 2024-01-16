package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"testing"
)

func TestIsEmptyHash(t *testing.T) {
	empty, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	t.Log(empty)
}
