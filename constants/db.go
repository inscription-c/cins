package constants

import (
	"github.com/btcsuite/btcd/btcutil"
	"path/filepath"
)

const (
	DefaultDBName               = "insc"
	DefaultWithFlushNum         = 1_000
	DefaultFlushCacheNum        = 50_000
	DefaultFlushOutputTraversed = 50_000
	DefaultDBListenPort         = "4000"
	DefaultDbStatusListenPort   = "10080"
	DefaultDBUser               = "root"
	DefaultDBPass               = ""
)

const (
	TidbSessionMemLimit = 3
)

func DBDatDir(testnet bool) string {
	if testnet {
		return btcutil.AppDataDir(filepath.Join(AppName, "inscription", "index", "testnet"), false)
	} else {
		return btcutil.AppDataDir(filepath.Join(AppName, "inscription", "index", "mainnet"), false)
	}
}
