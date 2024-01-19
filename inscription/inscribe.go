package inscription

import (
	"errors"
	"fmt"
	"github.com/inscription-c/insc/config"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/inscription/server"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/wallet"
	"github.com/spf13/cobra"
	"os"
)

// InsufficientBalanceError is an error that represents an insufficient balance.
var InsufficientBalanceError = errors.New("InsufficientBalanceError")

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

func init() {
	Cmd.AddCommand(server.Cmd)
}

// inscribe is a function that performs the inscription process.
// It checks the configuration, gets the UTXO, creates the commit and reveal transactions, and signs and sends the transactions.
// It also handles any errors that occur during these processes.
func inscribe() error {
	// Check the configuration
	if err := configCheck(); err != nil {
		return err
	}

	// Get the database
	db, err := dao.DB(config.IndexDir)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		if err := db.Close(); err != nil {
			log.Log.Error("db.Close", "err", err)
		}
	})

	// Create a new wallet client
	walletCli, err := wallet.NewWalletClient(
		config.RpcConnect,
		config.Username,
		config.Password,
		true,
	)
	if err != nil {
		return err
	}
	batchCli, err := wallet.NewBatchClient(
		config.RpcConnect,
		config.Username,
		config.Password,
		true,
	)
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		walletCli.Shutdown()
		batchCli.Shutdown()
	})

	// indexer is an instance of the Indexer struct from the index package.
	// The NewIndexer function is used to create this instance.
	// The WithDB method is used to set the database for the indexer, taking the db variable as an argument.
	// The WithClient method is used to set the wallet client for the indexer, taking the walletCli variable as an argument.
	indexer := index.NewIndexer(
		index.WithDB(db),
		index.WithClient(walletCli),
		index.WithBatchClient(batchCli),
		index.WithFlushNum(constants.DefaultWithFlushNum),
	)

	// Create a new inscription from the file path
	inscription, err := NewFromPath(config.FilePath,
		WithWalletClient(walletCli),
		WithIndexer(indexer),
		WithPostage(config.Postage),
		WithDstChain(config.DstChain),
		WithWalletPass(config.WalletPass),
		WithCborMetadata(config.CborMetadata),
		WithJsonMetadata(config.JsonMetadata),
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
	if config.DryRun {
		log.Log.Info("dry run success")
		log.Log.Info("commitTx: ", inscription.CommitTxId())
		log.Log.Info("revealTx: ", inscription.RevealTxId())
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
