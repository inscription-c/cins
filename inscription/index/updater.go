package index

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/dotbitHQ/insc/constants"
	"github.com/dotbitHQ/insc/inscription/log"
	"github.com/dotbitHQ/insc/wallet"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang/protobuf/proto"
	"github.com/nutsdb/nutsdb"
	"math"
	"sort"
	"time"
)

type Curse int

const (
	CurseDuplicateField        Curse = 1
	CurseIncompleteField       Curse = 2
	CurseNotAtOffsetZero       Curse = 3
	CurseNotInFirstInput       Curse = 4
	CursePointer               Curse = 5
	CursePushNum               Curse = 6
	CurseReInscription         Curse = 7
	CurseStutter               Curse = 8
	CurseUnrecognizedEvenField Curse = 9
)

type Flotsam struct {
	InscriptionId *InscriptionId
	Offset        int64
	Origin        Origin
}

type Origin struct {
	New OriginNew
	Old OriginOld
}

type OriginNew struct {
	Cursed        bool
	Fee           int64
	Hidden        bool
	Pointer       []byte
	ReInscription bool
	Unbound       bool
}

type OriginOld struct {
	OldSatPoint SatPoint
}

func (idx *Indexer) UpdateIndex() error {
	var err error
	idx.height, err = idx.BlockCount()
	if err != nil {
		return err
	}

	startingHeight, err := idx.opts.cli.GetBlockCount()
	if err != nil {
		return err
	}
	startingHeight += 1

	if err := idx.Begin(func(tx *Tx) error {
		if err := tx.Put(constants.BucketWriteTransactionStartingBlockCountToTimestamp,
			gconv.Bytes(fmt.Sprint(idx.height)),
			gconv.Bytes(fmt.Sprint(time.Now().Unix()))); err != nil {
			return err
		}

		//blockCh := idx.fetchBlockFrom()
		//outpointCh, valueCh := idx.spawnFetcher()
		//uncommitted := 0
		//valueCache := make(map[*wire.OutPoint]int64)
		//
		//for block := range blockCh {
		//
		//}
		return nil
	}, true); err != nil {
		return err
	}
	return nil
}

func (idx *Indexer) spawnFetcher() (outpointCh chan *wire.OutPoint, valueCh chan int64) {
	bufferSize := 20_000
	batchSize := 2048
	parallelRequests := 12
	outpointCh = make(chan *wire.OutPoint, bufferSize)
	valueCh = make(chan int64, bufferSize)

	go func() {
		for {
			outpoint, ok := <-outpointCh
			if !ok {
				log.Srv.Debug("outpointCh closed")
				break
			}

			outpoints := make([]*wire.OutPoint, 0, batchSize)
			outpoints = append(outpoints, outpoint)
			for i := 0; i < batchSize-1; i++ {
				select {
				case outpoint, ok := <-outpointCh:
					if !ok {
						break
					}
					outpoints = append(outpoints, outpoint)
				default:
					break
				}
			}

			getTxByTxids := func(txids []string) ([]*btcutil.Tx, error) {
				txs, err := idx.getTransactions(txids)
				if err != nil {
					return nil, err
				}
				return txs, nil
			}

			chunkSize := (len(outpoints) / parallelRequests) + 1
			futs := make([]*btcutil.Tx, 0, parallelRequests)
			txids := make([]string, 0, chunkSize)
			for i := 0; i < len(outpoints); i++ {
				txids = append(txids, outpoints[i].Hash.String())
				if i != 0 && i%chunkSize == 0 {
					txs, err := getTxByTxids(txids)
					if err != nil {
						log.Srv.Error("getTxByTxids", err)
						return
					}
					futs = append(futs, txs...)
					txids = make([]string, 0, chunkSize)
				}
			}
			if len(txids) > 0 {
				txs, err := getTxByTxids(txids)
				if err != nil {
					log.Srv.Error("getTxByTxids", err)
					return
				}
				futs = append(futs, txs...)
			}

			for i, tx := range futs {
				valueCh <- tx.MsgTx().TxOut[outpoints[i].Index].Value
			}
		}
	}()
	return
}

func (idx *Indexer) getTransactions(txids []string) (resp []*btcutil.Tx, err error) {
	if len(txids) == 0 {
		return
	}
	retries := 0
	for {
		if retries > 0 {
			time.Sleep(100 * time.Millisecond * time.Duration(math.Pow(float64(2), float64(retries))))
		}
		var rawTxGetResp wallet.FutureBatchGetRawTransactionResult
		for _, v := range txids {
			cmd := btcjson.NewGetRawTransactionCmd(v, btcjson.Int(0))
			rawTxGetResp = idx.opts.batchCli.SendCmd(cmd)
		}
		if err = idx.opts.batchCli.Send(); err != nil {
			retries++
			if retries >= 5 {
				err = fmt.Errorf("failed to fetch raw transactions after 5 retries: %s", err)
				return
			}
			continue
		}
		return rawTxGetResp.Receive()
	}
}

