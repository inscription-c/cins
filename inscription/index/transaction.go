package index

import (
	"github.com/dotbitHQ/insc/constants"
	"github.com/nutsdb/nutsdb"
)

type Tx struct {
	*nutsdb.Tx
}

func (tx *Tx) Put(bucket string, key, val []byte) error {
	if err := tx.Tx.Put(bucket, key, val, 0); err != nil {
		return err
	}
	return tx.RPush(constants.BucketKeyList, []byte(bucket), key)
}

func (tx *Tx) GetLatestValue(bucket string) ([]byte, error) {
	key, err := tx.RPeek(constants.BucketKeyList, []byte(bucket))
	if err != nil {
		return nil, err
	}
	val, err := tx.Tx.Get(bucket, key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (tx *Tx) GetLatestKey(bucket string) ([]byte, error) {
	key, err := tx.RPeek(constants.BucketKeyList, []byte(bucket))
	if err != nil {
		return nil, err
	}
	return key, nil
}
