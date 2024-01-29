package index

import (
	"bytes"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/go-sql-driver/mysql"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/signal"
	"sync/atomic"
	"time"
)

// Options is a struct that holds configuration options for the Indexer.
type Options struct {
	// db is a pointer to a dao.DB instance which represents the database connection.
	db *dao.DB
	// cli is a pointer to a rpcclient.Client instance which represents the RPC client.
	cli *rpcclient.Client
	// batchCli is a pointer to a rpcclient.Client instance which represents the batch RPC client.
	batchCli *rpcclient.Client

	// indexSats is a boolean that indicates whether to index satoshis or not.
	indexSats bool
	// indexTransaction is a boolean that indicates whether to index transactions or not.
	indexTransaction bool
	// noIndexInscriptions is a boolean that indicates whether to index inscriptions or not.
	noIndexInscriptions bool
	// startHeight is an uint32 that represents the starting height of the blockchain to index from.
	firstInscriptionHeight uint32
}

// Option is a function type that takes a pointer to an Options struct.
type Option func(*Options)

// WithDB is a function that returns an Option.
// This Option sets the db field of the Options struct to the provided dao.DB instance.
func WithDB(db *dao.DB) func(*Options) {
	return func(options *Options) {
		options.db = db
	}
}

// WithClient is a function that returns an Option.
// This Option sets the cli field of the Options struct to the provided rpcclient.Client instance.
func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

// WithBatchClient is a function that returns an Option.
// This Option sets the batchCli field of the Options struct to the provided rpcclient.Client instance.
func WithBatchClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.batchCli = cli
	}
}

// WithIndexSats is a function that returns an Option.
func WithIndexSats(indexSats bool) func(*Options) {
	return func(options *Options) {
		options.indexSats = indexSats
	}
}

// Indexer is a struct that holds the configuration options and state for the Indexer.
type Indexer struct {
	// opts is a pointer to an Options struct which holds the configuration options for the Indexer.
	opts *Options
	// rangeCache is a map that caches the range of satoshis for each transaction.
	rangeCache map[string][]*model.SatRange
	// height is an uint32 that represents the current height of the blockchain being indexed.
	height uint32
	// satRangesSinceFlush is an uint32 that represents the number of satoshi ranges since the last flush.
	satRangesSinceFlush uint64
	// outputsCached is an uint64 that represents the number of outputs cached.
	outputsCached uint64
	// outputsInsertedSinceFlush is an uint64 that represents the number of outputs inserted since the last flush.
	outputsInsertedSinceFlush uint64
	// outputsTraversed is an uint32 that represents the number of outputs traversed.
	outputsTraversed uint64
	indexSpentSats   bool
}

// NewIndexer is a function that returns a pointer to a new Indexer instance.
// It takes a variadic number of Option functions as arguments, which are used to set the configuration options for the Indexer.
func NewIndexer(opts ...Option) *Indexer {
	// Create a new Indexer instance with default options.
	idx := &Indexer{
		opts: &Options{},
	}
	// Apply each Option function to the Indexer's options.
	for _, v := range opts {
		v(idx.opts)
	}
	// Initialize the rangeCache map.
	idx.rangeCache = make(map[string][]*model.SatRange)
	return idx
}

func (idx *Indexer) Start() {
	go func() {
		for {
			select {
			case <-signal.InterruptChannel:
			default:
				err := idx.UpdateIndex()
				if errors.Is(err, signal.ErrInterrupted) ||
					errors.Is(err, mysql.ErrInvalidConn) {
					return
				}
				if err != nil {
					log.Srv.Error("UpdateIndex", "err", err)
				}
				time.Sleep(time.Second * 5)
			}
		}
	}()
}

func (idx *Indexer) Stop() {
}

// DB is a method that returns a pointer to the dao.DB instance associated with the Indexer.
func (idx *Indexer) DB() *dao.DB {
	return idx.opts.db
}