func (idx *Indexer) fetchBlockFrom() chan *wire.MsgBlock {
	ch := make(chan *wire.MsgBlock, 32)
	go func() {
		for {
			block, err := idx.getBlockWithRetries(idx.height)
			if err != nil {
				log.Srv.Error(err)
				break
			}
			ch <- block
			idx.height++
		}
	}()
	return ch
}

func (idx *Indexer) getBlockWithRetries(height int64) (*wire.MsgBlock, error) {
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
		hash, err := idx.opts.cli.GetBlockHash(height)
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
	wtx *Tx,
	outpointCh chan *wire.OutPoint,
	valueCh chan int64,
	block *wire.MsgBlock,
	valueCache map[string]int64) error {
	if err := idx.detectReorg(block, idx.height); err != nil {
		return err
	}
	txids := make(map[string]struct{}, len(block.Transactions))
	for _, tx := range block.Transactions {
		txids[tx.TxHash().String()] = struct{}{}
	}

	// index inscriptions
	for _, tx := range block.Transactions {
		for _, input := range tx.TxIn {
			preOutput := input.PreviousOutPoint
			newHash := make([]byte, chainhash.HashSize)
			// We don't need coinbase input value
			if bytes.Compare(preOutput.Hash[:], newHash) == 0 {
				continue
			}
			// We don't need input values from txs earlier in the block, since they'll be added to value_cache
			// when the tx is indexed
			if _, ok := txids[preOutput.Hash.String()]; ok {
				continue
			}
			// We don't need input values we already have in our value_cache from earlier blocks
			if _, ok := valueCache[preOutput.String()]; ok {
				continue
			}
			// We don't need input values we already have in our outpoint_to_value table from earlier blocks that
			// were committed to db already
			if _, err := wtx.GetValueByOutpoint(preOutput.String()); err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
				return err
			} else if err == nil {
				continue
			}
			// We don't know the value of this tx input. Send this outpoint to background thread to be fetched
			outpointCh <- &preOutput
		}
	}

	cursedInscriptionCount, err := idx.GetStatisticCount(constants.StatisticCursedInscriptions)
	if err != nil {
		return err
	}
	blessedInscriptionCount, err := idx.GetStatisticCount(constants.StatisticBlessedInscriptions)
	if err != nil {
		return err
	}
	unboundInscriptions, err := idx.GetStatisticCount(constants.StatisticUnboundInscriptions)
	if err != nil {
		return err
	}
	nextSequenceNumber, err := idx.NextSequenceNumber()
	if err != nil {
		return err
	}

	inscriptionUpdater := &InscriptionUpdater{
		wtx:                     wtx,
		blessedInscriptionCount: blessedInscriptionCount,
		cursedInscriptionCount:  cursedInscriptionCount,
		height:                  idx.height,
		nextSequenceNumber:      nextSequenceNumber,
		timestamp:               block.Header.Timestamp.Unix(),
		unboundInscriptions:     unboundInscriptions,
		valueCache:              valueCache,
		valueCh:                 valueCh,
	}

	txs := append([]*wire.MsgTx{block.Transactions[len(block.Transactions)-1]}, block.Transactions[1:]...)
	for i := range txs {
		tx := txs[i]
		if err := inscriptionUpdater.indexEnvelopers(tx); err != nil {
			return err
		}
	}

	return nil
}

type InscriptionUpdater struct {
	wtx                     *Tx
	flotsam                 []*Flotsam
	lostSats                int64
	reward                  int64
	blessedInscriptionCount int64
	cursedInscriptionCount  int64
	height                  int64
	nextSequenceNumber      int64
	timestamp               int64
	unboundInscriptions     int64
	valueCache              map[string]int64
	valueCh                 chan int64
}

type inscribedOffsetEntity struct {
	inscriptionId *InscriptionId
	count         int64
}

type locationsInscription struct {
	satpoint *SatPoint
	flotsam  *Flotsam
}

type rangeToVout struct {
	outputValue int64
	end         int64
}

