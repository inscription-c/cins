package server

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/insc/btcd"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/inscription/server/handle"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	srvOptions = &SrvOptions{}

	mainNetRPCListen  = ":8335"
	testNetRPCListen  = ":18335"
	mainNetRPCConnect = "localhost:8334"
	testNetRPCConnect = "localhost:18334"
)

type SrvOptions struct {
	username           string
	password           string
	rpcListen          string
	testnet            bool
	rpcConnect         string
	embedDB            bool
	noApi              bool
	dataDir            string
	mysqlAddr          string
	mysqlUser          string
	mysqlPassword      string
	mysqlDBName        string
	dbListenPort       string
	dbStatusListenPort string
	enablePProf        bool
	indexSats          string
	indexSpendSats     string
	indexNoSyncBlock   bool
}

type SrvOption func(*SrvOptions)

func WithUserName(username string) SrvOption {
	return func(options *SrvOptions) {
		options.username = username
	}
}

func WithPassword(password string) SrvOption {
	return func(options *SrvOptions) {
		options.password = password
	}
}

func WithRpcListen(rpcListen string) SrvOption {
	return func(options *SrvOptions) {
		options.rpcListen = rpcListen
	}
}

func WithTestNet(testnet bool) SrvOption {
	return func(options *SrvOptions) {
		options.testnet = testnet
	}
}

func WithRpcConnect(rpcConnect string) SrvOption {
	return func(options *SrvOptions) {
		options.rpcConnect = rpcConnect
	}
}

func WithEmbedDB(embedDB bool) SrvOption {
	return func(options *SrvOptions) {
		options.embedDB = embedDB
	}
}

func WithNoApi(noApi bool) SrvOption {
	return func(options *SrvOptions) {
		options.noApi = noApi
	}
}

func WithDataDir(dataDir string) SrvOption {
	return func(options *SrvOptions) {
		options.dataDir = dataDir
	}
}

func WithMysqlAddr(mysqlAddr string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlAddr = mysqlAddr
	}
}

func WithMysqlUser(mysqlUser string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlUser = mysqlUser
	}
}

func WithMysqlPassword(mysqlPassword string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlPassword = mysqlPassword
	}
}

func WithMysqlDBName(mysqlDBName string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlDBName = mysqlDBName
	}
}

func WithDBListenPort(dbListenPort string) SrvOption {
	return func(options *SrvOptions) {
		options.dbListenPort = dbListenPort
	}
}

func WithDBStatusListenPort(dbStatusListenPort string) SrvOption {
	return func(options *SrvOptions) {
		options.dbStatusListenPort = dbStatusListenPort
	}
}

func WithEnablePProf(enablePProf bool) SrvOption {
	return func(options *SrvOptions) {
		options.enablePProf = enablePProf
	}
}

var Cmd = &cobra.Command{
	Use:   "srv",
	Short: "inscription index server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := IndexSrv(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		<-signal.InterruptHandlersDone
	},
}

