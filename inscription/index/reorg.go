// Package index provides the implementation of blockchain reorganization detection.
package index

import (
	"errors"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/dao"
)

// detectReorg is a function that detects a blockchain reorganization.
// It takes a pointer to a DB, a pointer to a MsgBlock, and a uint32 as parameters.
// The function first gets the previous block hash from the block header and assigns it to bitcoindPrevBlockHash.
// If the height is zero, it returns nil because there is no previous block to compare with.
// The function then gets the block hash of the block at the height minus one from the DB and assigns it to indexPreBlockHash.
// If there is an error getting the block hash, it returns the error.
// The function then compares indexPreBlockHash with bitcoindPrevBlockHash.
// If they are equal, it returns nil because there is no reorganization.
// If they are not equal, it returns an error because a reorganization is detected.
func detectReorg(wtx *dao.DB, block *wire.MsgBlock, height uint32) error {
	bitcoindPrevBlockHash := block.Header.PrevBlock.String() // Get the previous block hash from the block header
	if height == 0 {
		return nil // Return nil if there is no previous block to compare with
	}
	indexPreBlockHash, err := wtx.BlockHash(height - 1) // Get the block hash of the block at the height minus one from the DB
	if err != nil {
		return err // Return the error if there is an error getting the block hash
	}
	if indexPreBlockHash == bitcoindPrevBlockHash {
		return nil // Return nil if there is no reorganization
	}
	return errors.New("unrecoverable reorg detected") // Return an error if a reorganization is detected
}
