// Package index provides the implementation of blockchain reorganization detection.
package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
)

const (
	maxSavepoint      uint32 = 2
	savepointInterval uint32 = 10
	chainTipDistance  uint32 = 21
)

var ErrDetectReorg = errors.New("unrecoverable reorg detected")

type ErrRecoverable struct {
	Height uint32
	Depth  uint32
}

func (r *ErrRecoverable) Error() string {
	return fmt.Sprintf("recoverable reorg detected at height %d and depth %d", r.Height, r.Depth)
}

// detectReorg is a function that detects a blockchain reorganization.
// It takes a pointer to a DB, a pointer to a MsgBlock, and a uint32 as parameters.
// The function first gets the previous block hash from the block header and assigns it to bitcoindPrevBlockHash.
// If the height is zero, it returns nil because there is no previous block to compare with.
// The function then gets the block hash of the block at the height minus one from the DB and assigns it to indexPreBlockHash.
// If there is an error getting the block hash, it returns the error.
// The function then compares indexPreBlockHash with bitcoindPrevBlockHash.
// If they are equal, it returns nil because there is no reorganization.
// If they are not equal, it returns an error because a reorganization is detected.
func detectReorg(index *Indexer, wtx *dao.DB, block *wire.MsgBlock, height uint32) error {
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

	maxRecoverableReorgDepth := (maxSavepoint-1)*savepointInterval + height%savepointInterval
	for depth := uint32(1); depth <= maxRecoverableReorgDepth; depth++ {
		if height < depth {
			return errors.New("depth is greater than height")
		}
		indexBlockHash, err := wtx.BlockHash(height - depth)
		if err != nil {
			return err
		}
		bitcoindBlockHash, err := index.opts.cli.GetBlockHash(int64(height - depth))
		if err != nil {
			return err
		}
		if indexBlockHash == bitcoindBlockHash.String() {
			return &ErrRecoverable{Height: height, Depth: depth}
		}
	}
	return ErrDetectReorg
}

func updateSavePoints(index *Indexer, wtx *dao.DB, height uint32) error {
	chainInfo, err := index.opts.cli.GetBlockChainInfo()
	if err != nil {
		return err
	}
	savepoints, err := wtx.ListSavepoint()
	if err != nil {
		return err
	}

	if (height < savepointInterval || height%savepointInterval == 0) &&
		chainInfo.Headers-int32(height) <= int32(chainTipDistance) {
		if len(savepoints) >= int(maxSavepoint) {
			if err := wtx.Delete(&tables.SavePoint{
				Id: savepoints[0].Id,
			}).Error; err != nil {
				return err
			}
			if err := wtx.Where("id<=?", savepoints[0].UndoLogId).
				Delete(&tables.UndoLog{}).Error; err != nil {
				return err
			}
		}
		latest := &tables.UndoLog{}
		if err := wtx.Last(latest).Error; err != nil {
			return err
		}

		log.Srv.Infof("creating savepoint at height %d", height)
		return wtx.Create(&tables.SavePoint{
			Height:    height,
			UndoLogId: latest.Id,
		}).Error
	}
	if len(savepoints) == 0 {
		if err := wtx.DeleteUndoLog(); err != nil {
			return err
		}
	}
	return nil
}

func handleReorg(index *Indexer, height, depth uint32) error {
	log.Srv.Infof("rolling back database after reorg of depth %d at height %d", depth, height)
	if err := index.DB().Transaction(func(tx *dao.DB) error {
		undoLog, err := tx.FindUndoLog()
		if err != nil {
			return err
		}
		for _, v := range undoLog {
			if err := tx.Exec(v.Sql).Error; err != nil {
				return err
			}
		}
		if err := tx.DeleteUndoLog(); err != nil {
			return err
		}
		return tx.DeleteSavepoint()
	}); err != nil {
		return err
	}

	indexHeight, err := index.DB().BlockCount()
	if err != nil {
		return err
	}
	log.Srv.Infof("successfully rolled back database to height %d", indexHeight)
	return nil
}
