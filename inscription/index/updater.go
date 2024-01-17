package index

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang/protobuf/proto"
	"github.com/nutsdb/nutsdb"
	"sort"
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
	Offset        uint64
	Origin        Origin
}

type Origin struct {
	New *OriginNew
	Old *OriginOld
}

type OriginNew struct {
	Cursed        bool
	Fee           uint64
	Hidden        bool
	Pointer       []byte
	ReInscription bool
	Unbound       bool
}

type OriginOld struct {
	OldSatPoint SatPoint
}

type InscriptionUpdater struct {
	wtx                     *Tx
	flotsam                 []*Flotsam
	height                  uint64
	lostSats                uint64
	reward                  uint64
	blessedInscriptionCount *uint64
	cursedInscriptionCount  *uint64
	nextSequenceNumber      *int64
	unboundInscriptions     *uint64
	valueCache              map[string]uint64
	valueCh                 chan uint64
	timestamp               int64
}

type inscribedOffsetEntity struct {
	inscriptionId *InscriptionId
	count         int64
}

type locationsInscription struct {
	satpoint *SatPoint
	flotsam  *Flotsam
}

type SatRange struct {
	start uint64
	end   uint64
}

func (u *InscriptionUpdater) indexEnvelopers(
	tx *wire.MsgTx,
	inputSatRange []*SatRange) error {

	totalInputValue := uint64(0)
	idCounter := uint64(0)
	floatingInscriptions := make([]*Flotsam, 0)
	inscribedOffsets := make(map[uint64]*inscribedOffsetEntity)
	envelopes := ParsedEnvelopeFromTransaction(tx)
	totalOutputValue := uint64(0)
	for _, v := range tx.TxOut {
		totalOutputValue += uint64(v.Value)
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
			offset := totalInputValue + v.SatPoint.Offset
			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: v.Id,
				Offset:        offset,
				Origin: Origin{
					Old: &OriginOld{OldSatPoint: *v.SatPoint},
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
				currentInputValue = gconv.Uint64(string(v))
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
				gconv.Uint64(string(inscription.payload.Pointer)) < totalOutputValue {
				offset = gconv.Uint64(string(inscription.payload.Pointer))
			}

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

	rangeToVoutMap := make(map[SatRange]int)
	outputValue := uint64(0)
	newLocations := make([]*locationsInscription, 0)
	for vout, txOut := range tx.TxOut {
		end := outputValue + uint64(txOut.Value)
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

		rangeToVoutMap[SatRange{
			start: outputValue,
			end:   end,
		}] = vout

		outputValue = end

		outpoint := NewOutPoint(tx.TxHash().String(), uint32(vout)).String()
		u.valueCache[outpoint] = uint64(txOut.Value)
	}

	for _, flotsam := range newLocations {
		if len(flotsam.flotsam.Origin.New.Pointer) > 0 {
			pointer := gconv.Uint64(string(flotsam.flotsam.Origin.New.Pointer))
			if pointer < outputValue {
				for rangeEntity, vout := range rangeToVoutMap {
					if pointer >= rangeEntity.start && pointer < rangeEntity.end {
						flotsam.flotsam.Offset = pointer
						flotsam.satpoint = &SatPoint{
							Outpoint: &wire.OutPoint{
								Hash:  tx.TxHash(),
								Index: uint32(vout),
							},
							Offset: pointer - rangeEntity.start,
						}
					}
				}
			}
		}
		if err := u.updateInscriptionLocation(inputSatRange, flotsam.flotsam, flotsam.satpoint); err != nil {
			return err
		}
	}

	if isCoinBase {
		for _, flotsam := range floatingInscriptions {
			newSatpoint := &SatPoint{
				Outpoint: nil,
				Offset:   u.lostSats + flotsam.Offset - outputValue,
			}
			if err := u.updateInscriptionLocation(inputSatRange, flotsam, newSatpoint); err != nil {
				return err
			}
		}
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
	pattern := fmt.Sprintf("%s%s.*", output.String(), constants.OutpointDelimiter)
	if err = u.wtx.SKeys(constants.BucketSatpointToSequenceNumber, pattern, func(key string) bool {
		var sequenceNumbers [][]byte
		sequenceNumbers, err = u.wtx.SMembers(constants.BucketSatpointToSequenceNumber, []byte(key))
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

func (u *InscriptionUpdater) updateInscriptionLocation(
	inputSatRanges []*SatRange,
	flotsam *Flotsam,
	newSatpoint *SatPoint,
) error {
	inscriptionid := flotsam.InscriptionId
	var unbound bool
	var sequenceNumber int64
	if flotsam.Origin.Old != nil {
		if err := u.wtx.Delete(constants.BucketSatpointToSequenceNumber, []byte(flotsam.Origin.Old.OldSatPoint.String())); err != nil {
			return err
		}
		v, err := u.wtx.Get(constants.BucketInscriptionIdToSequenceNumber, []byte(inscriptionid.String()))
		if err != nil {
			return err
		}
		sequenceNumber = gconv.Int64(string(v))
	} else if flotsam.Origin.New != nil {
		unbound = flotsam.Origin.New.Unbound
		inscriptionNumber := int64(0)
		if flotsam.Origin.New.Cursed {
			num := *u.cursedInscriptionCount
			*u.cursedInscriptionCount++
			inscriptionNumber = -(int64(num) + 1)
		} else {
			inscriptionNumber = int64(*u.blessedInscriptionCount)
			*u.blessedInscriptionCount++
		}

		sequenceNumber = *u.nextSequenceNumber
		*u.nextSequenceNumber++

		if err := u.wtx.Put(constants.BucketInscriptionNumberToSequenceNumber,
			[]byte(gconv.String(inscriptionNumber)),
			[]byte(gconv.String(sequenceNumber))); err != nil {
			return err
		}

		var sat *Sat
		if !unbound {
			sat = u.calculateSat(inputSatRanges, flotsam.Offset)
		}
		charms := uint16(0)
		if flotsam.Origin.New.Cursed {
			constants.CharmCursed.Set(&charms)
		}
		if flotsam.Origin.New.ReInscription {
			constants.CharmReInscription.Set(&charms)
		}
		if sat != nil {
			if sat.NineBall() {
				constants.CharmNineBall.Set(&charms)
			}
			if sat.Coin() {
				constants.CharmCoin.Set(&charms)
			}

			// TODO rarity sat
		}

		if newSatpoint.Outpoint == nil || IsEmptyHash(newSatpoint.Outpoint.Hash) {
			constants.CharmLost.Set(&charms)
		}

		if unbound {
			constants.CharmUnbound.Set(&charms)
		}

		if sat != nil {
			if err := u.wtx.SAdd(
				constants.BucketSatToSequenceNumber,
				[]byte(gconv.String(uint64(*sat))),
				[]byte(gconv.String(sequenceNumber))); err != nil {
				return err
			}
		}

		// TODO parent

		entry := &InscriptionEntry{
			Charms:            uint32(charms),
			Fee:               flotsam.Origin.New.Fee,
			Height:            u.height,
			Id:                []byte(inscriptionid.String()),
			InscriptionNumber: inscriptionNumber,
			SequenceNumber:    sequenceNumber,
			Timestamp:         u.timestamp,
		}
		if sat != nil {
			satNum := uint64(*sat)
			entry.Sat = &satNum
		}
		entryData, err := proto.Marshal(entry)
		if err != nil {
			return err
		}
		if err := u.wtx.Put(
			constants.BucketSequenceNumberToInscriptionEntry,
			[]byte(gconv.String(sequenceNumber)),
			entryData); err != nil {
			return err
		}
		if err := u.wtx.Put(
			constants.BucketInscriptionIdToSequenceNumber,
			[]byte(inscriptionid.String()),
			[]byte(gconv.String(sequenceNumber))); err != nil {
			return err
		}
		// TODO home_inscriptions
	}

	satPoint := newSatpoint
	if unbound {
		satPoint = &SatPoint{
			Outpoint: nil,
			Offset:   *u.unboundInscriptions,
		}
		*u.unboundInscriptions++
	}

	if err := u.wtx.SAdd(
		constants.BucketSatpointToSequenceNumber,
		[]byte(satPoint.String()),
		[]byte(gconv.String(sequenceNumber))); err != nil {
		return err
	}
	if err := u.wtx.Put(
		constants.BucketSequenceNumberToSatpoint,
		[]byte(gconv.String(sequenceNumber)),
		[]byte(satPoint.String())); err != nil {
		return err
	}
	return nil
}

func (u *InscriptionUpdater) calculateSat(
	inputSatRanges []*SatRange,
	inputOffset uint64,
) *Sat {
	offset := uint64(0)
	for _, v := range inputSatRanges {
		size := v.end - v.start
		if offset+size > inputOffset {
			n := Sat(v.start + inputOffset - offset)
			return &n
		}
		offset += size
	}
	return nil
}
