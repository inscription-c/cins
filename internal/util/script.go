package util

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// AddressScript is a function that converts a given address to a script.
// It first decodes the address, then generates a pay-to-address script from the decoded address.
// It returns the generated script and any error that occurred during the process.
func AddressScript(address string, params *chaincfg.Params) ([]byte, error) {
	decodeAddress, err := btcutil.DecodeAddress(address, params)
	if err != nil {
		return nil, err
	}
	return txscript.PayToAddrScript(decodeAddress)
}
