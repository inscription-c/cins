package index

import (
	"errors"
	"fmt"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) BlockHeight() (height int64, err error) {
	err = idx.Begin(func(tx *Tx) error {
		val, err := tx.GetLatestKey(constants.BucketHeightToBlockHeader)
		if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
			return err
		}
		if err != nil {
			return nil
		}
		height = gconv.Int64(string(val))
		return nil
	})
	return
}

func (idx *Indexer) BlockCount() (height int64, err error) {
	h, err := idx.BlockHeight()
	if err != nil {
		return 0, err
	}
	return h + 1, nil
}

func (idx *Indexer) BlockHash(height ...int64) (blockHash string, err error) {
	err = idx.Begin(func(tx *Tx) error {
		var value []byte
		if len(height) == 0 {
			value, err = tx.GetLatestValue(constants.BucketHeightToBlockHeader)
			if err != nil && !errors.Is(err, nutsdb.ErrNotFoundKey) {
				return err
			}
			if err != nil {
				return nil
			}
		} else {
			value, err = tx.Get(constants.BucketHeightToBlockHeader, gconv.Bytes(fmt.Sprint(height[0])))
			if err != nil && !errors.Is(err, nutsdb.ErrNotFoundKey) {
				return err
			}
			if err != nil {
				return nil
			}
		}

		blockHeader, err := LoadHeader(value)
		if err != nil {
			return err
		}
		blockHash = blockHeader.BlockHash().String()
		return nil
	})
	return
}
