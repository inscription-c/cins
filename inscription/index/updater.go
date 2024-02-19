package index

import (
	"bytes"
	"errors"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/gutil"
	"github.com/inscription-c/insc/btcd/rpcclient"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/internal/util"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"sort"
	"sync"
	"sync/atomic"
)

// Curse represents the type of curse that can be applied to an inscription.
type Curse int

// These constants represent the different types of curses that can be applied to an inscription.
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

// Flotsam represents a floating inscription.
type Flotsam struct {
	// InscriptionId is a pointer to the unique identifier of the inscription.
	InscriptionId *tables.InscriptionId

	// Offset is the position of the inscription within the transaction.
	Offset uint64

	// Origin is the source of the inscription. It can be either new or old.
	Origin Origin
}

// Origin represents the origin of an inscription.
//
// It is a struct that contains two pointers, one to an OriginNew and one to an OriginOld.
// These pointers represent the new and old origins of an inscription respectively.
// Only one of these pointers should be non-nil at a time, depending on whether the inscription is new or old.
//
// Fields:
//
//	New (*OriginNew): A pointer to an OriginNew struct, representing a new origin of an inscription.
//	Old (*OriginOld): A pointer to an OriginOld struct, representing an old origin of an inscription.
type Origin struct {
	New *OriginNew
	Old *OriginOld
}

// OriginNew represents a new origin of an inscription.
type OriginNew struct {
	// Cursed is a boolean flag indicating whether the inscription is cursed.
	Cursed bool

	// Fee is the fee associated with the inscription. It is represented as an int64.
	Fee int64

	// Hidden is a boolean flag indicating whether the inscription is hidden.
	Hidden bool

	// Pointer is an int32 that points to the location of the inscription.
	Pointer []byte

	// ReInscription is a boolean flag indicating whether the inscription is a re-inscription.
	ReInscription bool

	// Unbound is a boolean flag indicating whether the inscription is unbound.
	Unbound bool

	// Inscription is a pointer to the Envelope struct that contains the inscription.
	Inscription *Envelope
}

// OriginOld represents an old origin of an inscription.
type OriginOld struct {
	OldSatPoint tables.SatPointToSequenceNum
}

// InscriptionUpdater is responsible for updating inscriptions.
type InscriptionUpdater struct {
	// idx is a pointer to the Indexer struct that is used for indexing inscriptions.
	idx *Indexer

	// wtx is a pointer to the DB struct that is used for database operations.
	wtx *dao.DB

	// flotsam is a slice of pointers to Flotsam structs. Each Flotsam represents a floating inscription.
	flotsam []*Flotsam

	// lostSats is an uint64 that represents the total number of lost Satoshis.
	lostSats *uint64

	// reward is an uint64 that represents the total reward for mining a block.
	reward uint64

	// valueCache is a map where the key is a string and the value is an int64. It is used for caching the values of transactions.
	valueCache *ValueCache

	// timestamp is an int64 that represents the timestamp of the last block.
	timestamp int64

	// nextSequenceNumber is a pointer to an uint64 that represents the next sequence number to be used for a transaction.
	nextSequenceNumber *int64

	// unboundInscriptions is a pointer to an uint32 that represents the total number of unbound inscriptions.
	unboundInscriptions *uint64

	// cursedInscriptionCount is a pointer to an uint32 that represents the total number of cursed inscriptions.
	cursedInscriptionCount *uint64

	// blessedInscriptionCount is a pointer to an uint32 that represents the total number of blessed inscriptions.
	blessedInscriptionCount *uint64
}

type ValueCache struct {
	sync.RWMutex
	size int
	m    map[string]int64
}

func NewValueCache() *ValueCache {
	return &ValueCache{
		m: make(map[string]int64),
	}
}

func (c *ValueCache) Read(outpoint string) (int64, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.m[outpoint]
	return v, ok
}

func (c *ValueCache) Write(outpoint string, value int64) {
	c.Lock()
	c.m[outpoint] = value
	c.Unlock()
}

func (c *ValueCache) Delete(outpoint string, height ...uint32) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.m[outpoint]; ok {
		c.size -= len(outpoint)
		c.size -= 8
		delete(c.m, outpoint)
		return
	}
}

func (c *ValueCache) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.m)
}

