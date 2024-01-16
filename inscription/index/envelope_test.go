package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"gotest.tools/assert"
	"testing"
)

func TestParsedEnvelopeFromTransaction(t *testing.T) {
	hash, err := chainhash.NewHashFromStr("107a102a2b61ed16dbe65cb59fc59fd7547d853a079eaa4c17d4221cd207d243")
	assert.Equal(t, err, nil)
	tx, err := rpcClient.GetRawTransaction(hash)
	assert.Equal(t, err, nil)
	ParsedEnvelopeFromTransaction(tx.MsgTx())
}