func (u *InscriptionUpdater) indexEnvelopers(tx *wire.MsgTx) error {
	totalInputValue := int64(0)
	idCounter := int64(0)
	floatingInscriptions := make([]*Flotsam, 0)
	inscribedOffsets := make(map[int64]*inscribedOffsetEntity)
	envelopes := ParsedEnvelopeFromTransaction(tx)
	totalOutputValue := int64(0)
	for _, v := range tx.TxOut {
		totalOutputValue += v.Value
	}

	for inputIndex := range tx.TxIn {
		txIn := tx.TxIn[inputIndex]
		if IsEmptyHash(txIn.PreviousOutPoint.Hash) {
			h := Height{Height: u.height}
			totalInputValue += h.Subsidy()
			continue
		}

		inscriptions, err := u.inscriptionsOnOutput(txIn.PreviousOutPoint)
		if err != nil {
			return err
		}
		for _, v := range inscriptions {
			offset := totalInputValue + int64(v.SatPoint.Offset)
			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: v.Id,
				Offset:        offset,
				Origin: Origin{
					Old: OriginOld{OldSatPoint: *v.SatPoint},
				},
			})

			offsetEntity, ok := inscribedOffsets[offset]
			if !ok {
				offsetEntity = &inscribedOffsetEntity{
					inscriptionId: v.Id,
				}
				inscribedOffsets[offset] = offsetEntity
			}
			offsetEntity.count++
		}

		offset := totalInputValue
		currentInputValue, ok := u.valueCache[txIn.PreviousOutPoint.String()]
		if ok {
			delete(u.valueCache, txIn.PreviousOutPoint.String())
		} else {
			v, err := u.wtx.Get(constants.BucketOutpointToValue, []byte(txIn.PreviousOutPoint.String()))
			if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
				return err
			}
			if err != nil {
				err = fmt.Errorf("failed to get transaction for %s", txIn.PreviousOutPoint.String())
				select {
				case currentInputValue, ok = <-u.valueCh:
					if !ok {
						return err
					}
				default:
					return err
				}
			} else {
				if err := u.wtx.Delete(constants.BucketOutpointToValue, []byte(txIn.PreviousOutPoint.String())); err != nil {
					return err
				}
				currentInputValue = gconv.Int64(string(v))
			}
		}
		totalInputValue += currentInputValue

		for _, inscription := range envelopes {
			if inscription.input != inputIndex {
				break
			}
			inscriptionId := InscriptionId{OutPoint{OutPoint: btcjson.OutPoint{
				Hash:  tx.TxHash().String(),
				Index: uint32(idCounter),
			}}}

			// TODO chain jubilee_height check
			var curse Curse
			if inscription.payload.UnRecognizedEvenField {
				curse = CurseUnrecognizedEvenField
			} else if inscription.payload.DuplicateField {
				curse = CurseDuplicateField
			} else if inscription.payload.IncompleteField {
				curse = CurseIncompleteField
			} else if inscription.input != 0 {
				curse = CurseNotInFirstInput
			} else if inscription.offset != 0 {
				curse = CurseNotAtOffsetZero
			} else if len(inscription.payload.Pointer) > 0 {
				curse = CursePointer
			} else if inscription.pushNum {
				curse = CursePushNum
			} else if inscription.stutter {
				curse = CurseStutter
			} else {
				inscribedEntity, ok := inscribedOffsets[offset]
				if ok {
					if inscribedEntity.count > 1 {
						curse = CurseReInscription
					} else {
						initialInscriptionSequenceNumber, err := u.wtx.Get(constants.BucketInscriptionIdToSequenceNumber, []byte(inscribedEntity.inscriptionId.String()))
						if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
							return err
						}
						initialInscriptionIsCursed := false
						inscriptionEntryData, err := u.wtx.Get(constants.BucketSequenceNumberToInscriptionEntry, initialInscriptionSequenceNumber)
						if err != nil && !errors.Is(err, nutsdb.ErrKeyNotFound) {
							return err
						}
						if err == nil {
							inscriptionEntry := &InscriptionEntry{}
							if err := proto.Unmarshal(inscriptionEntryData, inscriptionEntry); err != nil {
								return err
							}
							initialInscriptionIsCursed = inscriptionEntry.InscriptionNumber < 0
						}
						if !initialInscriptionIsCursed {
							curse = CurseReInscription
						}
					}
				}
			}

			unbound := currentInputValue == 0 ||
				curse == CurseUnrecognizedEvenField ||
				inscription.payload.UnRecognizedEvenField

			if len(inscription.payload.Pointer) > 0 &&
				gconv.Int64(string(inscription.payload.Pointer)) < totalOutputValue {
				offset = gconv.Int64(string(inscription.payload.Pointer))
			}

			_, reInscription := inscribedOffsets[offset]

			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: &inscriptionId,
				Offset:        offset,
				Origin: Origin{
					New: OriginNew{
						Cursed:        curse > 0,
						Fee:           0,
						Hidden:        false,
						Pointer:       inscription.payload.Pointer,
						ReInscription: reInscription,
						Unbound:       unbound,
					},
				},
			})

			inscribedOffset, ok := inscribedOffsets[offset]
			if !ok {
				inscribedOffset = &inscribedOffsetEntity{
					inscriptionId: &inscriptionId,
				}
				inscribedOffsets[offset] = inscribedOffset
			}
			inscribedOffset.count++
			idCounter++
		}
	}

	for _, flotsam := range floatingInscriptions {
		flotsam.Origin.New.Fee = (totalInputValue - totalOutputValue) / idCounter
	}

	isCoinBase := IsEmptyHash(tx.TxIn[0].PreviousOutPoint.Hash)
	if isCoinBase {
		floatingInscriptions = append(floatingInscriptions, u.flotsam...)
	}
	sort.Slice(floatingInscriptions, func(i, j int) bool {
		return floatingInscriptions[i].Offset < floatingInscriptions[j].Offset
	})

	rangeToVoutMap := make(map[rangeToVout]int)
	outputValue := int64(0)
	newLocations := make([]*locationsInscription, 0)
	for vout, txOut := range tx.TxOut {
		end := outputValue + txOut.Value
		for _, flotsam := range floatingInscriptions {
			if flotsam.Offset >= end {
				break
			}
			newSatpoint := &SatPoint{
				Outpoint: &wire.OutPoint{
					Hash:  tx.TxHash(),
					Index: uint32(vout),
				},
				Offset: flotsam.Offset - outputValue,
			}
			newLocations = append(newLocations, &locationsInscription{
				satpoint: newSatpoint,
				flotsam:  flotsam,
			})
		}

		rangeToVoutMap[rangeToVout{
			outputValue: outputValue,
			end:         end,
		}] = vout

		outputValue = end

		outpoint := NewOutPoint(tx.TxHash().String(), uint32(vout)).String()
		u.valueCache[outpoint] = txOut.Value
	}

	for _, flotsam := range newLocations {
		if len(flotsam.flotsam.Origin.New.Pointer) > 0 {
			pointer := gconv.Int64(string(flotsam.flotsam.Origin.New.Pointer))
			if pointer < outputValue {
				for rangeEntity, vout := range rangeToVoutMap {
					if pointer >= rangeEntity.outputValue && pointer < rangeEntity.end {
						flotsam.flotsam.Offset = pointer
						flotsam.satpoint = &SatPoint{
							Outpoint: &wire.OutPoint{
								Hash:  tx.TxHash(),
								Index: uint32(vout),
							},
							Offset: pointer - rangeEntity.outputValue,
						}
					}
				}
			}
		}

		// TODO
		// u.updateInscriptionLocation()
	}

	if isCoinBase {
		//for _, flotsam := range floatingInscriptions {
		//newSatpoint := &SatPoint{
		//	Outpoint: nil,
		//	Offset:   u.lostSats + flotsam.Offset - outputValue,
		//}
		// TODO
		// u.updateInscriptionLocation()
		//}
		u.lostSats += u.reward - outputValue
		return nil
	}

	for _, inscriptions := range floatingInscriptions {
		inscriptions.Offset = u.reward + inscriptions.Offset - outputValue
	}
	u.flotsam = append(u.flotsam, floatingInscriptions...)
	u.reward += totalInputValue - outputValue

	return nil
}

