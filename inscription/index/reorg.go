package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
)

const (
	maxSavePoints     int64 = 2
	savePointInterval int64 = 10
)

func (idx *Indexer) detectReorg(block *wire.MsgBlock, height int64) error {
	bitcoindPrevBlockHash := block.Header.PrevBlock.String()
	indexPreBlockHash, err := idx.BlockHash(height - 1)
	if err != nil {
		return err
	}
	if indexPreBlockHash == bitcoindPrevBlockHash {
		return nil
	}
	maxRecoverableReorgDepth := (maxSavePoints-1)*savePointInterval + height%savePointInterval
	for depth := int64(1); depth < maxRecoverableReorgDepth; depth++ {
		indexBlockHash, err := idx.BlockHash(height - depth)
		if err != nil {
			return err
		}

		h := height - depth
		if h < 0 {
			h = 0
		}
		bitcoinBlockHash, err := idx.opts.cli.GetBlockHash(h)
		if err != nil {
			return err
		}
		if indexBlockHash == bitcoinBlockHash.String() {
			return fmt.Errorf("%d block deep reorg detected at height %d", depth, height)
		}
	}
	return errors.New("unrecoverable reorg detected")
}
