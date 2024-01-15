package config

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/dotbitHQ/insc/constants"
	"path/filepath"
)

var (
	// Username rpc server user name
	Username string
	// Password rpc server password
	Password string
	// WalletPass wallet password
	WalletPass string
	// Testnet is bitcoin testnet3
	Testnet bool
	// NoBtcd wallet no embed btcd
	NoBtcd bool
	// RpcConnect wallet rpc server url
	RpcConnect string
	// RPCCert wallet rpc server cert path
	RPCCert string
	// TLSSkipVerify skip verify server tls
	TLSSkipVerify bool
	// FilePath inscription filepath
	FilePath string
	// Postage inscribe postage default is 10000sat
	Postage uint64
	// IndexDir inscriptions index database dir
	IndexDir string
	// Compress compress inscription
	Compress bool
	// CborMetadata cbor metadata file path
	CborMetadata string
	// JsonMetadata json metadata file path
	JsonMetadata string
	// DryRun Don't sign or broadcast transactions.
	DryRun bool
	// IsBrc20C is brc20c protocol
	IsBrc20C bool
	// DstChain target chain coin_type
	DstChain string
	// Destination destination address
	Destination string
	// Rpclisten rpc server listen address
	Rpclisten string
)

func init() {
	Postage = constants.DefaultPostage
	IndexDir = btcutil.AppDataDir(filepath.Join(constants.AppName, "inscription", "index"), false)
}
