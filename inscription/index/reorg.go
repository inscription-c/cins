package index

import (
	"errors"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/dao"
)

func detectReorg(wtx *dao.DB, block *wire.MsgBlock, height uint64) error {
	bitcoindPrevBlockHash := block.Header.PrevBlock.String()
	if height == 0 {
		return nil
	}
	indexPreBlockHash, err := wtx.BlockHash(height - 1)
	if err != nil {
		return err
	}
	if indexPreBlockHash == bitcoindPrevBlockHash {
		return nil
	}
	return errors.New("unrecoverable reorg detected")
}
