package server

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/getsentry/sentry-go"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index"
	"github.com/inscription-c/cins/inscription/index/dao"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/inscription/log"
	"github.com/inscription-c/cins/inscription/server/config"
	"github.com/inscription-c/cins/inscription/server/handle"
	sentry2 "github.com/inscription-c/cins/internal/sentry"
	"github.com/inscription-c/cins/pkg/signal"
	"github.com/inscription-c/cins/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	mainNetRPCListen  = ":8335"
	testNetRPCListen  = ":18335"
	mainNetRPCConnect = "http://localhost:8334"
	testNetRPCConnect = "http://localhost:18334"
)

// SrvOption is a function type that takes a pointer to a config.SrvConfigs struct as a parameter.
// It is used to set the fields of the config.SrvConfigs struct.
type SrvOption func(*config.SrvConfigs)

// WithUserName is a function that returns a SrvOption.
// The returned SrvOption sets the username field of the config.SrvConfigs struct to the provided username.
func WithUserName(username string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Chain.Username = username
	}
}

// WithPassword is a function that returns a SrvOption.
// The returned SrvOption sets the password field of the config.SrvConfigs struct to the provided password.
func WithPassword(password string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Chain.Password = password
	}
}

// WithRpcListen is a function that returns a SrvOption.
// The returned SrvOption sets the rpcListen field of the config.SrvConfigs struct to the provided rpcListen.
func WithRpcListen(rpcListen string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Server.RpcListen = rpcListen
	}
}

// WithTestNet is a function that returns a SrvOption.
// The returned SrvOption sets the testnet field of the config.SrvConfigs struct to the provided testnet.
func WithTestNet(testnet bool) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Server.Testnet = testnet
	}
}

// WithRpcConnect is a function that returns a SrvOption.
// The returned SrvOption sets the rpcConnect field of the config.SrvConfigs struct to the provided rpcConnect.
func WithRpcConnect(rpcConnect string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Chain.Url = rpcConnect
	}
}

// WithNoApi is a function that returns a SrvOption.
// The returned SrvOption sets the noApi field of the config.SrvConfigs struct to the provided noApi.
func WithNoApi(noApi bool) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Server.NoApi = noApi
	}
}

// WithMysqlAddr is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlAddr field of the config.SrvConfigs struct to the provided mysqlAddr.
func WithMysqlAddr(mysqlAddr string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.DB.Mysql.Addr = mysqlAddr
	}
}

// WithMysqlUser is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlUser field of the config.SrvConfigs struct to the provided mysqlUser.
func WithMysqlUser(mysqlUser string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.DB.Mysql.User = mysqlUser
	}
}

// WithMysqlPassword is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlPassword field of the config.SrvConfigs struct to the provided mysqlPassword.
func WithMysqlPassword(mysqlPassword string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.DB.Mysql.Password = mysqlPassword
	}
}

// WithMysqlDBName is a function that returns a SrvOption.
// The returned SrvOption sets the mysqlDBName field of the config.SrvConfigs struct to the provided mysqlDBName.
func WithMysqlDBName(mysqlDBName string) SrvOption {
	return func(options *config.SrvConfigs) {
		options.DB.Mysql.DB = mysqlDBName
	}
}

