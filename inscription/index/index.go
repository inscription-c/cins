package index

import (
	"bytes"
	"errors"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"golang.org/x/sync/errgroup"
	"time"
)

type Options struct {
	db  *dao.DB
	cli *rpcclient.Client

	indexSats           bool
	indexTransaction    bool
	noIndexInscriptions bool
	flushNum            uint64
}

type Option func(*Options)

func WithDB(db *dao.DB) func(*Options) {
	return func(options *Options) {
		options.db = db
	}
}

func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

func WithFlushNum(flushNum uint64) func(*Options) {
	return func(options *Options) {
		options.flushNum = flushNum
	}
}

type Indexer struct {
	opts                      *Options
	rangeCache                map[string]model.SatRange
	height                    uint32
	satRangesSinceFlush       uint32
	outputsCached             uint64
	outputsInsertedSinceFlush uint64
	outputsTraversed          uint32
}

func NewIndexer(opts ...Option) *Indexer {
	idx := &Indexer{
		opts: &Options{},
	}
	for _, v := range opts {
		v(idx.opts)
	}
	idx.rangeCache = make(map[string]model.SatRange)
	return idx
}

func (idx *Indexer) UpdateIndex() error {
	var err error
	wtx := idx.opts.db.Begin()
	defer func() {
		if err != nil {
			wtx.Rollback()
		}
	}()

	idx.height, err = wtx.BlockCount()
	if err != nil {
		return err
	}

	// latest block height
	startingHeight, err := idx.opts.cli.GetBlockCount()
	if err != nil {
		return err
	}
	valueCache := make(map[string]int64)

	blocks, err := idx.fetchBlockFrom(idx.height, uint32(startingHeight))
	if err != nil {
		return err
	}
	for _, block := range blocks {
		if err = idx.indexBlock(wtx, block, valueCache); err != nil {
			return err
		}

		if err := idx.commit(wtx, valueCache); err != nil {
			return err
		}
		valueCache = make(map[string]int64)

		wtx = wtx.Begin()
		height, err := wtx.BlockCount()
		if err != nil {
			return err
		}

		if height != idx.height {
			log.Srv.Warn("height != idx.height", "height", height, "idx.height", idx.height)
			break
		}
	}
	return nil
}

//
//func (idx *Indexer) spawnFetcher() (outpointCh chan *wire.OutPoint, valueCh chan uint64) {
//	bufferSize := 20_000
//	batchSize := 2048
//	parallelRequests := 12
//	outpointCh = make(chan *wire.OutPoint, bufferSize)
//	valueCh = make(chan uint64, bufferSize)
//
//	go func() {
//		for {
//			outpoint, ok := <-outpointCh
//			if !ok {
//				log.Srv.Debug("outpointCh closed")
//				break
//			}
//
//			outpoints := make([]*wire.OutPoint, 0, batchSize)
//			outpoints = append(outpoints, outpoint)
//			for i := 0; i < batchSize-1; i++ {
//				select {
//				case outpoint, ok := <-outpointCh:
//					if !ok {
//						break
//					}
//					outpoints = append(outpoints, outpoint)
//				default:
//					break
//				}
//			}
//
//			getTxByTxids := func(txids []string) ([]*btcutil.Tx, error) {
//				txs, err := idx.getTransactions(txids)
//				if err != nil {
//					return nil, err
//				}
//				return txs, nil
//			}
//
//			chunkSize := (len(outpoints) / parallelRequests) + 1
//			futs := make([]*btcutil.Tx, 0, parallelRequests)
//			txids := make([]string, 0, chunkSize)
//			for i := 0; i < len(outpoints); i++ {
//				txids = append(txids, outpoints[i].Hash.String())
//				if i != 0 && i%chunkSize == 0 {
//					txs, err := getTxByTxids(txids)
//					if err != nil {
//						log.Srv.Error("getTxByTxids", err)
//						return
//					}
//					futs = append(futs, txs...)
//					txids = make([]string, 0, chunkSize)
//				}
//			}
//			if len(txids) > 0 {
//				txs, err := getTxByTxids(txids)
//				if err != nil {
//					log.Srv.Error("getTxByTxids", err)
//					return
//				}
//				futs = append(futs, txs...)
//			}
//
//			for i, tx := range futs {
//				valueCh <- uint64(tx.MsgTx().TxOut[outpoints[i].Index].Value)
//			}
//		}
//	}()
//	return
//}

func (idx *Indexer) fetchBlockFrom(start, end uint32) ([]*wire.MsgBlock, error) {
	if start > end {
		return nil, nil
	}

	maxFetch := uint32(32)
	if end-start+1 < maxFetch {
		maxFetch = end - start + 1
	}

	errWg := errgroup.Group{}
	blocks := make([]*wire.MsgBlock, maxFetch)
	for i := start; i < start+maxFetch; i++ {
		height := i
		errWg.Go(func() error {
			block, err := idx.getBlockWithRetries(height)
			if err != nil {
				return err
			}
			blocks[height-start] = block
			return nil
		})
	}
	if err := errWg.Wait(); err != nil {
		return nil, err
	}
	return blocks, nil
}

