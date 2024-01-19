package config

var (
	// Username rpc server user name
	Username string
	// Password rpc server password
	Password string
	// WalletPass wallet password
	WalletPass string
	// Testnet is bitcoin testnet3
	Testnet bool
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
	// MysqlAddr inscription index database addr
	MysqlAddr string
	// MysqlUser inscription index database user
	MysqlUser string
	// MysqlPassword inscription index database password
	MysqlPassword string
	// MysqlDBName inscription index database name
	MysqlDBName string
	// NoEmbedDB no embed db
	NoEmbedDB bool
	// DataDir data dir
	DataDir string
	// DBListenPort db listen port
	DBListenPort string
	// DBStatusListenPort db status listen port
	DBStatusListenPort string
)
