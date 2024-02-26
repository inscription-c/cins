package server

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index"
	"github.com/inscription-c/cins/inscription/index/dao"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/inscription/log"
	"github.com/inscription-c/cins/inscription/server/handle"
	"github.com/inscription-c/cins/internal/signal"
	"github.com/inscription-c/cins/internal/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	srvOptions = &SrvOptions{}

	mainNetRPCListen  = ":8335"
	testNetRPCListen  = ":18335"
	mainNetRPCConnect = "http://localhost:8334"
	testNetRPCConnect = "http://localhost:18334"
)

// SrvOptions is a struct that holds the configuration options for the server.
// It includes fields for the username, password, rpcListen, testnet, rpcConnect,
// embedDB, noApi, dataDir, mysqlAddr, mysqlUser, mysqlPassword, mysqlDBName,
// dbListenPort, dbStatusListenPort, enablePProf, indexSats, indexSpendSats, and indexNoSyncBlock.
type SrvOptions struct {
	configFile     string
	Testnet        bool   `yaml:"testnet"`
	RpcListen      string `yaml:"rpc_listen"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	RpcConnect     string `yaml:"rpc_connect"`
	NoApi          bool   `yaml:"no_api"`
	IndexSats      string `yaml:"index_sats"`
	IndexSpendSats string `yaml:"index_spend_sats"`
	Mysql          struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DB       string `yaml:"db"`
	} `yaml:"mysql"`
	EnablePProf bool `yaml:"pprof"`
}

// SrvOption is a function type that takes a pointer to a SrvOptions struct as a parameter.
// It is used to set the fields of the SrvOptions struct.
type SrvOption func(*SrvOptions)

// WithUserName is a function that returns a SrvOption.
// The returned SrvOption sets the username field of the SrvOptions struct to the provided username.
func WithUserName(username string) SrvOption {
	return func(options *SrvOptions) {
		options.Username = username
	}
}

// WithPassword is a function that returns a SrvOption.
// The returned SrvOption sets the password field of the SrvOptions struct to the provided password.
func WithPassword(password string) SrvOption {
	return func(options *SrvOptions) {
		options.Password = password
	}
}

// WithRpcListen is a function that returns a SrvOption.
// The returned SrvOption sets the rpcListen field of the SrvOptions struct to the provided rpcListen.
func WithRpcListen(rpcListen string) SrvOption {
	return func(options *SrvOptions) {
		options.RpcListen = rpcListen
	}
}

// WithTestNet is a function that returns a SrvOption.
// The returned SrvOption sets the testnet field of the SrvOptions struct to the provided testnet.
func WithTestNet(testnet bool) SrvOption {
	return func(options *SrvOptions) {
		options.Testnet = testnet
	}
}

// WithRpcConnect is a function that returns a SrvOption.
// The returned SrvOption sets the rpcConnect field of the SrvOptions struct to the provided rpcConnect.
func WithRpcConnect(rpcConnect string) SrvOption {
	return func(options *SrvOptions) {
		options.RpcConnect = rpcConnect
	}
}

// WithNoApi is a function that returns a SrvOption.
// The returned SrvOption sets the noApi field of the SrvOptions struct to the provided noApi.
func WithNoApi(noApi bool) SrvOption {
	return func(options *SrvOptions) {
		options.NoApi = noApi
	}
}

// WithMysqlAddr is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlAddr field of the SrvOptions struct to the provided mysqlAddr.
func WithMysqlAddr(mysqlAddr string) SrvOption {
	return func(options *SrvOptions) {
		options.Mysql.Addr = mysqlAddr
	}
}

// WithMysqlUser is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlUser field of the SrvOptions struct to the provided mysqlUser.
func WithMysqlUser(mysqlUser string) SrvOption {
	return func(options *SrvOptions) {
		options.Mysql.User = mysqlUser
	}
}

// WithMysqlPassword is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlPassword field of the SrvOptions struct to the provided mysqlPassword.
func WithMysqlPassword(mysqlPassword string) SrvOption {
	return func(options *SrvOptions) {
		options.Mysql.Password = mysqlPassword
	}
}

// WithMysqlDBName is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlDBName field of the SrvOptions struct to the provided mysqlDBName.
func WithMysqlDBName(mysqlDBName string) SrvOption {
	return func(options *SrvOptions) {
		options.Mysql.DB = mysqlDBName
	}
}

// WithEnablePProf is a function that returns a SrvOption.
// The returned SrvOption sets the enablePProf field of the SrvOptions struct to the provided enablePProf.
func WithEnablePProf(enablePProf bool) SrvOption {
	return func(options *SrvOptions) {
		options.EnablePProf = enablePProf
	}
}

var Cmd = &cobra.Command{
	Use:   "indexer",
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
	Cmd.Flags().StringVarP(&srvOptions.configFile, "config", "c", "", "config file path")
	Cmd.Flags().BoolVarP(&srvOptions.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&srvOptions.Username, "user", "u", "root", "bitcoin rpc server username")
	Cmd.Flags().StringVarP(&srvOptions.Password, "password", "P", "root", "bitcoin rpc server password")
	Cmd.Flags().StringVarP(&srvOptions.RpcListen, "rpc_listen", "l", "", "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().StringVarP(&srvOptions.RpcConnect, "rpc_connect", "s", "", "the bitcoin backend URL of RPC server to connect to (default http://localhost:8334, testnet: http://localhost:18334)")
	Cmd.Flags().BoolVarP(&srvOptions.NoApi, "no_api", "", false, "don't start api server")
	Cmd.Flags().StringVarP(&srvOptions.Mysql.Addr, "mysql_addr", "d", "", "inscription index mysql database addr")
	Cmd.Flags().StringVarP(&srvOptions.Mysql.User, "mysql_user", "", "root", "inscription index mysql database user")
	Cmd.Flags().StringVarP(&srvOptions.Mysql.Password, "mysql_pass", "", "root", "inscription index mysql database password")
	Cmd.Flags().StringVarP(&srvOptions.Mysql.DB, "db", "", "", "inscription index mysql database name")
	Cmd.Flags().BoolVarP(&srvOptions.EnablePProf, "pprof", "", false, "enable pprof")
	Cmd.Flags().StringVarP(&srvOptions.IndexSats, "index_sats", "", "", "Track location of all satoshis, true/false")
	Cmd.Flags().StringVarP(&srvOptions.IndexSpendSats, "index_spend_sats", "", "", "Keep sat index entries of spent outputs, true/false")
}

func IndexSrv(opts ...SrvOption) error {
	if srvOptions.configFile != "" {
		configFile, err := os.Open(srvOptions.configFile)
		if err != nil {
			return err
		}
		defer configFile.Close()
		if err := yaml.NewDecoder(configFile).Decode(srvOptions); err != nil {
			return err
		}
	}

	for _, v := range opts {
		v(srvOptions)
	}

	if srvOptions.Mysql.DB == "" {
		srvOptions.Mysql.DB = constants.DefaultDBName
	}
	if srvOptions.Mysql.Addr == "" {
		srvOptions.Mysql.Addr = "127.0.0.1:3306"
	}
	if srvOptions.Testnet {
		util.ActiveNet = &netparams.TestNet3Params
		if srvOptions.RpcListen == "" {
			srvOptions.RpcListen = testNetRPCListen
		}
		if srvOptions.RpcConnect == "" {
			srvOptions.RpcConnect = testNetRPCConnect
		}
	} else {
		if srvOptions.RpcListen == "" {
			srvOptions.RpcListen = mainNetRPCListen
		}
		if srvOptions.RpcConnect == "" {
			srvOptions.RpcConnect = mainNetRPCConnect
		}
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logDir := filepath.Join(constants.AppName, "inscription", "logs", "index.log")
	logFile := btcutil.AppDataDir(logDir, false)
	log.InitLogRotator(logFile)

	// Create a new database instance using the server options.
	// The database is configured with the MySQL address, user, password, and database name from the server options.
	// The data directory and embedded database flag from the server options are also used.
	// The server port and status port for the database are set from the server options.
	// The tables to auto-migrate in the database are set to the tables from the tables package.
	db, err := dao.NewDB(
		dao.WithAddr(srvOptions.Mysql.Addr),
		dao.WithUser(srvOptions.Mysql.User),
		dao.WithPassword(srvOptions.Mysql.Password),
		dao.WithDBName(srvOptions.Mysql.DB),
		dao.WithAutoMigrateTables(tables.Tables...),
	)
	if err != nil {
		return err
	}

	// Create a new RPC client using the server options.
	// The client is configured with the RPC connect, username, and password from the server options.
	cli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(srvOptions.RpcConnect),
		rpcclient.WithClientUser(srvOptions.Username),
		rpcclient.WithClientPassword(srvOptions.Password),
	)
	if err != nil {
		return err
	}

	// Create a new batch RPC client using the server options.
	// The batch client is configured with the RPC connect, username, and password from the server options.
	// The batch client is also set to operate in batch mode.
	batchCli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(srvOptions.RpcConnect),
		rpcclient.WithClientUser(srvOptions.Username),
		rpcclient.WithClientPassword(srvOptions.Password),
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
		index.WithIndexSats(srvOptions.IndexSats),
		index.WithIndexSpendSats(srvOptions.IndexSpendSats),
		index.WithTidbSessionMemLimit(constants.TidbSessionMemLimit),
	)
	// Start the indexer.
	indexer.Start()
	// Add an interrupt handler that stops the indexer when an interrupt signal is received.
	signal.AddInterruptHandler(func() {
		indexer.Stop()
	})

	// If the no API field of the server options is false, create and run a new handler.
	if !srvOptions.NoApi {
		// Create a new handler using the database, the client, the RPC listen, the testnet,
		//and to enable pprof from the server options.
		h, err := handle.New(
			handle.WithDB(db),
			handle.WithClient(cli),
			handle.WithAddr(srvOptions.RpcListen),
			handle.WithTestNet(srvOptions.Testnet),
			handle.WithEnablePProf(srvOptions.EnablePProf),
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
