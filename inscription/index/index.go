package index

import (
	"errors"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/dotbitHQ/insc/constants"
	"github.com/nutsdb/nutsdb"
)

type Options struct {
	db       *nutsdb.DB
	cli      *rpcclient.Client
	batchCli *rpcclient.Client
}

type Option func(*Options)

func WithDB(db *nutsdb.DB) func(*Options) {
	return func(options *Options) {
		options.db = db
	}
}

func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

func WithBatchClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.batchCli = cli
	}
}

type Indexer struct {
	opts   *Options
	height int64
}

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

func NewIndexer(opts ...Option) *Indexer {
	idx := &Indexer{
		opts: &Options{},
	}
	for _, v := range opts {
		v(idx.opts)
	}
	return idx
}

func DB(dir string) (*nutsdb.DB, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(dir),
	)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *nutsdb.Tx) error {
		for _, v := range constants.KVBuckets {
			if err := tx.NewKVBucket(v); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
				return err
			}
		}
		for _, v := range constants.ListBuckets {
			if err := tx.NewListBucket(v); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
				return err
			}
		}
		if err := tx.NewListBucket(constants.BucketKeyList); err != nil && !errors.Is(err, nutsdb.ErrBucketAlreadyExist) {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return db, nil
}

func (idx *Indexer) Begin(fn func(tx *Tx) error, writable ...bool) error {
	w := false
	if len(writable) > 0 {
		w = writable[0]
	}
	innerTx, err := idx.opts.db.Begin(w)
	if err != nil {
		return err
	}
	tx := &Tx{innerTx}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
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
