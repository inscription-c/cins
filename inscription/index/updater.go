package index

import (
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/inscription/index/tables"
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
	InscriptionId *model.InscriptionId
	Offset        uint64
	Origin        Origin
}

type Origin struct {
	New *OriginNew
	Old *OriginOld
}

type OriginNew struct {
	Cursed        bool
	Fee           int64
	Hidden        bool
	Pointer       int32
	ReInscription bool
	Unbound       bool
	Inscription   *Envelope
}

type OriginOld struct {
	OldSatPoint dao.SatPoint
}

type InscriptionUpdater struct {
	idx                     *Indexer
	wtx                     *dao.DB
	flotsam                 []*Flotsam
	height                  uint32
	lostSats                uint64
	reward                  uint64
	valueCache              map[string]int64
	valueCh                 chan uint64
	timestamp               int64
	nextSequenceNumber      *uint64
	unboundInscriptions     *uint32
	cursedInscriptionCount  *uint32
	blessedInscriptionCount *uint32
}

type inscribedOffsetEntity struct {
	inscriptionId *model.InscriptionId
	count         int64
}

type locationsInscription struct {
	satpoint *dao.SatPoint
	flotsam  *Flotsam
}

func (u *InscriptionUpdater) indexEnvelopers(
	tx *wire.MsgTx,
	inputSatRange []*model.SatRange) error {

	idCounter := int64(0)
	totalInputValue := int64(0)
	floatingInscriptions := make([]*Flotsam, 0)
	inscribedOffsets := make(map[uint64]*inscribedOffsetEntity)

	envelopes := ParsedEnvelopeFromTransaction(tx)
	//inscriptions := len(envelopes) > 0

	totalOutputValue := int64(0)
	for _, v := range tx.TxOut {
		totalOutputValue += v.Value
	}

	for inputIndex := range tx.TxIn {
		txIn := tx.TxIn[inputIndex]
		// is coin base
		if IsEmptyHash(txIn.PreviousOutPoint.Hash) {
			totalInputValue += int64(NewHeight(u.height).Subsidy())
			continue
		}

		// find existing inscriptions on input (transfers of inscriptions)
		inscriptions, err := u.wtx.InscriptionsByOutpoint(txIn.PreviousOutPoint.String())
		if err != nil {
			return err
		}
		for _, v := range inscriptions {
			offset := uint64(totalInputValue) + uint64(v.SatPoint.Offset)
			insId := model.StringToOutpoint(v.Inscriptions.Outpoint).InscriptionId()
			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				InscriptionId: insId,
				Offset:        offset,
				Origin: Origin{
					Old: &OriginOld{OldSatPoint: *v.SatPoint},
				},
			})

			offsetEntity, ok := inscribedOffsets[offset]
			if !ok {
				offsetEntity = &inscribedOffsetEntity{
					inscriptionId: insId,
				}
				inscribedOffsets[offset] = offsetEntity
			}
			offsetEntity.count++
		}

		offset := uint64(totalInputValue)
		preOutpoint := txIn.PreviousOutPoint.String()
		currentInputValue, ok := u.valueCache[txIn.PreviousOutPoint.String()]
		if ok {
			delete(u.valueCache, preOutpoint)
		} else {
			if currentInputValue, err = u.wtx.GetValueByOutpoint(preOutpoint); err == nil {
				if err := u.wtx.DeleteValueByOutpoint(preOutpoint); err != nil {
					return err
				}
			} else {
				tx, err := u.idx.opts.cli.GetRawTransaction(&txIn.PreviousOutPoint.Hash)
				if err != nil {
					return err
				}
				currentInputValue = tx.MsgTx().TxOut[txIn.PreviousOutPoint.Index].Value
			}
		}
		totalInputValue += currentInputValue

		// go through all inscriptions in this input
		for _, inscription := range envelopes {
			if inscription.input != inputIndex {
				break
			}

			inscriptionId := model.InscriptionId{
				OutPoint: model.OutPoint{
					OutPoint: wire.OutPoint{
						Hash:  tx.TxHash(),
						Index: uint32(idCounter),
					},
				},
			}

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
			} else if inscription.payload.Pointer >= 0 {
				curse = CursePointer
			} else if inscription.pushNum {
				curse = CursePushNum
			} else if inscription.stutter {
				curse = CurseStutter
			} else {
				offsetEntity, ok := inscribedOffsets[offset]
				if ok {
					if offsetEntity.count > 1 {
						curse = CurseReInscription
					} else {
						entry, err := u.wtx.GetInscriptionById(offsetEntity.inscriptionId.String())
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

			unbound := currentInputValue == 0 ||
				curse == CurseUnrecognizedEvenField ||
				inscription.payload.UnRecognizedEvenField

			pointer := int64(inscription.payload.Pointer)
			if pointer > 0 && pointer < totalOutputValue {
				offset = uint64(inscription.payload.Pointer)
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
						Inscription:   inscription,
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

	// TODO index transaction
	// TODO potential_parents

	// still have to normalize over inscription size
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

	outputValue := uint64(0)
	rangeToVoutMap := make(map[model.SatRange]int)
	newLocations := make([]*locationsInscription, 0)

	for vout, txOut := range tx.TxOut {
		end := outputValue + uint64(txOut.Value)

		for _, flotsam := range floatingInscriptions {
			if flotsam.Offset >= end {
				break
			}
			newSatpoint := &dao.SatPoint{
				Outpoint: wire.OutPoint{
					Hash:  tx.TxHash(),
					Index: uint32(vout),
				},
				Offset: uint32(flotsam.Offset - outputValue),
			}
			newLocations = append(newLocations, &locationsInscription{
				satpoint: newSatpoint,
				flotsam:  flotsam,
			})
		}

		rangeToVoutMap[model.SatRange{
			Start: outputValue,
			End:   end,
		}] = vout

		outputValue = end

		outpoint := model.NewOutPoint(tx.TxHash().String(), uint32(vout)).String()
		u.valueCache[outpoint] = txOut.Value
	}

	for _, location := range newLocations {
		flotsam := location.flotsam
		newSatpoint := location.satpoint
		pointer := uint64(flotsam.Origin.New.Pointer)
		if pointer >= 0 && pointer < outputValue {
			for rangeEntity, vout := range rangeToVoutMap {
				if pointer < rangeEntity.Start || pointer >= rangeEntity.End {
					continue
				}
				flotsam.Offset = pointer
				newSatpoint = &dao.SatPoint{
					Outpoint: wire.OutPoint{
						Hash:  tx.TxHash(),
						Index: uint32(vout),
					},
					Offset: uint32(pointer - rangeEntity.Start),
				}
			}
		}
		if err := u.updateInscriptionLocation(inputSatRange, flotsam, newSatpoint); err != nil {
			return err
		}
	}

	if isCoinBase {
		for _, flotsam := range floatingInscriptions {
			newSatpoint := &dao.SatPoint{
				Offset: uint32(uint64(u.lostSats) + flotsam.Offset - outputValue),
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
	u.reward += uint64(totalInputValue) - outputValue

	return nil
}

func (u *InscriptionUpdater) updateInscriptionLocation(
	inputSatRanges []*model.SatRange,
	flotsam *Flotsam,
	newSatpoint *dao.SatPoint,
) error {

	var err error
	var unbound bool
	var sequenceNumber uint64
	inscriptionId := flotsam.InscriptionId

	if flotsam.Origin.Old != nil {
		if err := u.wtx.DeleteAllBySatPoint(&flotsam.Origin.Old.OldSatPoint); err != nil {
			return err
		}
		sequenceNumber, err = u.wtx.DeleteInscriptionById(inscriptionId.String())
		if err != nil {
			return err
		}
	} else if flotsam.Origin.New != nil {
		unbound = flotsam.Origin.New.Unbound
		inscriptionNumber := int64(0)

		if flotsam.Origin.New.Cursed {
			number := *u.cursedInscriptionCount
			*u.cursedInscriptionCount++
			// because cursed numbers start at -1
			inscriptionNumber = -(int64(number) + 1)
		} else {
			number := *u.blessedInscriptionCount
			*u.blessedInscriptionCount++
			inscriptionNumber = int64(number)
		}

		sequenceNumber = *u.nextSequenceNumber
		*u.nextSequenceNumber++

		var sat *Sat
		if !unbound {
			sat = u.calculateSat(inputSatRanges, flotsam.Offset)
		}

		charms := uint16(0)
		if flotsam.Origin.New.Cursed {
			CharmCursed.Set(&charms)
		}

		if flotsam.Origin.New.ReInscription {
			CharmReInscription.Set(&charms)
		}

		if sat != nil {
			if sat.NineBall() {
				CharmNineBall.Set(&charms)
			}
			if sat.Coin() {
				CharmCoin.Set(&charms)
			}

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

		if IsEmptyHash(newSatpoint.Outpoint.Hash) {
			CharmLost.Set(&charms)
		}

		if unbound {
			CharmUnbound.Set(&charms)
		}

		if sat != nil {
			if err := u.wtx.SaveSatToSequenceNumber(uint64(*sat), sequenceNumber); err != nil {
				return err
			}
		}

		ins := flotsam.Origin.New.Inscription
		entry := &tables.Inscriptions{
			Outpoint:        inscriptionId.OutPoint.String(),
			InscriptionNum:  inscriptionNumber,
			Charms:          charms,
			Fee:             uint64(flotsam.Origin.New.Fee),
			Height:          u.height,
			Timestamp:       u.timestamp,
			Body:            ins.payload.Body,
			ContentEncoding: ins.payload.ContentEncoding,
			ContentType:     string(ins.payload.ContentType),
			DstChain:        ins.payload.DstChain,
			Metadata:        ins.payload.Metadata,
			Pointer:         ins.payload.Pointer,
		}
		if sat != nil {
			entry.Sat = uint64(*sat)
			entry.Offset = uint32(flotsam.Offset)
		}
		if err := u.wtx.CreateInscription(entry); err != nil {
			return err
		}
		if entry.Id != sequenceNumber {
			return fmt.Errorf("sequence number mismatch")
		}
	}

	satPoint := newSatpoint
	if unbound {
		satPoint = &dao.SatPoint{
			Offset: *u.unboundInscriptions,
		}
		*u.unboundInscriptions++
	}
	if err := u.wtx.SetSatPointToSequenceNum(satPoint, sequenceNumber); err != nil {
		return err
	}
	return nil
}

func (u *InscriptionUpdater) calculateSat(
	inputSatRanges []*model.SatRange,
	inputOffset uint64,
) *Sat {
	offset := uint64(0)
	for _, v := range inputSatRanges {
		size := v.End - v.Start
		if offset+size > inputOffset {
			n := Sat(v.Start + inputOffset - offset)
			return &n
		}
		offset += size
	}
	return nil
}
