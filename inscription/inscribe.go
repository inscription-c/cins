package inscription

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/index"
	"github.com/dotbitHQ/insc/internal/signal"
	"github.com/dotbitHQ/insc/wallet"
	"github.com/spf13/cobra"
	"os"
)

var (
	InsufficientBalanceError = errors.New("InsufficientBalanceError")

	Cmd = &cobra.Command{
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
)

func inscribe() error {
	if err := configCheck(); err != nil {
		return err
	}

	db := index.DB()
	signal.AddInterruptHandler(func() {
		if err := db.Close(); err != nil {
			log.Error("db.Close", "err", err)
		}
	})

	walletCli, err := wallet.NewWalletClient(config.RpcConnect, config.Username, config.Password, &rpcclient.NotificationHandlers{
		OnClientConnected: OnClientConnected,
	})
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		walletCli.Shutdown()
	})

	inscription, err := NewFromPath(config.FilePath,
		WithWalletClient(walletCli),
		WithPostage(config.Postage),
		WithDstChain(config.DstChain),
		WithWalletPass(config.WalletPass),
		WithCborMetadata(config.CborMetadata),
		WithJsonMetadata(config.JsonMetadata),
	)
	if err != nil {
		return err
	}

	// get all utxo for all unspent address and exclude the utxo where the inscription
	if err := inscription.getUtxo(); err != nil {
		return err
	}

	// create commit and reveal transaction
	if err := inscription.createInscriptionTx(); err != nil {
		return err
	}

	if config.DryRun {
		log.Info("dry run success")
		log.Info("commitTx: ", inscription.CommitTxId())
		log.Info("revealTx: ", inscription.RevealTxId())
		return nil
	}

	// sign reveal transaction
	if err := inscription.signRevealTx(); err != nil {
		return err
	}
	// sign commit transaction
	if err := inscription.signCommitTx(); err != nil {
		return err
	}

	commitTxHash, err := walletCli.SendRawTransaction(inscription.commitTx, false)
	if err != nil {
		return err
	}
	log.Info("commitTxSendSuccess", commitTxHash)

	revealTxHash, err := walletCli.SendRawTransaction(inscription.revealTx, false)
	if err != nil {
		return err
	}
	log.Info("revealTxSendSuccess", revealTxHash)

	return nil
}
