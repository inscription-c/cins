package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/internal/util"
	"io"
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
	indexSats string
	// indexSpentSats is a boolean that indicates whether to index spent satoshis or not.
	indexSpentSats string
	// indexTransaction is a boolean that indicates whether to index transactions or not.
	indexTransaction bool
	// noIndexInscriptions is a boolean that indicates whether to index inscriptions or not.
	noIndexInscriptions bool
	// startHeight is an uint32 that represents the starting height of the blockchain to index from.
	firstInscriptionHeight uint32
	// no sync block info
	noSyncBlock bool
	// limit the tidb tx session memory, default 3GB
	tidbSessionMemLimit int
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
func WithIndexSats(indexSats string) func(*Options) {
	return func(options *Options) {
		options.indexSats = indexSats
	}
}

// WithIndexSpendSats is a function that returns an Option.
func WithIndexSpendSats(indexSpendSats string) func(*Options) {
	return func(options *Options) {
		options.indexSpentSats = indexSpendSats
	}
}

func WithNoSyncBLockInfo(noSyncBlock bool) func(*Options) {
	return func(options *Options) {
		options.noSyncBlock = noSyncBlock
	}
}

func WithTidbSessionMemLimit(tidbSessionMemLimit int) func(*Options) {
	return func(options *Options) {
		options.tidbSessionMemLimit = tidbSessionMemLimit
	}
}

// Indexer is a struct that holds the configuration options and state for the Indexer.
type Indexer struct {
	// opts is a pointer to an Options struct which holds the configuration options for the Indexer.
	opts *Options
	// cache for outpoint value
	valueCache *ValueCache
	// rangeCache is a map that caches the range of satoshis for each transaction.
	rangeCache *RangeCaches
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
	// indexSats is a boolean that indicates whether to index satoshis or not.
	indexSats bool
	// indexSpentSats is a boolean that indicates whether to index spent satoshis or not.
	indexSpentSats bool
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
	idx.valueCache = NewValueCache()
	idx.rangeCache = NewRangeCaches()
	return idx
}

