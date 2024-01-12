package index

import (
	"errors"
	"fmt"
	"github.com/dotbitHQ/insc/config"
	"github.com/dotbitHQ/insc/constants"
	"github.com/dotbitHQ/insc/model"
	"github.com/nutsdb/nutsdb"
	"os"
	"sync"
)

var lock sync.Once
var db *nutsdb.DB

func DB() *nutsdb.DB {
	lock.Do(func() {
		var err error
		db, err = nutsdb.Open(
			nutsdb.DefaultOptions,
			nutsdb.WithDir(config.IndexDir),
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := db.Update(func(tx *nutsdb.Tx) error {
			if err := tx.NewKVBucket(constants.BucketOutpointToInscriptions); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
				return err
			}
			return nil
		}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})
	return db
}

func GetInscriptionByOutPoints(outpoints []*model.OutPoint) (map[string][]byte, error) {
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
