package inscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/inscription/log"
	"github.com/inscription-c/cins/pkg/indexer"
	"github.com/inscription-c/cins/pkg/signal"
	"github.com/inscription-c/cins/pkg/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	DefaultTestNet3IndexerUrl = "http://localhost:18335"
	DefaultMainNetIndexerUrl  = "http://localhost:8335"
)

var (
	indexerUrl           string
	walletUrl            string
	walletRpcUser        string
	walletRpcPass        string
	walletPass           string
	testnet              bool
	inscriptionsFilePath string
	postage              = uint64(constants.DefaultPostage)
	compress             bool
	cborMetadata         string
	jsonMetadata         string
	dryRun               bool
	cbrc20               bool
	destination          string
	cInsDescriptionFile  string
	noBackup             bool
)

// InsufficientBalanceError is an error that represents an insufficient balance.
var InsufficientBalanceError = errors.New("InsufficientBalanceError")

func init() {
	Cmd.Flags().StringVarP(&indexerUrl, "indexer_url", "", DefaultMainNetIndexerUrl, "the URL of indexer server (default http://localhost:8335, testnet: http://localhost:18335)")
	Cmd.Flags().StringVarP(&walletUrl, "wallet_url", "", "localhost:8332", "the URL of wallet RPC server to connect to (default http://localhost:8332, testnet: localhost:18332)")
	Cmd.Flags().StringVarP(&walletRpcUser, "wallet_rpc_user", "", "root", "wallet rpc server user")
	Cmd.Flags().StringVarP(&walletRpcPass, "wallet_rpc_pass", "", "root", "wallet rpc server password")
	Cmd.Flags().StringVarP(&walletPass, "wallet_pass", "", "root", "wallet password for master private key")
	Cmd.Flags().BoolVarP(&testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&inscriptionsFilePath, "filepath", "f", "", "inscription file path")
	Cmd.Flags().StringVarP(&cInsDescriptionFile, "c_ins_description", "", "", "cins protocol description.")
	Cmd.Flags().StringVarP(&destination, "dest", "", "", "Send inscription to <DESTINATION> address.")
	Cmd.Flags().Uint64VarP(&postage, "postage", "p", constants.DefaultPostage, "Amount of postage to include in the inscription.")
	Cmd.Flags().BoolVarP(&compress, "compress", "", false, "Compress inscription content with brotli.")
	Cmd.Flags().StringVarP(&cborMetadata, "cbor_metadata", "", "", "Include CBOR in file at <METADATA> as inscription metadata")
	Cmd.Flags().StringVarP(&jsonMetadata, "json_metadata", "", "", "Include JSON in file at <METADATA> converted to CBOR as inscription metadata")
	Cmd.Flags().BoolVarP(&dryRun, "dry_run", "", false, "Don't sign or broadcast transactions.")
	Cmd.Flags().BoolVarP(&cbrc20, "c_brc_20", "", false, "is c-brc-20 protocol, add this flag will auto check protocol content effectiveness")
	Cmd.Flags().BoolVarP(&noBackup, "no_backup", "", false, "Do not back up recovery key.")
	if err := Cmd.MarkFlagRequired("filepath"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("dest"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("c_ins_description"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func configCheck() error {
	if testnet {
		walletUrl = "http://localhost:18332"
		if indexerUrl == DefaultMainNetIndexerUrl {
			indexerUrl = DefaultTestNet3IndexerUrl
		}
		util.ActiveNet = &netparams.TestNet3Params
	}

	//if postage < constants.DustLimit {
	//	return fmt.Errorf("postage must be greater than or equal %d", constants.DustLimit)
	//}
	if postage > constants.MaxPostage {
		return fmt.Errorf("postage must be less than or equal %d", constants.MaxPostage)
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "inscription.log"), false)
	log.InitLogRotator(logFile)

	// unlock condition check
	if _, err := tables.CInsDescriptionFromFile(cInsDescriptionFile); err != nil {
		return err
	}
	return nil
}

// Cmd is a cobra command that runs the inscribe function when executed.
// It also handles any errors returned by the inscribe function.
var Cmd = &cobra.Command{
	Use:   "inscribe",
	Short: "inscription casting",
	Run: func(cmd *cobra.Command, args []string) {
		if err := inscribe(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		signal.SimulateInterrupt()
		<-signal.InterruptHandlersDone
	},
}

// inscribe is a function that performs the inscription process.
// It checks the configuration, gets the UTXO, creates the commit and reveal transactions, and signs and sends the transactions.
// It also handles any errors that occur during these processes.
func inscribe() error {
	// Check the configuration
	if err := configCheck(); err != nil {
		return err
	}

	// Create a new wallet client
	walletCli, err := rpcclient.NewClient(
		rpcclient.WithClientHost(walletUrl),
		rpcclient.WithClientUser(walletRpcUser),
		rpcclient.WithClientPassword(walletRpcPass),
	)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		walletCli.Shutdown()
	})

	// Get the unlock condition from the file path
	cInsDescription, err := tables.CInsDescriptionFromFile(cInsDescriptionFile)
	if err != nil {
		return err
	}

	// Create a new inscription from the file path
	inscription, err := NewFromPath(inscriptionsFilePath,
		WithIndexer(indexer.NewIndexer(indexerUrl)),
		WithWalletClient(walletCli),
		WithPostage(postage),
		WithCInsDescription(cInsDescription),
		WithWalletPass(walletPass),
		WithCborMetadata(cborMetadata),
		WithJsonMetadata(jsonMetadata),
	)
	if err != nil {
		return err
	}

	// Get all UTXO for all unspent addresses and exclude the UTXO where the inscription
	if err := inscription.getUtxo(); err != nil {
		return err
	}

	// Create commit and reveal transactions
	if err := inscription.CreateInscriptionTx(); err != nil {
		return err
	}

	// If it's a dry run, log the success and the transaction IDs and return
	if dryRun {
		log.Log.Info("dry run success")
		out := Output{
			Commit:    inscription.CommitTxId(),
			Reveal:    inscription.RevealTxId(),
			TotalFees: inscription.totalFee,
		}
		outData, _ := json.MarshalIndent(out, "", "\t")
		fmt.Println(string(outData))
		return nil
	}

	// Sign the reveal transaction
	if err := inscription.SignRevealTx(); err != nil {
		return err
	}
	// Sign the commit transaction
	if err := inscription.SignCommitTx(); err != nil {
		return err
	}

	// backup temporary private key
	if !noBackup {
		wif, err := btcutil.NewWIF(inscription.priKey, util.ActiveNet.Params, true)
		if err != nil {
			return err
		}
		if err := walletCli.ImportPrivKey(wif); err != nil {
			return err
		}
	}

	// Send the commit transaction
	commitTxHash, err := walletCli.SendRawTransaction(inscription.commitTx, false)
	if err != nil {
		return err
	}
	log.Log.Info("commitTxSendSuccess", commitTxHash)

	// Send the reveal transaction
	revealTxHash, err := walletCli.SendRawTransaction(inscription.revealTx, false)
	if err != nil {
		return err
	}
	log.Log.Info("revealTxSendSuccess", revealTxHash)

	return nil
}