type inscriptionEntry struct {
	SequenceNumber int64
	SatPoint       *SatPoint
	Id             *InscriptionId
}

func (u *InscriptionUpdater) inscriptionsOnOutput(output wire.OutPoint) (inscriptions []*inscriptionEntry, err error) {
	if err = u.wtx.LKeys(constants.BucketSatpointToSequenceNumber, fmt.Sprintf("%s:.*", output.String()), func(key string) bool {
		var sequenceNumbers [][]byte
		sequenceNumbers, err = u.wtx.LRange(constants.BucketSatpointToSequenceNumber, []byte(key), 0, -1)
		if err != nil {
			return false
		}
		for _, v := range sequenceNumbers {
			var entryVal []byte
			entryVal, err = u.wtx.Get(constants.BucketSequenceNumberToInscriptionEntry, v)
			if err != nil {
				return false
			}
			entry := &InscriptionEntry{}
			if err = proto.Unmarshal(entryVal, entry); err != nil {
				return false
			}
			var satPoint *SatPoint
			satPoint, err = NewSatPointFromString(key)
			if err != nil {
				return false
			}
			inscriptions = append(inscriptions, &inscriptionEntry{
				SequenceNumber: gconv.Int64(string(v)),
				SatPoint:       satPoint,
				Id:             StringToInscriptionId(string(entry.Id)),
			})
		}
		return true
	}); err != nil {
		return
	}

	sort.SliceIsSorted(inscriptions, func(i, j int) bool {
		return inscriptions[i].SequenceNumber < inscriptions[j].SequenceNumber
	})
	return
}

func (u *InscriptionUpdater) updateInscriptionLocation() error {
	return nil
}
