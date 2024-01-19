package index

import (
	"errors"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
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
