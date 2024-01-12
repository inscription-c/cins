package model

import (
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/wire"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/shopspring/decimal"
	"strings"
)

type Amount float64

func (a Amount) Sat() int64 {
	return decimal.NewFromFloat(float64(a)).
		Mul(decimal.NewFromInt(constants.OneBtc)).IntPart()
}

func Sat(amount float64) int64 {
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
	return fmt.Sprintf("%s%s%d", o.Hash, constants.InscriptionIdDelimiter, o.Index)
}

func (o *OutPoint) Outpoint() string {
	return fmt.Sprintf("%s%s%d", o.Hash, constants.OutpointDelimiter, o.Index)
}

func (o *OutPoint) WireOutpoint() (*wire.OutPoint, error) {
	return wire.NewOutPointFromString(o.Outpoint())
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
