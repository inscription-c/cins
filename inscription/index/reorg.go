// Package index provides the implementation of blockchain reorganization detection.
package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/cins/inscription/index/dao"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/inscription/log"
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
	for depth := uint32(1); depth < maxRecoverableReorgDepth; depth++ {
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

// updateSavePoints is a function that updates the savepoints in the blockchain.
// It takes a pointer to an Indexer, a pointer to a DB, and an uint32 as parameters.
// The function first gets the blockchain info from the Indexer and assigns it to chainInfo.
// If there is an error getting the blockchain info, it returns the error.
// The function then gets the list of savepoints from the DB and assigns it to savepoints.
// If there is an error getting the savepoints, it returns the error.
// The function then checks if the height is less than the savepoint interval or if the height is a multiple of the savepoint interval.
// It also checks if the difference between the number of headers in the blockchain info and the height is less than or equal to the chain tip distance.
// If both conditions are true, it checks if the length of the savepoints is greater than or equal to the maximum savepoint.
// If it is, it deletes the first savepoint and all undo logs with an id less than or equal to the id of the undo log of the first savepoint.
// It then gets the last undo log and assigns it to latest.
// If there is an error getting the last undo log, it returns the error.
// The function then creates a new savepoint with the height and the id of the latest undo log.
// If there is an error creating the new savepoint, it returns the error.
// If the length of the savepoints is zero, it deletes all undo logs.
// If there is an error deleting the undo logs, it returns the error.
// The function then returns nil.
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
			// delete oldest savepoint
			if err := wtx.Delete(&tables.SavePoint{
				Id: savepoints[0].Id,
			}).Error; err != nil {
				return err
			}
			if err := wtx.Where("id<=?", savepoints[1].UndoLogId).
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

// handleReorg is a function that handles a blockchain reorganization.
// It takes a pointer to an Indexer, a uint32 for the height, and a uint32 for the depth as parameters.
// The function first logs the depth and height of the reorganization.
// It then starts a transaction on the DB of the Indexer.
// Inside the transaction, it gets the undo log and assigns it to undoLog.
// If there is an error getting the undo log, it returns the error.
// The function then executes each SQL statement in the undo log.
// If there is an error executing a SQL statement, it returns the error.
// The function then deletes the undo log.
// If there is an error deleting the undo log, it returns the error.
// The function then deletes the savepoint.
// If there is an error deleting the savepoint, it returns the error.
// If there is an error with the transaction, it returns the error.
// The function then gets the block count from the DB of the Indexer and assigns it to indexHeight.
// If there is an error getting the block count, it returns the error.
// The function then logs the height of the blockchain after the rollback.
// The function then returns nil.
func handleReorg(index *Indexer, height, depth uint32) error {
	log.Srv.Infof("rolling back database after reorg of depth %d at height %d", depth, height)
	if err := index.DB().Transaction(func(tx *dao.DB) error {
		oldestSavepoint, err := tx.OldestSavepoint()
		if err != nil {
			return err
		}
		if oldestSavepoint.Id == 0 {
			return errors.New("no savepoint found")
		}

		rows, err := tx.FindUndoLog()
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var undoLog tables.UndoLog
			if err := tx.ScanRows(rows, &undoLog); err != nil {
				return err
			}
			if err := tx.Exec(undoLog.Sql).Error; err != nil {
				return err
			}
		}
		if err := tx.DeleteUndoLog(); err != nil {
			return err
		}
		return tx.DeleteSavepoint(oldestSavepoint.Id)
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