func init() {
	Cmd.Flags().StringVarP(&srvOptions.username, "user", "u", "root", "btcd rpc server username")
	Cmd.Flags().StringVarP(&srvOptions.password, "password", "P", "root", "btcd rpc server password")
	Cmd.Flags().StringVarP(&srvOptions.rpcListen, "rpc_listen", "l", mainNetRPCListen, "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().BoolVarP(&srvOptions.testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&srvOptions.rpcConnect, "rpc_connect", "s", mainNetRPCConnect, "the URL of RPC server to connect to (default localhost:8334, testnet: localhost:18334)")
	//Cmd.Flags().BoolVarP(&srvOptions.embedDB, "embed_db", "", false, "use embed db")
	Cmd.Flags().BoolVarP(&srvOptions.noApi, "no_api", "", false, "don't start api server")
	//Cmd.Flags().StringVarP(&srvOptions.dataDir, "data_dir", "", "", "embed database data dir")
	Cmd.Flags().StringVarP(&srvOptions.mysqlAddr, "mysql_addr", "d", "127.0.0.1:3306", "inscription index mysql database addr")
	Cmd.Flags().StringVarP(&srvOptions.mysqlUser, "mysql_user", "", "root", "inscription index mysql database user")
	Cmd.Flags().StringVarP(&srvOptions.mysqlPassword, "mysql_pass", "", "root", "inscription index mysql database password")
	Cmd.Flags().StringVarP(&srvOptions.mysqlDBName, "dbname", "", constants.DefaultDBName, "inscription index mysql database name")
	//Cmd.Flags().StringVarP(&srvOptions.dbListenPort, "db_listen", "", "4000", "inscription index database server listen port")
	//Cmd.Flags().StringVarP(&srvOptions.dbStatusListenPort, "db_status_listen", "", "10080", "inscription index database server status listen port")
	Cmd.Flags().BoolVarP(&srvOptions.enablePProf, "pprof", "", false, "enable pprof")
	Cmd.Flags().StringVarP(&srvOptions.indexSats, "index_sats", "", "", "Track location of all satoshis, true/false")
	Cmd.Flags().StringVarP(&srvOptions.indexSpendSats, "index_spend_sats", "", "", "Keep sat index entries of spent outputs, true/false")
	//Cmd.Flags().BoolVarP(&srvOptions.indexNoSyncBlock, "no_sync_block", "", false, "no sync block")
}

func IndexSrv(opts ...SrvOption) error {
	for _, v := range opts {
		v(srvOptions)
	}
	if srvOptions.testnet {
		util.ActiveNet = &netparams.TestNet3Params
		if srvOptions.rpcListen == mainNetRPCListen {
			srvOptions.rpcListen = testNetRPCListen
		}
		if srvOptions.rpcConnect == mainNetRPCConnect {
			srvOptions.rpcConnect = testNetRPCConnect
		}
	}
	srvOptions.dataDir = constants.DBDatDir(srvOptions.testnet)

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "index.log"), false)
	log.InitLogRotator(logFile)

	db, err := dao.NewDB(
		dao.WithAddr(srvOptions.mysqlAddr),
		dao.WithUser(srvOptions.mysqlUser),
		dao.WithPassword(srvOptions.mysqlPassword),
		dao.WithDBName(srvOptions.mysqlDBName),
		dao.WithDataDir(srvOptions.dataDir),
		dao.WithEmbedDB(srvOptions.embedDB),
		dao.WithServerPort(srvOptions.dbListenPort),
		dao.WithStatusPort(srvOptions.dbStatusListenPort),
		dao.WithAutoMigrateTables(tables.Tables...),
	)
	if err != nil {
		return err
	}

	cli, err := btcd.NewClient(srvOptions.rpcConnect, srvOptions.username, srvOptions.password)
	if err != nil {
		return err
	}
	batchCli, err := btcd.NewBatchClient(srvOptions.rpcConnect, srvOptions.username, srvOptions.password)
	if err != nil {
		return err
	}

	indexer := index.NewIndexer(
		index.WithDB(db),
		index.WithClient(cli),
		index.WithBatchClient(batchCli),
		index.WithIndexSats(srvOptions.indexSats),
		index.WithIndexSpendSats(srvOptions.indexSpendSats),
		index.WithTidbSessionMemLimit(constants.TidbSessionMemLimit),
	)
	indexer.Start()
	signal.AddInterruptHandler(func() {
		indexer.Stop()
	})

	if !srvOptions.noApi {
		h, err := handle.New(
			handle.WithDB(db),
			handle.WithClient(cli),
			handle.WithAddr(srvOptions.rpcListen),
			handle.WithTestNet(srvOptions.testnet),
			handle.WithEnablePProf(srvOptions.enablePProf),
		)
		if err != nil {
			return err
		}
		if err := h.Run(); err != nil {
			return err
		}
	}
	return nil
}
