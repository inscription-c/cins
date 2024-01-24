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
	"github.com/inscription-c/insc/internal/cfgutil"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
)

var (
	activeNet  = &netparams.MainNetParams
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
	noEmbedDB          bool
	dataDir            string
	mysqlAddr          string
	mysqlUser          string
	mysqlPassword      string
	mysqlDBName        string
	dbListenPort       string
	dbStatusListenPort string
	enablePProf        bool
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

func WithNoEmbedDB(noEmbedDB bool) SrvOption {
	return func(options *SrvOptions) {
		options.noEmbedDB = noEmbedDB
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
	Cmd.Flags().StringVarP(&srvOptions.username, "user", "u", "", "btcd rpc server username")
	Cmd.Flags().StringVarP(&srvOptions.password, "password", "P", "", "btcd rpc server password")
	Cmd.Flags().StringVarP(&srvOptions.rpcListen, "rpclisten", "l", mainNetRPCListen, "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().BoolVarP(&srvOptions.testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&srvOptions.rpcConnect, "rpcconnect", "s", mainNetRPCConnect, "the URL of btcd RPC server to connect to (default localhost:8334, testnet: localhost:18334)")
	Cmd.Flags().BoolVarP(&srvOptions.noEmbedDB, "noembeddb", "", false, "don't embed db")
	Cmd.Flags().StringVarP(&srvOptions.dataDir, "datadir", "", "", "embed database data dir")
	Cmd.Flags().StringVarP(&srvOptions.mysqlAddr, "mysqladdr", "d", "127.0.0.1:4000", "inscription index mysql database addr")
	Cmd.Flags().StringVarP(&srvOptions.mysqlUser, "mysqluser", "", "root", "inscription index mysql database user")
	Cmd.Flags().StringVarP(&srvOptions.mysqlPassword, "mysqlpass", "", "", "inscription index mysql database password")
	Cmd.Flags().StringVarP(&srvOptions.mysqlDBName, "dbname", "", constants.DefaultDBName, "inscription index mysql database name")
	Cmd.Flags().StringVarP(&srvOptions.dbListenPort, "dblisten", "", "4000", "inscription index database server listen port")
	Cmd.Flags().StringVarP(&srvOptions.dbStatusListenPort, "dbstatuslisten", "", "10080", "inscription index database server status listen port")
	Cmd.Flags().BoolVarP(&srvOptions.enablePProf, "enablepprof", "", false, "enable pprof")
}

func datDir() string {
	if srvOptions.testnet {
		return btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "index", "testnet"), false)
	} else {
		return btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "index", "mainnet"), false)
	}
}

func IndexSrv(opts ...SrvOption) error {
	for _, v := range opts {
		v(srvOptions)
	}
	if srvOptions.testnet {
		activeNet = &netparams.TestNet3Params
		if srvOptions.rpcListen == mainNetRPCListen {
			srvOptions.rpcListen = testNetRPCListen
		}
		if srvOptions.rpcConnect == mainNetRPCConnect {
			srvOptions.rpcConnect = testNetRPCConnect
		}
	}
	srvOptions.dataDir = datDir()

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "index.log"), false)
	log.InitLogRotator(logFile)

	disableTls := false
	if srvOptions.rpcConnect != "" {
		rpcConnect, err := cfgutil.NormalizeAddress(srvOptions.rpcConnect, activeNet.RPCClientPort)
		if err != nil {
			return err
		}
		RPCHost, _, err := net.SplitHostPort(rpcConnect)
		if err != nil {
			return err
		}
		localhostListeners := map[string]struct{}{
			"localhost": {},
			"127.0.0.1": {},
			"::1":       {},
		}
		if _, ok := localhostListeners[RPCHost]; ok {
			disableTls = true
		}
	}

	db, err := dao.NewDB(
		dao.WithAddr(srvOptions.mysqlAddr),
		dao.WithUser(srvOptions.mysqlUser),
		dao.WithPassword(srvOptions.mysqlPassword),
		dao.WithDBName(srvOptions.mysqlDBName),
		dao.WithLogger(log.Gorm),
		dao.WithDataDir(srvOptions.dataDir),
		dao.WithNoEmbedDB(srvOptions.noEmbedDB),
		dao.WithServerPort(srvOptions.dbListenPort),
		dao.WithStatusPort(srvOptions.dbStatusListenPort),
		dao.WithAutoMigrateTables(
			&tables.BlockInfo{},
			&tables.Inscriptions{},
			&tables.OutpointSatRange{},
			&tables.OutpointValue{},
			&tables.Sat{},
			&tables.SatPoint{},
			&tables.SatSatPoint{},
			&tables.Statistic{},
			&tables.Protocol{},
		),
	)
	if err != nil {
		return err
	}

	cli, err := btcd.NewClient(srvOptions.rpcConnect, srvOptions.username, srvOptions.password, disableTls)
	if err != nil {
		return err
	}
	batchCli, err := btcd.NewBatchClient(srvOptions.rpcConnect, srvOptions.username, srvOptions.password, disableTls)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		cli.Shutdown()
		batchCli.Shutdown()
	})

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

	idx := index.NewIndexer(
		index.WithDB(db),
		index.WithClient(cli),
		index.WithBatchClient(batchCli),
	)
	runner := NewRunner(
		WithClient(cli),
		WithIndex(idx),
	)
	runner.Start()
	signal.AddInterruptHandler(func() {
		runner.Stop()
	})
	if err := h.Run(); err != nil {
		return err
	}
	return nil
}