func (idx *Indexer) getBlockWithRetries(height uint32) (*wire.MsgBlock, error) {
	errs := -1
	for {
		errs++
		if errs > 0 {
			seconds := 1 << errs
			if seconds > 120 {
				err := errors.New("would sleep for more than 120s, giving up")
				log.Srv.Error(err)
			}
			time.Sleep(time.Second * time.Duration(seconds))
		}
		hash, err := idx.opts.cli.GetBlockHash(int64(height))
		if err != nil {
			log.Srv.Warn("GetBlockHash", err)
			continue
		}
		block, err := idx.opts.cli.GetBlock(hash)
		if err != nil {
			log.Srv.Warn("GetBlock", err)
			continue
		}
		return block, nil
	}
}

func (idx *Indexer) indexBlock(
	wtx *dao.DB,
	block *wire.MsgBlock,
	valueCache map[string]int64) error {

	if err := detectReorg(wtx, block, idx.height); err != nil {
		return err
	}

	start := time.Now()
	satRangesWritten := uint64(0)
	outputsInBlock := uint32(0)
	indexInscriptions :=
		/*idx.height >= index.first_inscription_height && */ !idx.opts.noIndexInscriptions

	unboundInscriptions, err := wtx.GetStatisticCountByName(tables.StatisticUnboundInscriptions)
	if err != nil {
		return err
	}
	nextSequenceNumber, err := wtx.NextSequenceNumber()
	if err != nil {
		return err
	}
	inscriptionUpdater := &InscriptionUpdater{
		wtx:                 wtx,
		height:              idx.height,
		valueCache:          valueCache,
		nextSequenceNumber:  &nextSequenceNumber,
		unboundInscriptions: &unboundInscriptions,
		timestamp:           block.Header.Timestamp.UnixMilli(),
	}

	if idx.opts.indexSats {
		//coinbaseInputs := make([]byte, 0)
		//h := Height{Height: idx.height}
		//if h.Subsidy() > 0 {
		//
		//}
	} else if indexInscriptions {
		txs := append([]*wire.MsgTx{block.Transactions[len(block.Transactions)-1]}, block.Transactions[1:]...)
		for i := range txs {
			tx := txs[i]
			if err := inscriptionUpdater.indexEnvelopers(tx, nil); err != nil {
				return err
			}
		}
	}

	if indexInscriptions {
		buf := bytes.NewBufferString("")
		if err := block.Header.Serialize(buf); err != nil {
			return err
		}
		if err := wtx.SaveBlockInfo(&tables.BlockInfo{
			Height:      idx.height,
			SequenceNum: *inscriptionUpdater.nextSequenceNumber,
			Header:      buf.Bytes(),
		}); err != nil {
			return err
		}
	}

	if err := wtx.IncrementStatistic(tables.StatisticCursedInscriptions, *inscriptionUpdater.cursedInscriptionCount); err != nil {
		return err
	}
	if err := wtx.IncrementStatistic(tables.StatisticBlessedInscriptions, *inscriptionUpdater.blessedInscriptionCount); err != nil {
		return err
	}
	if err := wtx.IncrementStatistic(tables.StatisticUnboundInscriptions, *inscriptionUpdater.unboundInscriptions); err != nil {
		return err
	}

	idx.height++
	idx.outputsTraversed += outputsInBlock

	log.Srv.Infof("Block Height %d Wrote %d sat ranges from %d outputs in %d ms", idx.height-1, satRangesWritten, outputsInBlock, time.Since(start).Milliseconds())

	return nil
}

func (idx *Indexer) commit(wtx *dao.DB, valueCache map[string]int64) error {
	log.Srv.Infof(
		"Committing at block %d, %d outputs traversed, %d in map, %d cached",
		idx.height-1, idx.outputsTraversed, len(valueCache), idx.outputsCached,
	)

	if idx.opts.indexSats {
		log.Srv.Infof(
			"Flushing %d entries (%.1f%% resulting from %d insertions) from memory to database",
			len(idx.rangeCache),
			float64(len(idx.rangeCache))/float64(idx.outputsInsertedSinceFlush)*100,
			idx.outputsInsertedSinceFlush,
		)

		for outpoint, satRange := range idx.rangeCache {
			if err := wtx.SetOutpointToSatRange(outpoint, satRange); err != nil {
				return err
			}
		}
		idx.outputsInsertedSinceFlush = 0
	}

	for outpoint, value := range valueCache {
		if err := wtx.SetOutpointToValue(outpoint, value); err != nil {
			return err
		}
	}

	if err := wtx.IncrementStatistic(tables.StatisticOutputsTraversed, idx.outputsTraversed); err != nil {
		return err
	}
	idx.outputsTraversed = 0
	if err := wtx.IncrementStatistic(tables.StatisticSatRanges, idx.satRangesSinceFlush); err != nil {
		return err
	}
	idx.satRangesSinceFlush = 0
	if err := wtx.IncrementStatistic(tables.StatisticCommits, 1); err != nil {
		return err
	}
	wtx.Commit()
	return nil
}
