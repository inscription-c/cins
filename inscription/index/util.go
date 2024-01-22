package index

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"strings"
)

// IsEmptyHash checks if the provided hash is empty.
//
// It creates a new hash with the same size as the provided hash but filled with zeros.
// Then it compares the string representation of the new hash with the string representation of the provided hash.
// If they are equal, it means the provided hash was empty (filled with zeros), so it returns true.
// If they are not equal, it means the provided hash had some data, so it returns false.
//
// Parameters:
//
//	hash (chainhash.Hash): The hash to check.
//
// Returns:
//
//	bool: True if the hash is empty, false otherwise.
func IsEmptyHash(hash chainhash.Hash) bool {
	h, _ := chainhash.NewHash(make([]byte, chainhash.HashSize))
	return strings.EqualFold(h.String(), hash.String())
}