// Begin is a method that starts a new transaction and returns a pointer to the dao.DB instance associated with the transaction.
func (idx *Indexer) Begin() *dao.DB {
	return &dao.DB{DB: idx.opts.db.DB.Begin()}
}

// RpcClient is a method that returns a pointer to the rpcclient.Client instance associated with the Indexer.
// This client is used for making RPC calls to the Bitcoin node.
func (idx *Indexer) RpcClient() *rpcclient.Client {
	return idx.opts.cli
}

// BatchRpcClient is a method that returns a pointer to the rpcclient.Client instance associated with the Indexer.
// This client is used for making batch RPC calls to the Bitcoin node.
func (idx *Indexer) BatchRpcClient() *rpcclient.Client {
	return idx.opts.batchCli
}

// UpdateIndex is a method that updates the index of the blockchain.
// It fetches blocks from the blockchain, starting from the current height of the indexer, and indexes them.
// If the indexer is configured to index satoshis, it flushes the satoshi range cache to the database.
// It also updates various statistics related to the indexing process.
// The method returns an error if there is any issue during the indexing process.
func (idx *Indexer) UpdateIndex() error {
	var err error
	wtx := idx.Begin()
	// Get the current block count from the database.
	idx.height, err = wtx.BlockCount()
	if err != nil {
		return err
	}

	// Get the latest block height from the Bitcoin node.
	var endHeight int64
	endHeight, err = idx.RpcClient().GetBlockCount()
	if err != nil {
		return err
	}

	// Fetch blocks from the blockchain, starting from the current height of the indexer.
	var blocksCh chan *wire.MsgBlock
	blocksCh, err = idx.fetchBlockFrom(idx.height, uint32(endHeight), 16)
	if err != nil {
		return err
	}
	if blocksCh == nil {
		return nil
	}

	unCommit := 0
	startTime := time.Now()
	valueCache := NewValueCache()

	// Iterate over the fetched blocks.
	for block := range blocksCh {
		unCommit++

		// Index the block.
		if err = idx.indexBlock(wtx, block, valueCache); err != nil {
			return err
		}

		if unCommit >= constants.DefaultWithFlushNum ||
			valueCache.Len() >= constants.DefaultFlushCacheNum ||
			time.Since(startTime) > time.Minute*30 {
			unCommit = 0
			startTime = time.Now()

			if err := idx.commit(wtx, valueCache); err != nil {
				return err
			}
			valueCache = NewValueCache()

			wtx = idx.Begin()

			// Check if the block count in the database matches the current height of the indexer.
			height, err := wtx.BlockCount()
			if err != nil {
				return err
			}
			if height != idx.height {
				return errors.New("height mismatch")
			}
		}
	}

	if unCommit > 0 {
		if err = idx.commit(wtx, valueCache); err != nil {
			return err
		}
	}
	return nil
}

