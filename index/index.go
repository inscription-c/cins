package index

import (
	"errors"
	"github.com/dotbitHQ/insc/client"
	"github.com/dotbitHQ/insc/constants"
	"github.com/nutsdb/nutsdb"
)

var db *nutsdb.DB

func Open(dir string) (*nutsdb.DB, error) {
	var err error
	db, err = nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(dir),
	)
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.NewKVBucket(constants.BucketOutpointToInscriptions); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return db, nil
}

func GetInscriptionByOutPoints(outpoints []client.OutPoint) (map[string][]byte, error) {
	res := make(map[string][]byte)
	if err := db.View(func(tx *nutsdb.Tx) error {
		for _, v := range outpoints {
			data, err := tx.Get(constants.BucketOutpointToInscriptions, []byte(v.String()))
			if err != nil &&
				!errors.Is(err, nutsdb.ErrKeyNotFound) &&
				!errors.Is(err, nutsdb.ErrNotFoundBucket) {
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
