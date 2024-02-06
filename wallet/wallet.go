package wallet

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/chain"
	"github.com/btcsuite/btcwallet/rpc/legacyrpc"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/inscription-c/insc/constants"
	log2 "github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/internal/util"
	"github.com/inscription-c/insc/wallet/log"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"
)

const (
	btcdRpcListenMainNet = "127.0.0.1:8334"
	btcdRpcListenTestNet = "127.0.0.1:18334"
)

var (
	cfg *Config
)

type walletOptions struct {
	Username   string
	Password   string
	BtcdUrl    string
	WalletPass string
	Testnet    bool
}

var Options = &walletOptions{}

var Cmd = &cobra.Command{
	Use:   "wallet",
	Short: "wallet embed btcd endpoint",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Main(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		<-signal.InterruptHandlersDone
	},
}

func init() {
	Cmd.Flags().StringVarP(&Options.Username, "user", "u", "root", "btcd rpc server username")
	Cmd.Flags().StringVarP(&Options.Password, "password", "P", "root", "btcd rpc server password")
	Cmd.Flags().StringVarP(&Options.BtcdUrl, "btcd_url", "", "localhost:8334", "Hostname/IP and port of btcd RPC server to connect to (default localhost:8334, testnet: localhost:18334)")
	Cmd.Flags().StringVarP(&Options.WalletPass, "wallet_pass", "", "root", "wallet password")
	Cmd.Flags().BoolVarP(&Options.Testnet, "testnet", "t", false, "bitcoin testnet3")
	if err := Cmd.MarkFlagRequired("wallet_pass"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Main() error {
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "inscription.log"), false)
	log2.InitLogRotator(logFile)
	if err := Wallet(nil); err != nil {
		return err
	}
	return nil
}

// Wallet is a work-around main function that is required since deferred
// functions (such as log flushing) are not called with calls to os.Exit.
// Instead, main runs this function and checks for a non-nil error, at which
// point any defers have already run, and if the error is non-nil, the program
// can be exited with an error exit status.
func Wallet(walletCh chan<- *wallet.Wallet) error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	signal.AddInterruptHandler(func() {
		if log.LogRotator != nil {
			log.LogRotator.Close()
		}
	})

	// Show version at startup.
	log.Log.Infof("Version %s", version())

	if cfg.Profile != "" {
		go func() {
			listenAddr := net.JoinHostPort("", cfg.Profile)
			log.Log.Infof("Profile server listening on %s", listenAddr)
			profileRedirect := http.RedirectHandler("/debug/pprof",
				http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			log.Log.Errorf("%v", http.ListenAndServe(listenAddr, nil))
		}()
	}

	dbDir := networkDir(cfg.AppDataDir.Value, util.ActiveNet.Params)
	loader := wallet.NewLoader(
		util.ActiveNet.Params, dbDir, true, cfg.DBTimeout, 250,
	)

	// Create and start HTTP server to serve wallet client connections.
	// This will be updated with the wallet and chain server RPC client
	// created below after each is created.
	legacyRPCServer, err := startRPCServers(loader)
	if err != nil {
		log.Log.Errorf("Unable to create RPC servers: %v", err)
		return err
	}

	// Create and start chain RPC client, so it's ready to connect to
	// the wallet when loaded later.
	if !cfg.NoInitialLoad {
		go rpcClientConnectLoop(legacyRPCServer, loader)
	}

	loader.RunAfterLoad(func(w *wallet.Wallet) {
		startWalletRPCServices(w, legacyRPCServer)
		if walletCh != nil {
			go func() {
				walletCh <- w
			}()
		}
	})

	if !cfg.NoInitialLoad {
		// Load the wallet database.  It must have been created already
		// or this will return an appropriate error.
		_, err = loader.OpenExistingWallet([]byte(cfg.WalletPass), true)
		if err != nil {
			log.Log.Error(err)
			return err
		}
	}

	// Add interrupt handlers shutdown the various process components
	// before exiting.  Interrupt handlers run in LIFO order, so the wallet
	// (which should be closed last) is added first.
	signal.AddInterruptHandler(func() {
		err := loader.UnloadWallet()
		if err != nil && !errors.Is(err, wallet.ErrNotLoaded) {
			log.Log.Errorf("Failed to close wallet: %v", err)
		}
	})
	if legacyRPCServer != nil {
		signal.AddInterruptHandler(func() {
			log.Log.Warn("Stopping legacy RPC server...")
			legacyRPCServer.Stop()
			log.Log.Info("Legacy RPC server shutdown")
		})
		go func() {
			<-legacyRPCServer.RequestProcessShutdown()
			signal.SimulateInterrupt()
		}()
	}
	return nil
}

// rpcClientConnectLoop continuously attempts a connection to the consensus RPC
// server.  When a connection is established, the client is used to sync the
// loaded wallet, either immediately or when loaded at a later time.
//
// The legacy RPC is optional.  If set, the connected RPC client will be
// associated with the server for RPC passthrough and to enable additional
// methods.
func rpcClientConnectLoop(legacyRPCServer *legacyrpc.Server, loader *wallet.Loader) {
	for {
		var (
			chainClient chain.Interface
			err         error
		)

		chainClient, err = startChainRPC(nil)
		if err != nil {
			log.Log.Errorf("Unable to open connection to consensus RPC server: %v", err)
			continue
		}

		// Rather than inlining this logic directly into the loader
		// callback, a function variable is used to avoid running any of
		// this after the client disconnects by setting it to nil.  This
		// prevents the callback from associating a wallet loaded at a
		// later time with a client that has already disconnected.  A
		// mutex is used to make this concurrent safe.
		associateRPCClient := func(w *wallet.Wallet) {
			w.SynchronizeRPC(chainClient)
			if legacyRPCServer != nil {
				legacyRPCServer.SetChainServer(chainClient)
			}
		}
		mu := new(sync.Mutex)
		loader.RunAfterLoad(func(w *wallet.Wallet) {
			mu.Lock()
			associate := associateRPCClient
			mu.Unlock()
			if associate != nil {
				associate(w)
			}
		})

		chainClient.WaitForShutdown()

		mu.Lock()
		associateRPCClient = nil
		mu.Unlock()

		loadedWallet, ok := loader.LoadedWallet()
		if ok {
			// Do not attempt a reconnect when the wallet was
			// explicitly stopped.
			if loadedWallet.ShuttingDown() {
				return
			}

			loadedWallet.SetChainSynced(false)

			loadedWallet.Stop()
			loadedWallet.WaitForShutdown()
			loadedWallet.Start()
		}
	}
}

// startChainRPC opens an RPC client connection to a btcd server for blockchain
// services.  This function uses the RPC options from the global config and
// there is no recovery in case the server is not available or if there is an
// authentication error.  Instead, all requests to the client will simply error.
func startChainRPC(certs []byte) (*chain.RPCClient, error) {
	log.Log.Infof("Attempting RPC client connection to %v", cfg.RPCConnect)
	rpcc, err := chain.NewRPCClient(util.ActiveNet.Params, cfg.RPCConnect, cfg.BtcdUsername, cfg.BtcdPassword, certs, cfg.DisableClientTLS, 0)
	if err != nil {
		return nil, err
	}
	err = rpcc.Start()
	return rpcc, err
}