func (c *ValueCache) Size() int {
	c.RLock()
	defer c.RUnlock()
	return c.size
}

func (c *ValueCache) Range(fn func(k string, v int64)) {
	c.RLock()
	defer c.RUnlock()

	for k, v := range c.m {
		fn(k, v)
	}
	return
}

func (c *ValueCache) Values() map[string]int64 {
	c.RLock()
	defer c.RUnlock()
	return c.m
}

type RangeCaches struct {
	sync.RWMutex
	size int
	m    map[string]*bytes.Buffer
}

func NewRangeCaches() *RangeCaches {
	return &RangeCaches{
		m: make(map[string]*bytes.Buffer),
	}
}

func (c *RangeCaches) Read(outpoint string) (*bytes.Buffer, bool) {
	c.RLock()
	defer c.RUnlock()

	v, ok := c.m[outpoint]
	return v, ok
}

func (c *RangeCaches) Write(outpoint string, ranges []byte) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.m[outpoint]; !ok {
		c.size += len(outpoint)
		c.m[outpoint] = bytes.NewBuffer(ranges)
	} else {
		c.m[outpoint].Write(ranges)
	}
	c.size += len(ranges)
}

func (c *RangeCaches) Delete(outpoint string) (*bytes.Buffer, bool) {
	c.Lock()
	defer c.Unlock()

	if v, ok := c.m[outpoint]; ok {
		c.size -= len(outpoint)
		c.size -= v.Len()
		delete(c.m, outpoint)
		return v, ok
	}
	return nil, false
}

func (c *RangeCaches) Range(fn func(k string, v *bytes.Buffer)) {
	c.RLock()
	defer c.RUnlock()

	for k, v := range c.m {
		fn(k, v)
	}
}

func (c *RangeCaches) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.m)
}

func (c *RangeCaches) Size() int {
	c.RLock()
	defer c.RUnlock()
	return c.size
}

// inscribedOffsetEntity represents an entity with an inscribed offset.
type inscribedOffsetEntity struct {
	inscriptionId *tables.InscriptionId
	count         int64
}

// locationsInscription represents the location of an inscription.
type locationsInscription struct {
	satpoint *tables.SatPointToSequenceNum
	flotsam  *Flotsam
}

