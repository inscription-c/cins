package server

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/insc/btcd/rpcclient"
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

// SrvOptions is a struct that holds the configuration options for the server.
// It includes fields for the username, password, rpcListen, testnet, rpcConnect,
// embedDB, noApi, dataDir, mysqlAddr, mysqlUser, mysqlPassword, mysqlDBName,
// dbListenPort, dbStatusListenPort, enablePProf, indexSats, indexSpendSats, and indexNoSyncBlock.
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

// SrvOption is a function type that takes a pointer to a SrvOptions struct as a parameter.
// It is used to set the fields of the SrvOptions struct.
type SrvOption func(*SrvOptions)

// WithUserName is a function that returns a SrvOption.
// The returned SrvOption sets the username field of the SrvOptions struct to the provided username.
func WithUserName(username string) SrvOption {
	return func(options *SrvOptions) {
		options.username = username
	}
}

// WithPassword is a function that returns a SrvOption.
// The returned SrvOption sets the password field of the SrvOptions struct to the provided password.
func WithPassword(password string) SrvOption {
	return func(options *SrvOptions) {
		options.password = password
	}
}

// WithRpcListen is a function that returns a SrvOption.
// The returned SrvOption sets the rpcListen field of the SrvOptions struct to the provided rpcListen.
func WithRpcListen(rpcListen string) SrvOption {
	return func(options *SrvOptions) {
		options.rpcListen = rpcListen
	}
}

// WithTestNet is a function that returns a SrvOption.
// The returned SrvOption sets the testnet field of the SrvOptions struct to the provided testnet.
func WithTestNet(testnet bool) SrvOption {
	return func(options *SrvOptions) {
		options.testnet = testnet
	}
}

// WithRpcConnect is a function that returns a SrvOption.
// The returned SrvOption sets the rpcConnect field of the SrvOptions struct to the provided rpcConnect.
func WithRpcConnect(rpcConnect string) SrvOption {
	return func(options *SrvOptions) {
		options.rpcConnect = rpcConnect
	}
}

// WithEmbedDB is a function that returns a SrvOption.
// The returned SrvOption sets the embedDB field of the SrvOptions struct to the provided embedDB.
func WithEmbedDB(embedDB bool) SrvOption {
	return func(options *SrvOptions) {
		options.embedDB = embedDB
	}
}

// WithNoApi is a function that returns a SrvOption.
// The returned SrvOption sets the noApi field of the SrvOptions struct to the provided noApi.
func WithNoApi(noApi bool) SrvOption {
	return func(options *SrvOptions) {
		options.noApi = noApi
	}
}

// WithDataDir is a function that returns a SrvOption.
// The returned SrvOption sets the dataDir field of the SrvOptions struct to the provided dataDir.
func WithDataDir(dataDir string) SrvOption {
	return func(options *SrvOptions) {
		options.dataDir = dataDir
	}
}

// WithMysqlAddr is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlAddr field of the SrvOptions struct to the provided mysqlAddr.
func WithMysqlAddr(mysqlAddr string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlAddr = mysqlAddr
	}
}

// WithMysqlUser is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlUser field of the SrvOptions struct to the provided mysqlUser.
func WithMysqlUser(mysqlUser string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlUser = mysqlUser
	}
}

// WithMysqlPassword is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlPassword field of the SrvOptions struct to the provided mysqlPassword.
func WithMysqlPassword(mysqlPassword string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlPassword = mysqlPassword
	}
}

// WithMysqlDBName is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlDBName field of the SrvOptions struct to the provided mysqlDBName.
func WithMysqlDBName(mysqlDBName string) SrvOption {
	return func(options *SrvOptions) {
		options.mysqlDBName = mysqlDBName
	}
}

// WithDBListenPort is a function that returns a SrvOption.
// The returned SrvOption sets the dbListenPort field of the SrvOptions struct to the provided dbListenPort.
func WithDBListenPort(dbListenPort string) SrvOption {
	return func(options *SrvOptions) {
		options.dbListenPort = dbListenPort
	}
}

// WithDBStatusListenPort is a function that returns a SrvOption.
// The returned SrvOption sets the dbStatusListenPort field of the SrvOptions struct to the provided dbStatusListenPort.
func WithDBStatusListenPort(dbStatusListenPort string) SrvOption {
	return func(options *SrvOptions) {
		options.dbStatusListenPort = dbStatusListenPort
	}
}

// WithEnablePProf is a function that returns a SrvOption.
// The returned SrvOption sets the enablePProf field of the SrvOptions struct to the provided enablePProf.
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

	// Create a new database instance using the server options.
	// The database is configured with the MySQL address, user, password, and database name from the server options.
	// The data directory and embedded database flag from the server options are also used.
	// The server port and status port for the database are set from the server options.
	// The tables to auto-migrate in the database are set to the tables from the tables package.
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

	// Create a new RPC client using the server options.
	// The client is configured with the RPC connect, username, and password from the server options.
	cli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(srvOptions.rpcConnect),
		rpcclient.WithClientUser(srvOptions.username),
		rpcclient.WithClientPassword(srvOptions.password),
	)
	if err != nil {
		return err
	}

	// Create a new batch RPC client using the server options.
	// The batch client is configured with the RPC connect, username, and password from the server options.
	// The batch client is also set to operate in batch mode.
	batchCli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(srvOptions.rpcConnect),
		rpcclient.WithClientUser(srvOptions.username),
		rpcclient.WithClientPassword(srvOptions.password),
		rpcclient.WithClientBatch(true),
	)
	if err != nil {
		return err
	}

	// Create a new indexer using the database, the client, the batch client, the index sats, the index spend sats, and the TiDB session memory limit.
	// The indexer is configured with the database, the client, and the batch client.
	// The indexer is also configured with the index sats and index spend sats from the server options.
	// The TiDB session memory limit is set to the TiDB session memory limit constant from the constants package.
	indexer := index.NewIndexer(
		index.WithDB(db),
		index.WithClient(cli),
		index.WithBatchClient(batchCli),
		index.WithIndexSats(srvOptions.indexSats),
		index.WithIndexSpendSats(srvOptions.indexSpendSats),
		index.WithTidbSessionMemLimit(constants.TidbSessionMemLimit),
	)
	// Start the indexer.
	indexer.Start()
	// Add an interrupt handler that stops the indexer when an interrupt signal is received.
	signal.AddInterruptHandler(func() {
		indexer.Stop()
	})

	// If the no API field of the server options is false, create and run a new handler.
	if !srvOptions.noApi {
		// Create a new handler using the database, the client, the RPC listen, the testnet,
		//and to enable pprof from the server options.
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
		// Run the handler.
		// If there is an error running the handler, return the error.
		if err := h.Run(); err != nil {
			return err
		}
	}
	return nil
}
