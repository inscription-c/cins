package server

import (
	"fmt"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/inscription/index"
	"github.com/dotbitHQ/insc/inscription/server/handle"
	"github.com/dotbitHQ/insc/internal/cfgutil"
	"github.com/dotbitHQ/insc/internal/signal"
	"github.com/dotbitHQ/insc/wallet"
	"github.com/spf13/cobra"
	"net"
	"os"
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
	Cmd.Flags().StringVarP(&config.Username, "user", "u", "", "btcd rpc server username")
	Cmd.Flags().StringVarP(&config.Password, "password", "P", "", "btcd rpc server password")
	Cmd.Flags().StringVarP(&config.IndexDir, "indexdir", "i", config.IndexDir, fmt.Sprintf("inscriptions index database dir. Default `%s`", config.IndexDir))
	Cmd.Flags().StringVarP(&config.Rpclisten, "rpclisten", "l", ":8335", "rpc server listen address. Default `mainnet :8335, testnet :18335`")
	Cmd.Flags().BoolVarP(&config.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&config.RpcConnect, "rpcconnect", "s", "localhost:8334", "the URL of wallet RPC server to connect to (default localhost:8334, testnet: localhost:18334)")
}

func srv() error {
	if config.Testnet {
		activeNet = &netparams.TestNet3Params
	}
	localhostListeners := map[string]struct{}{
		"localhost": {},
		"127.0.0.1": {},
		"::1":       {},
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
		if _, ok := localhostListeners[RPCHost]; ok {
			disableTls = true
		}
	}

	db, err := index.DB(config.IndexDir)
	if err != nil {
		return err
	}

	cli, err := wallet.NewWalletClient(config.RpcConnect, config.Username, config.Password, disableTls)
	if err != nil {
		return err
	}
	batchCli, err := wallet.NewBatchClient(config.RpcConnect, config.Username, config.Password, disableTls)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		cli.Shutdown()
		batchCli.Shutdown()
	})

	h, err := handle.New(
		handle.WithAddr(config.Rpclisten),
		handle.WithDB(db),
		handle.WithTestNet(config.Testnet),
		handle.WithClient(cli),
		handle.WithBatchClient(batchCli),
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
		WithBatchIndex(batchCli),
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
