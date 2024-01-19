package index

import (
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/shopspring/decimal"
	"strings"
)

type Inscription struct {
	Body            []byte
	ContentEncoding []byte
	ContentType     constants.ContentType
	DstChain        []byte
	Metadata        []byte
	Parent          []byte
	Pointer         uint64

	DuplicateField        bool
	IncompleteField       bool
	UnRecognizedEvenField bool
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
	return &InscriptionId{
		OutPoint: OutPoint{OutPoint: btcjson.OutPoint{
			Hash:  insId[0],
			Index: gconv.Uint32(insId[1]),
		}},
	}
}

type Amount float64

type Sat uint64

func (a Amount) Sat() Sat {
	return Sat(decimal.NewFromFloat(float64(a)).
		Mul(decimal.NewFromInt(constants.OneBtc)).IntPart())
}

func (s *Sat) NineBall() bool {
	return uint64(*s) >= 50*constants.OneBtc*9 && uint64(*s) < 50*constants.OneBtc*10
}

func (s *Sat) Coin() bool {
	return uint64(*s)%constants.OneBtc == 0
}

func AmountToSat(amount float64) Sat {
	return Amount(amount).Sat()
}

type OutPoint struct {
	btcjson.OutPoint
}

func NewOutPoint(txId string, index uint32) *OutPoint {
	return &OutPoint{
		OutPoint: btcjson.OutPoint{
			Hash:  txId,
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

func StringToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !constants.OutpointRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.OutpointDelimiter)
	return &OutPoint{
		OutPoint: btcjson.OutPoint{
			Hash:  insId[0],
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
	return &OutPoint{
		OutPoint: btcjson.OutPoint{
			Hash:  insId[0],
			Index: gconv.Uint32(insId[1]),
		},
	}
}