// WithEnablePProf is a function that returns a SrvOption.
// The returned SrvOption sets the enablePProf field of the config.SrvConfigs struct to the provided enablePProf.
func WithEnablePProf(enablePProf bool) SrvOption {
	return func(options *config.SrvConfigs) {
		options.Server.EnablePProf = enablePProf
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

var configFilePath string

func init() {
	Cmd.Flags().StringVarP(&configFilePath, "config", "c", "", "config file path")
	Cmd.Flags().BoolVarP(&config.SrvCfg.Server.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&config.SrvCfg.Server.RpcListen, "rpc_listen", "l", "", "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().StringVarP(&config.SrvCfg.Chain.Url, "chain_url", "s", "", "the bitcoin backend URL of RPC server to connect to (default http://localhost:8334, testnet: http://localhost:18334)")
	Cmd.Flags().StringVarP(&config.SrvCfg.Chain.Username, "chain_user", "u", "root", "bitcoin rpc server username")
	Cmd.Flags().StringVarP(&config.SrvCfg.Chain.Password, "chain_password", "P", "root", "bitcoin rpc server password")
	Cmd.Flags().BoolVarP(&config.SrvCfg.Server.NoApi, "no_api", "", false, "don't start api server")
	Cmd.Flags().StringVarP(&config.SrvCfg.DB.Mysql.Addr, "mysql_addr", "d", "", "inscription index mysql database addr")
	Cmd.Flags().StringVarP(&config.SrvCfg.DB.Mysql.User, "mysql_user", "", "root", "inscription index mysql database user")
	Cmd.Flags().StringVarP(&config.SrvCfg.DB.Mysql.Password, "mysql_pass", "", "root", "inscription index mysql database password")
	Cmd.Flags().StringVarP(&config.SrvCfg.DB.Mysql.DB, "db", "", "", "inscription index mysql database name")
	Cmd.Flags().BoolVarP(&config.SrvCfg.Server.EnablePProf, "pprof", "", false, "enable pprof")
	Cmd.Flags().StringVarP(&config.SrvCfg.Server.IndexSats, "index_sats", "", "", "Track location of all satoshis, true/false")
	Cmd.Flags().StringVarP(&config.SrvCfg.Server.IndexSpendSats, "index_spend_sats", "", "", "Keep sat index entries of spent outputs, true/false")
	Cmd.Flags().StringVarP(&config.SrvCfg.Sentry.Dsn, "sentry_dsn", "", "", "sentry dsn")
	Cmd.Flags().Float64VarP(&config.SrvCfg.Sentry.TracesSampleRate, "sentry_traces_sample_rate", "", 1.0, "sentry traces sample rate")
	Cmd.Flags().BoolVarP(&config.SrvCfg.Server.Prometheus, "prometheus", "", false, "enable prometheus metrics")
	Cmd.Flags().StringSliceVarP(&config.SrvCfg.Origins, "origins", "", []string{}, "allowed origins for CORS")
}

func IndexSrv(opts ...SrvOption) error {
	if configFilePath != "" {
		configFile, err := os.Open(configFilePath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		if err := yaml.NewDecoder(configFile).Decode(config.SrvCfg); err != nil {
			return err
		}
	}

	for _, v := range opts {
		v(config.SrvCfg)
	}

	if config.SrvCfg.DB.Mysql.DB == "" {
		config.SrvCfg.DB.Mysql.DB = constants.DefaultDBName
	}
	if config.SrvCfg.DB.Mysql.Addr == "" {
		config.SrvCfg.DB.Mysql.Addr = "127.0.0.1:3306"
	}
	if config.SrvCfg.Server.Testnet {
		util.ActiveNet = &netparams.TestNet3Params
		if config.SrvCfg.Server.RpcListen == "" {
			config.SrvCfg.Server.RpcListen = testNetRPCListen
		}
		if config.SrvCfg.Chain.Url == "" {
			config.SrvCfg.Chain.Url = testNetRPCConnect
		}
	} else {
		if config.SrvCfg.Server.RpcListen == "" {
			config.SrvCfg.Server.RpcListen = mainNetRPCListen
		}
		if config.SrvCfg.Chain.Url == "" {
			config.SrvCfg.Chain.Url = mainNetRPCConnect
		}
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logDir := filepath.Join(constants.AppName, "inscription", "logs", "index.log")
	logFile := btcutil.AppDataDir(logDir, false)
	log.InitLogRotator(logFile)

	// Initialize sentry error reporting.
	if config.SrvCfg.Sentry.Dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			EnableTracing:    true,
			Dsn:              config.SrvCfg.Sentry.Dsn,
			TracesSampleRate: config.SrvCfg.Sentry.TracesSampleRate,
		}); err != nil {
			return err
		}
		defer sentry2.RecoverPanic()
	}

	// Create a new database instance using the server options.
	// The database is configured with the MySQL address, user, password, and database name from the server options.
	// The data directory and embedded database flag from the server options are also used.
	// The server port and status port for the database are set from the server options.
	// The tables to auto-migrate in the database are set to the tables from the tables package.
	db, err := dao.NewDB(
		dao.WithAddr(config.SrvCfg.DB.Mysql.Addr),
		dao.WithUser(config.SrvCfg.DB.Mysql.User),
		dao.WithPassword(config.SrvCfg.DB.Mysql.Password),
		dao.WithDBName(config.SrvCfg.DB.Mysql.DB),
		dao.WithAutoMigrateTables(tables.Tables...),
	)
	if err != nil {
		return err
	}

	// Create a new RPC client using the server options.
	// The client is configured with the RPC connect, username, and password from the server options.
	cli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(config.SrvCfg.Chain.Url),
		rpcclient.WithClientUser(config.SrvCfg.Chain.Username),
		rpcclient.WithClientPassword(config.SrvCfg.Chain.Password),
	)
	if err != nil {
		return err
	}

	// Create a new batch RPC client using the server options.
	// The batch client is configured with the RPC connect, username, and password from the server options.
	// The batch client is also set to operate in batch mode.
	batchCli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(config.SrvCfg.Chain.Url),
		rpcclient.WithClientUser(config.SrvCfg.Chain.Username),
		rpcclient.WithClientPassword(config.SrvCfg.Chain.Password),
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
		index.WithIndexSats(config.SrvCfg.Server.IndexSats),
		index.WithIndexSpendSats(config.SrvCfg.Server.IndexSpendSats),
		index.WithTidbSessionMemLimit(constants.TidbSessionMemLimit),
	)
	// Start the indexer.
	indexer.Start()
	// Add an interrupt handler that stops the indexer when an interrupt signal is received.
	signal.AddInterruptHandler(func() {
		indexer.Stop()
	})

	// If the no API field of the server options is false, create and run a new handler.
	if !config.SrvCfg.Server.NoApi {
		// Create a new handler using the database, the client, the RPC listen, the testnet,
		//and to enable pprof from the server options.
		h, err := handle.New(
			handle.WithDB(db),
			handle.WithClient(cli),
			handle.WithAddr(config.SrvCfg.Server.RpcListen),
			handle.WithTestNet(config.SrvCfg.Server.Testnet),
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
