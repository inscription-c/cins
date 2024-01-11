package inscription

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/constants"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	Cmd.Flags().StringVarP(&config.Username, "user", "u", "", "wallet rpc server username")
	Cmd.Flags().StringVarP(&config.Password, "password", "P", "", "wallet rpc server password")
	Cmd.Flags().StringVarP(&config.WalletPass, "wallet_pass", "", "", "wallet password for master private key")
	Cmd.Flags().BoolVarP(&config.Testnet, "testnet", "t", false, "bitcoin testnet3")
	Cmd.Flags().StringVarP(&config.RPCCert, "rpc_cert", "c", "", "rpc cert file path")
	Cmd.Flags().BoolVarP(&config.TLSSkipVerify, "tls_skip_verify", "", false, "skip server tls verify")
	Cmd.Flags().StringVarP(&config.FilePath, "filepath", "f", "", "inscription file path")
	Cmd.Flags().StringVarP(&config.RpcConnect, "rpc_connect", "s", "", "the URL of wallet RPC server to connect to (default http://localhost:8332, testnet: http://localhost:18332)")
	Cmd.Flags().Uint64VarP(&config.Postage, "postage", "p", constants.DefaultPostage, "Amount of postage to include in the inscription. Default `10000sat`.")
	Cmd.Flags().StringVarP(&config.IndexDir, "index_dir", "d", config.IndexDir, fmt.Sprintf("inscriptions index database dir. Default `%s`", config.IndexDir))
	Cmd.Flags().BoolVarP(&config.Compress, "compress", "", false, "Compress inscription content with brotli.")
	Cmd.Flags().StringVarP(&config.CborMetadata, "cbor_metadata", "", "", "Include CBOR in file at <METADATA> as inscription metadata")
	Cmd.Flags().StringVarP(&config.JsonMetadata, "json_metadata", "", "", "Include JSON in file at <METADATA> converted to CBOR as inscription metadata")
	Cmd.Flags().BoolVarP(&config.DryRun, "dry_run", "", false, "Don't sign or broadcast transactions.")
	Cmd.Flags().BoolVarP(&config.IsBrc20C, "brc20c", "", false, "is brc-20-c protocol, add this flag will auto check protocol content effectiveness")
	Cmd.Flags().StringVarP(&config.ChainId, "chain_id", "", "", "The latest 8 chars of the first block hash of target chain. xxxxxxxx-xxxxxxxx is acceptable for forked chain.")
	Cmd.Flags().StringVarP(&config.Destination, "destination", "", "", "Send inscription to <DESTINATION> address.")
	if err := Cmd.MarkFlagRequired("filepath"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("chain_id"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := Cmd.MarkFlagRequired("destination"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func configCheck() error {
	if config.RpcConnect == "" {
		config.RpcConnect = "http://localhost:8332"
		if config.Testnet {
			config.RpcConnect = "http://localhost:18332"
		}
	}
	config.ChainId = strings.ToLower(config.ChainId)
	for _, v := range strings.Split(config.ChainId, "-") {
		hexBs, err := hex.DecodeString(v)
		if err != nil {
			return fmt.Errorf("chain_id %s invalid", config.ChainId)
		}
		if len(hexBs) != 8 {
			return fmt.Errorf("chain_id %s invalid", config.ChainId)
		}
	}
	if config.Postage <= constants.DustLimit {
		return fmt.Errorf("postage must be greater than %d", constants.DustLimit)
	}
	if config.Postage > constants.MaxPostage {
		return fmt.Errorf("postage must be less than or equal %d", constants.MaxPostage)
	}

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	logFile := btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "logs", "inscription.log"), false)
	initLogRotator(logFile)
	return nil
}
