package inscription

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/netparams"
	"github.com/inscription-c/insc/btcd"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/inscription/server"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/internal/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	username             string
	password             string
	walletPass           string
	testnet              bool
	inscriptionsFilePath string
	rpcConnect           string
	postage              = uint64(constants.DefaultPostage)
	compress             bool
	cborMetadata         string
	jsonMetadata         string
	dryRun               bool
	cbrc20               bool
	destination          string
	unlockConditionFile  string
	dbAddr               string
	dbUser               string
	dbPass               string
	dbName               string
	noBackup             bool
)

// InsufficientBalanceError is an error that represents an insufficient balance.
var InsufficientBalanceError = errors.New("InsufficientBalanceError")

func init() {
	Cmd.Flags().StringVarP(&username, "user", "u", "root", "wallet rpc server username")
	Cmd.Flags().StringVarP(&password, "password", "P", "root", "wallet rpc server password")
	Cmd.Flags().StringVarP(&walletPass, "wallet_pass", "", "root", "wallet password for master private key")
	Cmd.Flags().BoolVarP(&testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&inscriptionsFilePath, "filepath", "f", "", "inscription file path")
	Cmd.Flags().StringVarP(&unlockConditionFile, "unlock_condition", "d", "", "unlock condition file path.")
	Cmd.Flags().StringVarP(&destination, "dest", "", "", "Send inscription to <DESTINATION> address.")
	Cmd.Flags().StringVarP(&rpcConnect, "rpc_connect", "s", "localhost:8332", "the URL of wallet RPC server to connect to (default localhost:8332, testnet: localhost:18332)")
	Cmd.Flags().Uint64VarP(&postage, "postage", "p", constants.DefaultPostage, "Amount of postage to include in the inscription. Default `10000sat`.")
	Cmd.Flags().BoolVarP(&compress, "compress", "", false, "Compress inscription content with brotli.")
	Cmd.Flags().StringVarP(&cborMetadata, "cbor_metadata", "", "", "Include CBOR in file at <METADATA> as inscription metadata")
	Cmd.Flags().StringVarP(&jsonMetadata, "json_metadata", "", "", "Include JSON in file at <METADATA> converted to CBOR as inscription metadata")
	Cmd.Flags().BoolVarP(&dryRun, "dry_run", "", false, "Don't sign or broadcast transactions.")
	Cmd.Flags().BoolVarP(&cbrc20, "c_brc_20", "", false, "is c-brc-20 protocol, add this flag will auto check protocol content effectiveness")
	Cmd.Flags().StringVarP(&dbAddr, "mysql_addr", "", "127.0.0.1:3306", "index server database address")
	Cmd.Flags().StringVarP(&dbUser, "mysql_user", "", "root", "index server database user")
	Cmd.Flags().StringVarP(&dbPass, "mysql_pass", "", "root", "index server database password")
	Cmd.Flags().StringVarP(&dbName, "dbname", "", constants.DefaultDBName, "inscription index mysql database name")
	Cmd.Flags().BoolVarP(&noBackup, "no_backup", "", false, "Do not back up recovery key.")
	if err := Cmd.MarkFlagRequired("filepath"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("dest"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("unlock_condition"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func configCheck() error {
	if testnet {
		rpcConnect = "localhost:18332"
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
	if _, err := tables.UnlockConditionFromFile(unlockConditionFile); err != nil {
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

	// Create a new wallet client
	walletCli, err := btcd.NewClient(
		rpcConnect,
		username,
		password,
		true,
	)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	signal.AddInterruptHandler(func() {
		walletCli.Shutdown()
	})

	// Get the database
	db, err := dao.NewDB(
		dao.WithAddr(dbAddr),
		dao.WithUser(dbUser),
		dao.WithPassword(dbPass),
		dao.WithDBName(dbName),
	)
	if err != nil {
		return err
	}

	// Get the unlock condition from the file path
	unlockCondition, err := tables.UnlockConditionFromFile(unlockConditionFile)
	if err != nil {
		return err
	}

	// Create a new inscription from the file path
	inscription, err := NewFromPath(inscriptionsFilePath,
		WithDB(db),
		WithWalletClient(walletCli),
		WithPostage(postage),
		WithUnlockCondition(unlockCondition),
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
