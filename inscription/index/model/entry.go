package model

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
)

type Inscription struct {
	Body            []byte
	ContentEncoding string
	ContentType     constants.ContentType
	DstChain        string
	Metadata        []byte
	Pointer         int32

	UnRecognizedEvenField bool
	DuplicateField        bool
	IncompleteField       bool
}

type InscriptionId struct {
	OutPoint
}

func (i *InscriptionId) String() string {
	return fmt.Sprintf("%s%s%d", i.Hash, constants.InscriptionIdDelimiter, i.Index)
}

func StringToInscriptionId(s string) *InscriptionId {
	s = strings.ToLower(s)
	if !constants.InscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.InscriptionIdDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &InscriptionId{
		OutPoint: OutPoint{OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		}},
	}
}

func StringToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !constants.OutpointRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.OutpointDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		},
	}
}

func InscriptionIdToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !constants.InscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.InscriptionIdDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		},
	}
}

type OutPoint struct {
	wire.OutPoint
}

func NewOutPoint(txId string, index uint32) *OutPoint {
	h, _ := chainhash.NewHashFromStr(txId)
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: index,
		},
	}
}

func (o *OutPoint) String() string {
	return fmt.Sprintf("%s%s%d", o.Hash, constants.OutpointDelimiter, o.Index)
}

func (o *OutPoint) WireOutpoint() (*wire.OutPoint, error) {
	return wire.NewOutPointFromString(o.String())
}

func (o *OutPoint) InscriptionId() *InscriptionId {
	return &InscriptionId{
		OutPoint: *o,
	}
}
