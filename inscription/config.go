package inscription

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/constants"
	"github.com/dotbitHQ/insc/inscription/log"
	"os"
	"path/filepath"
)

func init() {
	Cmd.Flags().StringVarP(&config.Username, "user", "u", "", "wallet rpc server username")
	Cmd.Flags().StringVarP(&config.Password, "password", "P", "", "wallet rpc server password")
	Cmd.Flags().StringVarP(&config.WalletPass, "walletpass", "", "", "wallet password for master private key")
	Cmd.Flags().BoolVarP(&config.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&config.RPCCert, "rpccert", "c", "", "rpc cert file path")
	Cmd.Flags().BoolVarP(&config.TLSSkipVerify, "tlsskipverify", "", false, "skip server tls verify")
	Cmd.Flags().StringVarP(&config.FilePath, "filepath", "f", "", "inscription file path")
	Cmd.Flags().StringVarP(&config.RpcConnect, "rpcconnect", "s", "localhost:8332", "the URL of wallet RPC server to connect to (default localhost:8332, testnet: localhost:18332)")
	Cmd.Flags().Uint64VarP(&config.Postage, "postage", "p", constants.DefaultPostage, "Amount of postage to include in the inscription. Default `10000sat`.")
	Cmd.Flags().StringVarP(&config.IndexDir, "indexdir", "d", config.IndexDir, fmt.Sprintf("inscriptions index database dir. Default `%s`", config.IndexDir))
	Cmd.Flags().BoolVarP(&config.Compress, "compress", "", false, "Compress inscription content with brotli.")
	Cmd.Flags().StringVarP(&config.CborMetadata, "cbormetadata", "", "", "Include CBOR in file at <METADATA> as inscription metadata")
	Cmd.Flags().StringVarP(&config.JsonMetadata, "jsonmetadata", "", "", "Include JSON in file at <METADATA> converted to CBOR as inscription metadata")
	Cmd.Flags().BoolVarP(&config.DryRun, "dryrun", "", false, "Don't sign or broadcast transactions.")
	Cmd.Flags().BoolVarP(&config.IsBrc20C, "brc20c", "", false, "is brc-20-c protocol, add this flag will auto check protocol content effectiveness")
	Cmd.Flags().StringVarP(&config.DstChain, "dstchain", "", "", "target chain coin_type for https://github.com/satoshilabs/slips/blob/master/slip-0044.md")
	Cmd.Flags().StringVarP(&config.Destination, "destination", "", "", "Send inscription to <DESTINATION> address.")
	if err := Cmd.MarkFlagRequired("filepath"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("dstchain"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("destination"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func configCheck() error {
	if config.Testnet {
		config.RpcConnect = "localhost:18332"
	}
	if config.Postage < constants.DustLimit {
		return fmt.Errorf("postage must be greater than or equal %d", constants.DustLimit)
	}
	if config.Postage > constants.MaxPostage {
		return fmt.Errorf("postage must be less than or equal %d", constants.MaxPostage)
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "inscription.log"), false)
	log.InitLogRotator(logFile)
	return nil
}
