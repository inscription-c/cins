package index

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/dotbitHQ/insc/wallet"
	"testing"
)

var (
	rpcClient *rpcclient.Client

	host     = "test-btcd.d.id"
	username = "root"
	password = "root"
)

func TestMain(t *testing.M) {
	var err error
	rpcClient, err = wallet.NewWalletClient(host, username, password, false)
	if err != nil {
		panic(err)
	}
	t.Run()
}
