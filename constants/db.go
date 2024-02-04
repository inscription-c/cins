package constants

import (
	"github.com/btcsuite/btcd/btcutil"
	"path/filepath"
)

const (
	DefaultDBName               = "cins"
	DefaultWithFlushNum         = 1_000
	DefaultFlushCacheNum        = 50_000
	DefaultFlushOutputTraversed = 50_000
	DefaultDBListenPort         = "4000"
	DefaultDbStatusListenPort   = "10080"
	DefaultDBUser               = "root"
	DefaultDBPass               = ""
	MaxInsertDataSize           = 6 * 1024 * 1024
)

const (
	TidbSessionMemLimit = 1
)

func DBDatDir(testnet bool) string {
	if testnet {
		return btcutil.AppDataDir(filepath.Join(AppName, "inscription", "index", "testnet"), false)
	} else {
		return btcutil.AppDataDir(filepath.Join(AppName, "inscription", "index", "mainnet"), false)
	}
}
