package server

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/insc/config"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/inscription/server/handle"
	"github.com/inscription-c/insc/internal/cfgutil"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/wallet"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
)

var activeNet = &netparams.MainNetParams

var Cmd = &cobra.Command{
	Use:   "srv",
	Short: "inscription index server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := srv(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		signal.SimulateInterrupt()
		<-signal.InterruptHandlersDone
	},
}

func init() {
	datadir := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "index"), false)
	Cmd.Flags().StringVarP(&config.Username, "user", "u", "", "btcd rpc server username")
	Cmd.Flags().StringVarP(&config.Password, "password", "P", "", "btcd rpc server password")
	Cmd.Flags().StringVarP(&config.Rpclisten, "rpclisten", "l", ":8335", "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().BoolVarP(&config.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&config.RpcConnect, "rpcconnect", "s", "localhost:8334", "the URL of wallet RPC server to connect to (default localhost:8334, testnet: localhost:18334)")
	Cmd.Flags().BoolVarP(&config.NoEmbedDB, "noembeddb", "", false, "don't embed db")
	Cmd.Flags().StringVarP(&config.DataDir, "datadir", "", datadir, "embed database data dir")
	Cmd.Flags().StringVarP(&config.MysqlAddr, "mysqladdr", "d", "127.0.0.1:4000", "inscription index mysql database addr")
	Cmd.Flags().StringVarP(&config.MysqlUser, "mysqluser", "", "", "inscription index mysql database user")
	Cmd.Flags().StringVarP(&config.MysqlPassword, "mysqlpass", "", "", "inscription index mysql database password")
	Cmd.Flags().StringVarP(&config.MysqlDBName, "dbname", "", constants.DefaultDBName, "inscription index mysql database name")
	Cmd.Flags().StringVarP(&config.DBListenPort, "dblisten", "", "4000", "inscription index database server listen port")
	Cmd.Flags().StringVarP(&config.DBStatusListenPort, "dbstatuslisten", "", "10080", "inscription index database server status listen port")
}

func srv() error {
	if config.Testnet {
		activeNet = &netparams.TestNet3Params
	}

	disableTls := false
	if config.RpcConnect != "" {
		rpcConnect, err := cfgutil.NormalizeAddress(config.RpcConnect, activeNet.RPCClientPort)
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
		dao.WithAddr(config.MysqlAddr),
		dao.WithUser(config.MysqlUser),
		dao.WithPassword(config.MysqlPassword),
		dao.WithDBName(config.MysqlDBName),
		dao.WithLogger(log.Gorm),
		dao.WithDataDir(config.DataDir),
		dao.WithNoEmbedDB(config.NoEmbedDB),
		dao.WithServerPort(config.DBListenPort),
		dao.WithStatusPort(config.DBStatusListenPort),
		dao.WithAutoMigrateTables(
			&tables.BlockInfo{},
			&tables.Inscriptions{},
			&tables.OutpointSatRange{},
			&tables.OutpointValue{},
			&tables.Sat{},
			&tables.SatPoint{},
			&tables.SatSatPoint{},
			&tables.Statistic{},
		),
	)
	if err != nil {
		return err
	}

	cli, err := wallet.NewWalletClient(config.RpcConnect, config.Username, config.Password, disableTls)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		cli.Shutdown()
	})

	h, err := handle.New(
		handle.WithAddr(config.Rpclisten),
		handle.WithDB(db),
		handle.WithTestNet(config.Testnet),
		handle.WithClient(cli),
	)
	if err != nil {
		return err
	}

	idx := index.NewIndexer(
		index.WithDB(db),
		index.WithClient(cli),
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
