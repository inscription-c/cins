package index

import (
	"errors"
	"github.com/dotbitHQ/insc/constants"
	"github.com/nutsdb/nutsdb"
)

func DB(dir string) (*nutsdb.DB, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(dir),
	)
	if err != nil {
		return nil, err
	}

	initKVBuckets := map[string]struct{}{}
	initSetBuckets := map[string]struct{}{}
	initBucketKeyList := map[string]struct{}{}

	if err := db.Update(func(tx *nutsdb.Tx) error {
		for _, v := range constants.KVBuckets {
			if err = tx.NewKVBucket(v); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
				return err
			}
			if err == nil {
				initKVBuckets[v] = struct{}{}
			}
		}
		for _, v := range constants.SetBuckets {
			if err = tx.NewSetBucket(v); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
				return err
			}
			if err == nil {
				initSetBuckets[v] = struct{}{}
			}
		}
		if err = tx.NewListBucket(constants.BucketKeyList); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
			return err
		}
		if err == nil {
			initBucketKeyList[constants.BucketKeyList] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *nutsdb.Tx) error {
		for k := range initKVBuckets {
			if k == constants.BucketStatisticToCount {
				for _, v := range constants.Statistics {
					if err := tx.Put(k, []byte(v), []byte("0"), 0); err != nil {
						return err
					}
				}
				continue
			}
			if err := tx.Put(k, []byte(k), []byte(k), 0); err != nil {
				return err
			}
		}
		for k := range initSetBuckets {
			if err := tx.SAdd(k, []byte(k), []byte(k)); err != nil {
				return err
			}
		}
		for k := range initBucketKeyList {
			if err := tx.RPush(k, []byte(k), []byte(k)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return db, nil
}