// Start is a method that starts the Indexer.
func (idx *Indexer) Start() {
	if idx.opts.noSyncBlock {
		return
	}
	go func() {
		for {
			select {
			case <-signal.InterruptChannel:
			default:
				err := idx.UpdateIndex()
				if errors.Is(err, signal.ErrInterrupted) {
					return
				}
				if err != nil {
					log.Srv.Error("UpdateIndex", err)
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
	session := idx.opts.db.DB.Begin()
	if idx.opts.db.EmbedDB() {
		limit := 1
		if idx.opts.tidbSessionMemLimit > limit {
			limit = idx.opts.tidbSessionMemLimit
		}
		session.Exec(fmt.Sprintf("SET tidb_mem_quota_query = %d << 30;", limit))
	}
	return &dao.DB{DB: session}
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
	if idx.opts.indexSats != "" || idx.opts.indexSpentSats != "" {
		if err := idx.DB().Transaction(func(tx *dao.DB) error {
			indexSats := gconv.Uint64(gconv.Bool(idx.opts.indexSats) || gconv.Bool(idx.opts.indexSpentSats))
			if err := tx.SetStatistic(tables.StatisticIndexSats, indexSats); err != nil {
				return err
			}
			if idx.opts.indexSpentSats != "" {
				if err := tx.SetStatistic(tables.StatisticIndexSpentSats, gconv.Uint64(gconv.Bool(idx.opts.indexSpentSats))); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	wtx := idx.Begin()
	var err error
	defer func() {
		if err != nil {
			wtx.Rollback()
		}
	}()

	indexSats, err := wtx.GetStatisticCountByName(tables.StatisticIndexSats)
	if err != nil {
		return err
	}
	idx.indexSats = indexSats > 0

	indexSpentSats, err := wtx.GetStatisticCountByName(tables.StatisticIndexSpentSats)
	if err != nil {
		return err
	}
	idx.indexSpentSats = indexSpentSats > 0

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
		wtx.Commit()
		return nil
	}

	unCommit := 0
	startTime := time.Now()

	for block := range blocksCh {
		unCommit++

		// Index the block.
		if err = idx.indexBlock(wtx, block); err != nil {
			return err
		}

		if idx.needCommit(startTime, unCommit) {
			unCommit = 0
			startTime = time.Now()

			if err = idx.commit(wtx); err != nil {
				return err
			}

			wtx = idx.Begin()
			var height uint32
			height, err = wtx.BlockCount()
			if err != nil {
				return err
			}
			if height != idx.height {
				return errors.New("height mismatch")
			}
		}
	}

	if unCommit > 0 {
		if err = idx.commit(wtx); err != nil {
			return err
		}
	} else {
		wtx.Commit()
	}
	return nil
}

// indexBlock is a method that indexes a block from the blockchain.
// It updates various statistics related to the indexing process.
// The method returns an error if there is any issue during the indexing process.
func (idx *Indexer) indexBlock(
	wtx *dao.DB,
	block *wire.MsgBlock,
) error {

	// Detect if there is a reorganization in the blockchain.
	if err := detectReorg(wtx, block, idx.height); err != nil {
		return err
	}

	startTime := time.Now()
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
	sequenceNumber := nextSequenceNumber

	inscriptionUpdater := &InscriptionUpdater{
		wtx:                     wtx,
		idx:                     idx,
		valueCache:              idx.valueCache,
		reward:                  NewHeight(idx.height).Subsidy(),
		lostSats:                &lostSats,
		nextSequenceNumber:      &nextSequenceNumber,
		unboundInscriptions:     &unboundInscriptions,
		cursedInscriptionCount:  &cursedInscriptions,
		blessedInscriptionCount: &blessedInscriptions,
		timestamp:               block.Header.Timestamp.Unix(),
	}

	// If the indexer is configured to index satoshis, index the satoshis in the block.
	if idx.indexSats {
		coinbaseInputs := make(tables.SatRanges, 0)
		h := NewHeight(idx.height)

		if h.Subsidy() > 0 {
			start := h.StartingSat()
			end := Sat(start.N() + h.Subsidy())
			coinbaseInputs = append([]*tables.SatRange{
				{
					Start: start.N(),
					End:   end.N(),
				},
			}, coinbaseInputs...)
			idx.satRangesSinceFlush++
		}

		txSatsIdxStart := time.Now()
		latestSatsIdxStart := time.Now()
		commonTx := block.Transactions[1:]
		for i := range commonTx {
			tx := commonTx[i]

			inputSatRanges := make(tables.SatRanges, 0)

			for _, input := range tx.TxIn {
				select {
				case <-signal.InterruptChannel:
					return nil
				default:
					outpoint := input.PreviousOutPoint.String()

					var ok bool
					var satRanges *bytes.Buffer
					if idx.indexSpentSats {
						satRanges, ok = idx.rangeCache.Read(outpoint)
					} else {
						satRanges, ok = idx.rangeCache.Delete(outpoint)
					}
					if ok {
						idx.outputsCached++
					} else {
						var outpointSatRanges []*tables.OutpointSatRange
						if idx.indexSpentSats {
							outpointSatRanges, err = wtx.OutpointToSatRanges(outpoint)
							if err != nil {
								return err
							}
						} else {
							outpointSatRanges, err = wtx.DelSatRangesByOutpoint(outpoint)
							if err != nil {
								return err
							}
						}
						if len(outpointSatRanges) == 0 {
							return fmt.Errorf("could not find outpoint %s in index", outpoint)
						}

						for idx, v := range outpointSatRanges {
							if idx == 0 {
								satRanges = bytes.NewBuffer(v.SatRange)
								continue
							}
							satRanges.Write(v.SatRange)
						}
					}

					satRangesEntry, err := tables.NewSatRanges(satRanges.Bytes())
					if err != nil {
						return err
					}
					inputSatRanges = append(inputSatRanges, satRangesEntry...)
				}
			}

			if err := idx.indexTransactionSats(
				wtx,
				tx,
				&inputSatRanges,
				&satRangesWritten,
				&outputsInBlock,
				inscriptionUpdater,
				indexInscriptions,
			); err != nil {
				return err
			}
			coinbaseInputs = append(coinbaseInputs, inputSatRanges...)

			timeNow := time.Now()
			if timeNow.Sub(latestSatsIdxStart).Seconds() > 3 ||
				(timeNow.Sub(txSatsIdxStart).Seconds() > 3 && i+1 == len(commonTx)) {
				log.Srv.Infof("Block %d indexTransactionSats %d/%d in %s", idx.height, i+1, len(commonTx), time.Since(txSatsIdxStart)/1e6*1e6)
				latestSatsIdxStart = time.Now()
			}
		}

		coinBaseTx := block.Transactions[0]
		if err := idx.indexTransactionSats(
			wtx,
			coinBaseTx,
			&coinbaseInputs,
			&satRangesWritten,
			&outputsInBlock,
			inscriptionUpdater,
			indexInscriptions,
		); err != nil {
			return err
		}

		if len(coinbaseInputs) > 0 {
			nullOutpoint := util.NullOutpoint().String()
			lostSatRanges, err := wtx.DelSatRangesByOutpoint(nullOutpoint)
			if err != nil {
				return err
			}
			if len(lostSatRanges) == 0 {
				lostSatRanges = []*tables.OutpointSatRange{{
					Outpoint: nullOutpoint,
					SatRange: make([]byte, 0),
				}}
			}
			for i, coinBase := range coinbaseInputs {
				latestIdx := len(lostSatRanges) - 1
				start := Sat(coinBase.Start)
				if !start.Common() {
					if err := wtx.SatToSatPoint(&tables.SatSatPoint{
						Sat:      start.N(),
						Outpoint: nullOutpoint,
						Offset:   lostSats,
					}); err != nil {
						return err
					}
				}

				lostSatRanges[latestIdx].SatRange = append(lostSatRanges[latestIdx].SatRange, coinBase.Store()...)
				lostSats += coinBase.End - coinBase.Start

				if i < len(coinbaseInputs)-1 &&
					len(lostSatRanges[latestIdx].SatRange) >= constants.MaxSatRangesDataSize {
					lostSatRanges = append(lostSatRanges,
						&tables.OutpointSatRange{
							Outpoint: nullOutpoint,
							SatRange: make([]byte, 0),
						},
					)
				}
			}
			if err := wtx.SetOutpointToSatRange(lostSatRanges...); err != nil {
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
		buf := bytes.NewBufferString("")
		if err := block.Header.Serialize(buf); err != nil {
			return err
		}
		blockInfo := &tables.BlockInfo{
			Height:    idx.height,
			Header:    buf.Bytes(),
			Timestamp: block.Header.Timestamp.Unix(),
		}
		if *inscriptionUpdater.nextSequenceNumber > sequenceNumber {
			blockInfo.SequenceNum = *inscriptionUpdater.nextSequenceNumber
		}
		if err := wtx.SaveBlockInfo(blockInfo); err != nil {
			return err
		}
	}

	if idx.indexSats {
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

	atomic.AddUint32(&idx.height, 1)
	atomic.AddUint64(&idx.outputsTraversed, outputsInBlock)
	log.Srv.Infof("Block %d Wrote %d sat ranges from %d outputs in %s", idx.height-1, satRangesWritten, outputsInBlock, time.Since(startTime)/1e6*1e6)
	return nil
}

// indexTransactionSats is a method that indexes the satoshis in a transaction.
// It takes as arguments a pointer to a dao.DB instance (wtx), a pointer to a wire.MsgTx instance (tx),
// a slice of pointers to tables.OutpointSatRange instances (inputSatRanges), a pointer to an uint64 (satRangesWritten),
// a pointer to an uint64 (outputsInBlock), a pointer to an InscriptionUpdater instance (inscriptionUpdater),
// and a boolean (indexInscriptions).
// The method returns a slice of pointers to tables.OutpointSatRange instances and an error.
// The method first checks if the indexer is configured to index inscriptions. If so, it indexes the inscriptions in the transaction.
// Then, for each output in the transaction, it creates a new outpoint and a slice of pointers to tables.OutpointSatRange instances.
// It then iterates over the remaining satoshis in the output, and for each remaining satoshi, it checks if there are any satoshi ranges left.
// If there are no satoshi ranges left, it returns an error. Otherwise, it gets the first satoshi range and removes it from the slice of input satoshi ranges.
// It then checks if the start satoshi is common. If it is not, it creates a new tables.SatSatPoint instance and adds it to the database.
// It then checks if the count of satoshis in the first range is greater than the remaining satoshis. If it is, it creates a new satoshi range and appends it to the slice of input satoshi ranges.
// It then appends the assigned satoshi range to the slice of satoshis, and subtracts the end of the assigned satoshi range from the remaining satoshis.
// It increments the number of satoshi ranges written, and increments the number of outputs in the block.
// It then adds the slice of satoshis to the range cache, and increments the number of outputs inserted since the last flush.
// Finally, it sets the list of input satoshi ranges to the remaining input satoshi ranges, and returns the list and any error that occurred during the process.
func (idx *Indexer) indexTransactionSats(
	wtx *dao.DB,
	tx *wire.MsgTx,
	inputSatRanges *tables.SatRanges,
	satRangesWritten *uint64,
	outputsTraversed *uint64,
	inscriptionUpdater *InscriptionUpdater,
	indexInscriptions bool,
) error {
	if indexInscriptions {
		if err := inscriptionUpdater.indexEnvelopers(tx, *inputSatRanges); err != nil {
			return err
		}
	}

	for i, txOut := range tx.TxOut {
		select {
		case <-signal.InterruptChannel:
		default:
			outpoint := wire.OutPoint{
				Hash:  tx.TxHash(),
				Index: uint32(i),
			}.String()

			sats := bytes.NewBufferString("")
			remaining := txOut.Value

			for remaining > 0 {
				if len(*inputSatRanges) == 0 {
					return errors.New("insufficient inputs for transaction outputs")
				}
				firstRange := (*inputSatRanges)[0]
				*inputSatRanges = (*inputSatRanges)[1:]
				startSat := Sat(firstRange.Start)

				if !startSat.Common() {
					offset := txOut.Value - remaining
					if offset < 0 {
						return errors.New("negative offset")
					}
					if err := wtx.SatToSatPoint(&tables.SatSatPoint{
						Sat:      firstRange.Start,
						Outpoint: outpoint,
						Offset:   uint64(offset),
					}); err != nil {
						return err
					}
				}

				count := int64(firstRange.End) - int64(firstRange.Start)
				assigned := firstRange

				if count > remaining {
					idx.satRangesSinceFlush++
					middle := firstRange.Start + uint64(remaining)
					*inputSatRanges = append([]*tables.SatRange{
						{
							Start: middle,
							End:   firstRange.End,
						},
					}, *inputSatRanges...)
					assigned.End = middle
				}
				sats.Write(assigned.Store())
				remaining -= int64(assigned.End) - int64(assigned.Start)
				*satRangesWritten++
			}

			*outputsTraversed++

			idx.rangeCache.Write(outpoint, sats.Bytes())
			idx.outputsInsertedSinceFlush++
		}
	}
	return nil
}

// needCommit is a method that checks if a commit is needed.
func (idx *Indexer) needCommit(
	startTime time.Time,
	unCommit int,
) bool {
	return unCommit >= constants.DefaultWithFlushNum ||
		idx.valueCache.Len() >= constants.DefaultFlushCacheNum ||
		idx.outputsTraversed >= constants.DefaultFlushOutputTraversed ||
		time.Since(startTime) > time.Minute*30
	//if commit {
	//	return true
	//}
	//return idx.valueCache.Size() >= constants.MaxInsertDataSize ||
	//	idx.rangeCache.Size() >= constants.MaxInsertDataSize
}

// detectReorg is a method that detects if there is a reorganization in the blockchain.
func (idx *Indexer) commit(wtx *dao.DB) (err error) {
	log.Srv.Infof(
		"Committing at block %d, %d outputs traversed, %d in map, %d cached",
		idx.height-1, idx.outputsTraversed, idx.valueCache.Len(), idx.outputsCached,
	)

	// If the indexer is configured to index satoshis, flush the satoshi range cache to the database.
	if idx.indexSats {
		log.Srv.Infof(
			"Flushing %d entries, total %d bytes (%.1f%% resulting from %d insertions) from memory to database",
			idx.rangeCache.Len(),
			idx.rangeCache.Size(),
			float64(idx.rangeCache.Len())/float64(idx.outputsInsertedSinceFlush)*100,
			idx.outputsInsertedSinceFlush,
		)

		buf := make([]byte, constants.MaxSatRangesDataSize)
		setRanges := make([]*tables.OutpointSatRange, 0)
		idx.rangeCache.Range(func(k string, v *bytes.Buffer) {
			if v.Len() <= constants.MaxSatRangesDataSize {
				setRanges = append(setRanges, &tables.OutpointSatRange{
					Outpoint: k,
					SatRange: v.Bytes(),
				})
				return
			}

			for {
				n, err := v.Read(buf)
				if err == io.EOF {
					break
				}
				dest := make([]byte, n)
				copy(dest, buf[:n])
				setRanges = append(setRanges, &tables.OutpointSatRange{
					Outpoint: k,
					SatRange: dest,
				})
			}
		})
		if err = wtx.SetOutpointToSatRange(setRanges...); err != nil {
			return err
		}
		setRanges = nil
	}

	// Update the value cache.
	if err = wtx.SetOutpointToValue(idx.valueCache.Values()); err != nil {
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

	idx.outputsInsertedSinceFlush = 0
	idx.valueCache = NewValueCache()
	idx.rangeCache = NewRangeCaches()
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