// indexBlock is a method that indexes a block from the blockchain.
// It updates various statistics related to the indexing process.
// The method returns an error if there is any issue during the indexing process.
func (idx *Indexer) indexBlock(
	wtx *dao.DB,
	block *wire.MsgBlock,
	valueCache *ValueCache) error {

	// Detect if there is a reorganization in the blockchain.
	if err := detectReorg(wtx, block, idx.height); err != nil {
		return err
	}

	start := time.Now()
	satRangesWritten := uint64(0)
	outputsInBlock := uint64(0)
	indexInscriptions :=
		/*idx.height >= index.first_inscription_height && */ !idx.opts.noIndexInscriptions

	// Get various statistics related to the indexing process.
	lostSats, err := wtx.GetStatisticCountByName(tables.StatisticLostSats)
	if err != nil {
		return err
	}
	unboundInscriptions, err := wtx.GetStatisticCountByName(tables.StatisticUnboundInscriptions)
	if err != nil {
		return err
	}
	cursedInscriptions, err := wtx.GetStatisticCountByName(tables.StatisticCursedInscriptions)
	if err != nil {
		return err
	}
	blessedInscriptions, err := wtx.GetStatisticCountByName(tables.StatisticBlessedInscriptions)
	if err != nil {
		return err
	}
	nextSequenceNumber, err := wtx.NextSequenceNumber()
	if err != nil {
		return err
	}
	inscriptionUpdater := &InscriptionUpdater{
		wtx:                     wtx,
		idx:                     idx,
		valueCache:              valueCache,
		lostSats:                &lostSats,
		nextSequenceNumber:      &nextSequenceNumber,
		unboundInscriptions:     &unboundInscriptions,
		cursedInscriptionCount:  &cursedInscriptions,
		blessedInscriptionCount: &blessedInscriptions,
		timestamp:               block.Header.Timestamp.Unix(),
	}

	// If the indexer is configured to index satoshis, index the satoshis in the block.
	if idx.opts.indexSats {
		coinbaseInputs := make([]*model.SatRange, 0)
		h := NewHeight(idx.height)
		if h.Subsidy() > 0 {
			start := h.StartingSat()
			end := Sat(start.N() + h.Subsidy())
			coinbaseInputs = append([]*model.SatRange{
				{
					Start: start.N(),
					End:   end.N(),
				},
			}, coinbaseInputs...)
			idx.satRangesSinceFlush++
		}

		for _, tx := range block.Transactions[1:] {
			inputSatRanges := make([]*model.SatRange, 0)
			for _, input := range tx.TxIn {
				key := input.PreviousOutPoint.Hash.String()
				var satRanges []*model.SatRange
				if idx.indexSpentSats {
					satRanges = idx.rangeCache[key]
				} else {
					satRanges = idx.rangeCache[key]
					delete(idx.rangeCache, key)
				}
				if satRanges != nil {
					idx.outputsCached++
				} else {
					inputSatRanges, err = idx.DB().OutpointToSatRanges(key)
					if err != nil {
						return err
					}
					if !idx.indexSpentSats {
						if _, err := wtx.DelSatPointByOutpoint(key); err != nil {
							return err
						}
					}
				}
				for i := range satRanges {
					inputSatRanges = append(inputSatRanges, satRanges[i])
				}
			}
			inputSatRages, err := idx.indexTransactionSats(
				wtx,
				tx,
				inputSatRanges,
				&satRangesWritten,
				&outputsInBlock,
				inscriptionUpdater,
				indexInscriptions,
			)
			if err != nil {
				return err
			}
			coinbaseInputs = append(coinbaseInputs, inputSatRages...)
		}

		coinbaseInputs, err := idx.indexTransactionSats(
			wtx,
			block.Transactions[0],
			coinbaseInputs,
			&satRangesWritten,
			&outputsInBlock,
			inscriptionUpdater,
			indexInscriptions,
		)
		if err != nil {
			return err
		}

		if len(coinbaseInputs) > 0 {
			emptyOutpoint := &wire.OutPoint{
				Hash:  chainhash.Hash{},
				Index: 0,
			}
			lostSatRanges, err := wtx.DelSatPointByOutpoint(emptyOutpoint.String())
			if err != nil {
				return err
			}
			for _, coinBase := range coinbaseInputs {
				start := Sat(coinBase.Start)
				if !start.Common() {
					if err := wtx.SatToSatPointCreate(&tables.SatSatPoint{
						Sat:      start.N(),
						Outpoint: emptyOutpoint.String(),
						Offset:   lostSats,
					}); err != nil {
						return err
					}
				}

				lostSatRanges = append(lostSatRanges, &model.SatRange{
					Start: coinBase.Start,
					End:   coinBase.End,
				})
				lostSats += coinBase.End - coinBase.Start
			}

			if err := wtx.SetOutpointToSatRange(map[string][]*model.SatRange{
				emptyOutpoint.String(): lostSatRanges,
			}); err != nil {
				return err
			}
		}

	} else if indexInscriptions {
		// If the indexer is configured to index inscriptions, index the inscriptions in the block.
		txs := append(block.Transactions[1:], block.Transactions[0])
		// Iterate over the transactions in the block.
		for i := range txs {
			tx := txs[i]
			if err := inscriptionUpdater.indexEnvelopers(tx, nil); err != nil {
				return err
			}
		}
	}

	if indexInscriptions {
		// If the indexer is configured to index inscriptions, save the block information into the database.
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

	if idx.opts.indexSats {
		if err := wtx.SetStatistic(tables.StatisticLostSats, lostSats); err != nil {
			return err
		}
	} else {
		if err := wtx.SetStatistic(tables.StatisticLostSats, *inscriptionUpdater.lostSats); err != nil {
			return err
		}
	}
	if err := wtx.SetStatistic(tables.StatisticCursedInscriptions, *inscriptionUpdater.cursedInscriptionCount); err != nil {
		return err
	}
	if err := wtx.SetStatistic(tables.StatisticBlessedInscriptions, *inscriptionUpdater.blessedInscriptionCount); err != nil {
		return err
	}
	if err := wtx.SetStatistic(tables.StatisticUnboundInscriptions, *inscriptionUpdater.unboundInscriptions); err != nil {
		return err
	}

	// Increment the height of the indexer and the number of outputs traversed.
	atomic.AddUint32(&idx.height, 1)
	atomic.AddUint64(&idx.outputsTraversed, outputsInBlock)
	log.Srv.Infof("Block Height %d Wrote %d sat ranges from %d outputs in %d ms", idx.height-1, satRangesWritten, outputsInBlock, time.Since(start).Milliseconds())
	return nil
}

// detectReorg is a method that detects if there is a reorganization in the blockchain.
func (idx *Indexer) commit(wtx *dao.DB, valueCache *ValueCache) (err error) {
	log.Srv.Infof(
		"Committing at block %d, %d outputs traversed, %d in map, %d cached",
		idx.height-1, idx.outputsTraversed, valueCache.Len(), idx.outputsCached,
	)

	// If the indexer is configured to index satoshis, flush the satoshi range cache to the database.
	if idx.opts.indexSats {
		log.Srv.Infof(
			"Flushing %d entries (%.1f%% resulting from %d insertions) from memory to database",
			len(idx.rangeCache),
			float64(len(idx.rangeCache))/float64(idx.outputsInsertedSinceFlush)*100,
			idx.outputsInsertedSinceFlush,
		)
		if err = wtx.SetOutpointToSatRange(idx.rangeCache); err != nil {
			return err
		}
		idx.outputsInsertedSinceFlush = 0
	}

	// Update the value cache.
	if err = wtx.SetOutpointToValue(valueCache.Values()); err != nil {
		return err
	}

	// Update various statistics related to the indexing process.
	if err = wtx.IncrementStatistic(tables.StatisticOutputsTraversed, idx.outputsTraversed); err != nil {
		return err
	}
	idx.outputsTraversed = 0
	if err = wtx.IncrementStatistic(tables.StatisticSatRanges, idx.satRangesSinceFlush); err != nil {
		return err
	}
	idx.satRangesSinceFlush = 0
	if err = wtx.IncrementStatistic(tables.StatisticCommits, 1); err != nil {
		return err
	}

	wtx.Commit()
	return nil
}

// fetchBlockFrom is a method that fetches blocks from the blockchain, starting from a specified start height and ending at a specified end height.
// It returns a channel that emits the fetched blocks.
// The method returns an error if there is any issue during the fetching process.
func (idx *Indexer) fetchBlockFrom(start, end, current uint32) (chan *wire.MsgBlock, error) {
	if start > end {
		return nil, nil
	}
	blockCh := make(chan *wire.MsgBlock, current)

	// Start a goroutine to fetch the blocks and emit them to the channel.
	go func() {
		defer close(blockCh)
		for i := start; i <= end; i++ {
			var err error
			var block *wire.MsgBlock
			block, err = idx.getBlockWithRetries(i)
			if err != nil {
				log.Srv.Warn("getBlockWithRetries", err)
				return
			}
			blockCh <- block
		}
	}()
	return blockCh, nil
}

// getBlockWithRetries is a method that fetches a block from the blockchain at a specified height.
// It retries the fetching process if there is any issue, with an exponential backoff.
// The method returns an error if there is any issue during the fetching process.
func (idx *Indexer) getBlockWithRetries(height uint32) (*wire.MsgBlock, error) {
	errs := -1
	for {
		select {
		case <-signal.InterruptChannel:
			return nil, signal.ErrInterrupted
		default:
			errs++
			if errs > 0 {
				seconds := 1 << errs
				if seconds > 120 {
					err := errors.New("would sleep for more than 120s, giving up")
					log.Srv.Error(err)
				}
				time.Sleep(time.Second * time.Duration(seconds))
			}
			// Get the hash of the block at the specified height.
			hash, err := idx.RpcClient().GetBlockHash(int64(height))
			if err != nil && !errors.Is(err, rpcclient.ErrClientShutdown) {
				log.Srv.Warn("GetBlockHash", err)
				continue
			}
			if errors.Is(err, rpcclient.ErrClientShutdown) {
				return nil, signal.ErrInterrupted
			}

			// Get the block with the obtained hash.
			block, err := idx.opts.cli.GetBlock(hash)
			if err != nil && !errors.Is(err, rpcclient.ErrClientShutdown) {
				log.Srv.Warn("GetBlock", err)
				continue
			}
			if errors.Is(err, rpcclient.ErrClientShutdown) {
				return nil, signal.ErrInterrupted
			}
			return block, nil
		}
	}
}

func (idx *Indexer) indexTransactionSats(
	wtx *dao.DB,
	tx *wire.MsgTx,
	inputSatRanges []*model.SatRange,
	satRangesWritten *uint64,
	outputsInBlock *uint64,
	inscriptionUpdater *InscriptionUpdater,
	indexInscriptions bool,
) (list []*model.SatRange, err error) {
	if indexInscriptions {
		if err = inscriptionUpdater.indexEnvelopers(tx, inputSatRanges); err != nil {
			return
		}
	}

	for i, txOut := range tx.TxOut {
		outpoint := wire.OutPoint{
			Hash:  tx.TxHash(),
			Index: uint32(i),
		}
		sats := make([]*model.SatRange, 0)

		remaining := uint64(txOut.Value)
		for remaining > 0 {
			if len(inputSatRanges) == 0 {
				return
			}
			firstRange := inputSatRanges[0]
			startSat := Sat(firstRange.Start)

			if !startSat.Common() {
				if err := wtx.SatToSatPointCreate(&tables.SatSatPoint{
					Sat:       firstRange.Start,
					Outpoint:  outpoint.String(),
					Offset:    uint64(txOut.Value) - remaining,
					CreatedAt: time.Time{},
					UpdatedAt: time.Time{},
				}); err != nil {
					return nil, err
				}
			}

			count := firstRange.End - firstRange.Start
			assigned := firstRange
			if count > remaining {
				idx.satRangesSinceFlush++
				middle := firstRange.Start + remaining
				inputSatRanges = append([]*model.SatRange{
					{
						Start: middle,
						End:   firstRange.End,
					},
				}, inputSatRanges...)
				assigned.Start = firstRange.Start
				assigned.End = middle
			}

			sats = append(sats, assigned)

			remaining -= assigned.End - assigned.Start
			*satRangesWritten++
		}

		*outputsInBlock++

		idx.rangeCache[outpoint.String()] = sats
		idx.outputsInsertedSinceFlush++
	}

	list = inputSatRanges
	return
}
