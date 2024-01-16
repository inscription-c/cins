package index

import (
	"errors"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) GetStatisticCount(key constants.Statistic) (count int64, err error) {
	err = idx.Begin(func(tx *Tx) error {
		v, err := tx.Get(constants.BucketStatisticToCount, []byte(key))
		if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
			return err
		}
		if err != nil {
			return nil
		}
		count = gconv.Int64(string(v))
		return nil
	})
	return
}
