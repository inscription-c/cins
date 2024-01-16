package index

import (
	"errors"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) NextSequenceNumber() (netSequenceNumber int64, err error) {
	err = idx.Begin(func(tx *Tx) error {
		seqNum, err := tx.GetLatestKey(constants.BucketSequenceNumberToInscriptionEntry)
		if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
			return err
		}
		if err != nil {
			return nil
		}
		netSequenceNumber = gconv.Int64(string(seqNum)) + 1
		return nil
	})
	return
}
