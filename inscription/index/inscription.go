package index

import (
	"errors"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nutsdb/nutsdb"
)

func (idx *Indexer) NextSequenceNumber(tx *Tx) (netSequenceNumber int64, err error) {
	seqNum, err := tx.GetLatestKey(constants.BucketSequenceNumberToInscriptionEntry)
	if err != nil && !errors.Is(err, nutsdb.ErrListNotFound) {
		return 0, err
	}
	if err != nil {
		err = nil
		return
	}
	netSequenceNumber = gconv.Int64(string(seqNum)) + 1
	return
}

func (idx *Indexer) GetInscriptionByOutPoints(outpoints []*OutPoint) (map[string][]byte, error) {
	res := make(map[string][]byte)
	if err := idx.opts.db.View(func(tx *nutsdb.Tx) error {
		for _, v := range outpoints {
			data, err := tx.Get(constants.BucketOutpointToInscriptions, []byte(v.String()))
			if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
				return err
			}
			if err != nil {
				return nil
			}
			res[v.String()] = data
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}
