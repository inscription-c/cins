package util

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
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

// TapRootAddress is a function that converts a given public key script into a Taproot address.
// It first parses the public key from the script, then generates a Taproot address from the serialized public key.
// It returns the string representation of the Taproot address and any error that occurred during the process.
func TapRootAddress(pkScript []byte, params *chaincfg.Params) (string, error) {
	// Parse the public key from the script
	pk, err := btcec.ParsePubKey(pkScript)
	if err != nil {
		return "", err
	}
	// Generate a Taproot address from the serialized public key
	taprootAddress, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(pk),
		params,
	)
	if err != nil {
		return "", err
	}
	// Return the string representation of the Taproot address
	return taprootAddress.String(), nil
}
