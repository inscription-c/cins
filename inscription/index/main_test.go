package index

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/wallet"
	"path/filepath"
	"testing"
)

var (
	indexer   *Indexer
	rpcClient *rpcclient.Client
	batchCli  *rpcclient.Client

	host     = ""
	username = ""
	password = ""

	dbPath = "./test"
)

func TestMain(t *testing.M) {
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "inscription.log"), false)
	log.InitLogRotator(logFile)

	var err error
	rpcClient, err = wallet.NewWalletClient(host, username, password, false)
	if err != nil {
		panic(err)
	}

	batchCli, err = wallet.NewBatchClient(host, username, password, false)
	if err != nil {
		panic(err)
	}

	db, err := dao.DB(dbPath)
	if err != nil {
		panic(err)
	}
	indexer = NewIndexer(
		WithDB(db),
		WithClient(rpcClient),
		WithBatchClient(batchCli),
	)
	t.Run()
}
