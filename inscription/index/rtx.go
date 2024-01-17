package index

import (
	"errors"
	"fmt"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) BlockHeight(tx *Tx) (height uint64, err error) {
	val, err := tx.GetLatestKey(constants.BucketHeightToBlockHeader)
	if err != nil {
		return 0, err
	}
	height = gconv.Uint64(string(val))
	return
}

func (idx *Indexer) BlockCount(tx *Tx) (height uint64, err error) {
	h, err := idx.BlockHeight(tx)
	if err != nil && !errors.Is(err, nutsdb.ErrListNotFound) {
		return 0, err
	}
	if errors.Is(err, nutsdb.ErrListNotFound) {
		return 0, nil
	}
	return h + 1, nil
}

func (idx *Indexer) BlockHash(tx *Tx, height ...uint64) (blockHash string, err error) {
	var value []byte
	if len(height) == 0 {
		value, err = tx.GetLatestValue(constants.BucketHeightToBlockHeader)
		if err != nil {
			return
		}
	} else {
		value, err = tx.Get(constants.BucketHeightToBlockHeader, []byte(fmt.Sprint(height[0])))
		if err != nil {
			return
		}
	}

	blockHeader, err := LoadHeader(value)
	if err != nil {
		return "", err
	}
	blockHash = blockHeader.BlockHash().String()
	return
}
