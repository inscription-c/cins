package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
)

const (
	maxSavePoints     uint64 = 2
	savePointInterval uint64 = 10
	chainTipDistance  uint64 = 21
)

func detectReorg(wtx *Tx, idx *Indexer, block *wire.MsgBlock, height uint64) error {
	bitcoindPrevBlockHash := block.Header.PrevBlock.String()
	if height == 0 {
		return nil
	}
	indexPreBlockHash, err := idx.BlockHash(wtx, height-1)
	if err != nil {
		return err
	}
	if indexPreBlockHash == bitcoindPrevBlockHash {
		return nil
	}
	maxRecoverableReorgDepth := (maxSavePoints-1)*savePointInterval + height%savePointInterval
	for depth := uint64(1); depth < maxRecoverableReorgDepth; depth++ {
		indexBlockHash, err := idx.BlockHash(wtx, height-depth)
		if err != nil {
			return err
		}

		h := height - depth
		if h < 0 {
			h = 0
		}
		bitcoinBlockHash, err := idx.opts.cli.GetBlockHash(int64(h))
		if err != nil {
			return err
		}
		if indexBlockHash == bitcoinBlockHash.String() {
			return fmt.Errorf("%d block deep reorg detected at height %d", depth, height)
		}
	}
	return errors.New("unrecoverable reorg detected")
}
