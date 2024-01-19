package index

import (
	"errors"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) GetStatisticCount(tx *Tx, key constants.Statistic) (count uint64, err error) {
	v, err := tx.Get(constants.BucketStatisticToCount, []byte(key))
	if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
		return 0, err
	}
	if err != nil {
		err = nil
		return
	}
	count = gconv.Uint64(string(v))
	return
}
