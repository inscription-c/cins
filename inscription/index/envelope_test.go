package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"testing"
)

func TestParsedEnvelopFromTransaction(t *testing.T) {
	cli, err := rpcclient.NewClient(
		rpcclient.WithClientHost("https://palpable-evocative-tent.btc-testnet.quiknode.pro/6b721f0afc59b375a98d75b39cc5955701089064/"),
		rpcclient.WithClientUser("root"),
		rpcclient.WithClientPassword("root"),
	)
	if err != nil {
		t.Fatal(err)
	}
	txHash, _ := chainhash.NewHashFromStr("7acc8d793e096b5ffd7e3b0a3908ca2067b2bcc154421397dda75624f34e274e")
	tx, err := cli.GetRawTransaction(txHash)
	if err != nil {
		t.Fatal(err)
	}
	ParsedEnvelopFromTransaction(tx.MsgTx())
}