// indexEnvelopers indexes the envelopers of a transaction.
func (u *InscriptionUpdater) indexEnvelopers(
	tx *wire.MsgTx,
	inputSatRange tables.SatRanges) error {

	idCounter := uint32(0)

	totalInputValue := int64(0)

	floatingInscriptions := make([]*Flotsam, 0)

	inscribedOffsets := make(map[uint64]*inscribedOffsetEntity)

	envelopes, err := util.NewPeekable(ParsedEnvelopFromTransaction(tx))
	if err != nil {
		return err
	}

	totalOutputValue := int64(0)
	for _, v := range tx.TxOut {
		totalOutputValue += v.Value
	}

	valueCh, errCh, needDelOutpoints, err := u.fetchOutputValues(tx, 32)
	if err != nil {
		return err
	}

	for inputIndex := range tx.TxIn {
		txIn := tx.TxIn[inputIndex]

		// Check if the input is a coinbase transaction.
		// If it is, add the subsidy for the current block height to the total input value and skip to the next iteration.
		if util.IsNullOutpoint(txIn.PreviousOutPoint) {
			totalInputValue += int64(NewHeight(u.idx.height).Subsidy())
			continue
		}

		// Fetch existing inscriptions on the input (transfers of inscriptions).
		inscriptions, err := u.wtx.InscriptionsByOutpoint(txIn.PreviousOutPoint.String())
		if err != nil {
			return err
		}

		// Loop over each fetched inscription.
		for _, v := range inscriptions {
			offset := uint64(totalInputValue) + v.SatPointToSequenceNum.Offset
			insId := &tables.InscriptionId{
				TxId:   v.TxId,
				Offset: v.Inscriptions.Offset,
			}
			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: insId,
				Offset:        offset,
				Origin: Origin{
					Old: &OriginOld{OldSatPoint: *v.SatPointToSequenceNum},
				},
			})

			// Check if the offset already exists in the inscribed offsets map.
			offsetEntity, ok := inscribedOffsets[offset]
			if !ok {
				// If it doesn't exist, create a new inscribed offset entity with the inscription ID and add it to the map.
				offsetEntity = &inscribedOffsetEntity{
					inscriptionId: insId,
				}
				inscribedOffsets[offset] = offsetEntity
			}
			// Increment the count of the inscribed offset entity.
			offsetEntity.count++
		}

		// Initialize the offset as the total input value.
		offset := uint64(totalInputValue)
		preOutpoint := txIn.PreviousOutPoint.String()

		// Try to get the current input value from the value cache.
		currentInputValue, ok := u.valueCache.Read(txIn.PreviousOutPoint.String())
		if ok {
			// If the value is in the cache, delete it from the cache.
			u.valueCache.Delete(preOutpoint)
		} else {
			if _, ok := needDelOutpoints[preOutpoint]; !ok {
				select {
				case <-signal.InterruptChannel:
					return signal.ErrInterrupted
				case currentInputValue, ok = <-valueCh:
					if !ok {
						return errors.New("valueCh closed")
					}
				case err = <-errCh:
					return err
				}
			}
		}
		// Add the current input value to the total input value.
		totalInputValue += currentInputValue

		// Loop over each inscription in the input.
		for v := envelopes.Peek(); ; v = envelopes.Peek() {
			if v == nil {
				break
			}
			inscription := v.(*Envelope)
			if inscription.index != uint32(inputIndex) {
				break
			}

			// Create a new inscription ID for the current inscription.
			inscriptionId := tables.InscriptionId{
				TxId:   tx.TxHash().String(),
				Offset: idCounter,
			}

			// Initialize a variable to store the type of curse, if any, that applies to the inscription.
			var curse Curse
			// Check each possible curse condition and set the curse variable accordingly.
			if inscription.payload.UnRecognizedEvenField {
				curse = CurseUnrecognizedEvenField
			} else if inscription.payload.DuplicateField {
				curse = CurseDuplicateField
			} else if inscription.payload.IncompleteField {
				curse = CurseIncompleteField
			} else if inscription.index != 0 {
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
				// If none of the above conditions are met, check if the inscription is a re-inscription.
				offsetEntity, ok := inscribedOffsets[offset]
				if ok {
					if offsetEntity.count > 1 {
						curse = CurseReInscription
					} else {
						entry, err := u.wtx.GetInscriptionById(offsetEntity.inscriptionId)
						if err != nil {
							return err
						}
						if entry.Id > 0 {
							iniInscriptionWasCursedOrVindicated := entry.InscriptionNum < 0 || CharmVindicated.IsSet(entry.Charms)
							if !iniInscriptionWasCursedOrVindicated {
								curse = CurseReInscription
							}
						}
					}
				}
			}

			// Determine if the inscription is unbound.
			unbound := currentInputValue == 0 ||
				curse == CurseUnrecognizedEvenField ||
				inscription.payload.UnRecognizedEvenField

			// Get the pointer from the inscription payload.
			if len(inscription.payload.Pointer) > 0 {
				pointer := gconv.Int64(string(inscription.payload.Pointer))
				if pointer < totalOutputValue {
					offset = uint64(pointer)
				}
			}
			// Check if the offset is a re-inscription.
			_, reInscription := inscribedOffsets[offset]

			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: &inscriptionId,
				Offset:        offset,
				Origin: Origin{
					New: &OriginNew{
						Cursed:        curse > 0,
						Fee:           0,
						Hidden:        false,
						Pointer:       inscription.payload.Pointer,
						ReInscription: reInscription,
						Unbound:       unbound,
						Inscription:   inscription,
					},
				},
			})

			// Check if the offset already exists in the inscribed offsets map.
			inscribedOffset, ok := inscribedOffsets[offset]
			if !ok {
				// If it doesn't exist, create a new inscribed offset entity with the inscription ID and add it to the map.
				inscribedOffset = &inscribedOffsetEntity{
					inscriptionId: &inscriptionId,
				}
				inscribedOffsets[offset] = inscribedOffset
			}
			// Increment the count of the inscribed offset entity.
			inscribedOffset.count++
			idCounter++
			envelopes.Next()
		}
	}

	// If the value is in the database, delete it from the database.
	if err := u.wtx.DeleteValueByOutpoint(gutil.Keys(needDelOutpoints)...); err != nil {
		return err
	}

	// TODO index transaction
	// TODO potential_parents

	// still have to normalize over inscription size
	for _, flotsam := range floatingInscriptions {
		if flotsam.Origin.New == nil {
			continue
		}
		flotsam.Origin.New.Fee = (totalInputValue - totalOutputValue) / int64(idCounter)
	}

	// Check if the transaction is a coinbase transaction.
	// A coinbase transaction is a unique type of bitcoin transaction that can only be created by a miner.
	// This is done by checking if the hash of the previous outpoint of the first input in the transaction is empty.
	isCoinBase := blockchain.IsCoinBaseTx(tx)

	// If the transaction is a coinbase transaction, append all the floating inscriptions from the updater to the current floating inscriptions.
	// Floating inscriptions are inscriptions that are not yet bound to a specific location in the blockchain.
	// They are stored in the updater and are bound to a location when a new block is mined.
	if isCoinBase {
		floatingInscriptions = append(floatingInscriptions, u.flotsam...)
		u.flotsam = make([]*Flotsam, 0)
	}

	// Sort the floating inscriptions by their offset.
	// The offset is the position of the inscription within the transaction.
	// Sorting is done in ascending order, so inscriptions with a lower offset will come before inscriptions with a higher offset.
	sort.Slice(floatingInscriptions, func(i, j int) bool {
		return floatingInscriptions[i].Offset < floatingInscriptions[j].Offset
	})

	// Initialize outputValue as zero. This will be used to keep track of the total value of the outputs processed so far.
	outputValue := uint64(0)

	// Initialize a map to associate each range of Satoshis in the transaction outputs with the corresponding output index (vout).
	rangeToVoutMap := make(map[tables.SatRange]int)

	// Initialize a slice to store the new locations of the inscriptions.
	newLocations := make([]*locationsInscription, 0)

	inscriptions, err := util.NewPeekable(floatingInscriptions)
	if err != nil {
		return err
	}
	for index, txOut := range tx.TxOut {
		end := outputValue + uint64(txOut.Value)

		for v := inscriptions.Peek(); ; v = inscriptions.Peek() {
			if v == nil {
				break
			}
			flotsam := v.(*Flotsam)
			if flotsam.Offset >= end {
				break
			}
			newSatpoint := &tables.SatPointToSequenceNum{
				Outpoint: tables.FormatOutpoint(tx.TxHash().String(), uint32(index)),
				Offset:   flotsam.Offset - outputValue,
			}
			newLocations = append(newLocations, &locationsInscription{
				satpoint: newSatpoint,
				flotsam:  inscriptions.Next().(*Flotsam),
			})
		}

		rangeToVoutMap[tables.SatRange{
			Start: outputValue,
			End:   end,
		}] = index

		// Update the total value of the outputs processed so far.
		outputValue = end

		// Cache the value of the current output.
		outpoint := model.NewOutPoint(tx.TxHash().String(), uint32(index)).String()
		u.valueCache.Write(outpoint, txOut.Value)
	}

	for _, location := range newLocations {
		flotsam := location.flotsam
		newSatpoint := location.satpoint

		if flotsam.Origin.New != nil {
			pointer := gconv.Uint64(string(flotsam.Origin.New.Pointer))
			if len(flotsam.Origin.New.Pointer) > 0 && pointer < outputValue {
				for rangeEntity, vout := range rangeToVoutMap {
					if pointer < rangeEntity.Start || pointer >= rangeEntity.End {
						continue
					}
					flotsam.Offset = pointer
					newSatpoint = &tables.SatPointToSequenceNum{
						Outpoint: tables.FormatOutpoint(tx.TxHash().String(), uint32(vout)),
						Offset:   pointer - rangeEntity.Start,
					}
				}
			}
		}

		// Update the location of the inscription in the database.
		if err := u.updateInscriptionLocation(inputSatRange, flotsam, newSatpoint); err != nil {
			return err
		}
	}

	// If the transaction is a coinbase transaction,
	// update the offset of each floating inscription
	// and add them to the list of floating inscriptions in the updater.
	if isCoinBase {
		if err := inscriptions.Range(func(i int, v interface{}) error {
			flotsam := v.(*Flotsam)
			newSatPoint := &tables.SatPointToSequenceNum{
				Offset: *u.lostSats + flotsam.Offset - outputValue,
			}
			if err := u.updateInscriptionLocation(inputSatRange, flotsam, newSatPoint); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		*u.lostSats += u.reward - outputValue
		return nil
	}

	// If the transaction is not a coinbase transaction,
	// update the offset of each floating inscription
	// and add them to the list of floating inscriptions in the updater.
	if err := inscriptions.Range(func(i int, v interface{}) error {
		flotsam := v.(*Flotsam)
		flotsam.Offset = u.reward + flotsam.Offset - outputValue
		u.flotsam = append(u.flotsam, flotsam)
		return nil
	}); err != nil {
		return err
	}
	// Update the total reward by subtracting the total value
	// of the outputs from the total value of the inputs.
	u.reward += uint64(totalInputValue) - outputValue

	return nil
}

// updateInscriptionLocation updates the location of an inscription.
func (u *InscriptionUpdater) updateInscriptionLocation(
	inputSatRanges tables.SatRanges,
	flotsam *Flotsam,
	newSatPoint *tables.SatPointToSequenceNum,
) error {

	// Initialize error, unbound flag, and sequence number.
	var unbound bool
	var sequenceNumber int64
	inscriptionId := flotsam.InscriptionId

	// If the origin of the flotsam is old, delete all by SatPoint and delete the inscription by ID.
	if flotsam.Origin.Old != nil {
		if err := u.wtx.DeleteAllBySatPoint(&flotsam.Origin.Old.OldSatPoint); err != nil {
			return err
		}
		inscription, err := u.wtx.GetInscriptionById(inscriptionId)
		if err != nil {
			return err
		}
		sequenceNumber = inscription.SequenceNum
	} else if flotsam.Origin.New != nil { // If the origin of the flotsam is new, process it.
		unbound = flotsam.Origin.New.Unbound
		inscriptionNumber := int64(0)

		// If the flotsam is cursed, increment the cursed inscription count.
		if flotsam.Origin.New.Cursed {
			number := *u.cursedInscriptionCount
			if !atomic.CompareAndSwapUint64(u.cursedInscriptionCount, number, number+1) {
				return errors.New("cursedInscriptionCount compare and swap failed")
			}
			// because cursed numbers start at -1
			inscriptionNumber = -(int64(number) + 1)
		} else { // If the flotsam is not cursed, increment the blessed inscription count.
			number := *u.blessedInscriptionCount
			if !atomic.CompareAndSwapUint64(u.blessedInscriptionCount, number, number+1) {
				return errors.New("blessedInscriptionCount compare and swap failed")
			}
			inscriptionNumber = int64(number)

		}
		// Increment the sequence number.
		sequenceNumber = *u.nextSequenceNumber
		if !atomic.CompareAndSwapInt64(u.nextSequenceNumber, sequenceNumber, sequenceNumber+1) {
			return errors.New("nextSequenceNumber compare and swap failed")
		}

		// If the flotsam is not unbound, calculate its Sat.
		var sat *Sat
		if !unbound {
			sat = u.calculateSat(inputSatRanges, flotsam.Offset)
		}

		// Initialize the charms.
		charms := uint16(0)
		// If the flotsam is cursed, set the cursed charm.
		if flotsam.Origin.New.Cursed {
			CharmCursed.Set(&charms)
		}

		// If the flotsam is a re-inscription, set the re-inscription charm.
		if flotsam.Origin.New.ReInscription {
			CharmReInscription.Set(&charms)
		}

		// If the Sat is not nil, set the appropriate charms based on its properties.
		if sat != nil {
			if sat.NineBall() {
				CharmNineBall.Set(&charms)
			}
			if sat.Coin() {
				CharmCoin.Set(&charms)
			}

			// Set the charm based on the rarity of the Sat.
			switch sat.Rarity() {
			case RarityCommon, RarityMythic:
			case RarityUncommon:
				CharmUncommon.Set(&charms)
			case RarityRare:
				CharmRare.Set(&charms)
			case RarityEpic:
				CharmEpic.Set(&charms)
			case RarityLegendary:
				CharmLegendary.Set(&charms)
			}
		}

		// If the new newSatPoint is empty, set the lost charm.
		outpoint, err := wire.NewOutPointFromString(newSatPoint.Outpoint)
		if err != nil {
			return err
		}
		if util.IsNullOutpoint(*outpoint) {
			CharmLost.Set(&charms)
		}

		// If the flotsam is unbound, set the unbound charm.
		if unbound {
			CharmUnbound.Set(&charms)
		}

		// If the Sat is not nil, save it to the sequence number in the database.
		if sat != nil {
			if err := u.wtx.SaveSatToSequenceNumber(uint64(*sat), sequenceNumber); err != nil {
				return err
			}
		}

		inscription := flotsam.Origin.New.Inscription
		// Create a new inscription entry.
		entry := &tables.Inscriptions{
			InscriptionId:   *inscriptionId,
			Index:           inscription.index,
			SequenceNum:     sequenceNumber,
			InscriptionNum:  inscriptionNumber,
			Owner:           inscription.owner,
			CInsDescription: inscription.payload.CInsDescription,
			Charms:          charms,
			Fee:             uint64(flotsam.Origin.New.Fee),
			Height:          u.idx.height,
			Timestamp:       u.timestamp,
			Body:            inscription.payload.Body,
			ContentEncoding: string(inscription.payload.ContentEncoding),
			ContentType:     string(inscription.payload.ContentType),
			MediaType:       string(inscription.payload.ContentType.MediaType()),
			ContentSize:     uint32(len(inscription.payload.Body)),
			Metadata:        inscription.payload.Metadata,
			Pointer:         gconv.Int32(string(inscription.payload.Pointer)),
		}
		// If the Sat is not nil, set the Sat and offset in the entry.
		if sat != nil {
			entry.Sat = uint64(*sat)
			entry.Offset = uint32(flotsam.Offset)
		}

		protocol, err := util.NewProtocolFromBytes(inscription.payload.Body)
		if err == nil {
			entry.ContentProtocol = protocol.Name()
		}

		// Create the inscription in the database.
		if err := u.wtx.CreateInscription(entry); err != nil {
			return err
		}
		// Create protocol entry
		if err := NewProtocol(u.wtx, entry).SaveProtocol(); err != nil && !errors.Is(err, util.NotSupportedProtocol) {
			return err
		}
	}

	satPoint := newSatPoint
	if unbound {
		unboundNum := *u.unboundInscriptions
		if !atomic.CompareAndSwapUint64(u.unboundInscriptions, unboundNum, unboundNum+1) {
			return errors.New("unboundInscriptions compare and swap failed")
		}
		satPoint = &tables.SatPointToSequenceNum{
			Offset: unboundNum,
		}
	}
	if err := u.wtx.SetSatPointToSequenceNum(satPoint, sequenceNumber); err != nil {
		return err
	}
	return nil
}

// calculateSat calculates the Sat of an inscription.
func (u *InscriptionUpdater) calculateSat(
	inputSatRanges tables.SatRanges,
	inputOffset uint64,
) *Sat {
	// Initialize an offset counter starting from 0.
	offset := uint64(0)

	// Loop over each range in the inputSatRanges slice.
	for _, v := range inputSatRanges {
		// Calculate the size of the current range by subtracting the start from the end.
		size := v.End - v.Start

		// If the offset plus the size of the current range is greater than the inputOffset,
		// calculate the Sat of the inscription and return a pointer to it.
		if offset+size > inputOffset {
			// Calculate the Sat by adding the start of the current range to the inputOffset,
			// and subtracting the current offset.
			n := Sat(v.Start + inputOffset - offset)

			// Return a pointer to the calculated Sat.
			return &n
		}

		// If the offset plus the size of the current range is not greater than the inputOffset,
		// increment the offset by the size of the current range.
		offset += size
	}

	// If no Sat could be calculated for any of the ranges in the inputSatRanges slice,
	// return nil.
	return nil
}

// fetchOutputValues fetches the output values of a transaction.
func (u *InscriptionUpdater) fetchOutputValues(tx *wire.MsgTx, maxCurrentNum int) (valueCh chan int64, errCh chan error, needDelOutpoints map[string]struct{}, err error) {
	// Calculate the current number of outpoints to fetch.
	currentNum := len(tx.TxIn)/2 + 2
	if currentNum > maxCurrentNum {
		currentNum = maxCurrentNum
	}

	// Create a channel to receive errors.
	errCh = make(chan error)

	// Create a channel to receive the fetched values.
	valueCh = make(chan int64, currentNum)

	// Create a channel to send the outpoints that need to be fetched.
	needFetchOutpointsCh := make(chan *wire.OutPoint, currentNum)

	needDelOutpointsLock := &sync.Mutex{}
	needDelOutpoints = make(map[string]struct{})

	errWg := &errgroup.Group{}
	for inputIndex := range tx.TxIn {
		select {
		case <-signal.InterruptChannel:
			err = signal.ErrInterrupted
			return
		default:
			txIn := tx.TxIn[inputIndex]

			// If the input is a coinbase, skip it.
			if util.IsNullOutpoint(txIn.PreviousOutPoint) {
				continue
			}

			// If the value of the input is already cached, skip it.
			if _, ok := u.valueCache.Read(txIn.PreviousOutPoint.String()); ok {
				continue
			}

			errWg.Go(func() error {
				if signal.InterruptRequested() {
					return nil
				}
				// Try to get the value of the input from the database.
				if _, err := u.idx.DB().GetValueByOutpoint(txIn.PreviousOutPoint.String()); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err == nil {
					needDelOutpointsLock.Lock()
					needDelOutpoints[txIn.PreviousOutPoint.String()] = struct{}{}
					needDelOutpointsLock.Unlock()
				}
				return nil
			})
		}
	}

	if err = errWg.Wait(); err != nil {
		return
	}

	go func() {
		defer close(needFetchOutpointsCh)
		// Loop over each input in the transaction.
		for inputIndex := range tx.TxIn {
			select {
			case <-signal.InterruptChannel:
				err = signal.ErrInterrupted
				return
			default:
				txIn := tx.TxIn[inputIndex]

				// If the input is a coinbase, skip it.
				if util.IsNullOutpoint(txIn.PreviousOutPoint) {
					continue
				}

				// If the value of the input is already cached, skip it.
				if _, ok := u.valueCache.Read(txIn.PreviousOutPoint.String()); ok {
					continue
				}

				// Try to get the value of the input from the database.
				if _, ok := needDelOutpoints[txIn.PreviousOutPoint.String()]; !ok {
					needFetchOutpointsCh <- &txIn.PreviousOutPoint
				}
			}
		}
	}()

	// Start a goroutine to fetch the values in batches.
	go func() {
		commitNum := 0
		fetchOutpointsNum := 0
		needFetchOutpoints := make([]*wire.OutPoint, currentNum)
		batchResult := make([]*rpcclient.FutureGetRawTransactionResult, currentNum)

		defer func() {
			close(valueCh)
			batchResult = nil
			needFetchOutpoints = nil
		}()

		for outpoint := range needFetchOutpointsCh {
			select {
			case <-signal.InterruptChannel:
				return
			default:
				i := fetchOutpointsNum % currentNum
				needFetchOutpoints[i] = outpoint
				res := u.idx.BatchRpcClient().GetRawTransactionAsync(&outpoint.Hash)
				batchResult[i] = &res

				if (fetchOutpointsNum+1)%currentNum == 0 {
					if err := u.idx.BatchRpcClient().Send(); err != nil {
						errCh <- err
						close(errCh)
						return
					}
					// Loop over the results of the batch.
					for ii := 0; ii < currentNum; ii++ {
						// Receive the result of the fetch operation.
						tx, err := batchResult[ii].Receive()
						if err != nil {
							errCh <- err
							close(errCh)
							return
						}
						// Send the value of the output to the value channel.
						valueCh <- tx.MsgTx().TxOut[needFetchOutpoints[ii].Index].Value
						batchResult[ii] = nil
					}
					commitNum++
				}
			}
			fetchOutpointsNum++
		}

		lastNum := fetchOutpointsNum % currentNum
		if lastNum > 0 {
			if err := u.idx.BatchRpcClient().Send(); err != nil {
				errCh <- err
				close(errCh)
				return
			}
			for i := 0; i < lastNum; i++ {
				// Receive the result of the fetch operation.
				tx, err := batchResult[i].Receive()
				if err != nil {
					errCh <- err
					close(errCh)
					return
				}
				// Send the value of the output to the value channel.
				valueCh <- tx.MsgTx().TxOut[needFetchOutpoints[i].Index].Value
				batchResult[i] = nil
			}
		}
	}()
	return
}
